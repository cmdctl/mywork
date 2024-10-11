package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
)

const titleLimit = 120

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

func truncateString(s string, length int) string {
	if len(s) > length {
		return s[:length] + "..."
	}
	return s
}

func display(cmd *exec.Cmd) {
	var workItems AzureWorkItems
	output, err := cmd.Output()
	if err != nil {
		log.Println("could not get command output")
		log.Fatal(err)
	}

	if len(output) == 0 {
		fmt.Println("No work items found")
		return
	}

	err = json.Unmarshal(output, &workItems)
	if err != nil {
		log.Println("could not unmarshal json")
		log.Println(string(output))
		log.Fatal(err)
	}

  // Sort by state rank
	sort.Slice(workItems, func(i, j int) bool {
    state1 := SystemState(workItems[i].Fields.SystemState )
    state2 := SystemState(workItems[j].Fields.SystemState )
		return state1.rank < state2.rank
	})

  // find space padding for the title column
	var longestTitle int
	for _, workItem := range workItems {
		title := truncateString(workItem.Fields.SystemTitle, titleLimit)
		if len(title) > longestTitle {
			longestTitle = len(title)
		}
	}

  // Print to console
	for _, workItem := range workItems {
		title := truncateString(workItem.Fields.SystemTitle, titleLimit)
		padding := longestTitle - len(title)
		fmt.Printf("%d | %s", workItem.Fields.SystemID, title)
		for i := 0; i < padding; i++ {
			fmt.Print(" ")
		}
    state := SystemState(workItem.Fields.SystemState)
		fmt.Printf(" | %s\n", state)
	}
}

type Rank uint
const (
  RankNew Rank = iota + 1
  RankTodo
  RankInProgress
  RankInReview
  RankReadyForTesting
  RankInTesting
  RankBusinessAcceptance
  RankReadyForDeployment
)

type State struct {
  name string
  rank Rank
}

func (s State) String() string {
  return s.name
}

func SystemState(azState string) State {
  state := &State{
    name: azState,
  }
  switch azState {
  case "New":
    state.rank = RankNew
  case "To Do":
    state.rank = RankTodo
  case "In Progress":
    state.rank = RankInProgress
  case "In Review":
    state.rank = RankInReview
  case "Ready For Testing":
    state.rank = RankReadyForTesting
  case "In Testing":
    state.rank = RankInTesting
  case "Business Acceptance":
    state.rank = RankBusinessAcceptance
  case "Ready For Deployment":
    state.rank = RankReadyForDeployment
  }
  return *state
}

func ListWorkItems(name string) {
	const query = `az boards query --wiql "SELECT [System.Id], [System.Title], [System.State] FROM WorkItems WHERE [System.AssignedTo] = '%s' AND [System.State] <> 'Done' AND [System.State] <> 'Resolved' AND [System.State] <> 'Closed' AND [System.State] <> 'Removed' AND [System.State] <> 'Design'"`
	cmd := exec.Command("bash", "-c", fmt.Sprintf(query, name))
	display(cmd)
}

func ListMyWorkItems() {
	const query = `az boards query --wiql "SELECT [System.Id], [System.Title], [System.State] FROM WorkItems WHERE [System.AssignedTo] = @Me AND [System.State] <> 'Done' AND [System.State] <> 'Resolved' AND [System.State] <> 'Closed' AND [System.State] <> 'Removed' AND [System.State] <> 'Design'"`
	cmd := exec.Command("bash", "-c", query)
	display(cmd)
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
					name := c.Args().Slice()
					ListWorkItems(strings.Join(name, " "))
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
