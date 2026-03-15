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

// VendorCategory represents the category of a vendor.
type VendorCategory struct {
	DisplayName string `json:"displayName"`
}

// VendorAuthDetails holds authentication details for a vendor.
type VendorAuthDetails struct {
	Method                 string `json:"method"`
	PasswordMFA            bool   `json:"passwordMFA"`
	PasswordRequiresNumber bool   `json:"passwordRequiresNumber"`
	PasswordRequiresSymbol bool   `json:"passwordRequiresSymbol"`
	PasswordMinimumLength  int    `json:"passwordMinimumLength"`
}

// VendorContractAmount holds the contract amount and currency for a vendor.
type VendorContractAmount struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// VendorLatestDecision holds the latest review decision for a vendor.
type VendorLatestDecision struct {
	Status        string    `json:"status"`
	LastUpdatedAt time.Time `json:"lastUpdatedAt"`
}

// VendorLinkedTaskTrackerTask holds a linked task tracker task for a vendor.
type VendorLinkedTaskTrackerTask struct {
	Service string `json:"service"`
	URL     string `json:"url"`
}

// Vendor represents a single vendor returned by the Vanta API.
type Vendor struct {
	ID                                      string                       `json:"id"`
	Name                                    string                       `json:"name"`
	WebsiteURL                              string                       `json:"websiteUrl"`
	AccountManagerName                      string                       `json:"accountManagerName"`
	AccountManagerEmail                     string                       `json:"accountManagerEmail"`
	ServicesProvided                        string                       `json:"servicesProvided"`
	AdditionalNotes                         string                       `json:"additionalNotes"`
	AuthDetails                             *VendorAuthDetails           `json:"authDetails"`
	SecurityOwnerUserID                     string                       `json:"securityOwnerUserId"`
	BusinessOwnerUserID                     string                       `json:"businessOwnerUserId"`
	ContractStartDate                       *time.Time                   `json:"contractStartDate"`
	ContractRenewalDate                     *time.Time                   `json:"contractRenewalDate"`
	ContractTerminationDate                 *time.Time                   `json:"contractTerminationDate"`
	LastSecurityReviewCompletionDate        *time.Time                   `json:"lastSecurityReviewCompletionDate"`
	NextSecurityReviewDueDate               *time.Time                   `json:"nextSecurityReviewDueDate"`
	IsVisibleToAuditors                     bool                         `json:"isVisibleToAuditors"`
	IsRiskAutoScored                        bool                         `json:"isRiskAutoScored"`
	Category                                *VendorCategory              `json:"category"`
	RiskAttributeIDs                        []string                     `json:"riskAttributeIds"`
	Status                                  string                       `json:"status"`
	InherentRiskLevel                       string                       `json:"inherentRiskLevel"`
	ResidualRiskLevel                       string                       `json:"residualRiskLevel"`
	VendorHeadquarters                      string                       `json:"vendorHeadquarters"`
	ContractAmount                          *VendorContractAmount        `json:"contractAmount"`
	LatestDecision                          *VendorLatestDecision        `json:"latestDecision"`
	LinkedTaskTrackerTaskProcurementRequest *VendorLinkedTaskTrackerTask `json:"linkedTaskTrackerTaskProcurementRequest"`
}

// Vendors fetches all vendors from the Vanta API, following pagination.
func (c *Client) Vendors() ([]Vendor, error) {
	return fetchAll[Vendor](c, "/vendors")
}

var vendorHeaders = []string{
	"id", "name", "status", "category", "inherentRiskLevel", "residualRiskLevel",
	"websiteUrl", "accountManagerName", "accountManagerEmail",
	"contractStartDate", "contractRenewalDate", "contractTerminationDate",
	"lastSecurityReviewCompletionDate", "nextSecurityReviewDueDate",
	"vendorHeadquarters", "isVisibleToAuditors",
}

func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func vendorRow(v Vendor) []string {
	category := ""
	if v.Category != nil {
		category = v.Category.DisplayName
	}
	return []string{
		v.ID,
		v.Name,
		v.Status,
		category,
		v.InherentRiskLevel,
		v.ResidualRiskLevel,
		v.WebsiteURL,
		v.AccountManagerName,
		v.AccountManagerEmail,
		formatDate(v.ContractStartDate),
		formatDate(v.ContractRenewalDate),
		formatDate(v.ContractTerminationDate),
		formatDate(v.LastSecurityReviewCompletionDate),
		formatDate(v.NextSecurityReviewDueDate),
		v.VendorHeadquarters,
		fmt.Sprintf("%v", v.IsVisibleToAuditors),
	}
}

func printVendorsCSV(vendors []Vendor) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(vendorHeaders); err != nil {
		return err
	}
	for _, v := range vendors {
		if err := w.Write(vendorRow(v)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printVendorsTSV(vendors []Vendor) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(vendorHeaders); err != nil {
		return err
	}
	for _, v := range vendors {
		if err := w.Write(vendorRow(v)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printVendorsMarkdown(vendors []Vendor) {
	fmt.Println("| " + strings.Join(vendorHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(vendorHeaders)))
	for _, v := range vendors {
		row := vendorRow(v)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}

func newVendorsCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "vendors",
		Short: "Fetch vendors",
		Run: func(cmd *cobra.Command, args []string) {
			vendors, err := newClient().Vendors()
			if err != nil {
				log.Fatal(err)
			}
			switch format {
			case "json":
				if err := printJSON(vendors); err != nil {
					log.Fatal(err)
				}
			case "csv":
				if err := printVendorsCSV(vendors); err != nil {
					log.Fatal(err)
				}
			case "tsv":
				if err := printVendorsTSV(vendors); err != nil {
					log.Fatal(err)
				}
			case "markdown":
				printVendorsMarkdown(vendors)
			default:
				log.Fatalf("unknown format %q: must be json, csv, tsv, or markdown", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, tsv, markdown")
	return cmd
}
