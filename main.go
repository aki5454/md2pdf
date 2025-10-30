package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/jung-kurt/gofpdf"
)

const (
	version = "1.0.0"
)

type Config struct {
	InputFile  string
	OutputFile string
	PageSize   string
	FontSize   float64
	ShowHelp   bool
	ShowVer    bool
}

func main() {
	cfg := parseFlags()

	if cfg.ShowHelp {
		flag.Usage()
		os.Exit(0)
	}

	if cfg.ShowVer {
		fmt.Printf("md2pdf version %s\n", version)
		os.Exit(0)
	}

	if cfg.InputFile == "" {
		log.Fatal("Error: input file is required. Use -i flag to specify input file.")
	}

	if cfg.OutputFile == "" {
		cfg.OutputFile = strings.TrimSuffix(cfg.InputFile, filepath.Ext(cfg.InputFile)) + ".pdf"
	}

	if err := convertMarkdownToPDF(cfg); err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	fmt.Printf("Successfully converted %s to %s\n", cfg.InputFile, cfg.OutputFile)
}

func parseFlags() Config {
	cfg := Config{}

	flag.StringVar(&cfg.InputFile, "i", "", "Input Markdown file (required)")
	flag.StringVar(&cfg.InputFile, "input", "", "Input Markdown file (required)")
	flag.StringVar(&cfg.OutputFile, "o", "", "Output PDF file (default: input filename with .pdf extension)")
	flag.StringVar(&cfg.OutputFile, "output", "", "Output PDF file (default: input filename with .pdf extension)")
	flag.StringVar(&cfg.PageSize, "page", "A4", "Page size (A4, Letter, Legal)")
	flag.Float64Var(&cfg.FontSize, "font-size", 12, "Base font size")
	flag.BoolVar(&cfg.ShowHelp, "h", false, "Show help message")
	flag.BoolVar(&cfg.ShowHelp, "help", false, "Show help message")
	flag.BoolVar(&cfg.ShowVer, "v", false, "Show version")
	flag.BoolVar(&cfg.ShowVer, "version", false, "Show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "md2pdf - Convert Markdown files to PDF\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  md2pdf -i input.md [-o output.pdf] [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  md2pdf -i README.md -o output.pdf\n")
		fmt.Fprintf(os.Stderr, "  md2pdf -i document.md --font-size 14 --page Letter\n")
	}

	flag.Parse()
	return cfg
}

func convertMarkdownToPDF(cfg Config) error {
	// Read markdown file
	mdContent, err := os.ReadFile(cfg.InputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Convert markdown to HTML
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(mdContent)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	htmlContent := markdown.Render(doc, renderer)

	// Create PDF
	pdf := gofpdf.New("P", "mm", cfg.PageSize, "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", cfg.FontSize)

	// Parse and write content
	text := stripHTML(string(htmlContent))
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			pdf.Ln(5)
			continue
		}

		// Detect headings (simple heuristic)
		if strings.HasPrefix(line, "#") {
			level := strings.Count(strings.Split(line, " ")[0], "#")
			headingText := strings.TrimSpace(strings.TrimLeft(line, "#"))
			fontSize := cfg.FontSize + float64(4-level)*2
			if fontSize < cfg.FontSize {
				fontSize = cfg.FontSize
			}
			pdf.SetFont("Arial", "B", fontSize)
			pdf.MultiCell(0, 10, headingText, "", "", false)
			pdf.SetFont("Arial", "", cfg.FontSize)
			pdf.Ln(3)
		} else if strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "- ") {
			// List items
			itemText := strings.TrimPrefix(strings.TrimPrefix(line, "* "), "- ")
			pdf.MultiCell(0, 6, "  • "+itemText, "", "", false)
		} else {
			pdf.MultiCell(0, 6, line, "", "", false)
		}
	}

	// Save PDF
	if err := pdf.OutputFileAndClose(cfg.OutputFile); err != nil {
		return fmt.Errorf("failed to create PDF: %w", err)
	}

	return nil
}

// stripHTML removes HTML tags from string (basic implementation)
func stripHTML(s string) string {
	// Simple HTML tag removal
	inTag := false
	result := strings.Builder{}

	for _, char := range s {
		if char == '<' {
			inTag = true
			continue
		}
		if char == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(char)
		}
	}

	// Replace HTML entities
	text := result.String()
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")

	return text
}
