package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"golang.org/x/net/html"
)

// renderMarkdownToPDF converts a Markdown audit report to a formatted PDF.
// Uses goldmark for MD→HTML, x/net/html for DOM parsing, and fpdf for rendering.
func renderMarkdownToPDF(reportMD, repo string) ([]byte, error) {
	// 1. Markdown → HTML
	gm := goldmark.New(goldmark.WithExtensions(extension.GFM))
	var htmlBuf bytes.Buffer
	if err := gm.Convert([]byte(reportMD), &htmlBuf); err != nil {
		return nil, fmt.Errorf("md→html: %w", err)
	}

	// 2. Parse HTML DOM
	doc, err := html.Parse(&htmlBuf)
	if err != nil {
		return nil, fmt.Errorf("html parse: %w", err)
	}

	// 3. Render to PDF
	r := newPDFRenderer(repo)
	r.walkChildren(doc)

	var out bytes.Buffer
	if err := r.pdf.Output(&out); err != nil {
		return nil, fmt.Errorf("pdf output: %w", err)
	}
	return out.Bytes(), nil
}

// ── PDF renderer ─────────────────────────────────────────────────────────────

const (
	pdfMargin    = 20.0
	pdfPageW     = 210.0 // A4
	pdfPageH     = 297.0
	pdfBodyW     = pdfPageW - 2*pdfMargin
	pdfFontBody  = 10.0
	pdfLineH     = 5.5
	pdfCodeLineH = 4.8
)

type pdfRenderer struct {
	pdf         *fpdf.Fpdf
	repo        string
	bold        bool
	italic      bool
	inPre       bool
	listDepth   int
	listOrdered []bool // stack: true = ordered
	listCounts  []int  // current counter per level
	firstPage   bool
}

func newPDFRenderer(repo string) *pdfRenderer {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfMargin, pdfMargin, pdfMargin)
	pdf.SetAutoPageBreak(true, pdfMargin)
	pdf.AddPage()

	// Header / footer
	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(150, 150, 150)
		pdf.CellFormat(0, 8, "Security Audit Report — "+repo, "", 1, "R", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	})
	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(150, 150, 150)
		pdf.CellFormat(0, 5, fmt.Sprintf("Page %d", pdf.PageNo()), "", 0, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	})

	pdf.SetFont("Helvetica", "", pdfFontBody)
	return &pdfRenderer{pdf: pdf, repo: repo, firstPage: true}
}

// ── DOM walker ────────────────────────────────────────────────────────────────

func (r *pdfRenderer) walkChildren(n *html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		r.walk(c)
	}
}

func (r *pdfRenderer) walk(n *html.Node) {
	if n.Type == html.TextNode {
		r.emitText(n.Data)
		return
	}
	if n.Type != html.ElementNode {
		r.walkChildren(n)
		return
	}

	switch n.Data {
	case "h1":
		r.renderHeading(n, 18, false)
	case "h2":
		r.renderHeading(n, 14, true)
	case "h3":
		r.renderHeading(n, 12, false)
	case "h4", "h5", "h6":
		r.renderHeading(n, 11, false)
	case "p":
		r.renderParagraph(n)
	case "strong", "b":
		r.bold = true
		r.walkChildren(n)
		r.bold = false
	case "em", "i":
		r.italic = true
		r.walkChildren(n)
		r.italic = false
	case "code":
		if r.inPre {
			r.walkChildren(n)
		} else {
			r.renderInlineCode(n)
		}
	case "pre":
		r.renderCodeBlock(n)
	case "ul":
		r.listDepth++
		r.listOrdered = append(r.listOrdered, false)
		r.listCounts = append(r.listCounts, 0)
		r.walkChildren(n)
		r.listDepth--
		r.listOrdered = r.listOrdered[:len(r.listOrdered)-1]
		r.listCounts = r.listCounts[:len(r.listCounts)-1]
		if r.listDepth == 0 {
			r.pdf.Ln(2)
		}
	case "ol":
		r.listDepth++
		r.listOrdered = append(r.listOrdered, true)
		r.listCounts = append(r.listCounts, 0)
		r.walkChildren(n)
		r.listDepth--
		r.listOrdered = r.listOrdered[:len(r.listOrdered)-1]
		r.listCounts = r.listCounts[:len(r.listCounts)-1]
		if r.listDepth == 0 {
			r.pdf.Ln(2)
		}
	case "li":
		r.renderListItem(n)
	case "table":
		r.renderTable(n)
	case "hr":
		r.pdf.SetDrawColor(200, 200, 200)
		r.pdf.Line(pdfMargin, r.pdf.GetY(), pdfPageW-pdfMargin, r.pdf.GetY())
		r.pdf.SetDrawColor(0, 0, 0)
		r.pdf.Ln(3)
	case "blockquote":
		r.renderBlockquote(n)
	case "a":
		r.renderLink(n)
	case "br":
		r.pdf.Ln(pdfLineH)
	default:
		r.walkChildren(n)
	}
}

