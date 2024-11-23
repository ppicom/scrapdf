package cmd

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ppicom/scrapedf/internal/scraper"
	"github.com/spf13/cobra"
)

var (
	outputDir string
	stripHTML bool
	force     bool
	clean     bool
)

// openDirectory opens the specified directory in the default file manager
func openDirectory(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin": // macOS
		cmd = exec.Command("open", path)
	default: // Linux and other Unix-like systems
		cmd = exec.Command("xdg-open", path)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open directory: %w", err)
	}
	return nil
}

var scrapeCmd = &cobra.Command{
	Use:   "scrape [url]",
	Short: "Scrape a website and convert pages to PDF",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		inputURL := args[0]

		parsedURL, err := url.Parse(inputURL)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}

		outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.zip", parsedURL.Host))
		absOutputPath, err := filepath.Abs(outputPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Check if file exists and prompt for confirmation
		if _, err := os.Stat(outputPath); err == nil && !force {
			fmt.Printf("Warning: The file %s already exists.\n", outputPath)
			fmt.Print("Do you want to replace it? [y/N]: ")

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read user input: %w", err)
			}

			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Operation cancelled")
				return nil
			}
		}

		s := scraper.NewScraper(stripHTML, clean)
		fmt.Printf("Starting to scrape %s\n", inputURL)
		if err := s.ScrapeAndSave(inputURL, outputPath); err != nil {
			return fmt.Errorf("failed to scrape website: %w", err)
		}

		dir, file := filepath.Split(absOutputPath)
		fmt.Printf("Successfully created ZIP file:\n")
		fmt.Printf("  Directory: %s\n", dir)
		fmt.Printf("  File:      %s\n", file)

		// Try to open the directory
		if err := openDirectory(dir); err != nil {
			fmt.Printf("Note: Could not open the output directory automatically: %v\n", err)
		}

		return nil
	},
}

func init() {
	scrapeCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for the ZIP file")
	scrapeCmd.Flags().BoolVar(&stripHTML, "strip", false, "Strip HTML tags from content before creating PDF")
	scrapeCmd.Flags().BoolVarP(&force, "force", "f", false, "Force overwrite if output file exists")
	scrapeCmd.Flags().BoolVar(&clean, "clean", false, "Remove lines with two words or less (requires --strip)")

	// Make clean flag require strip flag
	scrapeCmd.MarkFlagsRequiredTogether("clean", "strip")
}
