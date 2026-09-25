package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows/registry"
)

// collectDiskHardware reads Win32_DiskDrive for physical disk details.
func collectDiskHardware(rep *Report) {
	type wmiDiskDrive struct {
		Model            string
		Manufacturer     string
		SerialNumber     string
		InterfaceType    string
		MediaType        string
		FirmwareRevision string
		Size             uint64
		Partitions       uint32
		PNPDeviceID      string
		Status           string
		BytesPerSector   uint32
		Index            uint32
	}

	client := &wmi.Client{NonePtrZero: true, AllowMissingFields: true}
	var drives []wmiDiskDrive
	err := client.Query(
		`SELECT Model, Manufacturer, SerialNumber, InterfaceType, MediaType, FirmwareRevision, Size, Partitions, PNPDeviceID, Status, BytesPerSector, Index FROM Win32_DiskDrive`,
		&drives, "",
	)
	if err != nil {
		addError(rep, fmt.Sprintf("wmi disk drive: %v", err))
		return
	}

	for _, d := range drives {
		if d.Size == 0 && d.Model == "" {
			continue
		}
		rep.DiskDevices = append(rep.DiskDevices, DiskDeviceInfo{
			Model:          CleanString(d.Model),
			Manufacturer:   CleanString(d.Manufacturer),
			Serial:         CleanString(d.SerialNumber),
			Interface:      CleanString(d.InterfaceType),
			Media:          diskMediaType(d.Model, d.MediaType, d.PNPDeviceID, d.InterfaceType),
			Firmware:       CleanString(d.FirmwareRevision),
			Size:           ByteCountSI(d.Size),
			Partitions:     d.Partitions,
			BytesPerSector: d.BytesPerSector,
			Status:         CleanString(d.Status),
		})
	}
}

// diskMediaType classifies the physical media as SSD/NVMe/HDD.
func diskMediaType(model, mediaType, pnpID, iface string) string {
	all := strings.ToLower(model + " " + mediaType + " " + pnpID)
	switch {
	case strings.Contains(all, "nvme"):
		return "SSD (NVMe)"
	case strings.Contains(all, "solid state") || strings.Contains(all, "ssd"):
		return "SSD"
	case strings.Contains(all, "flash") || strings.Contains(all, "sata ssd"):
		return "SSD"
	default:
		return "HDD (disco rígido)"
	}
}

