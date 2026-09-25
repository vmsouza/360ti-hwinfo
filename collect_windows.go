package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows/registry"
)

var errMu sync.Mutex

func addError(rep *Report, s string) {
	errMu.Lock()
	rep.Errors = append(rep.Errors, s)
	errMu.Unlock()
}

// ---- WMI structs ----

type wmiBIOS struct {
	Manufacturer         string
	SMBIOSBIOSVersion    string
	Version              string
	SerialNumber         string
	ReleaseDate          time.Time
	Status               string
	Name                 string
	CurrentLanguage      string
	SMBIOSMajorVersion   uint16
	SMBIOSMinorVersion   uint16
	InstallableLanguages uint16
	BIOSVersion          []string
}

type wmiBaseBoard struct {
	Manufacturer string
	Product      string
	Version      string
	SerialNumber string
	PartNumber   string
	SKU          string
}

type wmiComputerSystemProduct struct {
	Vendor            string
	Name              string
	IdentifyingNumber string
	UUID              string
	SKUNumber         string
}

type wmiComputerSystem struct {
	Manufacturer            string
	Model                   string
	Domain                  string
	Workgroup               string
	PartOfDomain            bool
	UserName                string
	SystemType              string
	PCSystemType            uint16
	TotalPhysicalMemory     uint64
	NumberOfProcessors      uint32
	NumberOfLogicalProcessors uint32
	DomainRole              uint16
}

type wmiPhysicalMemory struct {
	BankLabel    string
	DeviceLocator string
	Capacity     uint64
	Speed        uint32
	Manufacturer string
	PartNumber   string
	SerialNumber string
	FormFactor   uint16
	MemoryType   uint16
}

type wmiProcessor struct {
	Name                      string
	Manufacturer              string
	NumberOfCores             uint32
	NumberOfEnabledCore       uint32
	NumberOfLogicalProcessors uint32
	SocketDesignation         string
	Family                    uint16
	CurrentClockSpeed         uint32
	MaxClockSpeed             uint32
	CurrentVoltage            uint16
	L2CacheSize               uint32
	L3CacheSize               uint32
	Architecture              uint16
	ProcessorId               string
}

type wmiVideoController struct {
	Name                  string
	AdapterRAM            uint32
	DriverVersion         string
	DriverDate            time.Time
	VideoProcessor        string
	AdapterCompatibility  string
	Status                string
}

type wmiDesktopMonitor struct {
	Name               string
	MonitorManufacturer string
	MonitorType        string
	ScreenWidth        uint32
	ScreenHeight       uint32
}

type wmiUSBController struct {
	Name         string
	Manufacturer string
	Description  string
	Status       string
	DeviceID     string
}

func collectWindowsInfo(rep *Report) {
	collectOSDisplay(rep)

	jobs := []struct {
		label string
		d     time.Duration
		fn    func()
	}{
		{"WMI principal", 6 * time.Second, func() { collectWMI(rep) }},
		{"Software", 6 * time.Second, func() { collectSoftware(rep) }},
		{"Sensores", 3 * time.Second, func() { collectSensors(rep) }},
		{"PCI", 6 * time.Second, func() { collectPCI(rep) }},
		{"Gabinete", 4 * time.Second, func() { collectEnclosure(rep) }},
		{"Discos físicos", 5 * time.Second, func() { collectDiskHardware(rep) }},
		{"Rede detalhada", 6 * time.Second, func() { collectNetworkDetailed(rep) }},
		{"Pastas de rede", 4 * time.Second, func() { collectNetworkFolders(rep) }},
	}

	var wg sync.WaitGroup
	wg.Add(len(jobs))
	for _, j := range jobs {
		go func(job struct {
			label string
			d     time.Duration
			fn    func()
		}) {
			defer wg.Done()
			runWithTimeout(job.d, job.label, job.fn)
		}(j)
	}

	// wait for all with a single overall budget so a hanging query cannot
	// hold the report for long.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(18 * time.Second):
	}
}

