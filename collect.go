package main

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	gops "github.com/shirou/gopsutil/v3/process"
)

type Report struct {
	GeneratedAt string
	Host        HostInfo
	CPU         []CPUInfo
	CPUDetail   CPUIDetail
	Mem         MemInfo
	MemModules  []MemModuleInfo
	Disks       []DiskInfo
	NetFolders  []NetFolder
	DiskDevices []DiskDeviceInfo
	Net         []NetInfo
	NetDetailed []NetAdapterInfo
	Software    []SoftwareInfo
	BIOS        []BIOSInfo
	Board       []BoardInfo
	Product     []ProductInfo
	GPU         []GPUInfo
	Monitors    []MonitorInfo
	USB         []USBInfo
	Processes   []ProcessInfo
	Temps       []TempSensor
	PCI         []PCIInfo
	Chipset     []PCIInfo
	Enclosures  []EnclosureInfo
	Errors      []string
}

// TempSensor holds a temperature reading from the system.
type TempSensor struct {
	Label string
	TempC string
}

// PCIInfo holds a PCI/PCIe hardware device.
type PCIInfo struct {
	VendorID     string
	DeviceID     string
	Name         string
	Manufacturer string
	Class        string
	Status       string
}

// EnclosureInfo describes the physical chassis.
type EnclosureInfo struct {
	Manufacturer string
	ChassisType  string
	Serial       string
	AssetTag     string
	Version      string
}

type HostInfo struct {
	Hostname string
	OS       string
	Platform string
	Kernel   string
	Arch     string
	Uptime   string
	BootTime string
	HostID   string
}

type CPUInfo struct {
	Model     string
	Vendor    string
	Cores     int32
	Threads   int32
	Mhz       float64
	CacheSize int32
	Socket    string
	Family    string
	Current   uint32
	Max       uint32
	Voltage   string
}

// CPUIDetail holds deep processor details collected via CPUID, HWiNFO-style.
type CPUIDetail struct {
	Brand          string
	Vendor         string
	FamilyModel    string
	Stepping       string
	HWThreads      int
	PhysicalCores  int
	LogicalCores   int
	ThreadsPerCore int
	CacheLine      int
	L1I            string
	L1D            string
	L2             string
	L3             string
	BaseClock      string
	BoostClock     string
	X64Level       string
	VM             string
	Features       []string
	HasDetail      bool
}

type MemInfo struct {
	Total       string
	Available   string
	Used        string
	UsedPercent float64
}

type MemModuleInfo struct {
	Slot         string
	Bank         string
	Capacity     string
	Speed        uint32
	Manufacturer string
	PartNumber   string
	Serial       string
	FormFactor   string
	MemoryType   string
}

type DiskInfo struct {
	Device     string
	FileSystem string
	Total      string
	Used       string
	Free       string
	Percent    float64
}

// NetFolder holds a mapped network drive (\\server\share).
type NetFolder struct {
	Device     string
	Remote     string
	Volume     string
	FileSystem string
	Total      string
	Used       string
	Free       string
	Percent    float64
}

// DiskDeviceInfo holds the physical/hardware details of a disk drive.
type DiskDeviceInfo struct {
	Model          string
	Manufacturer   string
	Serial         string
	Interface      string
	Media          string
	Firmware       string
	Size           string
	Partitions     uint32
	BytesPerSector uint32
	Status         string
}

type NetInfo struct {
	Name string
	MAC  string
	MTU  int
	Addrs string
	Up   bool
}

// NetAdapterInfo holds the full configuration of a physical/logical NIC
// (IP, mask, gateway, DHCP, DNS, speed, link state), HWiNFO/HWiD-like.
type NetAdapterInfo struct {
	Name         string
	Manufacturer string
	Type         string
	Speed        string
	Status       string
	Physical     bool
	MAC          string
	IPs          string
	Subnets      string
	Gateways     string
	DHCP         string
	DHCPServer   string
	Lease        string
	DNS          string
}