// collectNetworkDetailed builds a full NIC configuration table from
// Win32_NetworkAdapterConfiguration joined with Win32_NetworkAdapter.
func collectNetworkDetailed(rep *Report) {
	type wmiNetConfig struct {
		Index                int32
		Description          string
		MACAddress           string
		IPEnabled            bool
		DHCPEnabled          bool
		DHCPServer           string
		DHCPLeaseObtained    time.Time
		DHCPLeaseExpires     time.Time
		IPAddress            []string
		IPSubnet             []string
		DefaultIPGateway     []string
		DNSServerSearchOrder []string
		DNSDomain            string
		DNSHostName          string
	}
	type wmiNetAdapter struct {
		Index               uint32
		Name                string
		Manufacturer        string
		AdapterType         string
		MACAddress          string
		Speed               uint64
		MaxSpeed            uint64
		NetConnectionID     string
		NetConnectionStatus uint16
		PhysicalAdapter     bool
		NetEnabled          bool
	}

	client := &wmi.Client{NonePtrZero: true, AllowMissingFields: true}

	var cfg []wmiNetConfig
	if err := client.Query(
		`SELECT Index, Description, MACAddress, IPEnabled, DHCPEnabled, DHCPServer, DHCPLeaseObtained, DHCPLeaseExpires, IPAddress, IPSubnet, DefaultIPGateway, DNSServerSearchOrder, DNSDomain, DNSHostName FROM Win32_NetworkAdapterConfiguration`,
		&cfg, "",
	); err != nil {
		addError(rep, fmt.Sprintf("wmi net config: %v", err))
		return
	}

	var adapts []wmiNetAdapter
	client.Query(
		`SELECT Index, Name, Manufacturer, AdapterType, MACAddress, Speed, MaxSpeed, NetConnectionID, NetConnectionStatus, PhysicalAdapter, NetEnabled FROM Win32_NetworkAdapter`,
		&adapts, "",
	)

	byIdx := make(map[int32]*wmiNetAdapter)
	for i := range adapts {
		if adapts[i].MACAddress == "" {
			continue
		}
		byIdx[int32(adapts[i].Index)] = &adapts[i]
	}

	for _, c := range cfg {
		if !c.IPEnabled {
			continue
		}
		ad := byIdx[int32(c.Index)]

		info := NetAdapterInfo{
			Name:    CleanString(c.Description),
			MAC:     CleanString(c.MACAddress),
			IPs:     joinStrings(c.IPAddress),
			Subnets: joinStrings(c.IPSubnet),
			DNS:     joinStrings(c.DNSServerSearchOrder),
		}
		if len(c.DefaultIPGateway) > 0 {
			info.Gateways = joinStrings(c.DefaultIPGateway)
		}
		if c.DHCPEnabled {
			info.DHCP = "Sim"
			info.DHCPServer = CleanString(c.DHCPServer)
			obt := c.DHCPLeaseObtained
			exp := c.DHCPLeaseExpires
			if !obt.IsZero() || !exp.IsZero() {
				o := "n/d"
				e := "n/d"
				if !obt.IsZero() {
					o = obt.Format("02/01/2006 15:04:05")
				}
				if !exp.IsZero() {
					e = exp.Format("02/01/2006 15:04:05")
				}
				info.Lease = fmt.Sprintf("obtida %s · expira %s", o, e)
			}
		} else {
			info.DHCP = "Não (IP fixo)"
		}

		if ad != nil {
			info.Name = CleanString(firstNonEmpty(ad.Name, ad.NetConnectionID, c.Description))
			info.Manufacturer = CleanString(ad.Manufacturer)
			info.Type = CleanString(ad.AdapterType)
			info.Speed = formatLinkSpeed(ad.Speed)
			info.Status = netConnStatus(ad.NetConnectionStatus)
			info.Physical = ad.PhysicalAdapter
			if ad.MACAddress != "" {
				info.MAC = CleanString(ad.MACAddress)
			}
			if !ad.NetEnabled {
				if info.Status == "" {
					info.Status = "Desabilitada"
				}
			}
		}
		if info.Status == "" {
			info.Status = "n/a"
		}
		if info.Manufacturer == "" {
			info.Manufacturer = "n/a"
		}
		if info.Type == "" {
			info.Type = "n/a"
		}
		if info.Speed == "" {
			info.Speed = "n/a"
		}
		rep.NetDetailed = append(rep.NetDetailed, info)
	}
}

func joinStrings(list []string) string {
	var out []string
	for _, s := range list {
		if s == "" {
			continue
		}
		out = append(out, CleanString(s))
	}
	return strings.Join(out, ", ")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// formatLinkSpeed converts a bit-rate value into a readable speed.
func formatLinkSpeed(bits uint64) string {
	if bits == 0 || bits == ^uint64(0) {
		return ""
	}
	if bits >= 1_000_000_000 {
		return fmt.Sprintf("%.0f Gbps", float64(bits)/1_000_000_000)
	}
	if bits >= 1_000_000 {
		return fmt.Sprintf("%.0f Mbps", float64(bits)/1_000_000)
	}
	return fmt.Sprintf("%d Kbps", bits/1000)
}

// netConnStatus maps Win32_NetworkAdapter.NetConnectionStatus codes.
func netConnStatus(code uint16) string {
	m := map[uint16]string{
		0: "Desconectada", 1: "Conectando", 2: "Conectada", 3: "Desconectando",
		4: "Hardware ausente", 5: "Hardware desabilitado", 6: "Falha de hardware",
		7: "Mídia desconectada", 8: "Autenticando", 9: "Autenticação ok",
		10: "Falha de autenticação", 11: "Endereço inválido", 12: "Credenciais necessárias",
	}
	if s, ok := m[code]; ok {
		return s
	}
	return fmt.Sprintf("%d", code)
}

// collectBIOSMode detects whether the machine boots via UEFI or legacy BIOS.
func collectBIOSMode() string {
	if _, err := os.Stat(`C:\Windows\Boot\EFI\bootmgfw.efi`); err == nil {
		return "UEFI"
	}
	if _, err := os.Stat(`C:\Windows\Boot\PCAT\bootmgr`); err == nil {
		return "Legacy (BIOS)"
	}
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\SecureBoot\State`, registry.READ); err == nil {
		k.Close()
		return "UEFI"
	}
	return "n/a"
}