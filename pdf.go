package main

import (
	_ "embed"

	"github.com/go-pdf/fpdf"
)

//go:embed fonts/DejaVuSans.ttf
var dejavuSans []byte

//go:embed fonts/DejaVuSans-Bold.ttf
var dejavuSansBold []byte

const pdfPageW = 182.0 // usable width (A4 - margins)

// writePDF generates a legible A4 PDF report from the collected data.
func writePDF(path string, rep *Report, cfg *resolvedConfig) error {
	accent := [3]int{37, 99, 235}
	appName := "360ti HWiNFO"
	company := "360ti"
	if cfg != nil {
		if cfg.AccentColor != [3]int{} {
			accent = cfg.AccentColor
		}
		if cfg.AppName != "" {
			appName = cfg.AppName
		}
		if cfg.CompanyName != "" {
			company = cfg.CompanyName
		}
	}
	darker := [3]int{accent[0] / 2, accent[1] / 2, accent[2] / 2}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(14, 14, 14)
	pdf.SetAutoPageBreak(true, 16)
	pdf.AddUTF8FontFromBytes("DjVu", "", dejavuSans)
	pdf.AddUTF8FontFromBytes("DjVu", "B", dejavuSansBold)

	const lineH = 5.2
	const cellPad = 1.0

	drawHeader := func() {
		pdf.SetFillColor(accent[0], accent[1], accent[2])
		pdf.SetDrawColor(accent[0], accent[1], accent[2])
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("DjVu", "B", 16)
		pdf.CellFormat(0, 9, appName, "", 0, "L", true, 0, "")
		pdf.Ln(9)
		pdf.SetFont("DjVu", "", 9)
		pdf.CellFormat(0, 5, company, "", 0, "L", true, 0, "")
		pdf.Ln(5)
		pdf.CellFormat(0, 5, "Máquina: "+rep.Host.Hostname, "", 0, "L", true, 0, "")
		pdf.Ln(5)
		pdf.CellFormat(0, 5, "Gerado em: "+rep.GeneratedAt, "", 0, "L", true, 0, "")
		pdf.Ln(5)
		pdf.SetTextColor(20, 30, 50)
	}

	// wrapLines breaks text into lines that fit the given width.
	wrapLines := func(text string, width float64) []string {
		var lines []string
		var cur []rune
		for _, r := range text {
			if r == '\n' {
				lines = append(lines, string(cur))
				cur = cur[:0]
				continue
			}
			if len(cur) > 0 && pdf.GetStringWidth(string(cur)+string(r)) > width {
				lines = append(lines, string(cur))
				cur = []rune{r}
				continue
			}
			cur = append(cur, r)
		}
		if len(cur) > 0 {
			lines = append(lines, string(cur))
		}
		if len(lines) == 0 {
			lines = append(lines, "")
		}
		return lines
	}

	drawSectionTitle := func(title string) {
		if pdf.GetY() > 252 {
			pdf.AddPage()
		}
		pdf.Ln(3)
		pdf.SetFillColor(darker[0], darker[1], darker[2])
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("DjVu", "B", 11)
		pdf.CellFormat(0, 8, title, "", 1, "L", true, 0, "")
		pdf.SetTextColor(20, 30, 50)
	}

	drawTable := func(headers []string, rows [][]string) {
		n := len(headers)
		if n == 0 {
			return
		}
		widths := tableWidths(n)
		colX := make([]float64, n)
		x := 14.0
		for i := 0; i < n; i++ {
			colX[i] = x
			x += widths[i]
		}

		startX := 14.0
		startY := pdf.GetY()

		// header row
		pdf.SetFont("DjVu", "B", 8)
		headerH := 6.5
		pdf.SetFillColor(accent[0], accent[1], accent[2])
		pdf.SetTextColor(255, 255, 255)
		pdf.SetDrawColor(160, 170, 190)
		for c := 0; c < n; c++ {
			pdf.SetXY(colX[c], startY)
			pdf.CellFormat(widths[c], headerH, headers[c], "", 0, "L", true, 0, "")
		}
		pdf.Line(startX, startY+headerH, startX+pdfPageW, startY+headerH)
		pdf.SetTextColor(20, 30, 50)

		rowY := startY + headerH
		for r := 0; r < len(rows); r++ {
			if rowY > 250 {
				pdf.AddPage()
				rowY = pdf.GetY() + 4
				pdf.SetFont("DjVu", "B", 8)
				pdf.SetFillColor(accent[0], accent[1], accent[2])
				pdf.SetTextColor(255, 255, 255)
				for c := 0; c < n; c++ {
					pdf.SetXY(colX[c], rowY)
					pdf.CellFormat(widths[c], headerH, headers[c], "", 0, "L", true, 0, "")
				}
				pdf.Line(startX, rowY+headerH, startX+pdfPageW, rowY+headerH)
				rowY += headerH
				pdf.SetTextColor(20, 30, 50)
				pdf.SetFont("DjVu", "", 8)
			}

			// compute the wrapped lines for each cell
			cellLines := make([][]string, n)
			maxN := 1
			for c := 0; c < n; c++ {
				var cell string
				if c < len(rows[r]) {
					cell = rows[r][c]
				}
				cellLines[c] = wrapLines(cell, widths[c]-2*cellPad)
				if len(cellLines[c]) > maxN {
					maxN = len(cellLines[c])
				}
			}

			rowH := float64(maxN) * lineH
			if rowY+rowH > 270 {
				pdf.AddPage()
				rowY = pdf.GetY() + 4
				pdf.SetFont("DjVu", "", 8)
			}

			// background
			pdf.SetFillColor(248, 249, 252)
			if r%2 == 1 {
				pdf.SetFillColor(240, 243, 248)
			}
			pdf.Rect(startX, rowY, pdfPageW, rowH, "F")

			// text
			pdf.SetTextColor(30, 40, 60)
			for c := 0; c < n; c++ {
				for j := 0; j < len(cellLines[c]); j++ {
					pdf.SetXY(colX[c]+cellPad, rowY+float64(j)*lineH)
					pdf.CellFormat(widths[c]-2*cellPad, lineH, cellLines[c][j], "", 0, "L", false, 0, "")
				}
			}

			// separator
			pdf.SetDrawColor(200, 208, 220)
			pdf.Line(startX, rowY+rowH, startX+pdfPageW, rowY+rowH)
			rowY += rowH
		}

		// vertical separators
		pdf.SetDrawColor(200, 208, 220)
		for c := 1; c < n; c++ {
			pdf.Line(colX[c], startY, colX[c], rowY)
		}
		pdf.SetY(rowY + 2)
	}

	pdf.AddPage()
	drawHeader()

	for _, s := range buildSections(rep) {
		pdf.SetFont("DjVu", "", 8)
		drawSectionTitle(s.Title)
		drawTable(s.Headers, s.Rows)
	}

	return pdf.OutputFileAndClose(path)
}

// tableWidths splits the usable page width into column widths,
// giving the first column more space.
func tableWidths(n int) []float64 {
	w := make([]float64, n)
	if n <= 0 {
		return w
	}
	if n == 1 {
		w[0] = pdfPageW
		return w
	}
	w[0] = pdfPageW * 0.24
	rest := pdfPageW - w[0]
	for i := 1; i < n; i++ {
		w[i] = rest / float64(n-1)
	}
	return w
}