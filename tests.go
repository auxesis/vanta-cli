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

// TestsFiltered fetches tests filtered by framework and/or status.
// The API only accepts a single value per filter per request, so multiple values
// result in separate requests that are merged and deduplicated by test ID.
func (c *Client) TestsFiltered(frameworks, statuses []string) ([]Test, error) {
	if len(frameworks) == 0 && len(statuses) == 0 {
		return c.Tests()
	}

	fws := frameworks
	sts := statuses
	if len(fws) == 0 {
		fws = []string{""}
	}
	if len(sts) == 0 {
		sts = []string{""}
	}

	seen := map[string]bool{}
	var all []Test
	for _, fw := range fws {
		for _, st := range sts {
			p := url.Values{}
			if fw != "" {
				p.Set("frameworkFilter", fw)
			}
			if st != "" {
				p.Set("statusFilter", st)
			}
			results, err := fetchAll[Test](c, "/tests", p)
			if err != nil {
				return nil, err
			}
			for _, t := range results {
				if !seen[t.ID] {
					seen[t.ID] = true
					all = append(all, t)
				}
			}
		}
	}
	return all, nil
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

var validFrameworks = map[string]bool{
	"soc2": true,
}

var validStatuses = map[string]bool{
	"OK": true, "DEACTIVATED": true, "NEEDS_ATTENTION": true,
	"IN_PROGRESS": true, "INVALID": true, "NOT_APPLICABLE": true,
}

func filterTestsDueBefore(tests []Test, dueBefore time.Time) []Test {
	if dueBefore.IsZero() {
		return tests
	}
	filtered := tests[:0]
	for _, t := range tests {
		d := t.RemediationStatusInfo
		if d != nil && d.SoonestRemediateByDate != nil && d.SoonestRemediateByDate.Before(dueBefore) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func filterTestsDueAfter(tests []Test, dueAfter time.Time) []Test {
	if dueAfter.IsZero() {
		return tests
	}
	filtered := tests[:0]
	for _, t := range tests {
		d := t.RemediationStatusInfo
		if d != nil && d.SoonestRemediateByDate != nil && d.SoonestRemediateByDate.After(dueAfter) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func newTestsCmd() *cobra.Command {
	var format string
	var frameworkFlag string
	var statusFlag string
	var dueBeforeFlag string
	var dueAfterFlag string

	cmd := &cobra.Command{
		Use:   "tests",
		Short: "Fetch tests",
		Run: func(cmd *cobra.Command, args []string) {
			var frameworks, statuses []string
			for _, f := range splitFilter(frameworkFlag) {
				if !validFrameworks[f] {
					log.Fatalf("unknown framework %q: valid values are soc2", f)
				}
				frameworks = append(frameworks, f)
			}
			for _, s := range splitFilter(statusFlag) {
				if !validStatuses[s] {
					log.Fatalf("unknown status %q: valid values are OK, DEACTIVATED, NEEDS_ATTENTION, IN_PROGRESS, INVALID, NOT_APPLICABLE", s)
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
			tests, err := newClient().TestsFiltered(frameworks, statuses)
			if err != nil {
				log.Fatal(err)
			}
			tests = filterTestsDueBefore(tests, dueBefore)
			tests = filterTestsDueAfter(tests, dueAfter)
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
	cmd.Flags().StringVar(&frameworkFlag, "framework", "", "comma-separated framework filter (valid: soc2)")
	cmd.Flags().StringVar(&statusFlag, "status", "", "comma-separated status filter (valid: OK, DEACTIVATED, NEEDS_ATTENTION, IN_PROGRESS, INVALID, NOT_APPLICABLE)")
	cmd.Flags().StringVar(&dueBeforeFlag, "due-before", "", "only show tests due before this date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&dueAfterFlag, "due-after", "", "only show tests due after this date (YYYY-MM-DD)")
	return cmd
}

// splitFilter splits a comma-separated filter string, trimming spaces, returning nil for empty input.
func splitFilter(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
