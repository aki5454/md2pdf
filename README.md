# md2pdf

A simple command-line tool to convert Markdown files to PDF format, written in Go.

## Features

- Convert Markdown files to PDF with simple command
- Support for common Markdown syntax (headings, lists, paragraphs)
- Customizable page size (A4, Letter, Legal)
- Adjustable font size
- Automatic output filename generation

## Installation

### From Source

```bash
git clone https://github.com/torcheees/md2pdf.git
cd md2pdf
go build -o md2pdf
```

### Using go install

```bash
go install github.com/torcheees/md2pdf@latest
```

## Usage

### Basic Usage

```bash
# Convert README.md to README.pdf
md2pdf -i README.md

# Specify output file
md2pdf -i input.md -o output.pdf
```

### Advanced Options

```bash
# Custom page size and font size
md2pdf -i document.md --page Letter --font-size 14

# Show help
md2pdf -h

# Show version
md2pdf -v
```

### Command-Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-i, --input` | Input Markdown file (required) | - |
| `-o, --output` | Output PDF file | `<input>.pdf` |
| `--page` | Page size (A4, Letter, Legal) | A4 |
| `--font-size` | Base font size | 12 |
| `-h, --help` | Show help message | - |
| `-v, --version` | Show version | - |

## Examples

### Convert a single file

```bash
md2pdf -i README.md
```

Output: `README.pdf`

### Specify output location

```bash
md2pdf -i docs/guide.md -o output/guide.pdf
```

### Use Letter size with larger font

```bash
md2pdf -i document.md --page Letter --font-size 14
```

## Supported Markdown Features

- Headings (H1-H6)
- Paragraphs
- Lists (bullet and numbered)
- Bold and italic text (converted to plain text in PDF)
- Links (text is preserved, URLs are removed)
- Code blocks (formatted as plain text)

## Dependencies

- [gomarkdown/markdown](https://github.com/gomarkdown/markdown) - Markdown parser
- [jung-kurt/gofpdf](https://github.com/jung-kurt/gofpdf) - PDF generation

## Development

### Build

```bash
go build -o md2pdf
```

### Test

Create a sample Markdown file:

```bash
echo "# Test Document

This is a test paragraph.

## Features

- Feature 1
- Feature 2
- Feature 3

### Conclusion

End of test." > test.md

./md2pdf -i test.md
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Author

torcheees