type SoftwareInfo struct {
	Name        string
	Version     string
	Publisher   string
	InstallDate string
	Size        string
	Arch        string
}

type BIOSInfo struct {
	Manufacturer    string
	SMBIOSVersion   string
	Version         string
	ReleaseDate     string
	Serial          string
	BiosMode        string
	BiosVersion     string
	SmbiosSpec      string
	CurrentLanguage string
	Name            string
	Status          string
}

type BoardInfo struct {
	Manufacturer string
	Product      string
	Version      string
	Serial       string
	PartNumber   string
	SKU          string
	Family       string
	SystemType   string
	Domain       string
	Workgroup    string
	PartOfDomain bool
	DomainRole   string
	NumProc      uint32
	NumLogical   uint32
	ChassisSKU   string
	PcType       string
}

type ProductInfo struct {
	Vendor string
	Name   string
	Serial string
	UUID   string
	SKU    string
}

type GPUInfo struct {
	Name            string
	AdapterRAM      string
	DriverVersion   string
	DriverDate      string
	VideoProcessor  string
	Status          string
}

type MonitorInfo struct {
	Manufacturer string
	Name         string
	ScreenWidth  uint32
	ScreenHeight uint32
	MonitorType  string
}

type USBInfo struct {
	Name        string
	Manufacturer string
	Description string
	Status      string
	DeviceID    string
}

type ProcessInfo struct {
	Num  int
	Name string
	PID  int32
	Mem  string
	CPU  string
	User string
}

// Collect gathers all machine data and returns a populated report.
func Collect() *Report {
	rep := &Report{GeneratedAt: time.Now().Format("02/01/2006 15:04:05")}
	CollectInto(rep)
	return rep
}

// CollectInto fills the report with all available machine data.
func CollectInto(rep *Report) {
	defer writeTimingLog()
	timedStep("host", func() { collectHostInfo(rep) })
	timedStep("cpu (gopsutil)", func() { collectCPU(rep) })
	timedStep("cpu (cpuid)", func() { collectCPUID(rep) })
	timedStep("memoria", func() { collectMemory(rep) })
	timedStep("temperaturas do sistema", func() { collectTemps(rep) })
	timedStep("discos e particoes", func() { collectDisks(rep) })
	timedStep("rede (sistema)", func() { collectNetwork(rep) })
	timedStep("processos", func() { collectProcesses(rep) })
	timedStep("windows extras", func() { collectWindowsInfo(rep) })
}

// runWithTimeout executes fn in a goroutine, bounding its duration and
// recovering from panics so that a slow or broken collector can never
// block the report generation.
func runWithTimeout(d time.Duration, label string, fn func()) {
	done := make(chan struct{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("coletor [%s] recuperado de erro: %v\n", label, r)
			}
			close(done)
		}()
		fn()
	}()
	select {
	case <-done:
	case <-time.After(d):
		fmt.Printf("coletor [%s] demorou demais (>%v), seguindo sem esses dados\n", label, d)
	}
}

func collectHostInfo(rep *Report) {
	h, err := host.Info()
	if err != nil {
		rep.Errors = append(rep.Errors, fmt.Sprintf("host info: %v", err))
		return
	}
	rep.Host = HostInfo{
		Hostname: h.Hostname,
		Platform: h.Platform,
		Kernel:   h.KernelVersion,
		Arch:     runtime.GOARCH,
		Uptime:   formatDuration(h.Uptime),
		BootTime: time.Unix(int64(h.BootTime), 0).Format("02/01/2006 15:04:05"),
		HostID:   h.HostID,
	}
	rep.Host.OS = fmt.Sprintf("%s %s", h.Platform, h.PlatformVersion)
}

func collectCPU(rep *Report) {
	infos, err := cpu.Info()
	if err != nil {
		rep.Errors = append(rep.Errors, fmt.Sprintf("cpu info: %v", err))
		return
	}
	for _, c := range infos {
		rep.CPU = append(rep.CPU, CPUInfo{
			Model:     c.ModelName,
			Vendor:    c.VendorID,
			Cores:     c.Cores,
			Mhz:       c.Mhz,
			CacheSize: c.CacheSize,
			Socket:    c.PhysicalID,
		})
	}
}

