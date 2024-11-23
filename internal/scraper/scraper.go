package scraper

import (
	"archive/zip"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gocolly/colly/v2"
	"github.com/jung-kurt/gofpdf"
	"golang.org/x/net/html"
)

type Scraper struct {
	visited   sync.Map
	pdfs      map[string]string // map[url]pdfPath
	stripHTML bool
	clean     bool
}

func NewScraper(stripHTML bool, clean bool) *Scraper {
	if clean && !stripHTML {
		// This shouldn't happen due to cobra flag requirements, but let's be safe
		clean = false
	}
	return &Scraper{
		visited:   sync.Map{},
		pdfs:      make(map[string]string),
		stripHTML: stripHTML,
		clean:     clean,
	}
}

func (s *Scraper) ScrapeAndSave(startURL string, outputPath string) error {
	// Parse the starting URL to get the domain
	parsedURL, err := url.Parse(startURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create a temporary directory for PDFs
	tmpDir, err := os.MkdirTemp("", "webscraper")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	// Ensure cleanup of temporary directory
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			fmt.Printf("Warning: failed to clean up temporary directory: %v\n", err)
		}
	}()

	// Initialize the collector
	c := colly.NewCollector(
		colly.AllowedDomains(parsedURL.Host),
		colly.MaxDepth(5),
		colly.IgnoreRobotsTxt(),
	)

	// Set timeouts
	c.SetRequestTimeout(5 * time.Second)

	// Handle each page
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if err := e.Request.Visit(link); err != nil {
			// We can safely ignore the error here as it's usually due to:
			// - Already visited URLs (handled by colly)
			// - URLs outside allowed domain (handled by colly)
			// - Malformed URLs (handled by colly)
			return
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Failed to fetch %s: %v\n", r.Request.URL, err)
	})

	c.OnResponse(func(r *colly.Response) {
		// Skip if already processed
		if _, exists := s.visited.LoadOrStore(r.Request.URL.String(), true); exists {
			return
		}

		// Create sanitized filename from URL
		urlPath := r.Request.URL.Path
		if urlPath == "" || urlPath == "/" {
			urlPath = "index"
		}
		urlPath = strings.Trim(urlPath, "/")
		urlPath = strings.ReplaceAll(urlPath, "/", "_")

		filename := path.Join(tmpDir, fmt.Sprintf("%s_%s.pdf", r.Request.URL.Host, urlPath))

		// Create PDF directly from response body
		if err := s.createPDF(filename, string(r.Body)); err != nil {
			fmt.Printf("Failed to create PDF for %s: %v\n", r.Request.URL, err)
			// Clean up the failed PDF file if it exists
			if err := os.Remove(filename); err != nil {
				fmt.Printf("Warning: failed to clean up failed PDF file: %v\n", err)
			}
			return
		}

		s.pdfs[r.Request.URL.String()] = filename
		fmt.Printf("Created PDF for %s\n", r.Request.URL)
	})

	// Start scraping
	if err := c.Visit(startURL); err != nil {
		return fmt.Errorf("failed to start scraping: %w", err)
	}

	// Create ZIP file only if we have PDFs to store
	if len(s.pdfs) > 0 {
		if err := s.createZip(outputPath); err != nil {
			return fmt.Errorf("failed to create ZIP file: %w", err)
		}
	} else {
		return fmt.Errorf("no pages were successfully scraped")
	}

	return nil
}

func (s *Scraper) createPDF(filename, htmlContent string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	content := htmlContent
	if s.stripHTML {
		var err error
		content, err = stripHTMLTags(htmlContent)
		if err != nil {
			return fmt.Errorf("failed to strip HTML tags: %w", err)
		}

		if s.clean {
			// Clean up lines with two or fewer words
			var cleanedLines []string
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" {
					cleanedLines = append(cleanedLines, line) // Keep empty lines
					continue
				}
				words := strings.Fields(trimmed)
				if len(words) > 2 {
					cleanedLines = append(cleanedLines, line)
				}
			}
			content = strings.Join(cleanedLines, "\n")
		}
	}

	// Split content into lines and write to PDF
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			pdf.MultiCell(190, 10, line, "0", "L", false)
		}
	}

	return pdf.OutputFileAndClose(filename)
}

