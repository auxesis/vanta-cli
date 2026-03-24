package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// Framework represents a single framework returned by the Vanta API.
type Framework struct {
	ID                   string `json:"id"`
	DisplayName          string `json:"displayName"`
	ShorthandName        string `json:"shorthandName"`
	Description          string `json:"description"`
	NumControlsCompleted int    `json:"numControlsCompleted"`
	NumControlsTotal     int    `json:"numControlsTotal"`
	NumDocumentsPassing  int    `json:"numDocumentsPassing"`
	NumDocumentsTotal    int    `json:"numDocumentsTotal"`
	NumTestsPassing      int    `json:"numTestsPassing"`
	NumTestsTotal        int    `json:"numTestsTotal"`
}

// Frameworks fetches all frameworks from the Vanta API, following pagination.
func (c *Client) Frameworks() ([]Framework, error) {
	return fetchAll[Framework](c, "/frameworks")
}

var frameworkHeaders = []string{
	"id", "displayName", "shorthandName", "numControlsCompleted", "numControlsTotal",
	"numDocumentsPassing", "numDocumentsTotal", "numTestsPassing", "numTestsTotal",
}

func frameworkRow(f Framework) []string {
	return []string{
		f.ID,
		f.DisplayName,
		f.ShorthandName,
		strconv.Itoa(f.NumControlsCompleted),
		strconv.Itoa(f.NumControlsTotal),
		strconv.Itoa(f.NumDocumentsPassing),
		strconv.Itoa(f.NumDocumentsTotal),
		strconv.Itoa(f.NumTestsPassing),
		strconv.Itoa(f.NumTestsTotal),
	}
}

func printFrameworksCSV(frameworks []Framework) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(frameworkHeaders); err != nil {
		return err
	}
	for _, f := range frameworks {
		if err := w.Write(frameworkRow(f)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printFrameworksTSV(frameworks []Framework) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(frameworkHeaders); err != nil {
		return err
	}
	for _, f := range frameworks {
		if err := w.Write(frameworkRow(f)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printFrameworksMarkdown(frameworks []Framework) {
	fmt.Println("| " + strings.Join(frameworkHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(frameworkHeaders)))
	for _, f := range frameworks {
		row := frameworkRow(f)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}

}
func printFrameworksPrettyMarkdown(frameworks []Framework) {
	rows := make([][]string, len(frameworks))
	for i, item := range frameworks {
		rows[i] = frameworkRow(item)
	}
	printPrettyMarkdown(frameworkHeaders, rows)
}

func newFrameworksCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "frameworks",
		Short: "Fetch frameworks",
		Run: func(cmd *cobra.Command, args []string) {
			frameworks, err := newClient().Frameworks()
			if err != nil {
				log.Fatal(err)
			}
			switch format {
			case "json":
				if err := printJSON(frameworks); err != nil {
					log.Fatal(err)
				}
			case "csv":
				if err := printFrameworksCSV(frameworks); err != nil {
					log.Fatal(err)
				}
			case "tsv":
				if err := printFrameworksTSV(frameworks); err != nil {
					log.Fatal(err)
				}
			case "markdown":
				printFrameworksMarkdown(frameworks)
			case "pretty_markdown":
				printFrameworksPrettyMarkdown(frameworks)
			default:
				log.Fatalf("unknown format %q: must be json, csv, tsv, markdown, or pretty_markdown", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, tsv, markdown, pretty_markdown")
	return cmd
}