// ── Heading ───────────────────────────────────────────────────────────────────

func (r *pdfRenderer) renderHeading(n *html.Node, size float64, newPage bool) {
	if newPage && !r.firstPage {
		// Start H2 on new page if less than 60mm remaining
		if r.pdf.GetY() > pdfPageH-pdfMargin-60 {
			r.pdf.AddPage()
		}
	}
	r.firstPage = false
	r.pdf.Ln(3)
	r.pdf.SetFont("Helvetica", "B", size)
	r.pdf.SetTextColor(26, 35, 60)
	r.pdf.MultiCell(pdfBodyW, size*0.5, r.textOf(n), "", "L", false)
	r.pdf.SetTextColor(0, 0, 0)
	r.pdf.SetFont("Helvetica", "", pdfFontBody)
	r.pdf.Ln(2)
}

// ── Paragraph ─────────────────────────────────────────────────────────────────

func (r *pdfRenderer) renderParagraph(n *html.Node) {
	// Flush current paragraph text as a MultiCell
	var sb strings.Builder
	r.collectText(n, &sb)
	text := strings.TrimSpace(sb.String())
	if text == "" {
		return
	}
	r.pdf.SetFont("Helvetica", "", pdfFontBody)
	r.pdf.MultiCell(pdfBodyW, pdfLineH, text, "", "L", false)
	r.pdf.Ln(1)
}

// ── Inline text / code ────────────────────────────────────────────────────────

func (r *pdfRenderer) emitText(text string) {
	text = strings.ReplaceAll(text, "\n", " ")
	if text == "" {
		return
	}
	style := ""
	if r.bold {
		style += "B"
	}
	if r.italic {
		style += "I"
	}
	r.pdf.SetFont("Helvetica", style, pdfFontBody)
	r.pdf.Write(pdfLineH, text)
}

func (r *pdfRenderer) renderInlineCode(n *html.Node) {
	text := r.textOf(n)
	r.pdf.SetFont("Courier", "", 8.5)
	r.pdf.SetFillColor(245, 245, 245)
	r.pdf.Write(pdfLineH, text)
	r.pdf.SetFont("Helvetica", "", pdfFontBody)
}

// ── Code block ────────────────────────────────────────────────────────────────

func (r *pdfRenderer) renderCodeBlock(n *html.Node) {
	r.inPre = true
	var sb strings.Builder
	r.collectText(n, &sb)
	r.inPre = false
	code := strings.TrimRight(sb.String(), "\n")
	lines := strings.Split(code, "\n")

	r.pdf.Ln(2)
	r.pdf.SetFont("Courier", "", 8)
	r.pdf.SetFillColor(245, 245, 245)
	r.pdf.SetDrawColor(220, 220, 220)

	blockH := float64(len(lines)) * pdfCodeLineH
	x, y := r.pdf.GetX(), r.pdf.GetY()
	if y+blockH > pdfPageH-pdfMargin-5 {
		r.pdf.AddPage()
		x, y = r.pdf.GetX(), r.pdf.GetY()
	}
	r.pdf.Rect(x-1, y, pdfBodyW+2, blockH+2, "F")
	for _, line := range lines {
		r.pdf.CellFormat(pdfBodyW, pdfCodeLineH, line, "", 1, "L", false, 0, "")
	}
	r.pdf.SetDrawColor(0, 0, 0)
	r.pdf.SetFont("Helvetica", "", pdfFontBody)
	r.pdf.Ln(3)
}

