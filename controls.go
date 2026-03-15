package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// ControlOwner is the owner of a control.
type ControlOwner struct {
	ID           string `json:"id"`
	EmailAddress string `json:"emailAddress"`
	DisplayName  string `json:"displayName"`
}

// ControlCustomField is a custom field on a control.
type ControlCustomField struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// Control represents a single control returned by the Vanta API.
type Control struct {
	ID               string               `json:"id"`
	ExternalID       string               `json:"externalId"`
	Name             string               `json:"name"`
	Description      string               `json:"description"`
	Source           string               `json:"source"`
	Domains          []string             `json:"domains"`
	Owner            *ControlOwner        `json:"owner"`
	Role             string               `json:"role"`
	CustomFields     []ControlCustomField `json:"customFields"`
	CreationDate     *time.Time           `json:"creationDate"`
	ModificationDate *time.Time           `json:"modificationDate"`
}

// Controls fetches all controls from the Vanta API, following pagination.
func (c *Client) Controls() ([]Control, error) {
	return fetchAll[Control](c, "/controls")
}

var controlHeaders = []string{
	"id", "externalId", "name", "source", "role", "owner", "domains",
}

func controlRow(v Control) []string {
	owner := ""
	if v.Owner != nil {
		owner = v.Owner.DisplayName
	}
	return []string{
		v.ID,
		v.ExternalID,
		v.Name,
		v.Source,
		v.Role,
		owner,
		strings.Join(v.Domains, "|"),
	}
}

func printControlsCSV(controls []Control) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(controlHeaders); err != nil {
		return err
	}
	for _, v := range controls {
		if err := w.Write(controlRow(v)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printControlsTSV(controls []Control) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(controlHeaders); err != nil {
		return err
	}
	for _, v := range controls {
		if err := w.Write(controlRow(v)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printControlsMarkdown(controls []Control) {
	fmt.Println("| " + strings.Join(controlHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(controlHeaders)))
	for _, v := range controls {
		row := controlRow(v)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}

func newControlsCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "controls",
		Short: "Fetch controls",
		Run: func(cmd *cobra.Command, args []string) {
			controls, err := newClient().Controls()
			if err != nil {
				log.Fatal(err)
			}
			switch format {
			case "json":
				if err := printJSON(controls); err != nil {
					log.Fatal(err)
				}
			case "csv":
				if err := printControlsCSV(controls); err != nil {
					log.Fatal(err)
				}
			case "tsv":
				if err := printControlsTSV(controls); err != nil {
					log.Fatal(err)
				}
			case "markdown":
				printControlsMarkdown(controls)
			default:
				log.Fatalf("unknown format %q: must be json, csv, tsv, or markdown", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, tsv, markdown")
	return cmd
}
