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

// PersonName holds the name components of a person.
type PersonName struct {
	Display string `json:"display"`
	Last    string `json:"last"`
	First   string `json:"first"`
}

// PersonEmployment holds employment details for a person.
type PersonEmployment struct {
	EndDate   *time.Time `json:"endDate"`
	JobTitle  string     `json:"jobTitle"`
	StartDate *time.Time `json:"startDate"`
	Status    string     `json:"status"`
}

// PersonSourceRef is a reference to the integration source of a person's data.
type PersonSourceRef struct {
	IntegrationID string `json:"integrationId"`
	ResourceID    string `json:"resourceId"`
	Type          string `json:"type"`
}

// PersonEmploymentSources holds sources for employment start/end dates.
type PersonEmploymentSources struct {
	StartDate *PersonSourceRef `json:"startDate"`
	EndDate   *PersonSourceRef `json:"endDate"`
}

// PersonSources holds the integration sources for a person's fields.
type PersonSources struct {
	EmailAddress *PersonSourceRef         `json:"emailAddress"`
	Employment   *PersonEmploymentSources `json:"employment"`
}

// PersonTaskDisabled holds info about why a task is disabled.
type PersonTaskDisabled struct {
	Date   time.Time `json:"date"`
	Reason string    `json:"reason"`
}

// PersonTaskItem is a named training, policy, or task item.
type PersonTaskItem struct {
	Name string `json:"name"`
}

// PersonCompleteTrainingsTask holds the complete trainings task detail.
type PersonCompleteTrainingsTask struct {
	TaskType            string              `json:"taskType"`
	Status              string              `json:"status"`
	DueDate             *time.Time          `json:"dueDate"`
	CompletionDate      *time.Time          `json:"completionDate"`
	Disabled            *PersonTaskDisabled `json:"disabled"`
	IncompleteTrainings []PersonTaskItem    `json:"incompleteTrainings"`
	CompletedTrainings  []PersonTaskItem    `json:"completedTrainings"`
}

// PersonAcceptPoliciesTask holds the accept policies task detail.
type PersonAcceptPoliciesTask struct {
	TaskType           string              `json:"taskType"`
	Status             string              `json:"status"`
	DueDate            *time.Time          `json:"dueDate"`
	CompletionDate     *time.Time          `json:"completionDate"`
	Disabled           *PersonTaskDisabled `json:"disabled"`
	UnacceptedPolicies []PersonTaskItem    `json:"unacceptedPolicies"`
	AcceptedPolicies   []PersonTaskItem    `json:"acceptedPolicies"`
}

// PersonCompleteCustomTasksTask holds the complete custom tasks task detail.
type PersonCompleteCustomTasksTask struct {
	TaskType              string              `json:"taskType"`
	Status                string              `json:"status"`
	DueDate               *time.Time          `json:"dueDate"`
	CompletionDate        *time.Time          `json:"completionDate"`
	Disabled              *PersonTaskDisabled `json:"disabled"`
	IncompleteCustomTasks []PersonTaskItem    `json:"incompleteCustomTasks"`
	CompletedCustomTasks  []PersonTaskItem    `json:"completedCustomTasks"`
}

// PersonCompleteOffboardingCustomTasksTask holds the offboarding custom tasks detail.
type PersonCompleteOffboardingCustomTasksTask struct {
	TaskType                         string              `json:"taskType"`
	Status                           string              `json:"status"`
	DueDate                          *time.Time          `json:"dueDate"`
	CompletionDate                   *time.Time          `json:"completionDate"`
	Disabled                         *PersonTaskDisabled `json:"disabled"`
	IncompleteCustomOffboardingTasks []PersonTaskItem    `json:"incompleteCustomOffboardingTasks"`
	CompletedCustomOffboardingTasks  []PersonTaskItem    `json:"completedCustomOffboardingTasks"`
}

// PersonInstallDeviceMonitoringTask holds the install device monitoring task detail.
type PersonInstallDeviceMonitoringTask struct {
	TaskType       string              `json:"taskType"`
	Status         string              `json:"status"`
	DueDate        *time.Time          `json:"dueDate"`
	CompletionDate *time.Time          `json:"completionDate"`
	Disabled       *PersonTaskDisabled `json:"disabled"`
}

// PersonCompleteBackgroundChecksTask holds the complete background checks task detail.
type PersonCompleteBackgroundChecksTask struct {
	TaskType       string              `json:"taskType"`
	Status         string              `json:"status"`
	DueDate        *time.Time          `json:"dueDate"`
	CompletionDate *time.Time          `json:"completionDate"`
	Disabled       *PersonTaskDisabled `json:"disabled"`
}

