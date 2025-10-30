package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
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

	// Create HTML with Japanese support
	htmlTemplate := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: "Hiragino Sans", "Hiragino Kaku Gothic ProN", "Noto Sans JP", sans-serif;
            font-size: %fpt;
            line-height: 1.6;
            max-width: 800px;
            margin: 40px auto;
            padding: 0 20px;
        }
        h1 { font-size: %fpt; margin-top: 20px; }
        h2 { font-size: %fpt; margin-top: 18px; }
        h3 { font-size: %fpt; margin-top: 16px; }
        h4 { font-size: %fpt; margin-top: 14px; font-weight: bold; }
        ul, ol {
            margin-left: 20px;
            padding-left: 20px;
        }
        li {
            margin-bottom: 4px;
            white-space: nowrap;
            overflow: visible;
        }
        p {
            margin: 8px 0;
            white-space: pre-wrap;
            word-wrap: break-word;
        }
        code {
            background-color: #f4f4f4;
            padding: 2px 4px;
            white-space: nowrap;
        }
        pre {
            background-color: #f4f4f4;
            padding: 10px;
            overflow-x: auto;
            white-space: pre-wrap;
        }
    </style>
</head>
<body>
%s
</body>
</html>`

	h1Size := cfg.FontSize + 8
	h2Size := cfg.FontSize + 6
	h3Size := cfg.FontSize + 4
	h4Size := cfg.FontSize + 2

	htmlContent = []byte(fmt.Sprintf(htmlTemplate, cfg.FontSize, h1Size, h2Size, h3Size, h4Size, string(htmlContent)))

	// Create temporary HTML file
	tmpHTML := strings.TrimSuffix(cfg.OutputFile, filepath.Ext(cfg.OutputFile)) + "_tmp.html"
	if err := os.WriteFile(tmpHTML, htmlContent, 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}
	defer os.Remove(tmpHTML)

	// Try to use wkhtmltopdf if available
	if _, err := exec.LookPath("wkhtmltopdf"); err == nil {
		cmd := exec.Command("wkhtmltopdf",
			"--page-size", cfg.PageSize,
			"--encoding", "UTF-8",
			tmpHTML, cfg.OutputFile)

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("wkhtmltopdf failed: %w", err)
		}
		return nil
	}

	// Fallback: Try Chrome/Chromium headless
	chromePaths := []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"chromium",
		"google-chrome",
	}

	var chromeCmd *exec.Cmd
	for _, chromePath := range chromePaths {
		if _, err := exec.LookPath(chromePath); err == nil {
			chromeCmd = exec.Command(chromePath,
				"--headless",
				"--disable-gpu",
				"--print-to-pdf="+cfg.OutputFile,
				tmpHTML)
			break
		}
	}

	if chromeCmd == nil {
		return fmt.Errorf("no PDF renderer found. Please install wkhtmltopdf or Chrome:\n  brew install wkhtmltopdf")
	}

	if err := chromeCmd.Run(); err != nil {
		return fmt.Errorf("chrome headless failed: %w", err)
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
