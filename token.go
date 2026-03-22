package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var tokenHeaders = []string{"accessToken"}

func tokenRow(t *TokenResponse) []string {
	return []string{t.AccessToken}
}

func newTokenCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "token",
		Short: "Fetch an OAuth access token",
		Run: func(cmd *cobra.Command, args []string) {
			token, err := getOAuthToken(os.Getenv("VANTA_CLIENT_ID"), os.Getenv("VANTA_CLIENT_SECRET"))
			if err != nil {
				log.Fatal(err)
			}
			switch format {
			case "json":
				if err := printJSON(struct {
					AccessToken string `json:"accessToken"`
				}{token.AccessToken}); err != nil {
					log.Fatal(err)
				}
			case "csv":
				w := csv.NewWriter(os.Stdout)
				if err := w.Write(tokenHeaders); err != nil {
					log.Fatal(err)
				}
				if err := w.Write(tokenRow(token)); err != nil {
					log.Fatal(err)
				}
				w.Flush()
				if err := w.Error(); err != nil {
					log.Fatal(err)
				}
			case "tsv":
				w := csv.NewWriter(os.Stdout)
				w.Comma = '\t'
				if err := w.Write(tokenHeaders); err != nil {
					log.Fatal(err)
				}
				if err := w.Write(tokenRow(token)); err != nil {
					log.Fatal(err)
				}
				w.Flush()
				if err := w.Error(); err != nil {
					log.Fatal(err)
				}
			case "markdown":
				fmt.Println("| " + strings.Join(tokenHeaders, " | ") + " |")
				fmt.Println("|" + strings.Repeat(" --- |", len(tokenHeaders)))
				fmt.Println("| " + strings.Join(tokenRow(token), " | ") + " |")
			case "plain":
				fmt.Println(token.AccessToken)
			default:
				log.Fatalf("unknown format %q: must be json, csv, tsv, markdown, or plain", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, tsv, markdown, plain")
	return cmd
}
