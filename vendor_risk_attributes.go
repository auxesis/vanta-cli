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

// VendorRiskAttribute represents a single vendor risk attribute returned by the Vanta API.
type VendorRiskAttribute struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	VendorCategories []string `json:"vendorCategories"`
	Enabled          bool     `json:"enabled"`
	RiskLevel        string   `json:"riskLevel"`
}

// VendorRiskAttributes fetches all vendor risk attributes from the Vanta API, following pagination.
func (c *Client) VendorRiskAttributes() ([]VendorRiskAttribute, error) {
	return fetchAll[VendorRiskAttribute](c, "/vendor-risk-attributes")
}

var vendorRiskAttributeHeaders = []string{
	"id", "name", "riskLevel", "enabled", "vendorCategories",
}

func vendorRiskAttributeRow(v VendorRiskAttribute) []string {
	return []string{
		v.ID,
		v.Name,
		v.RiskLevel,
		strconv.FormatBool(v.Enabled),
		strings.Join(v.VendorCategories, "|"),
	}
}

func printVendorRiskAttributesCSV(items []VendorRiskAttribute) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(vendorRiskAttributeHeaders); err != nil {
		return err
	}
	for _, v := range items {
		if err := w.Write(vendorRiskAttributeRow(v)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printVendorRiskAttributesTSV(items []VendorRiskAttribute) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(vendorRiskAttributeHeaders); err != nil {
		return err
	}
	for _, v := range items {
		if err := w.Write(vendorRiskAttributeRow(v)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printVendorRiskAttributesMarkdown(items []VendorRiskAttribute) {
	fmt.Println("| " + strings.Join(vendorRiskAttributeHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(vendorRiskAttributeHeaders)))
	for _, v := range items {
		row := vendorRiskAttributeRow(v)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}

}
func printVendorRiskAttributesPrettyMarkdown(items []VendorRiskAttribute) {
	rows := make([][]string, len(items))
	for i, item := range items {
		rows[i] = vendorRiskAttributeRow(item)
	}
	printPrettyMarkdown(vendorRiskAttributeHeaders, rows)
}

func newVendorRiskAttributesCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "vendor-risk-attributes",
		Short: "Fetch vendor risk attributes",
		Run: func(cmd *cobra.Command, args []string) {
			items, err := newClient().VendorRiskAttributes()
			if err != nil {
				log.Fatal(err)
			}
			switch format {
			case "json":
				if err := printJSON(items); err != nil {
					log.Fatal(err)
				}
			case "csv":
				if err := printVendorRiskAttributesCSV(items); err != nil {
					log.Fatal(err)
				}
			case "tsv":
				if err := printVendorRiskAttributesTSV(items); err != nil {
					log.Fatal(err)
				}
			case "markdown":
				printVendorRiskAttributesMarkdown(items)
			case "pretty_markdown":
				printVendorRiskAttributesPrettyMarkdown(items)
			default:
				log.Fatalf("unknown format %q: must be json, csv, tsv, markdown, or pretty_markdown", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, tsv, markdown, pretty_markdown")
	return cmd
}
