package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func printPrettyMarkdown(headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	var sb strings.Builder
	for i, h := range headers {
		fmt.Fprintf(&sb, "| %-*s ", widths[i], h)
	}
	sb.WriteString("|")
	fmt.Println(sb.String())
	sb.Reset()
	for _, w := range widths {
		sb.WriteString("|")
		sb.WriteString(strings.Repeat("-", w+2))
	}
	sb.WriteString("|")
	fmt.Println(sb.String())
	for _, row := range rows {
		sb.Reset()
		for i, cell := range row {
			fmt.Fprintf(&sb, "| %-*s ", widths[i], cell)
		}
		sb.WriteString("|")
		fmt.Println(sb.String())
	}
}

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:     "vanta-cli",
		Short:   "Fetch data from the Vanta API",
		Version: version,
	}

	root.AddCommand(
		newTokenCmd(),
		newVulnerabilitiesCmd(),
		newPoliciesCmd(),
		newDocumentsCmd(),
		newDiscoveredVendorsCmd(),
		newVendorsCmd(),
		newControlsCmd(),
		newFrameworksCmd(),
		newGroupsCmd(),
		newIntegrationsCmd(),
		newMonitoredComputersCmd(),
		newPeopleCmd(),
		newRiskScenariosCmd(),
		newTestsCmd(),
		newVendorRiskAttributesCmd(),
		newVulnerabilityRemediationsCmd(),
		newVulnerableAssetsCmd(),
		newSchemaCmd(root),
	)

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
