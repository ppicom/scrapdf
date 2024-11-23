# scrapdf

A command-line tool that scrapes web pages and converts them to PDF format. It recursively follows links within the same domain and packages all PDFs into a ZIP file.

## Features
- Scrapes web pages and converts them to PDF
- Follows links recursively within the same domain (up to 5 levels deep)
- Strips HTML formatting (optional)
- Cleans up short lines (optional)
- Packages all PDFs into a single ZIP file
- Cross-platform support (Windows, macOS, Linux)

## Installation

### Prerequisites
- Go 1.21 or higher

### Building from source
1. Clone the repository:
   ```bash
   git clone https://github.com/ppicom/scrapedf.git
   cd scrapedf
   ```

2. Build the application:
   ```bash
   make build
   ```
   The binary will be created in the `bin` directory.

## Usage

### Basic usage:
```bash
scrapedf https://example.com
```

### Options
- `-o, --output <dir>`: Output directory for the ZIP file (default: current directory)
- `--strip`: Strip HTML tags from content before creating PDF
- `--clean`: Remove lines with two words or less (requires `--strip`)
- `-f, --force`: Force overwrite if output file exists

### Examples
```bash
# Basic scraping
scrapedf https://example.com

# Save to specific directory
scrapedf -o ~/Downloads https://example.com

# Strip HTML and clean short lines
scrapedf --strip --clean https://example.com

# Force overwrite existing files
scrapedf -f https://example.com
```


## Output
The tool creates a ZIP file named after the domain (e.g., `example.com.zip`) containing PDF files for each scraped page. The PDFs are named based on the URL path.

Example structure:
```
example.com.zip
├── example.com_index.pdf
├── example.com_about.pdf
├── example.com_contact.pdf
```

## Development
```bash
# Running Tests
make test

# Linting
make lint

# Clean Build Files
make clean
```

## Limitations
- Maximum crawl depth of 5 levels
- Only follows links within the same domain
- 5-second timeout for each page request

## Contributing
1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