// PersonTaskDetails holds the detailed breakdown of a person's tasks.
type PersonTaskDetails struct {
	CompleteTrainings              *PersonCompleteTrainingsTask              `json:"completeTrainings"`
	AcceptPolicies                 *PersonAcceptPoliciesTask                 `json:"acceptPolicies"`
	CompleteCustomTasks            *PersonCompleteCustomTasksTask            `json:"completeCustomTasks"`
	CompleteOffboardingCustomTasks *PersonCompleteOffboardingCustomTasksTask `json:"completeOffboardingCustomTasks"`
	InstallDeviceMonitoring        *PersonInstallDeviceMonitoringTask        `json:"installDeviceMonitoring"`
	CompleteBackgroundChecks       *PersonCompleteBackgroundChecksTask       `json:"completeBackgroundChecks"`
}

// PersonTasksSummary holds the tasks summary for a person.
type PersonTasksSummary struct {
	CompletionDate *time.Time         `json:"completionDate"`
	DueDate        *time.Time         `json:"dueDate"`
	Status         string             `json:"status"`
	Details        *PersonTaskDetails `json:"details"`
}

// Person represents a single person returned by the Vanta API.
type Person struct {
	ID           string              `json:"id"`
	EmailAddress string              `json:"emailAddress"`
	Employment   *PersonEmployment   `json:"employment"`
	LeaveInfo    *PersonTasksSummary `json:"leaveInfo"`
	GroupIDs     []string            `json:"groupIds"`
	Name         PersonName          `json:"name"`
	Sources      *PersonSources      `json:"sources"`
	TasksSummary *PersonTasksSummary `json:"tasksSummary"`
}

// People fetches all people from the Vanta API, following pagination.
func (c *Client) People() ([]Person, error) {
	return fetchAll[Person](c, "/people")
}

var personHeaders = []string{
	"id", "emailAddress", "displayName", "jobTitle", "employmentStatus",
	"employmentStartDate", "taskStatus", "taskDueDate",
}

func personRow(p Person) []string {
	jobTitle := ""
	employmentStatus := ""
	employmentStartDate := ""
	if p.Employment != nil {
		jobTitle = p.Employment.JobTitle
		employmentStatus = p.Employment.Status
		employmentStartDate = formatDate(p.Employment.StartDate)
	}
	taskStatus := ""
	taskDueDate := ""
	if p.TasksSummary != nil {
		taskStatus = p.TasksSummary.Status
		taskDueDate = formatDate(p.TasksSummary.DueDate)
	}
	return []string{
		p.ID,
		p.EmailAddress,
		p.Name.Display,
		jobTitle,
		employmentStatus,
		employmentStartDate,
		taskStatus,
		taskDueDate,
	}
}

func printPeopleCSV(people []Person) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(personHeaders); err != nil {
		return err
	}
	for _, p := range people {
		if err := w.Write(personRow(p)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printPeopleTSV(people []Person) error {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	if err := w.Write(personHeaders); err != nil {
		return err
	}
	for _, p := range people {
		if err := w.Write(personRow(p)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printPeopleMarkdown(people []Person) {
	fmt.Println("| " + strings.Join(personHeaders, " | ") + " |")
	fmt.Println("|" + strings.Repeat(" --- |", len(personHeaders)))
	for _, p := range people {
		row := personRow(p)
		fmt.Println("| " + strings.Join(row, " | ") + " |")
	}

}
func printPeoplePrettyMarkdown(people []Person) {
	rows := make([][]string, len(people))
	for i, item := range people {
		rows[i] = personRow(item)
	}
	printPrettyMarkdown(personHeaders, rows)
}

func newPeopleCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "people",
		Short: "Fetch people",
		Run: func(cmd *cobra.Command, args []string) {
			people, err := newClient().People()
			if err != nil {
				log.Fatal(err)
			}
			switch format {
			case "json":
				if err := printJSON(people); err != nil {
					log.Fatal(err)
				}
			case "csv":
				if err := printPeopleCSV(people); err != nil {
					log.Fatal(err)
				}
			case "tsv":
				if err := printPeopleTSV(people); err != nil {
					log.Fatal(err)
				}
			case "markdown":
				printPeopleMarkdown(people)
			case "pretty_markdown":
				printPeoplePrettyMarkdown(people)
			default:
				log.Fatalf("unknown format %q: must be json, csv, tsv, markdown, or pretty_markdown", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, tsv, markdown, pretty_markdown")
	return cmd
}
