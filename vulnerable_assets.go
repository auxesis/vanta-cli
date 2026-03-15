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

// VulnerableAssetTag represents a key/value tag on a vulnerable asset scanner.
type VulnerableAssetTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// VulnerableAssetScanner holds scanner details for a vulnerable asset.
type VulnerableAssetScanner struct {
	ResourceID                  string               `json:"resourceId"`
	IntegrationID               string               `json:"integrationId"`
	ImageDigest                 string               `json:"imageDigest"`
	ImagePushedAtDate           *time.Time           `json:"imagePushedAtDate"`
	ImageTags                   []string             `json:"imageTags"`
	AssetTags                   []VulnerableAssetTag `json:"assetTags"`
	ParentAccountOrOrganization string               `json:"parentAccountOrOrganization"`
	BiosUUID                    string               `json:"biosUuid"`
	IPv4s                       []string             `json:"ipv4s"`
	IPv6s                       []string             `json:"ipv6s"`
	MacAddresses                []string             `json:"macAddresses"`
	Hostnames                   []string             `json:"hostnames"`
	FQDNs                       []string             `json:"fqdns"`
	OperatingSystems            []string             `json:"operatingSystems"`
	TargetID                    string               `json:"targetId"`
}

// VulnerableAsset represents a single vulnerable asset returned by the Vanta API.
type VulnerableAsset struct {
	ID             string                   `json:"id"`
	Name           string                   `json:"name"`
	AssetType      string                   `json:"assetType"`
	HasBeenScanned bool                     `json:"hasBeenScanned"`
	ImageScanTag   string                   `json:"imageScanTag"`
	Scanners       []VulnerableAssetScanner `json:"scanners"`
}

// VulnerableAssets fetches all vulnerable assets from the Vanta API, following pagination.
func (c *Client) VulnerableAssets() ([]VulnerableAsset, error) {
	return fetchAll[VulnerableAsset](c, "/vulnerable-assets")
}

var vulnerableAssetHeaders = []string{
	"id", "name", "assetType", "hasBeenScanned", "imageScanTag", "scannerCount",
}

func vulnerableAssetRow(v VulnerableAsset) []string {
	return []string{
		v.ID,
		v.Name,
		v.AssetType,
		strconv.FormatBool(v.HasBeenScanned),
		v.ImageScanTag,
		strconv.Itoa(len(v.Scanners)),
	}
}

func printVulnerableAssetsCSV(items []VulnerableAsset) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(vulnerableAssetHeaders); err != nil {
		return err
	}
	for _, v := range items {
		if err := w.Write(vulnerableAssetRow(v)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printVulnerableAssetsTSV(items []VulnerableAsset) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(vulnerableAssetHeaders); err != nil {
		return err
	}
	for _, v := range items {
		if err := w.Write(vulnerableAssetRow(v)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printVulnerableAssetsMarkdown(items []VulnerableAsset) {
	fmt.Println("| " + strings.Join(vulnerableAssetHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(vulnerableAssetHeaders)))
	for _, v := range items {
		row := vulnerableAssetRow(v)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}
}

func newVulnerableAssetsCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "vulnerable-assets",
		Short: "Fetch vulnerable assets",
		Run: func(cmd *cobra.Command, args []string) {
			items, err := newClient().VulnerableAssets()
			if err != nil {
				log.Fatal(err)
			}
			switch format {
			case "json":
				if err := printJSON(items); err != nil {
					log.Fatal(err)
				}
			case "csv":
				if err := printVulnerableAssetsCSV(items); err != nil {
					log.Fatal(err)
				}
			case "tsv":
				if err := printVulnerableAssetsTSV(items); err != nil {
					log.Fatal(err)
				}
			case "markdown":
				printVulnerableAssetsMarkdown(items)
			default:
				log.Fatalf("unknown format %q: must be json, csv, tsv, or markdown", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, tsv, markdown")
	return cmd
}