func collectOSDisplay(rep *Report) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.READ)
	if err != nil {
		return
	}
	defer k.Close()
	product, _, _ := k.GetStringValue("ProductName")
	display, _, _ := k.GetStringValue("DisplayVersion")
	build, _, _ := k.GetStringValue("CurrentBuildNumber")
	if product != "" {
		rep.Host.OS = product + fmt.Sprintf(" (build %s)", build)
	}
	if display != "" {
		rep.Host.OS += fmt.Sprintf(" v%s", display)
	}
}

func collectWMI(rep *Report) {
	runQuery := func(dst interface{}, class, label string) error {
		q := wmi.CreateQuery(dst, "", class)
		if err := wmi.Query(q, dst); err != nil {
			addError(rep, fmt.Sprintf("wmi %s: %v", label, err))
			return err
		}
		return nil
	}

	var bios []wmiBIOS
	if runQuery(&bios, "Win32_BIOS", "BIOS") == nil {
		for _, b := range bios {
			smbiosVer := ""
			if b.SMBIOSMajorVersion > 0 || b.SMBIOSMinorVersion > 0 {
				smbiosVer = fmt.Sprintf("%d.%d", b.SMBIOSMajorVersion, b.SMBIOSMinorVersion)
			}
			rep.BIOS = append(rep.BIOS, BIOSInfo{
				Manufacturer:   b.Manufacturer,
				SMBIOSVersion:  b.SMBIOSBIOSVersion,
				Version:        b.Version,
				Serial:         b.SerialNumber,
				ReleaseDate:    b.ReleaseDate.Format("02/01/2006"),
				Status:         b.Status,
				CurrentLanguage: b.CurrentLanguage,
				BiosMode:       collectBIOSMode(),
				Name:           b.Name,
				SmbiosSpec:     smbiosVer,
				BiosVersion:    joinStrings(b.BIOSVersion),
			})
		}
	}

	var board []wmiBaseBoard
	if runQuery(&board, "Win32_BaseBoard", "BaseBoard") == nil {
		for _, b := range board {
			rep.Board = append(rep.Board, BoardInfo{
				Manufacturer: b.Manufacturer,
				Product:      b.Product,
				Version:      b.Version,
				Serial:       b.SerialNumber,
				PartNumber:   b.PartNumber,
				SKU:          b.SKU,
			})
		}
	}

	var prod []wmiComputerSystemProduct
	if runQuery(&prod, "Win32_ComputerSystemProduct", "ComputerSystemProduct") == nil {
		for _, p := range prod {
			rep.Product = append(rep.Product, ProductInfo{
				Vendor: p.Vendor,
				Name:   p.Name,
				Serial: p.IdentifyingNumber,
				UUID:   p.UUID,
				SKU:    p.SKUNumber,
			})
		}
	}

	var compSys []wmiComputerSystem
	if runQuery(&compSys, "Win32_ComputerSystem", "ComputerSystem") == nil {
		if len(compSys) > 0 {
			c := compSys[0]
			if rep.Host.Hostname == "" {
				rep.Host.Hostname = c.Model
			}
			rep.Host.OS += fmt.Sprintf(" | Modelo: %s %s", c.Manufacturer, c.Model)
			if len(rep.Board) > 0 {
				rep.Board[0].Domain = c.Domain
				rep.Board[0].Workgroup = c.Workgroup
				rep.Board[0].PartOfDomain = c.PartOfDomain
				rep.Board[0].DomainRole = domainRoleName(c.DomainRole)
				rep.Board[0].NumProc = c.NumberOfProcessors
				rep.Board[0].NumLogical = c.NumberOfLogicalProcessors
				rep.Board[0].PcType = pcSystemType(c.PCSystemType)
			}
		}
	}

	// Extended motherboard/system fields (SystemFamily is absent on very old OSes)
	var ext []struct {
		Manufacturer    string
		Model           string
		SystemFamily    string
		SkuNumber       string
		SystemType      string
		ChassisSKUNumber string
	}
	if wmi.Query("SELECT Manufacturer, Model, SystemFamily, SKUNumber, SystemType, ChassisSKUNumber FROM Win32_ComputerSystem", &ext) == nil {
		if len(ext) > 0 {
			if len(rep.Board) > 0 {
				rep.Board[0].Family = ext[0].SystemFamily
				rep.Board[0].SystemType = ext[0].SystemType
				rep.Board[0].SKU = ext[0].SkuNumber
				rep.Board[0].ChassisSKU = ext[0].ChassisSKUNumber
			}
		}
	}

	var memMods []wmiPhysicalMemory
	if runQuery(&memMods, "Win32_PhysicalMemory", "PhysicalMemory") == nil {
		for _, m := range memMods {
			rep.MemModules = append(rep.MemModules, MemModuleInfo{
				Slot:         m.DeviceLocator,
				Bank:         m.BankLabel,
				Capacity:     ByteCountSI(m.Capacity),
				Speed:        m.Speed,
				Manufacturer: m.Manufacturer,
				PartNumber:   m.PartNumber,
				Serial:       m.SerialNumber,
				FormFactor:   memoryFormFactor(m.FormFactor),
				MemoryType:   memoryType(m.MemoryType),
			})
		}
	}

	var procs []wmiProcessor
	if runQuery(&procs, "Win32_Processor", "Processor") == nil {
		// gopsutil cpu.Info already fills rep.CPU; enrich if empty with WMI detail
		if len(rep.CPU) == 0 {
			for _, p := range procs {
				rep.CPU = append(rep.CPU, CPUInfo{
					Model:   p.Name,
					Vendor:  p.Manufacturer,
					Cores:   int32(p.NumberOfCores),
					Threads: int32(p.NumberOfLogicalProcessors),
					Socket:  p.SocketDesignation,
					Current: p.CurrentClockSpeed,
					Max:     p.MaxClockSpeed,
				})
			}
		} else if len(procs) > 0 {
			// enrich the first CPU entry with socket/thread/voltage details
			p := procs[0]
			rep.CPU[0].Socket = p.SocketDesignation
			rep.CPU[0].Threads = int32(p.NumberOfLogicalProcessors)
			rep.CPU[0].Current = p.CurrentClockSpeed
			rep.CPU[0].Max = p.MaxClockSpeed
			if p.CurrentVoltage > 0 {
				rep.CPU[0].Voltage = fmt.Sprintf("%.3f V", float64(p.CurrentVoltage)/1000)
			}
		}
	}

	var gpus []wmiVideoController
	if runQuery(&gpus, "Win32_VideoController", "VideoController") == nil {
		for _, g := range gpus {
			rep.GPU = append(rep.GPU, GPUInfo{
				Name:           g.Name,
				AdapterRAM:     ByteCountSI(uint64(g.AdapterRAM)),
				DriverVersion:  g.DriverVersion,
				DriverDate:     g.DriverDate.Format("02/01/2006"),
				VideoProcessor: g.VideoProcessor,
				Status:         g.Status,
			})
		}
	}

	var monitors []wmiDesktopMonitor
	if runQuery(&monitors, "Win32_DesktopMonitor", "DesktopMonitor") == nil {
		for _, m := range monitors {
			rep.Monitors = append(rep.Monitors, MonitorInfo{
				Manufacturer: m.MonitorManufacturer,
				Name:         m.Name,
				MonitorType:  m.MonitorType,
				ScreenWidth:  m.ScreenWidth,
				ScreenHeight: m.ScreenHeight,
			})
		}
	}

	var usbs []wmiUSBController
	if runQuery(&usbs, "Win32_USBController", "USBController") == nil {
		for _, u := range usbs {
			rep.USB = append(rep.USB, USBInfo{
				Name:         u.Name,
				Manufacturer: u.Manufacturer,
				Description:  u.Description,
				Status:       u.Status,
				DeviceID:     u.DeviceID,
			})
		}
	}
}