// ── List ──────────────────────────────────────────────────────────────────────

func (r *pdfRenderer) renderListItem(n *html.Node) {
	indent := float64(r.listDepth) * 5.0
	r.pdf.SetX(pdfMargin + indent)

	bullet := "• "
	if len(r.listOrdered) > 0 && r.listOrdered[len(r.listOrdered)-1] {
		r.listCounts[len(r.listCounts)-1]++
		bullet = fmt.Sprintf("%d. ", r.listCounts[len(r.listCounts)-1])
	}

	var sb strings.Builder
	r.collectText(n, &sb)
	text := strings.TrimSpace(sb.String())

	r.pdf.SetFont("Helvetica", "", pdfFontBody)
	r.pdf.MultiCell(pdfBodyW-indent, pdfLineH, bullet+text, "", "L", false)
	r.pdf.SetX(pdfMargin)
}

// ── Table ─────────────────────────────────────────────────────────────────────

type tableData struct {
	headers []string
	rows    [][]string
}

func (r *pdfRenderer) renderTable(n *html.Node) {
	td := collectTableData(n)
	if len(td.headers) == 0 && len(td.rows) == 0 {
		return
	}

	nCols := len(td.headers)
	if nCols == 0 && len(td.rows) > 0 {
		nCols = len(td.rows[0])
	}
	if nCols == 0 {
		return
	}

	// Calculate column widths proportional to max content length
	maxLen := make([]int, nCols)
	for i, h := range td.headers {
		if len(h) > maxLen[i] {
			maxLen[i] = len(h)
		}
	}
	for _, row := range td.rows {
		for i, cell := range row {
			if i < nCols && len(cell) > maxLen[i] {
				maxLen[i] = len(cell)
			}
		}
	}
	total := 0
	for _, l := range maxLen {
		total += l
	}
	if total == 0 {
		total = 1
	}
	colW := make([]float64, nCols)
	for i, l := range maxLen {
		w := pdfBodyW * float64(l) / float64(total)
		if w < 15 {
			w = 15
		}
		if w > 90 {
			w = 90
		}
		colW[i] = w
	}

	r.pdf.Ln(2)
	r.pdf.SetFont("Helvetica", "B", 9)

	// Header row
	if len(td.headers) > 0 {
		r.pdf.SetFillColor(230, 234, 242)
		r.pdf.SetDrawColor(180, 180, 180)
		for i, h := range td.headers {
			if i < nCols {
				r.pdf.CellFormat(colW[i], 6, h, "1", 0, "L", true, 0, "")
			}
		}
		r.pdf.Ln(-1)
	}

	// Data rows
	r.pdf.SetFont("Helvetica", "", 8.5)
	for ri, row := range td.rows {
		fillColor := [3]int{255, 255, 255}
		if ri%2 == 1 {
			fillColor = [3]int{248, 249, 252}
		}
		r.pdf.SetFillColor(fillColor[0], fillColor[1], fillColor[2])

		// Calculate row height (max lines needed across cells)
		rowH := 5.5
		for i, cell := range row {
			if i >= nCols {
				break
			}
			lines := r.pdf.SplitLines([]byte(cell), colW[i]-2)
			h := float64(len(lines)) * 5.0
			if h > rowH {
				rowH = h
			}
		}
		// Page break if needed
		if r.pdf.GetY()+rowH > pdfPageH-pdfMargin {
			r.pdf.AddPage()
			// Re-print header on new page
			if len(td.headers) > 0 {
				r.pdf.SetFont("Helvetica", "B", 9)
				r.pdf.SetFillColor(230, 234, 242)
				for i, h := range td.headers {
					if i < nCols {
						r.pdf.CellFormat(colW[i], 6, h, "1", 0, "L", true, 0, "")
					}
				}
				r.pdf.Ln(-1)
				r.pdf.SetFont("Helvetica", "", 8.5)
			}
		}

		x0 := r.pdf.GetX()
		y0 := r.pdf.GetY()
		for i, cell := range row {
			if i >= nCols {
				break
			}
			r.pdf.SetXY(x0, y0)
			x0 += colW[i]
			r.pdf.MultiCell(colW[i], 5.0, cell, "1", "L", fillColor != [3]int{255, 255, 255})
		}
		r.pdf.SetXY(pdfMargin, y0+rowH)
	}

	r.pdf.SetDrawColor(0, 0, 0)
	r.pdf.SetFont("Helvetica", "", pdfFontBody)
	r.pdf.Ln(3)
}

