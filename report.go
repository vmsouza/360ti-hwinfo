package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"strings"
)

type reportCtx struct {
	Rep         *Report
	CSS         template.CSS
	JS          template.JS
	RepJSON     template.JS
	AppName     string
	CompanyName string
	LogoDataURI template.URL
	ThemeCSS    template.CSS
}

const reportTemplate = `<!DOCTYPE html>
<html lang="pt-BR" data-bs-theme="light">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>360ti HWiNFO - {{.Rep.Host.Hostname}}</title>
<style>
{{.CSS}}
</style>
<style>
:root {
	--bg: #f3f5f9;
	--sidebar: #ffffff;
	--surface: #ffffff;
	--surface-2: #f1f5f9;
	--border: #e5e9f0;
	--text: #0f172a;
	--muted: #64748b;
	--accent1: #2563eb;
	--accent2: #3b82f6;
	--accent3: #0ea5e9;
	--accent4: #60a5fa;
	--grad: linear-gradient(135deg, #2563eb, #3b82f6 45%, #0ea5e9);
	--radius: 18px;
}
html { scroll-behavior: smooth; }
body {
	background: var(--bg);
	color: var(--text);
	font-family: "Segoe UI", system-ui, -apple-system, "Inter", Roboto, Arial, sans-serif;
	overflow: hidden;
	height: 100vh;
}
/* soft blue glow */
body::before, body::after {
	content: ""; position: fixed; z-index: -1; border-radius: 50%;
	filter: blur(120px); opacity: .55; pointer-events: none;
}
body::before {
	width: 620px; height: 620px; top: -12%; left: -8%;
	background: radial-gradient(circle, rgba(37,99,235,.16), transparent 62%);
}
body::after {
	width: 760px; height: 760px; bottom: -18%; right: -10%;
	background: radial-gradient(circle, rgba(14,165,233,.14), transparent 62%);
	opacity: .4;
}

.grad-text {
	background: var(--grad);
	-webkit-background-clip: text;
	background-clip: text;
	color: transparent;
}

.app { display: flex; height: 100vh; }

/* sidebar */
.sidebar {
	width: 262px; flex-shrink: 0;
	position: sticky; top: 0; height: 100vh; overflow-y: auto;
	background: var(--sidebar);
	border-right: 1px solid var(--border);
	padding: 20px 14px;
}
.sidebar-brand { display: flex; flex-direction: column; align-items: center; gap: 8px; padding: 4px 6px 16px; }
.brand-img { max-height: 46px; max-width: 190px; object-fit: contain; }
.brand-title {
	font-size: 16px; font-weight: 800; color: var(--accent1); text-align: center;
	line-height: 1.2; text-transform: uppercase; letter-spacing: .03em;
}
.brand-company { font-size: 11px; font-weight: 600; color: var(--muted); letter-spacing: .05em; text-align: center; }
.side-label {
	font-size: 10px; text-transform: uppercase; letter-spacing: .16em;
	color: var(--muted); margin: 14px 10px 6px; font-weight: 700;
}
.nav-pills .nav-link {
	color: var(--text); border-radius: 12px; padding: .58rem .85rem;
	font-size: .89rem; font-weight: 500; margin-bottom: 3px;
	transition: all .16s ease; display: flex; align-items: center; gap: 10px;
}
.nav-pills .nav-link .ic { width: 20px; text-align: center; color: var(--muted); font-size: .95rem; }
.nav-pills .nav-link:hover { background: #eef2ff; color: var(--accent1); }
.nav-pills .nav-link.active {
	background: var(--grad); color: #fff; font-weight: 600;
	box-shadow: 0 8px 22px rgba(37,99,235,.3);
}
.nav-pills .nav-link.active .ic { color: #fff; }

/* content */
.content { flex: 1; min-width: 0; height: 100vh; overflow-y: auto; padding: 0 34px 40px; }
.topbar {
	position: sticky; top: 0; z-index: 120;
	display: flex; align-items: center; justify-content: space-between;
	gap: 14px; flex-wrap: wrap;
	background: var(--bg);
	padding: 18px 0 14px; margin-bottom: 20px;
	border-bottom: 1px solid var(--border);
}
.topbar h1 { font-size: 26px; font-weight: 800; margin: 0; letter-spacing: .01em; }
.topbar .meta { color: var(--muted); font-size: 13px; }
.topbar-actions { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.btn-action {
	background: linear-gradient(135deg, var(--accent1), var(--accent2));
	color: #fff; border: none; border-radius: 12px;
	padding: .62rem 1.35rem; font-size: .9rem; font-weight: 700; cursor: pointer;
	box-shadow: 0 8px 20px rgba(37,99,235,.25);
	transition: all .15s ease;
}
.btn-action:hover { transform: translateY(-1px); box-shadow: 0 10px 26px rgba(37,99,235,.35); }
.btn-glass {
	background: #fff; color: var(--text); border: 1px solid var(--border);
	border-radius: 12px; padding: .5rem 1rem; font-size: .87rem; font-weight: 600;
	transition: all .15s ease;
}
.btn-glass:hover { background: var(--surface-2); color: var(--accent1); border-color: #cbd5e1; }

.card {
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius);
	box-shadow: 0 8px 26px rgba(15,23,42,.07);
	margin-bottom: 24px;
	overflow: hidden;
}
.card-header {
	background: transparent; border-bottom: 1px solid var(--border);
	padding: 16px 22px; font-weight: 700; font-size: 13px; letter-spacing: .05em;
	text-transform: uppercase; color: #0f172a; display: flex; align-items: center; justify-content: space-between;
}
.card-header .count { color: var(--muted); font-size: 11px; letter-spacing: .02em; text-transform: none; font-weight: 500; }
.card-body { padding: 20px 22px; }
.hr { border-color: var(--border); opacity: 1; margin: 18px 0; }

.kv { display: grid; grid-template-columns: 210px 1fr; row-gap: 9px; font-size: .9rem; margin: 0; }
.kv dt { color: var(--muted); font-weight: 500; }
.kv dd { margin: 0; word-break: break-word; }

.table {
	--bs-table-bg: transparent; --bs-table-color: var(--text);
	--bs-table-border-color: var(--border); --bs-table-hover-bg: rgba(37,99,235,.05);
	font-size: .86rem; margin-bottom: 0;
}
.table thead th {
	color: var(--muted); text-transform: uppercase; font-size: 10.5px;
	letter-spacing: .08em; font-weight: 700; border-color: var(--border); white-space: nowrap;
}
.table tbody td { border-color: var(--border); vertical-align: middle; word-break: break-word; }
.table tbody tr:last-child td { border-bottom: none; }

.chip {
	display: inline-block; padding: 3px 11px; border-radius: 999px;
	background: rgba(37,99,235,.08); color: #1d4ed8;
	border: 1px solid rgba(37,99,235,.25); font-size: .72rem; font-weight: 600;
	margin: 0 4px 4px 0; letter-spacing: .02em;
}
.chip.cyan { background: rgba(16,185,129,.1); color: #065f46; border-color: rgba(16,185,129,.3); }
.chip.pink { background: rgba(239,68,68,.1); color: #b91c1c; border-color: rgba(239,68,68,.3); }

.badge-ok { background: rgba(16,185,129,.1); color: #065f46; border: 1px solid rgba(16,185,129,.3); }
.badge-soft { background: var(--surface-2); color: var(--muted); border: 1px solid var(--border); }

.progress { background: #e2e8f0; border-radius: 999px; height: 8px; }
.progress-bar { background-image: linear-gradient(90deg, var(--accent1), var(--accent3)); border-radius: 999px; }

.stat-card { text-align: center; padding: 6px 8px; }
.stat-card .v { font-size: 21px; font-weight: 800; }
.stat-card .l { color: var(--muted); font-size: 11px; text-transform: uppercase; letter-spacing: .08em; margin-top: 3px; }

.searchbox { padding: 14px 22px; border-bottom: 1px solid var(--border); }
.searchbox input {
	width: 100%; background: #fff; border: 1px solid #dbe3ee;
	border-radius: 12px; color: var(--text); padding: .55rem 1rem; font-size: .9rem; outline: none;
	transition: border-color .15s ease;
}
.searchbox input:focus { border-color: var(--accent1); box-shadow: 0 0 0 3px rgba(37,99,235,.15); }
.empty { color: var(--muted); font-size: .88rem; padding: 8px 0; }
.sensor-tile { text-align: center; padding: 14px 10px; border: 1px solid var(--border); border-radius: 16px; background: var(--surface); }
.sensor-tile .t { font-size: 18px; font-weight: 800; }
.sensor-tile .l { color: var(--muted); font-size: 11px; margin-top: 4px; word-break: break-word; }

/* instant tab switching (SaaS feel) */
.tab-pane.fade { transition: none !important; }
.tab-pane.fade:not(.show) { visibility: hidden; }
.tab-pane.fade.show { visibility: visible; opacity: 1; }

footer { text-align: center; color: var(--muted); font-size: 12px; padding: 8px 0 24px; }

@media (max-width: 991.98px) {
	body { overflow: auto; height: auto; }
	.app { flex-direction: column; height: auto; }
	.sidebar { width: 100%; height: auto; position: static; border-right: none; border-bottom: 1px solid var(--border); padding: 14px; }
	.sidebar .nav { flex-direction: row; flex-wrap: nowrap; overflow-x: auto; padding-bottom: 6px; }
	.sidebar .nav-link { white-space: nowrap; }
	.side-label { display: none; }
	.content { height: auto; overflow: visible; padding: 16px; }
	.topbar { position: static; margin-bottom: 16px; }
	.kv { grid-template-columns: 1fr; }
	.table thead { display: none; }
	.table, .table tbody, .table tr, .table td { display: block; width: 100%; }
	.table tr { border: 1px solid var(--border); border-radius: 14px; margin-bottom: 10px; padding: 6px 12px; }
	.table td { border: none; padding: 3px 2px; }
	.table td::before { content: attr(data-l); display: block; color: var(--muted); font-size: 10.5px; text-transform: uppercase; letter-spacing: .07em; }
}
@media print {
	body { overflow: visible; height: auto; }
	.sidebar { display: none; }
	.app { height: auto; }
	.content { height: auto; overflow: visible; padding: 0; }
	.topbar { position: static; background: none; border: none; }
	body::before, body::after { display: none; }
	body { background: #fff; }
}
</style>
<style>
{{.ThemeCSS}}
</style>
<body data-bs-theme="light">
<div class="app">
	<aside class="sidebar d-print-none">
		<div class="sidebar-brand">
			<img class="brand-img" src="{{.LogoDataURI}}" alt="360ti">
			<div class="brand-title">{{.AppName}}</div>
			<div class="brand-company">{{.CompanyName}}</div>
		</div>
		<div class="side-label">Navegação</div>
		<ul class="nav nav-pills flex-column" role="tablist">
			<li class="nav-item"><a class="nav-link active" data-bs-toggle="pill" data-bs-target="#pane-resumo" href="#pane-resumo"><span class="ic">◎</span> Resumo</a></li>
			{{if .Rep.CPUDetail.HasDetail}}
			<li class="nav-item"><a class="nav-link" data-bs-toggle="pill" data-bs-target="#pane-processador" href="#pane-processador"><span class="ic">⚙</span> Processador</a></li>
			{{end}}
			{{if .Rep.Temps}}
			<li class="nav-item"><a class="nav-link" data-bs-toggle="pill" data-bs-target="#pane-sensores" href="#pane-sensores"><span class="ic">◉</span> Sensores</a></li>
			{{end}}
			{{if .Rep.Board}}
			<li class="nav-item"><a class="nav-link" data-bs-toggle="pill" data-bs-target="#pane-placa" href="#pane-placa"><span class="ic">▦</span> Placa-Mãe</a></li>
			{{end}}
			<li class="nav-item"><a class="nav-link" data-bs-toggle="pill" data-bs-target="#pane-memoria" href="#pane-memoria"><span class="ic">▤</span> Memória</a></li>
			{{if or .Rep.Disks .Rep.DiskDevices}}
			<li class="nav-item"><a class="nav-link" data-bs-toggle="pill" data-bs-target="#pane-discos" href="#pane-discos"><span class="ic">⛁</span> Discos</a></li>
			{{end}}
			{{if or .Rep.NetDetailed .Rep.Net}}
			<li class="nav-item"><a class="nav-link" data-bs-toggle="pill" data-bs-target="#pane-rede" href="#pane-rede"><span class="ic">⌁</span> Rede</a></li>
			{{end}}
			{{if or .Rep.GPU .Rep.Monitors}}
			<li class="nav-item"><a class="nav-link" data-bs-toggle="pill" data-bs-target="#pane-video" href="#pane-video"><span class="ic">▨</span> Vídeo e Monitores</a></li>
			{{end}}
			{{if .Rep.USB}}
			<li class="nav-item"><a class="nav-link" data-bs-toggle="pill" data-bs-target="#pane-usb" href="#pane-usb"><span class="ic">⌖</span> USB</a></li>
			{{end}}
			{{if .Rep.PCI}}
			<li class="nav-item"><a class="nav-link" data-bs-toggle="pill" data-bs-target="#pane-pci" href="#pane-pci"><span class="ic">▣</span> PCI / Hardware</a></li>
			{{end}}
			{{if .Rep.BIOS}}
			<li class="nav-item"><a class="nav-link" data-bs-toggle="pill" data-bs-target="#pane-bios" href="#pane-bios"><span class="ic">◌</span> BIOS</a></li>
			{{end}}
			{{if .Rep.Software}}
			<li class="nav-item"><a class="nav-link" data-bs-toggle="pill" data-bs-target="#pane-software" href="#pane-software"><span class="ic">▧</span> Software</a></li>
			{{end}}
			{{if .Rep.Processes}}
			<li class="nav-item"><a class="nav-link" data-bs-toggle="pill" data-bs-target="#pane-processos" href="#pane-processos"><span class="ic">◪</span> Processos</a></li>
			{{end}}
		</ul>
	</aside>

	<main class="content">
		<div class="topbar">
			<div>
				<h1>{{.Rep.Host.Hostname}}</h1>
				<div class="meta">Relatório gerado em {{.Rep.GeneratedAt}}</div>
			</div>
			<div class="topbar-actions">
				<button class="btn-action" onclick="openReportPDF()">Abrir PDF</button>
				<button class="btn-action" onclick="copyReportJSON()">Copiar JSON</button>
				<span id="copyMsg" style="color:#059669;font-size:13px;font-weight:700"></span>
			</div>
		</div>

		<div class="tab-content">

			<!-- RESUMO -->
			<div class="tab-pane fade show active" id="pane-resumo" role="tabpanel">
				<div class="card">
					<div class="card-header">Resumo da Máquina</div>
					<div class="card-body">
						<div class="row g-3 mb-3">
							<div class="col-6 col-md-3"><div class="card stat-card"><div class="v grad-text">{{.Rep.Mem.Total}}</div><div class="l">RAM Total</div></div></div>
							{{if .Rep.CPU}}{{with (index .Rep.CPU 0)}}
							<div class="col-6 col-md-3"><div class="card stat-card"><div class="v grad-text">{{.Cores}} núcleos</div><div class="l">Processador</div></div></div>
							{{end}}{{end}}
							<div class="col-6 col-md-3"><div class="card stat-card"><div class="v grad-text">{{.Rep.Host.Arch}}</div><div class="l">Arquitetura</div></div></div>
							<div class="col-6 col-md-3"><div class="card stat-card"><div class="v grad-text">{{len .Rep.Software}}</div><div class="l">Softwares</div></div></div>
						</div>
						<dl class="kv">
							<dt>Hostname</dt><dd>{{.Rep.Host.Hostname}}</dd>
							<dt>Sistema Operacional</dt><dd>{{.Rep.Host.OS}}</dd>
							<dt>Plataforma</dt><dd>{{.Rep.Host.Platform}} · {{.Rep.Host.Arch}}</dd>
							<dt>Kernel</dt><dd>{{.Rep.Host.Kernel}}</dd>
							<dt>Uptime</dt><dd>{{.Rep.Host.Uptime}}</dd>
							<dt>Inicializado em</dt><dd>{{.Rep.Host.BootTime}}</dd>
							{{if .Rep.Host.HostID}}<dt>Machine ID</dt><dd>{{.Rep.Host.HostID}}</dd>{{end}}
						</dl>
						{{if .Rep.Errors}}
						<hr class="hr">
						<div class="empty">Observações: {{range .Rep.Errors}}<span>{{.}}; </span>{{end}}</div>
						{{end}}
					</div>
				</div>
			</div>

			<!-- PROCESSADOR -->
			{{if .Rep.CPUDetail.HasDetail}}
			<div class="tab-pane fade" id="pane-processador" role="tabpanel">
				<div class="card">
					<div class="card-header">Detalhes do Processador <span class="count">CPUID</span></div>
					<div class="card-body">
						<dl class="kv">
							<dt>Marca</dt><dd>{{.Rep.CPUDetail.Brand}}</dd>
							<dt>Fabricante</dt><dd>{{.Rep.CPUDetail.Vendor}}</dd>
							<dt>Família / Modelo</dt><dd>{{.Rep.CPUDetail.FamilyModel}}</dd>
							<dt>Stepping</dt><dd>{{.Rep.CPUDetail.Stepping}}</dd>
							<dt>Núcleos físicos</dt><dd>{{.Rep.CPUDetail.PhysicalCores}}</dd>
							<dt>Threads totais</dt><dd>{{.Rep.CPUDetail.LogicalCores}}</dd>
							<dt>Threads por núcleo</dt><dd>{{.Rep.CPUDetail.ThreadsPerCore}}</dd>
							<dt>Clock base</dt><dd>{{if .Rep.CPUDetail.BaseClock}}{{.Rep.CPUDetail.BaseClock}}{{else}}n/a{{end}}</dd>
							<dt>Clock boost</dt><dd>{{if .Rep.CPUDetail.BoostClock}}{{.Rep.CPUDetail.BoostClock}}{{else}}n/a{{end}}</dd>
							<dt>Microarquitetura</dt><dd>{{if .Rep.CPUDetail.X64Level}}{{.Rep.CPUDetail.X64Level}}{{else}}n/a{{end}}</dd>
							<dt>Linha de cache</dt><dd>{{if .Rep.CPUDetail.CacheLine}}{{.Rep.CPUDetail.CacheLine}} bytes{{else}}n/a{{end}}</dd>
							<dt>Virtualização</dt><dd>{{.Rep.CPUDetail.VM}}</dd>
						</dl>
						{{with .Rep.CPUDetail}}
						{{if .L1I}}
						<hr class="hr">
						<h6 class="text-uppercase text-muted mb-3" style="font-size:11px;letter-spacing:.08em;font-weight:700;">Cache</h6>
						<div class="table-responsive">
							<table class="table">
								<thead><tr><th>Nível</th><th>Tipo</th><th>Tamanho</th></tr></thead>
								<tbody>
									{{if .L1I}}<tr><td>L1</td><td>Instruções</td><td>{{.L1I}}</td></tr>{{end}}
									{{if .L1D}}<tr><td>L1</td><td>Dados</td><td>{{.L1D}}</td></tr>{{end}}
									{{if .L2}}<tr><td>L2</td><td>Cache</td><td>{{.L2}}</td></tr>{{end}}
									{{if .L3}}<tr><td>L3</td><td>Cache</td><td>{{.L3}}</td></tr>{{end}}
								</tbody>
							</table>
						</div>
						{{end}}
						{{if .Features}}
						<hr class="hr">
						<h6 class="text-uppercase text-muted mb-3" style="font-size:11px;letter-spacing:.08em;font-weight:700;">Conjunto de instruções ({{len .Features}})</h6>
						<div>{{range .Features}}<span class="chip">{{.}}</span>{{end}}</div>
						{{end}}
						{{end}}
					</div>
				</div>

				{{if .Rep.CPU}}
				<div class="card">
					<div class="card-header">Processadores <span class="count">{{len .Rep.CPU}} registro(s)</span></div>
					<div class="table-responsive">
						<table class="table">
							<thead><tr><th>Modelo</th><th>Fabricante</th><th>Núcleos</th><th>Threads</th><th>Clock</th><th>Voltagem</th><th>Soquete</th></tr></thead>
							<tbody>
							{{range .Rep.CPU}}
								<tr>
									<td data-l="Modelo">{{.Model}}</td>
									<td data-l="Fabricante">{{.Vendor}}</td>
									<td data-l="Núcleos">{{.Cores}}</td>
									<td data-l="Threads">{{if .Threads}}{{.Threads}}{{else}}n/a{{end}}</td>
									<td data-l="Clock">{{if .Max}}{{.Current}} / {{.Max}} MHz{{else}}{{.Mhz}} MHz{{end}}</td>
									<td data-l="Voltagem">{{if .Voltage}}{{.Voltage}}{{else}}n/a{{end}}</td>
									<td data-l="Soquete">{{.Socket}}</td>
								</tr>
							{{end}}
							</tbody>
						</table>
					</div>
				</div>
				{{end}}
			</div>
			{{end}}

			<!-- SENSORES -->
			{{if .Rep.Temps}}
			<div class="tab-pane fade" id="pane-sensores" role="tabpanel">
				<div class="card">
					<div class="card-header">Temperatura <span class="count">sensor(es) ACPI</span></div>
					<div class="card-body">
						<div class="row g-3">
							{{range .Rep.Temps}}
							<div class="col-6 col-md-4 col-lg-3">
								<div class="sensor-tile">
									<div class="t grad-text">{{.TempC}}</div>
									<div class="l">{{.Label}}</div>
								</div>
							</div>
							{{end}}
						</div>
						<div class="empty mt-3">Observação: muitos sistemas fixam a zona ACPI em ~27 °C quando o sensor real não é exposto ao Windows.</div>
					</div>
				</div>
			</div>
			{{end}}

			<!-- PLACA-MÃE -->
			{{if .Rep.Board}}
			<div class="tab-pane fade" id="pane-placa" role="tabpanel">
				{{range .Rep.Board}}
				<div class="card">
					<div class="card-header">Placa-Mãe</div>
					<div class="card-body">
						<dl class="kv">
							<dt>Fabricante</dt><dd>{{.Manufacturer}}</dd>
							<dt>Produto</dt><dd>{{.Product}}</dd>
							<dt>Versão</dt><dd>{{.Version}}</dd>
							<dt>Serial</dt><dd>{{.Serial}}</dd>
							<dt>Part Number</dt><dd>{{.PartNumber}}</dd>
							<dt>SKU</dt><dd>{{.SKU}}</dd>
							{{if .Family}}<dt>Família do sistema</dt><dd>{{.Family}}</dd>{{end}}
							{{if .SystemType}}<dt>Tipo de sistema</dt><dd>{{.SystemType}}</dd>{{end}}
							{{if .PcType}}<dt>Tipo de PC</dt><dd>{{.PcType}}</dd>{{end}}
							{{if .NumProc}}<dt>Processadores físicos</dt><dd>{{.NumProc}}</dd>{{end}}
							{{if .NumLogical}}<dt>Processadores lógicos</dt><dd>{{.NumLogical}}</dd>{{end}}
							{{if .Domain}}<dt>Domínio</dt><dd>{{.Domain}} {{if .PartOfDomain}}<span class="chip cyan">membro</span>{{end}}</dd>{{end}}
							{{if .Workgroup}}<dt>Workgroup</dt><dd>{{.Workgroup}}</dd>{{end}}
							{{if .DomainRole}}<dt>Função no domínio</dt><dd>{{.DomainRole}}</dd>{{end}}
							{{if .ChassisSKU}}<dt>Chassis SKU</dt><dd>{{.ChassisSKU}}</dd>{{end}}
						</dl>
					</div>
				</div>
				{{end}}

				{{range .Rep.Product}}
				<div class="card">
					<div class="card-header">Produto</div>
					<div class="card-body">
						<dl class="kv">
							<dt>Fabricante</dt><dd>{{.Vendor}}</dd>
							<dt>Modelo</dt><dd>{{.Name}}</dd>
							<dt>Nº de Série</dt><dd>{{.Serial}}</dd>
							<dt>UUID</dt><dd>{{.UUID}}</dd>
							<dt>SKU</dt><dd>{{.SKU}}</dd>
						</dl>
					</div>
				</div>
				{{end}}

				{{range .Rep.Enclosures}}
				<div class="card">
					<div class="card-header">Gabinete</div>
					<div class="card-body">
						<dl class="kv">
							<dt>Fabricante</dt><dd>{{.Manufacturer}}</dd>
							<dt>Tipo de chassi</dt><dd>{{.ChassisType}}</dd>
							<dt>Versão</dt><dd>{{.Version}}</dd>
							<dt>Serial</dt><dd>{{.Serial}}</dd>
							<dt>Asset Tag</dt><dd>{{.AssetTag}}</dd>
						</dl>
					</div>
				</div>
				{{end}}
			</div>
			{{end}}

			<!-- MEMÓRIA -->
			<div class="tab-pane fade" id="pane-memoria" role="tabpanel">
				<div class="card">
					<div class="card-header">Memória Física</div>
					<div class="card-body">
						<dl class="kv">
							<dt>Total</dt><dd>{{.Rep.Mem.Total}}</dd>
							<dt>Em uso</dt><dd>{{.Rep.Mem.Used}} ({{.Rep.Mem.UsedPercent}}%)</dd>
							<dt>Disponível</dt><dd>{{.Rep.Mem.Available}}</dd>
						</dl>
						<hr class="hr">
						<h6 class="text-uppercase text-muted mb-3" style="font-size:11px;letter-spacing:.08em;font-weight:700;">Módulos ({{len .Rep.MemModules}})</h6>
						<div class="table-responsive">
							<table class="table">
								<thead><tr><th>Slot</th><th>Banco</th><th>Capacidade</th><th>Velocidade</th><th>Tipo</th><th>Fabricante</th><th>Serial</th></tr></thead>
								<tbody>
								{{range .Rep.MemModules}}
									<tr>
										<td data-l="Slot">{{.Slot}}</td>
										<td data-l="Banco">{{.Bank}}</td>
										<td data-l="Capacidade">{{.Capacity}}</td>
										<td data-l="Velocidade">{{.Speed}} MHz</td>
										<td data-l="Tipo">{{.MemoryType}} ({{.FormFactor}})</td>
										<td data-l="Fabricante">{{.Manufacturer}}</td>
										<td data-l="Serial">{{.Serial}}</td>
									</tr>
								{{end}}
								</tbody>
							</table>
						</div>
					</div>
				</div>
			</div>

			<!-- DISCOS -->
			{{if or .Rep.Disks .Rep.DiskDevices}}
			<div class="tab-pane fade" id="pane-discos" role="tabpanel">
				{{if .Rep.DiskDevices}}
				<div class="card">
					<div class="card-header">Discos Físicos <span class="count">hardware</span></div>
					<div class="table-responsive">
						<table class="table">
							<thead><tr><th>Modelo</th><th>Fabricante</th><th>Serial</th><th>Interface</th><th>Mídia</th><th>Firmware</th><th>Tamanho</th><th>Partições</th><th>Status</th></tr></thead>
							<tbody>
							{{range .Rep.DiskDevices}}
								<tr>
									<td data-l="Modelo">{{.Model}}</td>
									<td data-l="Fabricante">{{.Manufacturer}}</td>
									<td data-l="Serial">{{.Serial}}</td>
									<td data-l="Interface">{{.Interface}}</td>
									<td data-l="Mídia">{{if eq .Media "SSD (NVMe)"}}<span class="chip cyan">{{.Media}}</span>{{else if eq .Media "SSD"}}<span class="chip cyan">{{.Media}}</span>{{else}}<span class="chip">{{.Media}}</span>{{end}}</td>
									<td data-l="Firmware">{{.Firmware}}</td>
									<td data-l="Tamanho">{{.Size}}</td>
									<td data-l="Partições">{{.Partitions}}</td>
									<td data-l="Status">{{.Status}}</td>
								</tr>
							{{end}}
							</tbody>
						</table>
					</div>
				</div>
				{{end}}
				{{if .Rep.NetFolders}}
				<div class="card">
					<div class="card-header">Pastas de Rede <span class="count">mapeadas</span></div>
					<div class="table-responsive">
						<table class="table">
							<thead><tr><th>Letra</th><th>Destino remoto</th><th>Volume</th><th>FS</th><th>Total</th><th>Usado</th><th>Livre</th><th>Uso</th></tr></thead>
							<tbody>
							{{range .Rep.NetFolders}}
								<tr>
									<td data-l="Letra"><code>{{.Device}}</code></td>
									<td data-l="Destino remoto">{{.Remote}}</td>
									<td data-l="Volume">{{.Volume}}</td>
									<td data-l="FS">{{.FileSystem}}</td>
									<td data-l="Total">{{.Total}}</td>
									<td data-l="Usado">{{.Used}}</td>
									<td data-l="Livre">{{.Free}}</td>
									<td data-l="Uso">
										<div class="d-flex align-items-center gap-2">
											<div class="progress flex-grow-1" style="max-width:180px"><div class="progress-bar" style="width:{{.Percent}}%"></div></div>
											<span class="text-nowrap">{{.Percent}}%</span>
										</div>
									</td>
								</tr>
							{{end}}
							</tbody>
						</table>
					</div>
				</div>
				{{end}}
				<div class="card">
					<div class="card-header">Partições e Uso</div>
					<div class="table-responsive">
						<table class="table">
							<thead><tr><th>Partição</th><th>FS</th><th>Total</th><th>Usado</th><th>Livre</th><th>Uso</th></tr></thead>
							<tbody>
							{{range .Rep.Disks}}
								<tr>
									<td data-l="Partição">{{.Device}}</td>
									<td data-l="FS">{{.FileSystem}}</td>
									<td data-l="Total">{{.Total}}</td>
									<td data-l="Usado">{{.Used}}</td>
									<td data-l="Livre">{{.Free}}</td>
									<td data-l="Uso">
										<div class="d-flex align-items-center gap-2">
											<div class="progress flex-grow-1" style="max-width:180px"><div class="progress-bar" style="width:{{.Percent}}%"></div></div>
											<span class="text-nowrap">{{.Percent}}%</span>
										</div>
									</td>
								</tr>
							{{end}}
							</tbody>
						</table>
					</div>
				</div>
			</div>
			{{end}}

			<!-- REDE -->
			{{if or .Rep.NetDetailed .Rep.Net}}
			<div class="tab-pane fade" id="pane-rede" role="tabpanel">
				{{if .Rep.NetDetailed}}
				<div class="card">
					<div class="card-header">Placas de Rede <span class="count">configuração</span></div>
					<div class="table-responsive">
						<table class="table">
							<thead><tr><th>Placa</th><th>Fabricante</th><th>Tipo</th><th>Velocidade</th><th>Estado</th><th>IP</th><th>Máscara</th><th>Gateway</th><th>DHCP</th><th>Servidor DHCP</th><th>DNS</th></tr></thead>
							<tbody>
							{{range .Rep.NetDetailed}}
								<tr>
									<td data-l="Placa">{{.Name}}</td>
									<td data-l="Fabricante">{{.Manufacturer}}</td>
									<td data-l="Tipo">{{.Type}}</td>
									<td data-l="Velocidade">{{.Speed}}</td>
									<td data-l="Estado">{{if eq .Status "Conectada"}}<span class="chip cyan">Conectada</span>{{else}}<span class="chip pink">{{.Status}}</span>{{end}}</td>
									<td data-l="IP">{{.IPs}}</td>
									<td data-l="Máscara">{{.Subnets}}</td>
									<td data-l="Gateway">{{.Gateways}}</td>
									<td data-l="DHCP">{{.DHCP}}</td>
									<td data-l="Servidor DHCP">{{.DHCPServer}}</td>
									<td data-l="DNS">{{.DNS}}</td>
								</tr>
								{{if .Lease}}<tr class="lease-row" style="border:none;background:transparent"><td colspan="11"><div class="empty">Lease DHCP: {{.Lease}}</div></td></tr>{{end}}
							{{end}}
							</tbody>
						</table>
					</div>
				</div>
				{{end}}

				{{if .Rep.Net}}
				<div class="card">
					<div class="card-header">Interfaces do Sistema</div>
					<div class="table-responsive">
						<table class="table">
							<thead><tr><th>Interface</th><th>Status</th><th>Endereços</th><th>MAC</th><th>MTU</th></tr></thead>
							<tbody>
							{{range .Rep.Net}}
								<tr>
									<td data-l="Interface">{{.Name}}</td>
									<td data-l="Status">{{if .Up}}<span class="chip cyan">Ativa</span>{{else}}<span class="chip pink">Inativa</span>{{end}}</td>
									<td data-l="Endereços">{{.Addrs}}</td>
									<td data-l="MAC">{{.MAC}}</td>
									<td data-l="MTU">{{.MTU}}</td>
								</tr>
							{{end}}
							</tbody>
						</table>
					</div>
				</div>
				{{end}}
			</div>
			{{end}}

			<!-- VÍDEO E MONITORES -->
			{{if or .Rep.GPU .Rep.Monitors}}
			<div class="tab-pane fade" id="pane-video" role="tabpanel">
				{{if .Rep.GPU}}
				<div class="card">
					<div class="card-header">Placas de Vídeo</div>
					<div class="table-responsive">
						<table class="table">
							<thead><tr><th>Nome</th><th>VRAM</th><th>Driver</th><th>Data</th><th>Status</th></tr></thead>
							<tbody>
							{{range .Rep.GPU}}
								<tr>
									<td data-l="Nome">{{.Name}}</td>
									<td data-l="VRAM">{{.AdapterRAM}}</td>
									<td data-l="Driver">{{.DriverVersion}}</td>
									<td data-l="Data">{{.DriverDate}}</td>
									<td data-l="Status">{{.Status}}</td>
								</tr>
							{{end}}
							</tbody>
						</table>
					</div>
				</div>
				{{end}}
				{{if .Rep.Monitors}}
				<div class="card">
					<div class="card-header">Monitores</div>
					<div class="table-responsive">
						<table class="table">
							<thead><tr><th>Fabricante</th><th>Modelo</th><th>Resolução</th></tr></thead>
							<tbody>
							{{range .Rep.Monitors}}
								<tr>
									<td data-l="Fabricante">{{.Manufacturer}}</td>
									<td data-l="Modelo">{{.Name}}</td>
									<td data-l="Resolução">{{if .ScreenWidth}}{{.ScreenWidth}}x{{.ScreenHeight}}{{else}}n/a{{end}}</td>
								</tr>
							{{end}}
							</tbody>
						</table>
					</div>
				</div>
				{{end}}
			</div>
			{{end}}

			<!-- USB -->
			{{if .Rep.USB}}
			<div class="tab-pane fade" id="pane-usb" role="tabpanel">
				<div class="card">
					<div class="card-header">Controladores USB</div>
					<div class="table-responsive">
						<table class="table">
							<thead><tr><th>Nome</th><th>Fabricante</th><th>Descrição</th><th>Status</th></tr></thead>
							<tbody>
							{{range .Rep.USB}}
								<tr>
									<td data-l="Nome">{{.Name}}</td>
									<td data-l="Fabricante">{{.Manufacturer}}</td>
									<td data-l="Descrição">{{.Description}}</td>
									<td data-l="Status">{{.Status}}</td>
								</tr>
							{{end}}
							</tbody>
						</table>
					</div>
				</div>
			</div>
			{{end}}

			<!-- PCI / HARDWARE -->
			{{if .Rep.PCI}}
			<div class="tab-pane fade" id="pane-pci" role="tabpanel">
				{{if .Rep.Chipset}}
				<div class="card">
					<div class="card-header">Chipset / Ponte Norte e Sul <span class="count">possíveis componentes</span></div>
					<div class="table-responsive">
						<table class="table">
							<thead><tr><th>Nome</th><th>Fabricante</th><th>Vendor</th><th>Device</th><th>Status</th></tr></thead>
							<tbody>
							{{range .Rep.Chipset}}
								<tr>
									<td data-l="Nome">{{.Name}}</td>
									<td data-l="Fabricante">{{.Manufacturer}}</td>
									<td data-l="Vendor"><code>{{.VendorID}}</code></td>
									<td data-l="Device"><code>{{.DeviceID}}</code></td>
									<td data-l="Status">{{.Status}}</td>
								</tr>
							{{end}}
							</tbody>
						</table>
					</div>
				</div>
				{{end}}

				<div class="card">
					<div class="card-header">Dispositivos PCI / PCIe <span class="count">{{len .Rep.PCI}} dispositivos</span></div>
					<div class="card-body">
						<div class="searchbox px-0">
							<input type="search" id="pciSearch" placeholder="Filtrar dispositivos PCI..." oninput="filterTable('pciSearch','pciTable')">
						</div>
					</div>
					<div class="table-responsive">
						<table class="table" id="pciTable">
							<thead><tr><th>Nome</th><th>Fabricante</th><th>Vendor</th><th>Device</th><th>Classe</th><th>Status</th></tr></thead>
							<tbody>
							{{range .Rep.PCI}}
								<tr>
									<td data-l="Nome">{{.Name}}</td>
									<td data-l="Fabricante">{{.Manufacturer}}</td>
									<td data-l="Vendor"><code>{{.VendorID}}</code></td>
									<td data-l="Device"><code>{{.DeviceID}}</code></td>
									<td data-l="Classe">{{.Class}}</td>
									<td data-l="Status">{{.Status}}</td>
								</tr>
							{{end}}
							</tbody>
						</table>
					</div>
					<div class="card-body"><div class="empty">A identificação por ID (ex.: VEN_8086 = Intel) permite consultar o banco do PCI para nome exato do chipset.</div></div>
				</div>
			</div>
			{{end}}

			<!-- BIOS -->
			{{if .Rep.BIOS}}
			<div class="tab-pane fade" id="pane-bios" role="tabpanel">
				{{range .Rep.BIOS}}
				<div class="card">
					<div class="card-header">BIOS / UEFI</div>
					<div class="card-body">
						<dl class="kv">
							<dt>Fabricante</dt><dd>{{.Manufacturer}}</dd>
							<dt>Nome</dt><dd>{{.Name}}</dd>
							<dt>Modo de boot</dt><dd>{{.BiosMode}}</dd>
							<dt>Versão SMBIOS</dt><dd>{{.SMBIOSVersion}}</dd>
							<dt>Especificação SMBIOS</dt><dd>{{.SmbiosSpec}}</dd>
							<dt>Versão</dt><dd>{{.Version}}</dd>
							<dt>Data de lançamento</dt><dd>{{.ReleaseDate}}</dd>
							<dt>Idioma atual</dt><dd>{{.CurrentLanguage}}</dd>
							<dt>Serial</dt><dd>{{.Serial}}</dd>
							<dt>Status</dt><dd>{{.Status}}</dd>
							{{if .BiosVersion}}<dt>Versão completa</dt><dd>{{.BiosVersion}}</dd>{{end}}
						</dl>
					</div>
				</div>
				{{end}}
			</div>
			{{end}}

			<!-- SOFTWARE -->
			{{if .Rep.Software}}
			<div class="tab-pane fade" id="pane-software" role="tabpanel">
				<div class="card">
					<div class="card-header">Software Instalado <span class="count">{{len .Rep.Software}} programa(s)</span></div>
					<div class="searchbox">
						<input type="search" id="swSearch" placeholder="Filtrar por nome, fabricante ou versão..." oninput="filterTable('swSearch','swTable')">
					</div>
					<div class="table-responsive">
						<table class="table" id="swTable">
							<thead><tr><th>Nome</th><th>Versão</th><th>Fabricante</th><th>Instalado em</th><th>Tamanho</th><th>Arquitetura</th></tr></thead>
							<tbody>
							{{range .Rep.Software}}
								<tr>
									<td data-l="Nome">{{.Name}}</td>
									<td data-l="Versão">{{.Version}}</td>
									<td data-l="Fabricante">{{.Publisher}}</td>
									<td data-l="Instalado em">{{.InstallDate}}</td>
									<td data-l="Tamanho">{{.Size}}</td>
									<td data-l="Arquitetura">{{.Arch}}</td>
								</tr>
							{{end}}
							</tbody>
						</table>
					</div>
				</div>
			</div>
			{{end}}

			<!-- PROCESSOS -->
			{{if .Rep.Processes}}
			<div class="tab-pane fade" id="pane-processos" role="tabpanel">
				<div class="card">
					<div class="card-header">Processos <span class="count">top {{len .Rep.Processes}} por memória</span></div>
					<div class="table-responsive">
						<table class="table">
							<thead><tr><th>#</th><th>Nome</th><th>PID</th><th>Memória</th><th>CPU %</th><th>Usuário</th></tr></thead>
							<tbody>
							{{range .Rep.Processes}}
								<tr>
									<td data-l="#">{{.Num}}</td>
									<td data-l="Nome">{{.Name}}</td>
									<td data-l="PID">{{.PID}}</td>
									<td data-l="Memória">{{.Mem}}</td>
									<td data-l="CPU">{{.CPU}}</td>
									<td data-l="Usuário">{{.User}}</td>
								</tr>
							{{end}}
							</tbody>
						</table>
					</div>
				</div>
			</div>
			{{end}}

		</div>

		<footer>360ti HWiNFO — relatório gerado localmente na máquina, sem envio de dados.</footer>
	</main>
</div>

<script>
{{.JS}}
</script>
<script>
function filterTable(inputId, tableId) {
	var q = document.getElementById(inputId).value.toLowerCase();
	document.querySelectorAll('#' + tableId + ' tbody tr').forEach(function (r) {
		r.style.display = r.textContent.toLowerCase().indexOf(q) >= 0 ? '' : 'none';
	});
}
function copyReportJSON() {
	var el = document.getElementById('report-json');
	if (!el) { return; }
	var txt = el.textContent.trim();
	var ok = function () { setCopyMsg('JSON copiado!'); };
	var fail = function () { setCopyMsg('Falha ao copiar'); };
	if (navigator.clipboard && window.isSecureContext) {
		navigator.clipboard.writeText(txt).then(ok, function () { legacyCopy(txt); });
	} else {
		legacyCopy(txt);
	}
}
function legacyCopy(txt) {
	var ta = document.createElement('textarea');
	ta.value = txt;
	ta.style.position = 'fixed';
	ta.style.opacity = '0';
	document.body.appendChild(ta);
	ta.focus();
	ta.select();
	var done = false;
	try { done = document.execCommand('copy'); } catch (e) {}
	document.body.removeChild(ta);
	if (done) { setCopyMsg('JSON copiado!'); } else { setCopyMsg('Selecione a aba JSON para copiar manualmente'); }
}
function setCopyMsg(msg) {
	var el = document.getElementById('copyMsg');
	if (el) { el.textContent = msg; setTimeout(function () { el.textContent = ''; }, 2200); }
}
function openReportPDF() {
	var url = window.location.href;
	var i = url.lastIndexOf('.');
	if (i < 0) { setCopyMsg('Arquivo PDF não localizado'); return; }
	window.open(url.substring(0, i) + '.pdf', '_blank');
}
</script>
<script id="report-json" type="application/json">{{.RepJSON}}</script>
</body>
</html>
`

// RenderReport produces the final HTML report as a string.
func RenderReport(rep *Report, cfg *resolvedConfig) (string, error) {
	var js string
	if b, err := json.MarshalIndent(rep, "", "  "); err == nil {
		js = string(b)
	}
	// avoid closing the embedding script tag from within the JSON
	js = strings.ReplaceAll(js, "</", "<\\/")

	if cfg == nil {
		cfg = &resolvedConfig{
			AppName:     "360ti HWiNFO",
			CompanyName: "360ti",
			LogoDataURI: "data:image/png;base64,iVBORw0KGgo=",
		}
	}

	ctx := reportCtx{
		Rep:         rep,
		CSS:         template.CSS(bootstrapCSS),
		JS:          template.JS(bootstrapJS),
		RepJSON:     template.JS(js),
		AppName:     cfg.AppName,
		CompanyName: cfg.CompanyName,
		LogoDataURI: template.URL(cfg.LogoDataURI),
		ThemeCSS:    template.CSS(cfg.ThemeCSS),
	}
	t := template.Must(template.New("report").Parse(reportTemplate))
	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
}