package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// sectionData is a titled table of data, shared by report.dat and the PDF.
type sectionData struct {
	Title   string
	Headers []string
	Rows    [][]string
}

// buildSections converts the report into a list of tables.
func buildSections(rep *Report) []sectionData {
	var out []sectionData

	add := func(title string, headers []string, rows [][]string) {
		if rows == nil {
			rows = [][]string{}
		}
		out = append(out, sectionData{Title: title, Headers: headers, Rows: rows})
	}

	kv := func(pairs [][2]string) [][]string {
		rows := make([][]string, 0, len(pairs))
		for _, p := range pairs {
			rows = append(rows, []string{p[0], p[1]})
		}
		return rows
	}

	add("Resumo", []string{"Propriedade", "Valor"}, kv([][2]string{
		{"Hostname", rep.Host.Hostname},
		{"Sistema Operacional", rep.Host.OS},
		{"Plataforma", rep.Host.Platform + " · " + rep.Host.Arch},
		{"Kernel", rep.Host.Kernel},
		{"Uptime", rep.Host.Uptime},
		{"Inicializado em", rep.Host.BootTime},
		{"Machine ID", rep.Host.HostID},
		{"RAM Total", rep.Mem.Total},
		{"RAM em uso", fmt.Sprintf("%s (%.1f%%)", rep.Mem.Used, rep.Mem.UsedPercent)},
		{"Softwares instalados", fmt.Sprintf("%d", len(rep.Software))},
		{"Relatório gerado em", rep.GeneratedAt},
	}))

	if rep.CPUDetail.HasDetail {
		add("Detalhes do Processador", []string{"Propriedade", "Valor"}, kv([][2]string{
			{"Marca", rep.CPUDetail.Brand},
			{"Fabricante", rep.CPUDetail.Vendor},
			{"Família / Modelo", rep.CPUDetail.FamilyModel},
			{"Stepping", rep.CPUDetail.Stepping},
			{"Núcleos físicos", fmt.Sprintf("%d", rep.CPUDetail.PhysicalCores)},
			{"Threads totais", fmt.Sprintf("%d", rep.CPUDetail.LogicalCores)},
			{"Threads por núcleo", fmt.Sprintf("%d", rep.CPUDetail.ThreadsPerCore)},
			{"Clock base", rep.CPUDetail.BaseClock},
			{"Clock boost", rep.CPUDetail.BoostClock},
			{"Microarquitetura", rep.CPUDetail.X64Level},
			{"Linha de cache", cacheLineStr(rep.CPUDetail.CacheLine)},
			{"Virtualização", rep.CPUDetail.VM},
			{"Recursos", strings.Join(rep.CPUDetail.Features, ", ")},
		}))
	}

	if len(rep.CPU) > 0 {
		rows := make([][]string, 0, len(rep.CPU))
		for _, c := range rep.CPU {
			clk := fmt.Sprintf("%.0f MHz", c.Mhz)
			if c.Max > 0 {
				clk = fmt.Sprintf("%d / %d MHz", c.Current, c.Max)
			}
			rows = append(rows, []string{c.Model, c.Vendor, intStr(int(c.Cores)), intStr(int(c.Threads)), clk, c.Voltage, c.Socket})
		}
		add("Processadores", []string{"Modelo", "Fabricante", "Núcleos", "Threads", "Clock", "Voltagem", "Soquete"}, rows)
	}

	if len(rep.Temps) > 0 {
		rows := make([][]string, len(rep.Temps))
		for i, t := range rep.Temps {
			rows[i] = []string{t.Label, t.TempC}
		}
		add("Sensores", []string{"Sensor", "Temperatura"}, rows)
	}

	if len(rep.Board) > 0 {
		b := rep.Board[0]
		pairs := [][2]string{
			{"Fabricante", b.Manufacturer},
			{"Produto", b.Product},
			{"Versão", b.Version},
			{"Serial", b.Serial},
			{"Part Number", b.PartNumber},
			{"SKU", b.SKU},
			{"Família do sistema", b.Family},
			{"Tipo de sistema", b.SystemType},
			{"Tipo de PC", b.PcType},
			{"Processadores físicos", uintStr(b.NumProc)},
			{"Processadores lógicos", uintStr(b.NumLogical)},
			{"Domínio", b.Domain},
			{"Workgroup", b.Workgroup},
			{"Função no domínio", b.DomainRole},
			{"Chassis SKU", b.ChassisSKU},
		}
		if len(rep.Product) > 0 {
			p := rep.Product[0]
			pairs = append(pairs,
				[2]string{"Produto - fabricante", p.Vendor},
				[2]string{"Produto - modelo", p.Name},
				[2]string{"Produto - serial", p.Serial},
				[2]string{"Produto - UUID", p.UUID},
				[2]string{"Produto - SKU", p.SKU},
			)
		}
		if len(rep.Enclosures) > 0 {
			e := rep.Enclosures[0]
			pairs = append(pairs,
				[2]string{"Gabinete - fabricante", e.Manufacturer},
				[2]string{"Gabinete - tipo de chassi", e.ChassisType},
				[2]string{"Gabinete - serial", e.Serial},
				[2]string{"Gabinete - asset tag", e.AssetTag},
				[2]string{"Gabinete - versão", e.Version},
			)
		}
		add("Placa-Mãe", []string{"Propriedade", "Valor"}, kv(pairs))
	}

	add("Resumo Memória", []string{"Propriedade", "Valor"}, kv([][2]string{
		{"Total", rep.Mem.Total},
		{"Em uso", fmt.Sprintf("%s (%.1f%%)", rep.Mem.Used, rep.Mem.UsedPercent)},
		{"Disponível", rep.Mem.Available},
	}))
	if len(rep.MemModules) > 0 {
		rows := make([][]string, len(rep.MemModules))
		for i, m := range rep.MemModules {
			rows[i] = []string{m.Slot, m.Bank, m.Capacity, fmt.Sprintf("%d MHz", m.Speed), m.MemoryType + " (" + m.FormFactor + ")", m.Manufacturer, m.Serial}
		}
		add("Módulos de Memória", []string{"Slot", "Banco", "Capacidade", "Velocidade", "Tipo", "Fabricante", "Serial"}, rows)
	}

	if len(rep.DiskDevices) > 0 {
		rows := make([][]string, len(rep.DiskDevices))
		for i, d := range rep.DiskDevices {
			rows[i] = []string{d.Model, d.Manufacturer, d.Serial, d.Interface, d.Media, d.Firmware, d.Size, uintStr(d.Partitions), d.Status}
		}
		add("Discos Físicos", []string{"Modelo", "Fabricante", "Serial", "Interface", "Mídia", "Firmware", "Tamanho", "Partições", "Status"}, rows)
	}
	if len(rep.Disks) > 0 {
		rows := make([][]string, len(rep.Disks))
		for i, d := range rep.Disks {
			rows[i] = []string{d.Device, d.FileSystem, d.Total, d.Used, d.Free, fmt.Sprintf("%.1f%%", d.Percent)}
		}
		add("Partições e Uso", []string{"Partição", "FS", "Total", "Usado", "Livre", "Uso"}, rows)
	}
	if len(rep.NetFolders) > 0 {
		rows := make([][]string, len(rep.NetFolders))
		for i, n := range rep.NetFolders {
			rows[i] = []string{n.Device, n.Remote, n.Volume, n.FileSystem, n.Total, n.Used, n.Free, fmt.Sprintf("%.1f%%", n.Percent)}
		}
		add("Pastas de Rede", []string{"Letra", "Destino remoto", "Volume", "FS", "Total", "Usado", "Livre", "Uso"}, rows)
	}

	if len(rep.NetDetailed) > 0 {
		rows := make([][]string, len(rep.NetDetailed))
		for i, n := range rep.NetDetailed {
			rows[i] = []string{n.Name, n.Manufacturer, n.Type, n.Speed, n.Status, n.IPs, n.Subnets, n.Gateways, n.DHCP, n.DHCPServer, n.DNS}
		}
		add("Placas de Rede", []string{"Placa", "Fabricante", "Tipo", "Velocidade", "Estado", "IP", "Máscara", "Gateway", "DHCP", "Servidor DHCP", "DNS"}, rows)
	}
	if len(rep.Net) > 0 {
		rows := make([][]string, len(rep.Net))
		for i, n := range rep.Net {
			state := "Inativa"
			if n.Up {
				state = "Ativa"
			}
			rows[i] = []string{n.Name, state, n.Addrs, n.MAC, intStr(n.MTU)}
		}
		add("Interfaces do Sistema", []string{"Interface", "Status", "Endereços", "MAC", "MTU"}, rows)
	}

	if len(rep.GPU) > 0 {
		rows := make([][]string, len(rep.GPU))
		for i, g := range rep.GPU {
			rows[i] = []string{g.Name, g.AdapterRAM, g.DriverVersion, g.DriverDate, g.Status}
		}
		add("Placas de Vídeo", []string{"Nome", "VRAM", "Driver", "Data", "Status"}, rows)
	}
	if len(rep.Monitors) > 0 {
		rows := make([][]string, len(rep.Monitors))
		for i, m := range rep.Monitors {
			res := "n/a"
			if m.ScreenWidth > 0 {
				res = fmt.Sprintf("%dx%d", m.ScreenWidth, m.ScreenHeight)
			}
			rows[i] = []string{m.Manufacturer, m.Name, res}
		}
		add("Monitores", []string{"Fabricante", "Modelo", "Resolução"}, rows)
	}

	if len(rep.USB) > 0 {
		rows := make([][]string, len(rep.USB))
		for i, u := range rep.USB {
			rows[i] = []string{u.Name, u.Manufacturer, u.Description, u.Status}
		}
		add("USB", []string{"Nome", "Fabricante", "Descrição", "Status"}, rows)
	}

	if len(rep.Chipset) > 0 {
		rows := make([][]string, len(rep.Chipset))
		for i, c := range rep.Chipset {
			rows[i] = []string{c.Name, c.Manufacturer, c.VendorID, c.DeviceID, c.Status}
		}
		add("Chipset (possíveis componentes)", []string{"Nome", "Fabricante", "Vendor", "Device", "Status"}, rows)
	}
	if len(rep.PCI) > 0 {
		rows := make([][]string, len(rep.PCI))
		for i, c := range rep.PCI {
			rows[i] = []string{c.Name, c.Manufacturer, c.VendorID, c.DeviceID, c.Class, c.Status}
		}
		add("Dispositivos PCI / PCIe", []string{"Nome", "Fabricante", "Vendor", "Device", "Classe", "Status"}, rows)
	}

	if len(rep.BIOS) > 0 {
		b := rep.BIOS[0]
		add("BIOS", []string{"Propriedade", "Valor"}, kv([][2]string{
			{"Fabricante", b.Manufacturer},
			{"Nome", b.Name},
			{"Modo de boot", b.BiosMode},
			{"Versão SMBIOS", b.SMBIOSVersion},
			{"Especificação SMBIOS", b.SmbiosSpec},
			{"Versão", b.Version},
			{"Data de lançamento", b.ReleaseDate},
			{"Idioma atual", b.CurrentLanguage},
			{"Serial", b.Serial},
			{"Status", b.Status},
			{"Versão completa", b.BiosVersion},
		}))
	}

	if len(rep.Software) > 0 {
		rows := make([][]string, len(rep.Software))
		for i, s := range rep.Software {
			rows[i] = []string{s.Name, s.Version, s.Publisher, s.InstallDate, s.Size, s.Arch}
		}
		add("Software Instalado", []string{"Nome", "Versão", "Fabricante", "Instalado em", "Tamanho", "Arquitetura"}, rows)
	}

	if len(rep.Processes) > 0 {
		rows := make([][]string, len(rep.Processes))
		for i, p := range rep.Processes {
			rows[i] = []string{intStr(p.Num), p.Name, intStr(int(p.PID)), p.Mem, p.CPU, p.User}
		}
		add("Processos", []string{"#", "Nome", "PID", "Memória", "CPU %", "Usuário"}, rows)
	}

	return out
}

// writeDataset emits the dataset format consumed by the native launcher:
//
//	#S <section title>
//	#C <header1>\t<header2>\t...
//	#R <value1>\t<value2>\t...
func writeDataset(path string, rep *Report) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	for _, s := range buildSections(rep) {
		_, _ = w.WriteString("#S " + clean(s.Title) + "\n")
		_, _ = w.WriteString("#C " + clean(strings.Join(s.Headers, "\t")) + "\n")
		for _, r := range s.Rows {
			_, _ = w.WriteString("#R " + clean(strings.Join(r, "\t")) + "\n")
		}
	}
	return w.Flush()
}

func clean(s string) string {
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func intStr(i int) string {
	if i == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%d", i)
}

func uintStr(u uint32) string {
	if u == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%d", u)
}

func cacheLineStr(bytes int) string {
	if bytes <= 0 {
		return "n/a"
	}
	return fmt.Sprintf("%d bytes", bytes)
}