func collectMemory(rep *Report) {
	m, err := mem.VirtualMemory()
	if err != nil {
		rep.Errors = append(rep.Errors, fmt.Sprintf("memory: %v", err))
		return
	}
	rep.Mem = MemInfo{
		Total:       ByteCountSI(m.Total),
		Available:   ByteCountSI(m.Available),
		Used:        ByteCountSI(m.Used),
		UsedPercent: math.Round(m.UsedPercent*10) / 10,
	}
}

func collectDisks(rep *Report) {
	parts, err := disk.Partitions(false)
	if err != nil {
		rep.Errors = append(rep.Errors, fmt.Sprintf("disks: %v", err))
		return
	}
	for _, p := range parts {
		u, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		rep.Disks = append(rep.Disks, DiskInfo{
			Device:     p.Device,
			FileSystem: p.Fstype,
			Total:      ByteCountSI(u.Total),
			Used:       ByteCountSI(u.Used),
			Free:       ByteCountSI(u.Free),
			Percent:    math.Round(u.UsedPercent*10) / 10,
		})
	}
}

func collectNetwork(rep *Report) {
	ifs, err := net.Interfaces()
	if err != nil {
		rep.Errors = append(rep.Errors, fmt.Sprintf("network: %v", err))
		return
	}
	for _, n := range ifs {
		var addrs []string
		for _, a := range n.Addrs {
			addrs = append(addrs, a.Addr)
		}
		isUp := false
		for _, f := range n.Flags {
			if f == "up" {
				isUp = true
			}
		}
		rep.Net = append(rep.Net, NetInfo{
			Name: n.Name,
			MAC:  n.HardwareAddr,
			MTU:  n.MTU,
			Addrs: strings.Join(addrs, ", "),
			Up:   isUp,
		})
	}
}

func collectProcesses(rep *Report) {
	procs, err := gops.Processes()
	if err != nil {
		rep.Errors = append(rep.Errors, fmt.Sprintf("processes: %v", err))
		return
	}

	type pInfo struct {
		ProcessInfo
		memBytes uint64
	}

	var list []pInfo
	for _, p := range procs {
		name, err := p.Name()
		if err != nil || name == "" {
			continue
		}
		m, err := p.MemoryInfo()
		if err != nil {
			continue
		}
		cpuP, _ := p.CPUPercent()
		user, _ := p.Username()
		list = append(list, pInfo{
			ProcessInfo: ProcessInfo{
				PID:  p.Pid,
				Name: name,
				Mem:  ByteCountSI(m.RSS),
				CPU:  strconv.FormatFloat(cpuP, 'f', 1, 64),
				User: user,
			},
			memBytes: m.RSS,
		})
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].memBytes > list[j].memBytes
	})
	const maxProcesses = 20
	if len(list) > maxProcesses {
		list = list[:maxProcesses]
	}
	for i, l := range list {
		l.Num = i + 1
		rep.Processes = append(rep.Processes, l.ProcessInfo)
	}
}

func collectTemps(rep *Report) {
	sensors, err := host.SensorsTemperatures()
	if err != nil {
		return
	}
	for _, s := range sensors {
		if s.Temperature <= 0 {
			continue
		}
		rep.Temps = append(rep.Temps, TempSensor{
			Label: s.SensorKey,
			TempC: strconv.FormatFloat(s.Temperature, 'f', 1, 64) + " °C",
		})
	}
}

func formatDuration(seconds uint64) string {
	if seconds == 0 {
		return "n/a"
	}
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	mins := (seconds % 3600) / 60
	return fmt.Sprintf("%d dia(s), %d hora(s), %d minuto(s)", days, hours, mins)
}

func ByteCountSI(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func CleanString(s string) string {
	r := strings.NewReplacer("\x00", "")
	s = r.Replace(s)
	return strings.ToValidUTF8(s, "")
}