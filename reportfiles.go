package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// writeReportFiles writes report.html, report.json, report.dat and report.pdf into dir.
func writeReportFiles(dir string, rep *Report, cfg *resolvedConfig) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	body, err := RenderReport(rep, cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "report.html"), []byte(body), 0644); err != nil {
		return err
	}

	js, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "report.json"), js, 0644); err != nil {
		return err
	}

	if err := writeDataset(filepath.Join(dir, "report.dat"), rep); err != nil {
		return err
	}

	return writePDF(filepath.Join(dir, "report.pdf"), rep, cfg)
}

// pdfPathFor derives the PDF path from an HTML path.
func pdfPathFor(htmlPath string) string {
	return strings.TrimSuffix(htmlPath, filepath.Ext(htmlPath)) + ".pdf"
}