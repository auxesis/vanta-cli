package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func main() {
	root := &cobra.Command{
		Use:   "vanta-cli",
		Short: "Fetch data from the Vanta API",
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
