package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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

// Documents fetches all documents from the Vanta API, following pagination.
func (c *Client) Documents() ([]Document, error) {
	return fetchAll[Document](c, "/documents")
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
