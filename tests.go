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

// TestVersion holds the major/minor version of a test.
type TestVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
}

// TestDeactivatedStatusInfo holds deactivation info for a test.
type TestDeactivatedStatusInfo struct {
	IsDeactivated     bool       `json:"isDeactivated"`
	DeactivatedReason *string    `json:"deactivatedReason"`
	LastUpdatedDate   *time.Time `json:"lastUpdatedDate"`
}

// TestRemediationStatusInfo holds remediation status info for a test.
type TestRemediationStatusInfo struct {
	Status                 string     `json:"status"`
	SoonestRemediateByDate *time.Time `json:"soonestRemediateByDate"`
	ItemCount              int        `json:"itemCount"`
}

// TestOwner holds the owner of a test.
type TestOwner struct {
	ID           string `json:"id"`
	EmailAddress string `json:"emailAddress"`
	DisplayName  string `json:"displayName"`
}

// Test represents a single test returned by the Vanta API.
type Test struct {
	ID                     string                     `json:"id"`
	Name                   string                     `json:"name"`
	LastTestRunDate        time.Time                  `json:"lastTestRunDate"`
	LatestFlipDate         *time.Time                 `json:"latestFlipDate"`
	Description            string                     `json:"description"`
	FailureDescription     string                     `json:"failureDescription"`
	RemediationDescription string                     `json:"remediationDescription"`
	Version                TestVersion                `json:"version"`
	Category               string                     `json:"category"`
	Integrations           []string                   `json:"integrations"`
	Status                 string                     `json:"status"`
	DeactivatedStatusInfo  *TestDeactivatedStatusInfo `json:"deactivatedStatusInfo"`
	RemediationStatusInfo  *TestRemediationStatusInfo `json:"remediationStatusInfo"`
	Owner                  *TestOwner                 `json:"owner"`
}

// Tests fetches all tests from the Vanta API, following pagination.
func (c *Client) Tests() ([]Test, error) {
	return fetchAll[Test](c, "/tests")
}

var testHeaders = []string{
	"id", "name", "status", "category", "lastTestRunDate",
	"integrations", "isDeactivated", "remediationStatus", "remediationItemCount",
}

func testRow(t Test) []string {
	integrations := strings.Join(t.Integrations, "|")
	isDeactivated := "false"
	if t.DeactivatedStatusInfo != nil {
		isDeactivated = strconv.FormatBool(t.DeactivatedStatusInfo.IsDeactivated)
	}
	remediationStatus := ""
	remediationItemCount := ""
	if t.RemediationStatusInfo != nil {
		remediationStatus = t.RemediationStatusInfo.Status
		remediationItemCount = strconv.Itoa(t.RemediationStatusInfo.ItemCount)
	}
	return []string{
		t.ID,
		t.Name,
		t.Status,
		t.Category,
		t.LastTestRunDate.Format(time.RFC3339),
		integrations,
		isDeactivated,
		remediationStatus,
		remediationItemCount,
	}
}

func printTestsCSV(tests []Test) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(testHeaders); err != nil {
		return err
	}
	for _, t := range tests {
		if err := w.Write(testRow(t)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printTestsTSV(tests []Test) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(testHeaders); err != nil {
		return err
	}
	for _, t := range tests {
		if err := w.Write(testRow(t)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printTestsMarkdown(tests []Test) {
	fmt.Println("| " + strings.Join(testHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(testHeaders)))
	for _, t := range tests {
		row := testRow(t)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}

func newTestsCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "tests",
		Short: "Fetch tests",
		Run: func(cmd *cobra.Command, args []string) {
			tests, err := newClient().Tests()
			if err != nil {
				log.Fatal(err)
			}
			switch format {
			case "json":
				if err := printJSON(tests); err != nil {
					log.Fatal(err)
				}
			case "csv":
				if err := printTestsCSV(tests); err != nil {
					log.Fatal(err)
				}
			case "tsv":
				if err := printTestsTSV(tests); err != nil {
					log.Fatal(err)
				}
			case "markdown":
				printTestsMarkdown(tests)
			default:
				log.Fatalf("unknown format %q: must be json, csv, tsv, or markdown", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, tsv, markdown")
	return cmd
}
