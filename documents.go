package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Document represents a single document returned by the Vanta API.
type Document struct {
	ID               string     `json:"id"`
	OwnerID          string     `json:"ownerId"`
	Category         string     `json:"category"`
	Description      string     `json:"description"`
	IsSensitive      bool       `json:"isSensitive"`
	Title            string     `json:"title"`
	UploadStatus     string     `json:"uploadStatus"`
	UploadStatusDate *time.Time `json:"uploadStatusDate"`
	URL              string     `json:"url"`
}

// documentStatusAPIValues maps CLI status flag values to the API's display strings.
var documentStatusAPIValues = map[string]string{
	"NEEDS_DOCUMENT": "Needs document",
	"NEEDS_UPDATE":   "Needs update",
	"NOT_RELEVANT":   "Not relevant",
	"OK":             "OK",
}

// Documents fetches all documents from the Vanta API, following pagination.
func (c *Client) Documents() ([]Document, error) {
	return fetchAll[Document](c, "/documents")
}

// DocumentsFiltered fetches documents filtered by framework and/or status.
// The API supports multiple values via repeated query parameters in a single request.
func (c *Client) DocumentsFiltered(frameworks, statuses []string) ([]Document, error) {
	if len(frameworks) == 0 && len(statuses) == 0 {
		return c.Documents()
	}
	p := url.Values{}
	for _, fw := range frameworks {
		p.Add("frameworkMatchesAny", fw)
	}
	for _, st := range statuses {
		p.Add("statusMatchesAny", documentStatusAPIValues[st])
	}
	return fetchAll[Document](c, "/documents", p)
}

var documentHeaders = []string{
	"id", "ownerId", "category", "title", "uploadStatus", "uploadStatusDate", "isSensitive", "url",
}

func documentRow(d Document) []string {
	uploadStatusDate := ""
	if d.UploadStatusDate != nil {
		uploadStatusDate = d.UploadStatusDate.Format(time.RFC3339)
	}
	return []string{d.ID, d.OwnerID, d.Category, d.Title, d.UploadStatus, uploadStatusDate, strconv.FormatBool(d.IsSensitive), d.URL}
}

func printDocumentsCSV(docs []Document) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(documentHeaders); err != nil {
		return err
	}
	for _, d := range docs {
		if err := w.Write(documentRow(d)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printDocumentsTSV(docs []Document) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(documentHeaders); err != nil {
		return err
	}
	for _, d := range docs {
		if err := w.Write(documentRow(d)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printDocumentsMarkdown(docs []Document) {
	fmt.Println("| " + strings.Join(documentHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(documentHeaders)))
	for _, d := range docs {
		row := documentRow(d)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}

}
func printDocumentsPrettyMarkdown(docs []Document) {
	rows := make([][]string, len(docs))
	for i, item := range docs {
		rows[i] = documentRow(item)
	}
	printPrettyMarkdown(documentHeaders, rows)
}

var validDocumentStatuses = map[string]bool{
	"NEEDS_DOCUMENT": true,
	"NEEDS_UPDATE":   true,
	"NOT_RELEVANT":   true,
	"OK":             true,
}

func filterDocumentsDueBefore(docs []Document, dueBefore time.Time) []Document {
	if dueBefore.IsZero() {
		return docs
	}
	filtered := docs[:0]
	for _, d := range docs {
		if d.UploadStatusDate != nil && d.UploadStatusDate.Before(dueBefore) {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func filterDocumentsDueAfter(docs []Document, dueAfter time.Time) []Document {
	if dueAfter.IsZero() {
		return docs
	}
	filtered := docs[:0]
	for _, d := range docs {
		if d.UploadStatusDate != nil && d.UploadStatusDate.After(dueAfter) {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func newDocumentsCmd() *cobra.Command {
	var format string
	var frameworkFlag string
	var statusFlag string
	var dueBeforeFlag string
	var dueAfterFlag string

	cmd := &cobra.Command{
		Use:   "documents",
		Short: "Fetch documents",
		Run: func(cmd *cobra.Command, args []string) {
			var frameworks, statuses []string
			for _, f := range splitFilter(frameworkFlag) {
				if !validFrameworks[f] {
					log.Fatalf("unknown framework %q: valid values are soc2", f)
				}
				frameworks = append(frameworks, f)
			}
			for _, s := range splitFilter(statusFlag) {
				if !validDocumentStatuses[s] {
					log.Fatalf("unknown status %q: valid values are NEEDS_DOCUMENT, NEEDS_UPDATE, NOT_RELEVANT, OK", s)
				}
				statuses = append(statuses, s)
			}
			var dueBefore time.Time
			if dueBeforeFlag != "" {
				var err error
				dueBefore, err = time.Parse(time.DateOnly, dueBeforeFlag)
				if err != nil {
					log.Fatalf("invalid --due-before %q: must be YYYY-MM-DD", dueBeforeFlag)
				}
			}
			var dueAfter time.Time
			if dueAfterFlag != "" {
				var err error
				dueAfter, err = time.Parse(time.DateOnly, dueAfterFlag)
				if err != nil {
					log.Fatalf("invalid --due-after %q: must be YYYY-MM-DD", dueAfterFlag)
				}
			}
			docs, err := newClient().DocumentsFiltered(frameworks, statuses)
			if err != nil {
				log.Fatal(err)
			}
			docs = filterDocumentsDueBefore(docs, dueBefore)
			docs = filterDocumentsDueAfter(docs, dueAfter)
			switch format {
			case "json":
				if err := printJSON(docs); err != nil {
					log.Fatal(err)
				}
			case "csv":
				if err := printDocumentsCSV(docs); err != nil {
					log.Fatal(err)
				}
			case "tsv":
				if err := printDocumentsTSV(docs); err != nil {
					log.Fatal(err)
				}
			case "markdown":
				printDocumentsMarkdown(docs)
			case "pretty_markdown":
				printDocumentsPrettyMarkdown(docs)
			default:
				log.Fatalf("unknown format %q: must be json, csv, tsv, markdown, or pretty_markdown", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, tsv, markdown, pretty_markdown")
	cmd.Flags().StringVar(&frameworkFlag, "framework", "", "comma-separated framework filter (valid: soc2)")
	cmd.Flags().StringVar(&statusFlag, "status", "", "comma-separated status filter (valid: NEEDS_DOCUMENT, NEEDS_UPDATE, NOT_RELEVANT, OK)")
	cmd.Flags().StringVar(&dueBeforeFlag, "due-before", "", "only show documents due before this date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&dueAfterFlag, "due-after", "", "only show documents due after this date (YYYY-MM-DD)")
	return cmd
}
