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

// Group represents a single group returned by the Vanta API.
type Group struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	CreationDate time.Time `json:"creationDate"`
}

// Groups fetches all groups from the Vanta API, following pagination.
func (c *Client) Groups() ([]Group, error) {
	return fetchAll[Group](c, "/groups")
}

var groupHeaders = []string{
	"id", "name", "creationDate",
}

func groupRow(g Group) []string {
	return []string{
		g.ID,
		g.Name,
		g.CreationDate.Format(time.RFC3339),
	}
}

func printGroupsCSV(groups []Group) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(groupHeaders); err != nil {
		return err
	}
	for _, g := range groups {
		if err := w.Write(groupRow(g)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printGroupsTSV(groups []Group) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(groupHeaders); err != nil {
		return err
	}
	for _, g := range groups {
		if err := w.Write(groupRow(g)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printGroupsMarkdown(groups []Group) {
	fmt.Println("| " + strings.Join(groupHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(groupHeaders)))
	for _, g := range groups {
		row := groupRow(g)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}

func newGroupsCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "groups",
		Short: "Fetch groups",
		Run: func(cmd *cobra.Command, args []string) {
			groups, err := newClient().Groups()
			if err != nil {
				log.Fatal(err)
			}
			switch format {
			case "json":
				if err := printJSON(groups); err != nil {
					log.Fatal(err)
				}
			case "csv":
				if err := printGroupsCSV(groups); err != nil {
					log.Fatal(err)
				}
			case "tsv":
				if err := printGroupsTSV(groups); err != nil {
					log.Fatal(err)
				}
			case "markdown":
				printGroupsMarkdown(groups)
			default:
				log.Fatalf("unknown format %q: must be json, csv, tsv, or markdown", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, tsv, markdown")
	return cmd
}
