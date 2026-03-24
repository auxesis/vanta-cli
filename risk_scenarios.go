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

// RiskScenario represents a single risk scenario returned by the Vanta API.
type RiskScenario struct {
	RiskID             string   `json:"riskId"`
	Description        string   `json:"description"`
	IsSensitive        *bool    `json:"isSensitive"`
	Likelihood         int      `json:"likelihood"`
	Impact             int      `json:"impact"`
	ResidualLikelihood int      `json:"residualLikelihood"`
	ResidualImpact     int      `json:"residualImpact"`
	Categories         []string `json:"categories"`
	CIACategories      []string `json:"ciaCategories"`
	Treatment          string   `json:"treatment"`
	Owner              string   `json:"owner"`
	Note               *string  `json:"note"`
	RiskRegister       string   `json:"riskRegister"`
	IsArchived         bool     `json:"isArchived"`
	ReviewStatus       string   `json:"reviewStatus"`
	Type               string   `json:"type"`
}

// RiskScenarios fetches all risk scenarios from the Vanta API, following pagination.
func (c *Client) RiskScenarios() ([]RiskScenario, error) {
	return fetchAll[RiskScenario](c, "/risk-scenarios")
}

var riskScenarioHeaders = []string{
	"riskId", "type", "reviewStatus", "treatment", "likelihood", "impact",
	"residualLikelihood", "residualImpact", "categories", "owner", "isArchived",
}

func riskScenarioRow(r RiskScenario) []string {
	return []string{
		r.RiskID,
		r.Type,
		r.ReviewStatus,
		r.Treatment,
		strconv.Itoa(r.Likelihood),
		strconv.Itoa(r.Impact),
		strconv.Itoa(r.ResidualLikelihood),
		strconv.Itoa(r.ResidualImpact),
		strings.Join(r.Categories, "|"),
		r.Owner,
		strconv.FormatBool(r.IsArchived),
	}
}

func printRiskScenariosCSV(scenarios []RiskScenario) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(riskScenarioHeaders); err != nil {
		return err
	}
	for _, r := range scenarios {
		if err := w.Write(riskScenarioRow(r)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printRiskScenariosTSV(scenarios []RiskScenario) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(riskScenarioHeaders); err != nil {
		return err
	}
	for _, r := range scenarios {
		if err := w.Write(riskScenarioRow(r)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printRiskScenariosMarkdown(scenarios []RiskScenario) {
	fmt.Println("| " + strings.Join(riskScenarioHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(riskScenarioHeaders)))
	for _, r := range scenarios {
		row := riskScenarioRow(r)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}

func newRiskScenariosCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "risk-scenarios",
		Short: "Fetch risk scenarios",
		Run: func(cmd *cobra.Command, args []string) {
			items, err := newClient().RiskScenarios()
			if err != nil {
				log.Fatal(err)
			}
			switch format {
			case "json":
				if err := printJSON(items); err != nil {
					log.Fatal(err)
				}
			case "csv":
				if err := printRiskScenariosCSV(items); err != nil {
					log.Fatal(err)
				}
			case "tsv":
				if err := printRiskScenariosTSV(items); err != nil {
					log.Fatal(err)
				}
			case "markdown":
				printRiskScenariosMarkdown(items)
			default:
				log.Fatalf("unknown format %q: must be json, csv, tsv, or markdown", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, tsv, markdown")
	return cmd
}
