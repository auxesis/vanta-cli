package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// DeactivateMetadata holds information about why and when a vulnerability was deactivated.
type DeactivateMetadata struct {
	DeactivatedBy                 string     `json:"deactivatedBy"`
	DeactivatedOnDate             time.Time  `json:"deactivatedOnDate"`
	DeactivationReason            string     `json:"deactivationReason"`
	DeactivatedUntilDate          *time.Time `json:"deactivatedUntilDate"`
	IsVulnDeactivatedIndefinitely bool       `json:"isVulnDeactivatedIndefinitely"`
}

// Vulnerability represents a single vulnerability returned by the Vanta API.
type Vulnerability struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	Description        string              `json:"description"`
	IntegrationID      string              `json:"integrationId"`
	PackageIdentifier  string              `json:"packageIdentifier"`
	VulnerabilityType  string              `json:"vulnerabilityType"`
	TargetID           string              `json:"targetId"`
	FirstDetectedDate  time.Time           `json:"firstDetectedDate"`
	SourceDetectedDate *time.Time          `json:"sourceDetectedDate"`
	LastDetectedDate   *time.Time          `json:"lastDetectedDate"`
	Severity           string              `json:"severity"`
	CVSSSeverityScore  float64             `json:"cvssSeverityScore"`
	ScannerScore       *int                `json:"scannerScore"`
	IsFixable          bool                `json:"isFixable"`
	RemediateByDate    time.Time           `json:"remediateByDate"`
	RelatedVulns       []string            `json:"relatedVulns"`
	RelatedURLs        []string            `json:"relatedUrls"`
	ExternalURL        string              `json:"externalURL"`
	ScanSource         string              `json:"scanSource"`
	DeactivateMetadata *DeactivateMetadata `json:"deactivateMetadata"`
}

// Vulnerabilities fetches all vulnerabilities from the Vanta API, following pagination.
func (c *Client) Vulnerabilities() ([]Vulnerability, error) {
	return fetchAll[Vulnerability](c, "/vulnerabilities")
}

var vulnerabilityHeaders = []string{
	"id", "name", "severity", "cvssSeverityScore", "isFixable",
	"firstDetectedDate", "remediateByDate", "packageIdentifier",
	"vulnerabilityType", "integrationId", "targetId", "scanSource", "externalURL",
}

func vulnerabilityRow(v Vulnerability) []string {
	return []string{
		v.ID,
		v.Name,
		v.Severity,
		strconv.FormatFloat(v.CVSSSeverityScore, 'f', -1, 64),
		strconv.FormatBool(v.IsFixable),
		v.FirstDetectedDate.Format(time.RFC3339),
		v.RemediateByDate.Format(time.RFC3339),
		v.PackageIdentifier,
		v.VulnerabilityType,
		v.IntegrationID,
		v.TargetID,
		v.ScanSource,
		v.ExternalURL,
	}
}

func printVulnerabilitiesCSV(vulns []Vulnerability) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(vulnerabilityHeaders); err != nil {
		return err
	}
	for _, v := range vulns {
		if err := w.Write(vulnerabilityRow(v)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printVulnerabilitiesTSV(vulns []Vulnerability) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(vulnerabilityHeaders); err != nil {
		return err
	}
	for _, v := range vulns {
		if err := w.Write(vulnerabilityRow(v)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printVulnerabilitiesMarkdown(vulns []Vulnerability) {
	fmt.Println("| " + strings.Join(vulnerabilityHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(vulnerabilityHeaders)))
	for _, v := range vulns {
		row := vulnerabilityRow(v)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}
