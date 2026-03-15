package main

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var timeType = reflect.TypeOf(time.Time{})

type fieldSchema struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable,omitempty"`
}

type modelSchema struct {
	Fields []fieldSchema `json:"fields"`
}

type flagSchema struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Default     string `json:"default"`
	Description string `json:"description"`
}

type commandSchema struct {
	Name        string       `json:"name"`
	Aliases     []string     `json:"aliases,omitempty"`
	Description string       `json:"description"`
	Flags       []flagSchema `json:"flags,omitempty"`
	OutputModel string       `json:"outputModel,omitempty"`
}

type schema struct {
	Commands []commandSchema        `json:"commands"`
	Models   map[string]modelSchema `json:"models"`
}

func fieldTypeName(t reflect.Type) (name string, nullable bool) {
	if t.Kind() == reflect.Ptr {
		name, _ = fieldTypeName(t.Elem())
		return name, true
	}
	if t == timeType {
		return "datetime", false
	}
	switch t.Kind() {
	case reflect.Slice:
		elem, _ := fieldTypeName(t.Elem())
		return elem + "[]", false
	case reflect.Struct:
		return t.Name(), false
	default:
		return t.Kind().String(), false
	}
}

func collectModels(t reflect.Type, models map[string]modelSchema) {
	if t.Kind() == reflect.Ptr {
		collectModels(t.Elem(), models)
		return
	}
	if t.Kind() != reflect.Struct || t == timeType {
		return
	}
	name := t.Name()
	if _, exists := models[name]; exists {
		return
	}
	// Reserve the name before recursing to avoid infinite loops
	models[name] = modelSchema{}

	var fields []fieldSchema
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		jsonTag := f.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		jsonName := strings.Split(jsonTag, ",")[0]
		tname, nullable := fieldTypeName(f.Type)
		fields = append(fields, fieldSchema{
			Name:     jsonName,
			Type:     tname,
			Nullable: nullable,
		})
		// Recurse into struct and slice-of-struct fields
		ft := f.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct && ft != timeType {
			collectModels(ft, models)
		}
		if ft.Kind() == reflect.Slice {
			elem := ft.Elem()
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			if elem.Kind() == reflect.Struct && elem != timeType {
				collectModels(elem, models)
			}
		}
	}
	models[name] = modelSchema{Fields: fields}
}

func buildSchema(root *cobra.Command, outputModels map[string]reflect.Type) schema {
	models := map[string]modelSchema{}
	for _, t := range outputModels {
		collectModels(t, models)
	}

	var commands []commandSchema
	for _, cmd := range root.Commands() {
		name := strings.SplitN(cmd.Use, " ", 2)[0]
		cs := commandSchema{
			Name:        name,
			Aliases:     cmd.Aliases,
			Description: cmd.Short,
		}
		if t, ok := outputModels[name]; ok {
			cs.OutputModel = t.Name()
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			cs.Flags = append(cs.Flags, flagSchema{
				Name:        f.Name,
				Type:        f.Value.Type(),
				Default:     f.DefValue,
				Description: f.Usage,
			})
		})
		commands = append(commands, cs)
	}

	return schema{Commands: commands, Models: models}
}

func newSchemaCmd(root *cobra.Command) *cobra.Command {
	outputModels := map[string]reflect.Type{
		"vulnerabilities":            reflect.TypeOf(Vulnerability{}),
		"policies":                   reflect.TypeOf(Policy{}),
		"documents":                  reflect.TypeOf(Document{}),
		"discovered-vendors":         reflect.TypeOf(DiscoveredVendor{}),
		"vendors":                    reflect.TypeOf(Vendor{}),
		"controls":                   reflect.TypeOf(Control{}),
		"frameworks":                 reflect.TypeOf(Framework{}),
		"groups":                     reflect.TypeOf(Group{}),
		"integrations":               reflect.TypeOf(Integration{}),
		"monitored-computers":        reflect.TypeOf(MonitoredComputer{}),
		"people":                     reflect.TypeOf(Person{}),
		"risk-scenarios":             reflect.TypeOf(RiskScenario{}),
		"tests":                      reflect.TypeOf(Test{}),
		"vendor-risk-attributes":     reflect.TypeOf(VendorRiskAttribute{}),
		"vulnerability-remediations": reflect.TypeOf(VulnerabilityRemediation{}),
		"vulnerable-assets":          reflect.TypeOf(VulnerableAsset{}),
	}

	return &cobra.Command{
		Use:   "schema",
		Short: "Emit machine-readable schema of subcommands and data models",
		Run: func(cmd *cobra.Command, args []string) {
			s := buildSchema(root, outputModels)
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(s); err != nil {
				log.Fatal(err)
			}
		},
	}
}