// stripHTMLTags removes HTML tags and extracts text content
func stripHTMLTags(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	var textBuilder strings.Builder
	var extractText func(*html.Node)
	var lastNodeWasBlock bool
	var lastNodeWasText bool

	// List of styling tags that should not add newlines
	stylingTags := map[string]bool{
		"strong": true,
		"b":      true,
		"em":     true,
		"i":      true,
		"u":      true,
		"span":   true,
		"mark":   true,
		"small":  true,
		"sub":    true,
		"sup":    true,
		"code":   true,
	}

	extractText = func(n *html.Node) {
		if n.Type == html.CommentNode ||
			(n.Type == html.ElementNode && (n.Data == "script" ||
				n.Data == "style" ||
				n.Data == "meta" ||
				n.Data == "link" ||
				n.Data == "noscript")) {
			return
		}

		if n.Type == html.ElementNode && stylingTags[n.Data] {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				extractText(c)
			}
			return
		}

		if n.Type == html.TextNode {
			var text string
			if n.Parent != nil && stylingTags[n.Parent.Data] {
				text = strings.TrimRightFunc(n.Data, unicode.IsSpace)
			} else {
				text = strings.TrimSpace(n.Data)
			}

			if text != "" {
				if n.Parent != nil {
					switch n.Parent.Data {
					case "p":
						textBuilder.WriteString(text)
						textBuilder.WriteString("\n\n")
						lastNodeWasBlock = true
						lastNodeWasText = false
					case "li":
						textBuilder.WriteString("• ")
						textBuilder.WriteString(text)
						textBuilder.WriteString("\n")
						lastNodeWasBlock = false
						lastNodeWasText = false
					case "h1", "h2", "h3", "h4", "h5", "h6":
						if !lastNodeWasBlock {
							textBuilder.WriteString("\n")
						}
						textBuilder.WriteString(text)
						textBuilder.WriteString("\n\n")
						lastNodeWasBlock = true
						lastNodeWasText = false
					case "a":
						if lastNodeWasText {
							textBuilder.WriteString("\n")
						}
						textBuilder.WriteString(text)
						if n.Parent.Parent != nil && n.Parent.Parent.Data == "nav" && n.Parent.NextSibling == nil {
							textBuilder.WriteString("\n")
						}
						lastNodeWasBlock = false
						lastNodeWasText = true
					default:
						if lastNodeWasText {
							textBuilder.WriteString(" ")
						}
						textBuilder.WriteString(text)
						lastNodeWasBlock = false
						lastNodeWasText = true
					}
				} else {
					if lastNodeWasText {
						textBuilder.WriteString(" ")
					}
					textBuilder.WriteString(text)
					lastNodeWasText = true
				}
			}
		}

		// Process child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}
	}

	extractText(doc)

	// Clean up multiple consecutive newlines and ensure proper trailing newlines
	content := textBuilder.String()
	content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
	content = strings.TrimSpace(content)

	// Ensure content ends with exactly two newlines for block elements
	if !strings.HasSuffix(content, "\n\n") && content != "" {
		if strings.HasSuffix(content, "\n") {
			content += "\n"
		} else {
			content += "\n\n"
		}
	}

	return content, nil
}

func (s *Scraper) createZip(zipname string) error {
	zipfile, err := os.Create(zipname)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	for urlStr, pdfPath := range s.pdfs {
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			return fmt.Errorf("failed to parse URL %s: %w", urlStr, err)
		}

		// Create clean filename for the ZIP entry
		urlPath := parsedURL.Path
		if urlPath == "" || urlPath == "/" {
			urlPath = "index"
		}
		urlPath = strings.Trim(urlPath, "/")
		urlPath = strings.ReplaceAll(urlPath, "/", "_")

		zipEntryName := fmt.Sprintf("%s_%s.pdf", parsedURL.Host, urlPath)

		file, err := os.Open(pdfPath)
		if err != nil {
			return fmt.Errorf("failed to open PDF file: %w", err)
		}

		writer, err := archive.Create(zipEntryName)
		if err != nil {
			file.Close()
			return fmt.Errorf("failed to create zip entry: %w", err)
		}

		if _, err := io.Copy(writer, file); err != nil {
			file.Close()
			return fmt.Errorf("failed to write to zip: %w", err)
		}

		file.Close()
	}

	return nil
}
