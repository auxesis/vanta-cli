package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
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

// DetailedControl extends Control with fields from the per-control detail endpoint.
type DetailedControl struct {
	Control
	NumDocumentsPassing int     `json:"numDocumentsPassing"`
	NumDocumentsTotal   int     `json:"numDocumentsTotal"`
	NumTestsPassing     int     `json:"numTestsPassing"`
	NumTestsTotal       int     `json:"numTestsTotal"`
	Status              string  `json:"status"`
	Note                *string `json:"note"`
}

// Controls fetches all controls from the Vanta API, following pagination.
func (c *Client) Controls() ([]Control, error) {
	return fetchAll[Control](c, "/controls")
}

// ControlByID fetches a single control by ID from the detail endpoint.
func (c *Client) ControlByID(id string) (DetailedControl, error) {
	return fetchByID[DetailedControl](c, "/controls/"+id)
}

// ControlsDetailed fetches all controls then enriches each with detail data.
func (c *Client) ControlsDetailed() ([]DetailedControl, error) {
	controls, err := c.Controls()
	if err != nil {
		return nil, err
	}
	detailed := make([]DetailedControl, 0, len(controls))
	for _, ctrl := range controls {
		d, err := c.ControlByID(ctrl.ID)
		if err != nil {
			return nil, fmt.Errorf("fetching detail for control %s: %w", ctrl.ID, err)
		}
		detailed = append(detailed, d)
	}
	return detailed, nil
}

var controlHeaders = []string{
	"id", "externalId", "name", "source", "role", "owner", "domains",
}

var controlDetailedHeaders = []string{
	"id", "externalId", "name", "source", "role", "owner", "domains",
	"status", "numTestsPassing", "numTestsTotal", "numDocumentsPassing", "numDocumentsTotal", "note",
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

func controlDetailedRow(v DetailedControl) []string {
	note := ""
	if v.Note != nil {
		note = *v.Note
	}
	return append(controlRow(v.Control),
		v.Status,
		strconv.Itoa(v.NumTestsPassing),
		strconv.Itoa(v.NumTestsTotal),
		strconv.Itoa(v.NumDocumentsPassing),
		strconv.Itoa(v.NumDocumentsTotal),
		note,
	)
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

func printDetailedControlsCSV(controls []DetailedControl) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(controlDetailedHeaders); err != nil {
		return err
	}
	for _, v := range controls {
		if err := w.Write(controlDetailedRow(v)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printDetailedControlsTSV(controls []DetailedControl) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(controlDetailedHeaders); err != nil {
		return err
	}
	for _, v := range controls {
		if err := w.Write(controlDetailedRow(v)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printDetailedControlsMarkdown(controls []DetailedControl) {
	fmt.Println("| " + strings.Join(controlDetailedHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(controlDetailedHeaders)))
	for _, v := range controls {
		row := controlDetailedRow(v)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}

func newControlsCmd() *cobra.Command {
	var format string
	var detailed bool

	cmd := &cobra.Command{
		Use:   "controls",
		Short: "Fetch controls",
		Run: func(cmd *cobra.Command, args []string) {
			client := newClient()
			if detailed {
				controls, err := client.ControlsDetailed()
				if err != nil {
					log.Fatal(err)
				}
				switch format {
				case "json":
					if err := printJSON(controls); err != nil {
						log.Fatal(err)
					}
				case "csv":
					if err := printDetailedControlsCSV(controls); err != nil {
						log.Fatal(err)
					}
				case "tsv":
					if err := printDetailedControlsTSV(controls); err != nil {
						log.Fatal(err)
					}
				case "markdown":
					printDetailedControlsMarkdown(controls)
				default:
					log.Fatalf("unknown format %q: must be json, csv, tsv, or markdown", format)
				}
				return
			}

			controls, err := client.Controls()
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
	cmd.Flags().BoolVar(&detailed, "detailed", false, "fetch per-control detail (status, test/document counts, note)")
	return cmd
}
