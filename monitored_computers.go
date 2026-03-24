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

// ComputerOutcome holds the outcome of a computer security check.
type ComputerOutcome struct {
	Outcome string `json:"outcome"`
}

// ComputerOperatingSystem holds OS info for a monitored computer.
type ComputerOperatingSystem struct {
	Type    string `json:"type"`
	Version string `json:"version"`
}

// ComputerOwner holds the owner of a monitored computer.
type ComputerOwner struct {
	ID           string `json:"id"`
	EmailAddress string `json:"emailAddress"`
	DisplayName  string `json:"displayName"`
}

// MonitoredComputer represents a single monitored computer returned by the Vanta API.
type MonitoredComputer struct {
	ID                    string                   `json:"id"`
	IntegrationID         string                   `json:"integrationId"`
	SerialNumber          string                   `json:"serialNumber"`
	UDID                  string                   `json:"udid"`
	LastCheckDate         time.Time                `json:"lastCheckDate"`
	Screenlock            *ComputerOutcome         `json:"screenlock"`
	DiskEncryption        *ComputerOutcome         `json:"diskEncryption"`
	PasswordManager       *ComputerOutcome         `json:"passwordManager"`
	AntivirusInstallation *ComputerOutcome         `json:"antivirusInstallation"`
	OperatingSystem       *ComputerOperatingSystem `json:"operatingSystem"`
	Owner                 *ComputerOwner           `json:"owner"`
}

// MonitoredComputers fetches all monitored computers from the Vanta API, following pagination.
func (c *Client) MonitoredComputers() ([]MonitoredComputer, error) {
	return fetchAll[MonitoredComputer](c, "/monitored-computers")
}

var monitoredComputerHeaders = []string{
	"id", "integrationId", "serialNumber", "lastCheckDate", "osType", "osVersion",
	"owner", "screenlock", "diskEncryption", "passwordManager", "antivirus",
}

func monitoredComputerRow(m MonitoredComputer) []string {
	osType := ""
	osVersion := ""
	if m.OperatingSystem != nil {
		osType = m.OperatingSystem.Type
		osVersion = m.OperatingSystem.Version
	}
	owner := ""
	if m.Owner != nil {
		owner = m.Owner.DisplayName
	}
	screenlock := ""
	if m.Screenlock != nil {
		screenlock = m.Screenlock.Outcome
	}
	diskEncryption := ""
	if m.DiskEncryption != nil {
		diskEncryption = m.DiskEncryption.Outcome
	}
	passwordManager := ""
	if m.PasswordManager != nil {
		passwordManager = m.PasswordManager.Outcome
	}
	antivirus := ""
	if m.AntivirusInstallation != nil {
		antivirus = m.AntivirusInstallation.Outcome
	}
	return []string{
		m.ID,
		m.IntegrationID,
		m.SerialNumber,
		m.LastCheckDate.Format(time.RFC3339),
		osType,
		osVersion,
		owner,
		screenlock,
		diskEncryption,
		passwordManager,
		antivirus,
	}
}

func printMonitoredComputersCSV(computers []MonitoredComputer) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(monitoredComputerHeaders); err != nil {
		return err
	}
	for _, m := range computers {
		if err := w.Write(monitoredComputerRow(m)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printMonitoredComputersTSV(computers []MonitoredComputer) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(monitoredComputerHeaders); err != nil {
		return err
	}
	for _, m := range computers {
		if err := w.Write(monitoredComputerRow(m)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printMonitoredComputersMarkdown(computers []MonitoredComputer) {
	fmt.Println("| " + strings.Join(monitoredComputerHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(monitoredComputerHeaders)))
	for _, m := range computers {
		row := monitoredComputerRow(m)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}

}
func printMonitoredComputersPrettyMarkdown(items []MonitoredComputer) {
	rows := make([][]string, len(items))
	for i, item := range items {
		rows[i] = monitoredComputerRow(item)
	}
	printPrettyMarkdown(monitoredComputerHeaders, rows)
}

func newMonitoredComputersCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "monitored-computers",
		Short: "Fetch monitored computers",
		Run: func(cmd *cobra.Command, args []string) {
			items, err := newClient().MonitoredComputers()
			if err != nil {
				log.Fatal(err)
			}
			switch format {
			case "json":
				if err := printJSON(items); err != nil {
					log.Fatal(err)
				}
			case "csv":
				if err := printMonitoredComputersCSV(items); err != nil {
					log.Fatal(err)
				}
			case "tsv":
				if err := printMonitoredComputersTSV(items); err != nil {
					log.Fatal(err)
				}
			case "markdown":
				printMonitoredComputersMarkdown(items)
			case "pretty_markdown":
				printMonitoredComputersPrettyMarkdown(items)
			default:
				log.Fatalf("unknown format %q: must be json, csv, tsv, markdown, or pretty_markdown", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, tsv, markdown, pretty_markdown")
	return cmd
}
