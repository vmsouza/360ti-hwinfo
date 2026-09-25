# 360ti HWiNFO

Windows hardware & software inventory in seconds. Collects all machine data locally (CPU, motherboard, memory, disks, network, BIOS, installed software, and more) and generates a **navigable HTML report** (left sidebar, SaaS-style), a **readable PDF**, and a **JSON** with everything structured — with nothing sent to any server. Customizable via `config.json` (logo as base64 + colors) and distributable with an NSIS installer.

> 🇧🇷 Versão em português: [README.ptbr.md](README.ptbr.md)

## How it works

1. Runs the collector (`360ti-hwinfo.exe`, no console window).
2. Reads `config.json` next to the executable (logo + colors) to style the report.
3. Collects data via WMI, CPUID and gopsutil.
4. Generates:
   - `reports/360ti-hwinfo.html` — navigable report with a left sidebar (HWiNFO-style), plus **Abrir PDF** and **Copiar JSON** buttons.
   - `reports/360ti-hwinfo.pdf` — A4 document with all sections as tables.
   - `reports/360ti-hwinfo.json` — structured data (also embedded in the HTML for the Copy JSON button).
5. Opens the HTML in the default browser.

The `reports` folder lives in the **user data** directory (`%AppData%\360ti-hwinfo\reports` on Windows), always writable even when installed under Program Files.

## What is collected

- Summary: hostname, OS, kernel, uptime, machine ID
- CPU: CPUID details (L1/L2/L3 cache, microarchitecture level, instruction set), per-socket processors
- Temperature sensors (ACPI)
- Motherboard, product (serial/UUID) and chassis (type)
- Memory: total/usage + modules (slot, capacity, speed, type)
- Physical disks (model, serial, interface, SSD/HDD/NVMe media, firmware) and partition usage
- Mapped network drives (remote UNC destination)
- Network adapters (IP, mask, gateway, DHCP, DNS, link speed)
- Video cards and monitors
- USB controllers
- PCI/PCIe devices and possible chipset components
- BIOS/UEFI (boot mode, SMBIOS spec, version)
- Installed software (from the registry)
- Processes (top by memory)

## Usage

```text
360ti-hwinfo.exe                   # collect and open the HTML in the browser
360ti-hwinfo.exe -out PATH         # generate the report at PATH (html + pdf)
360ti-hwinfo.exe -open=false       # generate without opening the browser
360ti-hwinfo.exe -report DIR       # write report.html/.pdf/.json/.dat into DIR
360ti-hwinfo.exe -timing           # write a timing log (diagnostics)
```

## Configuration (`config.json`)

The collector reads `config.json` next to the executable. The logo is embedded into the HTML as base64.

```json
{
  "app_name": "360ti HWiNFO",
  "company_name": "360ti",
  "logo": "logo360ti.png",
  "logo_base64": "",
  "colors": {
    "accent": "#2563eb",
    "accent2": "#3b82f6",
    "accent3": "#0ea5e9",
    "accent4": "#60a5fa",
    "background": "#f3f5f9",
    "sidebar": "#ffffff",
    "surface": "#ffffff",
    "surface2": "#f1f5f9",
    "border": "#e5e9f0",
    "text": "#0f172a",
    "muted": "#64748b"
  }
}
```

- `logo`: path to the logo file (relative to the exe). You can also use `logo_base64` with the base64 (or data URI) of the logo.
- `colors`: overrides the theme (`app_name` title and PDF accent follow the theme).

## Building

Requires Go 1.20+.

```bash
# build.sh cross-compiles for Windows (amd64 and 386) -> dist/360ti-hwinfo.exe
./build.sh

# or manually
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -H windowsgui" -o dist/360ti-hwinfo.exe .

# local test build (Linux)
go build -o tmp/360ti-hwinfo .
```

## Installer (NSIS)

See `installer/360ti-hwinfo.nsi`. On Windows, with NSIS 3.x:

```text
makensis installer\360ti-hwinfo.nsi
```

Produces `360ti-hwinfo-setup.exe` (shortcuts, uninstaller, Add/Remove Programs). The script expects `dist\` with the files next to `installer\`.

## Layout

```text
.
├── main.go              # entry point: flags and flow (collect -> HTML/PDF/JSON)
├── collect.go           # cross-platform collection (host, cpu, memory, disks, net, processes)
├── collect_windows.go   # Windows collection: WMI, sensors, PCI, network folders, etc.
├── collect_other.go     # non-Windows stub
├── network_windows.go   # physical disks and detailed NICs (WMI)
├── cpuid.go             # CPU details via CPUID
├── config.go            # reads config.json: logo (base64) + colors
├── dataset.go           # shared sections (report.dat and PDF)
├── pdf.go               # PDF generation (embedded DejaVu font)
├── reportfiles.go       # writes HTML/JSON/DAT/PDF
├── report.go            # HTML template (embedded Bootstrap) + embedded JSON
├── timing.go            # timing diagnostics (optional)
├── assets.go            # embedded bootstrap css/js
├── browser_windows.go   # open browser / error box (Windows)
├── browser_other.go     # non-Windows stub
├── fonts/               # TTF fonts embedded in the PDF
├── static/              # bootstrap.min.css / bootstrap.bundle.min.js
├── config.json          # logo + colors (read next to the exe)
├── logo360ti.png        # default logo
├── dist/                # generated binaries + distribution config/logo
├── installer/           # NSIS + generated setup
├── README.md            # this file (EN)
├── README.ptbr.md       # PT-BR version
└── build.sh
```

## Platform

Targets **Windows** (amd64 and 386). The main delivery is the HTML/PDF report; collection uses WMI/CPUID only on Windows.