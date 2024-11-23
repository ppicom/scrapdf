package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "webscraper",
	Short: "A web scraping tool that converts pages to PDF",
	Long: `webscraper is a CLI tool that scrapes web pages and converts them to PDF.
It recursively follows links within the same domain and creates a ZIP file
containing all scraped pages as PDFs.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(scrapeCmd)
}