// ── Blockquote ────────────────────────────────────────────────────────────────

func (r *pdfRenderer) renderBlockquote(n *html.Node) {
	var sb strings.Builder
	r.collectText(n, &sb)
	text := strings.TrimSpace(sb.String())
	if text == "" {
		return
	}
	r.pdf.Ln(1)
	r.pdf.SetFont("Helvetica", "I", pdfFontBody)
	r.pdf.SetFillColor(245, 245, 245)
	r.pdf.SetDrawColor(180, 180, 180)
	r.pdf.MultiCell(pdfBodyW-5, pdfLineH, text, "L", "L", true)
	r.pdf.SetDrawColor(0, 0, 0)
	r.pdf.SetFont("Helvetica", "", pdfFontBody)
	r.pdf.Ln(1)
}

// ── Link ──────────────────────────────────────────────────────────────────────

func (r *pdfRenderer) renderLink(n *html.Node) {
	text := r.textOf(n)
	href := attrVal(n, "href")
	r.pdf.SetTextColor(0, 80, 160)
	if href != "" {
		r.pdf.WriteLinkString(pdfLineH, text, href)
	} else {
		r.pdf.Write(pdfLineH, text)
	}
	r.pdf.SetTextColor(0, 0, 0)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (r *pdfRenderer) textOf(n *html.Node) string {
	var sb strings.Builder
	r.collectText(n, &sb)
	return strings.TrimSpace(sb.String())
}

func (r *pdfRenderer) collectText(n *html.Node, sb *strings.Builder) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			sb.WriteString(c.Data)
		case html.ElementNode:
			r.collectText(c, sb)
			if c.Data == "br" {
				sb.WriteString("\n")
			}
		}
	}
}

func attrVal(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// collectTableData extracts header and data rows from a <table> node.
func collectTableData(n *html.Node) tableData {
	var td tableData
	var collectRows func(*html.Node, bool)
	collectRows = func(node *html.Node, isHeader bool) {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Type != html.ElementNode {
				continue
			}
			switch c.Data {
			case "thead":
				collectRows(c, true)
			case "tbody", "tfoot":
				collectRows(c, false)
			case "tr":
				var row []string
				for td2 := c.FirstChild; td2 != nil; td2 = td2.NextSibling {
					if td2.Type == html.ElementNode && (td2.Data == "th" || td2.Data == "td") {
						row = append(row, cellText(td2))
					}
				}
				if isHeader {
					td.headers = row
				} else {
					td.rows = append(td.rows, row)
				}
			default:
				collectRows(c, isHeader)
			}
		}
	}
	collectRows(n, false)
	return td
}

func cellText(n *html.Node) string {
	var sb strings.Builder
	var collect func(*html.Node)
	collect = func(node *html.Node) {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			switch c.Type {
			case html.TextNode:
				sb.WriteString(c.Data)
			case html.ElementNode:
				collect(c)
			}
		}
	}
	collect(n)
	s := strings.TrimSpace(sb.String())
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
