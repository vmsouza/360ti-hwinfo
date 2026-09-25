package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	out := flag.String("out", "", "Output path for the HTML report (default: 360ti-hwinfo-<hostname>-<yymmdd-hhmmss>.html)")
	open := flag.Bool("open", true, "Open the report in the default browser after generation")
	report := flag.String("report", "", "Directory where report.html/.json/.dat are written (used by the native launcher; no browser is opened)")
	timing := flag.Bool("timing", false, "Write a timing log to 360ti-hwinfo-timing.log")
	flag.Parse()

	timingLogEnabled = *timing

	rep := &Report{GeneratedAt: time.Now().Format("02/01/2006 15:04:05")}

	fmt.Println("Collecting system information...")
	done := make(chan struct{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Coleta recuperada de erro:", r)
			}
			close(done)
		}()
		CollectInto(rep)
	}()
	select {
	case <-done:
	case <-time.After(60 * time.Second):
		fmt.Println("Aviso: coleta ainda em andamento, gerando relatório parcial...")
	}

	cfg := LoadConfig()

	if *report != "" {
		os.Exit(runReport(*report, rep, cfg))
	}

	path := *out
	if path == "" {
		dir, err := defaultReportsDir()
		if err != nil {
			showFatalError("Não foi possível criar a pasta de relatórios: " + err.Error())
			os.Exit(1)
		}
		path = filepath.Join(dir, "360ti-hwinfo.html")
	}

	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}

	body, err := RenderReport(rep, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error rendering report: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing report: %v\n", err)
		os.Exit(1)
	}
	if err := writePDF(pdfPathFor(path), rep, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error writing PDF: %v\n", err)
	}

	fmt.Printf("Report generated: %s\n", path)

	if *open {
		OpenInBrowser(path)
	}
}

// defaultReportsDir returns a writable folder for reports, preferring the
// per-user data folder (%AppData%\360ti\reports on Windows) and falling back
// to the executable's directory, Documents and finally the temp folder.
func defaultReportsDir() (string, error) {
	if confDir, err := os.UserConfigDir(); err == nil && confDir != "" {
		dir := filepath.Join(confDir, "360ti", "reports")
		if err := os.MkdirAll(dir, 0755); err == nil {
			return dir, nil
		}
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Join(filepath.Dir(exe), "reports")
		if err := os.MkdirAll(dir, 0755); err == nil {
			return dir, nil
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		dir := filepath.Join(home, "Documents", "360ti", "reports")
		if err := os.MkdirAll(dir, 0755); err == nil {
			return dir, nil
		}
	}
	if dir := os.TempDir(); dir != "" {
		dir = filepath.Join(dir, "360ti-hwinfo")
		if err := os.MkdirAll(dir, 0755); err == nil {
			return dir, nil
		}
	}
	return "", fmt.Errorf("nenhuma pasta gravável disponível")
}

// runReport writes report.html/.json/.dat into dir. On any failure it writes
// report-error.txt with the reason and returns a non-zero exit code.
func runReport(dir string, rep *Report, cfg *resolvedConfig) (exitCode int) {
	exitCode = 1
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("pânico: %v", r)
			_ = os.WriteFile(filepath.Join(dir, "report-error.txt"), []byte(msg), 0644)
			exitCode = 1
		}
	}()

	err := writeReportFiles(dir, rep, cfg)
	if err != nil {
		msg := "erro ao gerar relatório: " + err.Error()
		_ = os.WriteFile(filepath.Join(dir, "report-error.txt"), []byte(msg), 0644)
		fmt.Fprintln(os.Stderr, msg)
		return 1
	}
	_ = os.WriteFile(filepath.Join(dir, "report-ok.txt"), []byte("ok\n"), 0644)
	return 0
}