func collectSoftware(rep *Report) {
	baseKeys := []struct {
		path string
		arch string
	}{
		{`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`, "64-bit"},
		{`SOFTWARE\Wow6432Node\Microsoft\Windows\CurrentVersion\Uninstall`, "32-bit"},
	}

	seen := make(map[string]bool)

	for _, b := range baseKeys {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, b.path, registry.READ)
		if err != nil {
			continue
		}
		subs, err := k.ReadSubKeyNames(-1)
		if err != nil {
			k.Close()
			continue
		}
		for _, sub := range subs {
			sk, err := registry.OpenKey(registry.LOCAL_MACHINE, b.path+`\`+sub, registry.READ)
			if err != nil {
				continue
			}
			name, _, err := sk.GetStringValue("DisplayName")
			if err != nil || CleanString(name) == "" {
				sk.Close()
				continue
			}
			key := name + "|" + b.arch
			if seen[key] {
				sk.Close()
				continue
			}
			seen[key] = true

			ver, _, _ := sk.GetStringValue("DisplayVersion")
			pub, _, _ := sk.GetStringValue("Publisher")
			instDate, _, _ := sk.GetStringValue("InstallDate")
			est, _, _ := sk.GetIntegerValue("EstimatedSize")

			dateFmt := ""
			if t, err := time.Parse("20060102", instDate); err == nil {
				dateFmt = t.Format("02/01/2006")
			}
			size := ""
			if est > 0 {
				size = ByteCountSI(est * 1024)
			}

			rep.Software = append(rep.Software, SoftwareInfo{
				Name:        CleanString(name),
				Version:     CleanString(ver),
				Publisher:   CleanString(pub),
				InstallDate: dateFmt,
				Size:        size,
				Arch:        b.arch,
			})
			sk.Close()
		}
		k.Close()
	}

	sort.Slice(rep.Software, func(i, j int) bool {
		return rep.Software[i].Name < rep.Software[j].Name
	})
}

// collectNetworkFolders lists mapped network drives with their remote (UNC)
// destination and usage.
func collectNetworkFolders(rep *Report) {
	type wmiLogicalDisk struct {
		DeviceID     string
		DriveType    uint32
		ProviderName string
		VolumeName   string
		FileSystem   string
		Size         uint64
		FreeSpace    uint64
	}
	client := &wmi.Client{NonePtrZero: true, AllowMissingFields: true}
	var disks []wmiLogicalDisk
	if err := client.Query(
		`SELECT DeviceID, DriveType, ProviderName, VolumeName, FileSystem, Size, FreeSpace FROM Win32_LogicalDisk`,
		&disks, ""); err != nil {
		addError(rep, fmt.Sprintf("wmi logical disk: %v", err))
		return
	}
	for _, d := range disks {
		if d.DriveType != 4 || strings.TrimSpace(d.ProviderName) == "" {
			continue
		}
		used := d.Size - d.FreeSpace
		percent := 0.0
		if d.Size > 0 {
			percent = float64(used) / float64(d.Size) * 100
		}
		rep.NetFolders = append(rep.NetFolders, NetFolder{
			Device:     CleanString(d.DeviceID),
			Remote:     CleanString(d.ProviderName),
			Volume:     CleanString(d.VolumeName),
			FileSystem: CleanString(d.FileSystem),
			Total:      ByteCountSI(d.Size),
			Used:       ByteCountSI(used),
			Free:       ByteCountSI(d.FreeSpace),
			Percent:    math.Round(percent*10) / 10,
		})
	}
}

// collectSensors reads ACPI thermal zones (root\wmi). Many systems report
// 27.0 °C when no sensor is exposed; values may be placeholders.
func collectSensors(rep *Report) {
	type msAcpiThermalZone struct {
		InstanceName       string
		CurrentTemperature uint32
	}
	client := &wmi.Client{NonePtrZero: true, AllowMissingFields: true}
	var zones []msAcpiThermalZone
	err := client.Query(
		`SELECT InstanceName, CurrentTemperature FROM MSAcpi_ThermalZoneTemperature`,
		&zones, "", `root\wmi`, "", "",
	)
	if err != nil {
		// ACPI thermal zones are optional; do not pollute the report errors.
		return
	}
	for _, z := range zones {
		if z.CurrentTemperature == 0 {
			continue
		}
		celsius := float64(z.CurrentTemperature)/10.0 - 273.15
		if celsius < -50 || celsius > 250 {
			continue
		}
		label := StripAll(z.InstanceName)
		if label == "" {
			label = "Zona térmica"
		}
		rep.Temps = append(rep.Temps, TempSensor{
			Label: label,
			TempC: fmt.Sprintf("%.1f °C", celsius),
		})
	}
}

// collectPCI enumerates PCI/PCIe devices via WMI PnP, HWiNFO-style.
// It also flags the most likely chipset-related devices.
func collectPCI(rep *Report) {
	type pnpEntity struct {
		Name         string
		DeviceID     string
		Manufacturer string
		PNPClass     string
		Status       string
	}
	client := &wmi.Client{NonePtrZero: true, AllowMissingFields: true}
	var ents []pnpEntity
	err := client.Query(
		`SELECT Name, DeviceID, Manufacturer, PNPClass, Status FROM Win32_PnPEntity WHERE DeviceID LIKE 'PCI%'`,
		&ents, "",
	)
	if err != nil {
		addError(rep, fmt.Sprintf("wmi pci: %v", err))
		return
	}

	seen := make(map[string]bool)
	for _, e := range ents {
		id := CleanString(e.DeviceID)
		idx := strings.Index(id, `PCI\`)
		if idx < 0 {
			continue
		}
		rest := id[idx+4:]
		prefix := ""
		if n := strings.Index(rest, `\`); n >= 0 {
			prefix = rest[:n]
		} else {
			prefix = rest
		}
		if !strings.Contains(strings.ToUpper(prefix), "VEN_") {
			continue
		}
		vendor, device := parsePCIVenDev(prefix)
		key := vendor + "&" + device + "|" + CleanString(e.Name)
		if seen[key] {
			continue
		}
		seen[key] = true

		info := PCIInfo{
			VendorID:     vendor,
			DeviceID:     device,
			Name:         CleanString(e.Name),
			Manufacturer: CleanString(e.Manufacturer),
			Class:        CleanString(e.PNPClass),
			Status:       CleanString(e.Status),
		}
		rep.PCI = append(rep.PCI, info)

		if isChipsetCandidate(info.Name) {
			rep.Chipset = append(rep.Chipset, info)
		}
	}
}

// parsePCIVenDev extracts the VEN_/DEV_ hex values from a PnP prefix
// like "VEN_8086&DEV_8C50&SUBSYS_8C501462&REV_05".
func parsePCIVenDev(prefix string) (string, string) {
	parts := strings.Split(prefix, "&")
	var ven, dev string
	for _, p := range parts {
		up := strings.ToUpper(p)
		if strings.HasPrefix(up, "VEN_") {
			ven = strings.TrimPrefix(up, "VEN_")
		}
		if strings.HasPrefix(up, "DEV_") {
			dev = strings.TrimPrefix(up, "DEV_")
		}
	}
	if ven != "" {
		ven = "0x" + ven
	}
	if dev != "" {
		dev = "0x" + dev
	}
	return ven, dev
}

// isChipsetCandidate heuristically detects the devices that make up the
// motherboard chipset / northbridge southbridge, matching HWiNFO's listing.
func isChipsetCandidate(name string) bool {
	lower := strings.ToLower(name)
	for _, kw := range []string{
		"chipset",
		"host bridge",
		"lpc controller",
		"pch",
		"home agent",
		"memory controller",
		"southbridge",
		"northbridge",
		"system agent",
		"iommu",
		"root complex",
		"isa bridge",
		"pci-to-pci",
		"pci express root port",
		"pci express upstream",
		"pci express bridge",
		"processor dmi",
		"smbus",
		"thunderbolt bridge",
	} {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// collectEnclosure reads the physical chassis info of the machine.
func collectEnclosure(rep *Report) {
	type wmiEnclosure struct {
		Manufacturer string
		SerialNumber string
		SMBIOSAssetTag string
		Version      string
		ChassisTypes []uint16
	}
	client := &wmi.Client{NonePtrZero: true, AllowMissingFields: true}
	var enc []wmiEnclosure
	err := client.Query(
		`SELECT Manufacturer, SerialNumber, SMBIOSAssetTag, Version, ChassisTypes FROM Win32_SystemEnclosure`,
		&enc, "",
	)
	if err != nil {
		addError(rep, fmt.Sprintf("wmi enclosure: %v", err))
		return
	}
	for _, e := range enc {
		rep.Enclosures = append(rep.Enclosures, EnclosureInfo{
			Manufacturer: e.Manufacturer,
			ChassisType:  chassisTypeName(e.ChassisTypes),
			Serial:       e.SerialNumber,
			AssetTag:     e.SMBIOSAssetTag,
			Version:      e.Version,
		})
	}
}

func chassisTypeName(types []uint16) string {
	m := map[uint16]string{
		1: "Outro", 2: "Desconhecido", 3: "Desktop", 4: "Low Profile Desktop",
		5: "Pizza Box", 6: "Mini Tower", 7: "Tower", 8: "Portátil", 9: "Laptop",
		10: "Notebook", 11: "Handheld", 12: "Docking Station", 13: "All-in-One",
		14: "Sub Notebook", 15: "Space-saving", 16: "Lunch Box", 17: "Main Server",
		18: "Expansion", 19: "SubChassis", 20: "Bus Expansion", 21: "Peripheral",
		22: "RAID", 23: "Rack Mount", 24: "Sealed-case PC", 25: "Multi-system",
		26: "Compact PCI", 27: "Advanced TCA", 28: "Blade", 29: "Blade Enclosure",
		30: "Tablet", 31: "Convertible", 32: "Detachable", 33: "IoT Gateway",
		34: "Embedded", 35: "Mini PC", 36: "Stick PC",
	}
	var names []string
	for _, t := range types {
		if s, ok := m[t]; ok {
			names = append(names, s)
		} else {
			names = append(names, fmt.Sprintf("%d", t))
		}
	}
	return strings.Join(names, ", ")
}

func StripAll(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\n")
	s = strings.Trim(s, "\r")
	return s
}

func memoryFormFactor(f uint16) string {
	m := map[uint16]string{
		0: "Desconhecido", 1: "Outro", 2: "SIP", 3: "DIP", 4: "ZIP", 5: "Proprietário",
		6: "SIMM", 7: "DIMM", 8: "TSOP", 9: "PGA", 10: "RIMM", 11: "SODIMM",
		12: "SRIMM", 13: "FB-DIMM",
	}
	if s, ok := m[f]; ok {
		return s
	}
	return fmt.Sprintf("%d", f)
}

func memoryType(t uint16) string {
	m := map[uint16]string{
		0: "Desconhecido", 20: "DDR", 21: "DDR2", 22: "DDR2 FB-DIMM", 24: "DDR3",
		26: "DDR4", 27: "LPDDR", 28: "LPDDR2", 29: "LPDDR3", 30: "LPDDR4",
		32: "DDR5", 34: "LPDDR5",
	}
	if s, ok := m[t]; ok {
		return s
	}
	return fmt.Sprintf("%d", t)
}

func domainRoleName(r uint16) string {
	m := map[uint16]string{
		0: "Standalone Workstation", 1: "Member Workstation",
		2: "Standalone Server", 3: "Member Server",
		4: "Backup Domain Controller", 5: "Primary Domain Controller",
	}
	if s, ok := m[r]; ok {
		return s
	}
	return fmt.Sprintf("%d", r)
}

func pcSystemType(t uint16) string {
	m := map[uint16]string{
		0: "Desconhecido", 1: "Desktop", 2: "Notebook / Mobile",
		3: "Workstation", 4: "Enterprise Server", 5: "SOHO Server",
		6: "Appliance PC", 7: "Performance Server", 8: "Slate / Tablet",
	}
	if s, ok := m[t]; ok {
		return s
	}
	return fmt.Sprintf("%d", t)
}