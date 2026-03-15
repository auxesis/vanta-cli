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

// IntegrationConnection represents a single connection for an integration.
type IntegrationConnection struct {
	ConnectionID           string  `json:"connectionId"`
	IsDisabled             bool    `json:"isDisabled"`
	ConnectionErrorMessage *string `json:"connectionErrorMessage"`
}

// Integration represents a single integration returned by the Vanta API.
type Integration struct {
	IntegrationID string                  `json:"integrationId"`
	DisplayName   string                  `json:"displayName"`
	ResourceKinds []string                `json:"resourceKinds"`
	Connections   []IntegrationConnection `json:"connections"`
}

// Integrations fetches all integrations from the Vanta API, following pagination.
func (c *Client) Integrations() ([]Integration, error) {
	return fetchAll[Integration](c, "/integrations")
}

var integrationHeaders = []string{
	"integrationId", "displayName", "resourceKinds", "connectionCount", "disabledConnectionCount",
}

func integrationRow(i Integration) []string {
	disabledCount := 0
	for _, conn := range i.Connections {
		if conn.IsDisabled {
			disabledCount++
		}
	}
	return []string{
		i.IntegrationID,
		i.DisplayName,
		strings.Join(i.ResourceKinds, "|"),
		strconv.Itoa(len(i.Connections)),
		strconv.Itoa(disabledCount),
	}
}

func printIntegrationsCSV(integrations []Integration) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(integrationHeaders); err != nil {
		return err
	}
	for _, i := range integrations {
		if err := w.Write(integrationRow(i)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printIntegrationsTSV(integrations []Integration) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(integrationHeaders); err != nil {
		return err
	}
	for _, i := range integrations {
		if err := w.Write(integrationRow(i)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printIntegrationsMarkdown(integrations []Integration) {
	fmt.Println("| " + strings.Join(integrationHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(integrationHeaders)))
	for _, i := range integrations {
		row := integrationRow(i)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}

func newIntegrationsCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "integrations",
		Short: "Fetch integrations",
		Run: func(cmd *cobra.Command, args []string) {
			items, err := newClient().Integrations()
			if err != nil {
				log.Fatal(err)
			}
			switch format {
			case "json":
				if err := printJSON(items); err != nil {
					log.Fatal(err)
				}
			case "csv":
				if err := printIntegrationsCSV(items); err != nil {
					log.Fatal(err)
				}
			case "tsv":
				if err := printIntegrationsTSV(items); err != nil {
					log.Fatal(err)
				}
			case "markdown":
				printIntegrationsMarkdown(items)
			default:
				log.Fatalf("unknown format %q: must be json, csv, tsv, or markdown", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, tsv, markdown")
	return cmd
}
