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

// DiscoveredVendorCategory represents the category of a discovered vendor.
type DiscoveredVendorCategory struct {
	Name string `json:"name"`
}

// DiscoveredVendorIgnored holds information about why a vendor was ignored.
type DiscoveredVendorIgnored struct {
	IgnoredByUserID string    `json:"ignoredByUserId"`
	IgnoredReason   string    `json:"ignoredReason"`
	IgnoredAtDate   time.Time `json:"ignoredAtDate"`
}

// DiscoveredVendorRejected holds information about why a vendor was rejected.
type DiscoveredVendorRejected struct {
	RejectedByUserID string    `json:"rejectedByUserId"`
	RejectedReason   string    `json:"rejectedReason"`
	RejectedAtDate   time.Time `json:"rejectedAtDate"`
}

// DiscoveredVendor represents a single discovered vendor returned by the Vanta API.
type DiscoveredVendor struct {
	ID               string                    `json:"id"`
	Name             string                    `json:"name"`
	Category         DiscoveredVendorCategory  `json:"category"`
	Source           string                    `json:"source"`
	NormalizedName   string                    `json:"normalizedName"`
	DiscoveredDate   time.Time                 `json:"discoveredDate"`
	NumberOfAccounts int                       `json:"numberOfAccounts"`
	Ignored          *DiscoveredVendorIgnored  `json:"ignored"`
	Rejected         *DiscoveredVendorRejected `json:"rejected"`
}

// DiscoveredVendors fetches all discovered vendors from the Vanta API, following pagination.
func (c *Client) DiscoveredVendors() ([]DiscoveredVendor, error) {
	return fetchAll[DiscoveredVendor](c, "/discovered-vendors")
}

var discoveredVendorHeaders = []string{
	"id", "name", "category", "source", "normalizedName", "discoveredDate", "numberOfAccounts", "ignored", "rejected",
}

func discoveredVendorRow(v DiscoveredVendor) []string {
	ignored := ""
	if v.Ignored != nil {
		ignored = v.Ignored.IgnoredReason
	}
	rejected := ""
	if v.Rejected != nil {
		rejected = v.Rejected.RejectedReason
	}
	return []string{
		v.ID,
		v.Name,
		v.Category.Name,
		v.Source,
		v.NormalizedName,
		v.DiscoveredDate.Format(time.RFC3339),
		strconv.Itoa(v.NumberOfAccounts),
		ignored,
		rejected,
	}
}

func printDiscoveredVendorsCSV(vendors []DiscoveredVendor) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(discoveredVendorHeaders); err != nil {
		return err
	}
	for _, v := range vendors {
		if err := w.Write(discoveredVendorRow(v)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printDiscoveredVendorsTSV(vendors []DiscoveredVendor) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(discoveredVendorHeaders); err != nil {
		return err
	}
	for _, v := range vendors {
		if err := w.Write(discoveredVendorRow(v)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printDiscoveredVendorsMarkdown(vendors []DiscoveredVendor) {
	fmt.Println("| " + strings.Join(discoveredVendorHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(discoveredVendorHeaders)))
	for _, v := range vendors {
		row := discoveredVendorRow(v)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}

func newDiscoveredVendorsCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "discovered-vendors",
		Short: "Fetch discovered vendors",
		Run: func(cmd *cobra.Command, args []string) {
			vendors, err := newClient().DiscoveredVendors()
			if err != nil {
				log.Fatal(err)
			}
			switch format {
			case "json":
				if err := printJSON(vendors); err != nil {
					log.Fatal(err)
				}
			case "csv":
				if err := printDiscoveredVendorsCSV(vendors); err != nil {
					log.Fatal(err)
				}
			case "tsv":
				if err := printDiscoveredVendorsTSV(vendors); err != nil {
					log.Fatal(err)
				}
			case "markdown":
				printDiscoveredVendorsMarkdown(vendors)
			default:
				log.Fatalf("unknown format %q: must be json, csv, tsv, or markdown", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, tsv, markdown")
	return cmd
}
