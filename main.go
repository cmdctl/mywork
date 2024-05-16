package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

type AzureWorkItems []struct {
	CommentVersionRef CommentVersionRef `json:"commentVersionRef,omitempty"`
	Fields            Fields            `json:"fields"`
	ID                int               `json:"id"`
	Relations         any               `json:"relations"`
	Rev               int               `json:"rev"`
	URL               string            `json:"url"`
}
type CommentVersionRef struct {
	CommentID int    `json:"commentId"`
	URL       string `json:"url"`
	Version   int    `json:"version"`
}
type Fields struct {
	SystemID    int    `json:"System.Id"`
	SystemState string `json:"System.State"`
	SystemTitle string `json:"System.Title"`
}

func ListWorkItems(name string) {

	var workItems AzureWorkItems
	const query = `az boards query --wiql "SELECT [System.Id], [System.Title], [System.State] FROM WorkItems WHERE [System.AssignedTo] = '%s' AND [System.State] <> 'Done' AND [System.State] <> 'Resolved' AND [System.State] <> 'Closed' AND [System.State] <> 'Removed'"`
	cmd := exec.Command("bash", "-c", fmt.Sprintf(query, name))

	output, err := cmd.Output()
	if err != nil {
		log.Println("could not get command output")
		log.Fatal(err)
	}
	err = json.Unmarshal(output, &workItems)
	if err != nil {
		log.Println("could not unmarshal json")
		log.Fatal(err)
	}
	var longestTitle int
	for _, workItem := range workItems {
		if len(workItem.Fields.SystemTitle) > longestTitle {
			longestTitle = len(workItem.Fields.SystemTitle)
		}
	}

	for _, workItem := range workItems {
		padding := longestTitle - len(workItem.Fields.SystemTitle)
		fmt.Printf("%d | %s", workItem.Fields.SystemID, workItem.Fields.SystemTitle)
		for i := 0; i < padding; i++ {
			fmt.Print(" ")
		}
		fmt.Printf(" | %s\n", workItem.Fields.SystemState)
	}
}

func ListMyWorkItems() {
	var workItems AzureWorkItems
	const query = `az boards query --wiql "SELECT [System.Id], [System.Title], [System.State] FROM WorkItems WHERE [System.AssignedTo] = @Me AND [System.State] <> 'Done' AND [System.State] <> 'Resolved' AND [System.State] <> 'Closed' AND [System.State] <> 'Removed'"`
	cmd := exec.Command("bash", "-c", query)

	output, err := cmd.Output()
	if err != nil {
		log.Println("could not get command output")
		log.Fatal(err)
	}
	err = json.Unmarshal(output, &workItems)
	if err != nil {
		log.Println("could not unmarshal json")
		log.Fatal(err)
	}
	var longestTitle int
	for _, workItem := range workItems {
		if len(workItem.Fields.SystemTitle) > longestTitle {
			longestTitle = len(workItem.Fields.SystemTitle)
		}
	}

	for _, workItem := range workItems {
		padding := longestTitle - len(workItem.Fields.SystemTitle)
		fmt.Printf("%d | %s", workItem.Fields.SystemID, workItem.Fields.SystemTitle)
		for i := 0; i < padding; i++ {
			fmt.Print(" ")
		}
		fmt.Printf(" | %s\n", workItem.Fields.SystemState)
	}
}

func main() {

	app := &cli.App{
		Name:  "mywork",
		Usage: "List the work items assigned to me",
		Action: func(*cli.Context) error {
			ListMyWorkItems()
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "of",
				Usage: "List the work items assigned to a user",
				Action: func(c *cli.Context) error {
					name := c.Args().First()
					ListWorkItems(name)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
