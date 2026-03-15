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

// PolicyDocument represents a single document attachment on a policy version.
type PolicyDocument struct {
	Language string `json:"language"`
	SlugID   string `json:"slugId"`
	URL      string `json:"url"`
}

// PolicyVersion represents a version of a policy.
type PolicyVersion struct {
	Status    string           `json:"status"`
	Documents []PolicyDocument `json:"documents"`
}

// Policy represents a single policy returned by the Vanta API.
type Policy struct {
	ID                    string         `json:"id"`
	Name                  string         `json:"name"`
	Description           string         `json:"description"`
	Status                string         `json:"status"`
	ApprovedAtDate        *time.Time     `json:"approvedAtDate"`
	LatestVersion         *PolicyVersion `json:"latestVersion"`
	LatestApprovedVersion *PolicyVersion `json:"latestApprovedVersion"`
}

// Policies fetches all policies from the Vanta API, following pagination.
func (c *Client) Policies() ([]Policy, error) {
	return fetchAll[Policy](c, "/policies")
}

var policyHeaders = []string{
	"id", "name", "status", "approvedAtDate", "latestVersionStatus",
}

func policyRow(p Policy) []string {
	approvedAt := ""
	if p.ApprovedAtDate != nil {
		approvedAt = p.ApprovedAtDate.Format(time.RFC3339)
	}
	latestStatus := ""
	if p.LatestVersion != nil {
		latestStatus = p.LatestVersion.Status
	}
	return []string{p.ID, p.Name, p.Status, approvedAt, latestStatus}
}

func printPoliciesCSV(policies []Policy) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(policyHeaders); err != nil {
		return err
	}
	for _, p := range policies {
		if err := w.Write(policyRow(p)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printPoliciesTSV(policies []Policy) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(policyHeaders); err != nil {
		return err
	}
	for _, p := range policies {
		if err := w.Write(policyRow(p)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printPoliciesMarkdown(policies []Policy) {
	fmt.Println("| " + strings.Join(policyHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(policyHeaders)))
	for _, p := range policies {
		row := policyRow(p)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}

func newPoliciesCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "policies",
		Short: "Fetch policies",
		Run: func(cmd *cobra.Command, args []string) {
			policies, err := newClient().Policies()
			if err != nil {
				log.Fatal(err)
			}
			switch format {
			case "json":
				if err := printJSON(policies); err != nil {
					log.Fatal(err)
				}
			case "csv":
				if err := printPoliciesCSV(policies); err != nil {
					log.Fatal(err)
				}
			case "tsv":
				if err := printPoliciesTSV(policies); err != nil {
					log.Fatal(err)
				}
			case "markdown":
				printPoliciesMarkdown(policies)
			default:
				log.Fatalf("unknown format %q: must be json, csv, tsv, or markdown", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, tsv, markdown")
	return cmd
}
