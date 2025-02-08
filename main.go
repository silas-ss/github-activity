package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

type ApiRepo struct {
	Name string `json:"name"`
}

type ApiCommit struct {
	Message string `json:"message"`
}

type ApiForkee struct {
	FullName string `json:"full_name"`
}

type ApiPayload struct {
	Ref     string      `json:"ref"`
	RefType string      `json:"ref_type"`
	Commits []ApiCommit `json:"commits"`
	Forkee  ApiForkee   `json:"forkee"`
	Action  string      `json:"action"`
	Member  string      `json:"member"`
	Number  int         `json:"number"`
}

type ApiEvent struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Repo      ApiRepo    `json:"repo"`
	Payload   ApiPayload `json:"payload"`
	CreatedAt string     `json:"created_at"`
}

type Activity struct {
	Event   string `json:"event"`
	Message string `json:"message"`
}

func main() {
	var eventType string
	flag.StringVar(&eventType, "event", "", "Event type to filter")
	var formatOutput string
	flag.StringVar(&formatOutput, "formatOutput", "", "Format type to output")

	flag.Parse()

	if flag.NArg() == 0 {
		os.Exit(1)
	}

	username := flag.Arg(0)

	events, err := getEvents(username)
	if err != nil {
		log.Fatal(err)
	}

	if len(eventType) > 0 {
		events, err = filterEvent(events, eventType)
		if err != nil {
			log.Fatalf("failed on filter events. error: %s", err)
		}
	}

	activities, err := formatActivity(events)
	if err != nil {
		log.Fatal(err)
	}

	printEvents(activities, formatOutput)
}

func filterEvent(events []ApiEvent, evt string) ([]ApiEvent, error) {
	eventsFiltered := []ApiEvent{}
	for i := 0; i < len(events); i++ {
		e := events[i]
		if e.Type == evt {
			eventsFiltered = append(eventsFiltered, e)
		}
	}

	return eventsFiltered, nil
}

func formatActivity(events []ApiEvent) ([]Activity, error) {
	activities := []Activity{}
	for i := 0; i < len(events); i++ {
		evt := events[i]
		act, err := formatEvent(evt)
		if err != nil {
			return activities, err
		}
		activities = append(activities, act)
	}
	return activities, nil
}

func formatEvent(evt ApiEvent) (Activity, error) {
	act := Activity{Event: evt.Type}

	message := ""
	if evt.Type == "PushEvent" {
		message = fmt.Sprintf("Pushed %d commits to %s", len(evt.Payload.Commits), evt.Repo.Name)
	} else if evt.Type == "CreateEvent" {
		if evt.Payload.RefType == "repository" {
			message = fmt.Sprintf("Created a new repository called %s", evt.Repo.Name)
		} else if evt.Payload.RefType == "branch" {
			message = fmt.Sprintf("Created a new branch %s in %s", evt.Payload.Ref, evt.Repo.Name)
		} else {
			message = fmt.Sprintf("Created a new tag %s in %s", evt.Payload.Ref, evt.Repo.Name)
		}
	} else if evt.Type == "DeleteEvent" {
		if evt.Payload.RefType == "branch" {
			message = fmt.Sprintf("Deleted branch %s in %s", evt.Payload.Ref, evt.Repo.Name)
		} else {
			message = fmt.Sprintf("Deleted tag %s in %s", evt.Payload.Ref, evt.Repo.Name)
		}
	} else if evt.Type == "ForkEvent" {
		message = fmt.Sprintf("Forked repository to %s", evt.Payload.Forkee.FullName)
	} else if evt.Type == "GollumEvent" {
		message = fmt.Sprintf("Created page in wiki to %s", evt.Repo.Name)
	} else if evt.Type == "IssueCommentEvent" {
		if evt.Payload.Action == "created" {
			message = fmt.Sprintf("Created a new comment in %s", evt.Repo.Name)
		} else if evt.Payload.Action == "edited" {
			message = fmt.Sprintf("Edited a comment in %s", evt.Repo.Name)
		} else {
			message = fmt.Sprintf("Deleted a comment in %s", evt.Repo.Name)
		}
	} else if evt.Type == "IssuesEvent" {
		if evt.Payload.Action == "opened" {
			message = fmt.Sprintf("Opened a new issue in %s", evt.Repo.Name)
		} else if evt.Payload.Action == "edited" {
			message = fmt.Sprintf("Edited a issue in %s", evt.Repo.Name)
		} else if evt.Payload.Action == "closed" {
			message = fmt.Sprintf("Closed a issue in %s", evt.Repo.Name)
		} else if evt.Payload.Action == "reopened" {
			message = fmt.Sprintf("Reopened a issue in %s", evt.Repo.Name)
		} else if evt.Payload.Action == "assigned" {
			message = fmt.Sprintf("Assigned a issue in %s", evt.Repo.Name)
		} else if evt.Payload.Action == "unassigned" {
			message = fmt.Sprintf("Unassigned a issue in %s", evt.Repo.Name)
		} else if evt.Payload.Action == "labeled" {
			message = fmt.Sprintf("Labeled a issue in %s", evt.Repo.Name)
		} else if evt.Payload.Action == "unlabeled" {
			message = fmt.Sprintf("Unlabeled a issue in %s", evt.Repo.Name)
		}
	} else if evt.Type == "MemberEvent" {
		message = fmt.Sprintf("Added a member %s to %s", evt.Payload.Member, evt.Repo.Name)
	} else if evt.Type == "PublicEvent" {
		message = fmt.Sprintf("The repository %s is public", evt.Repo.Name)
	} else if evt.Type == "PullRequestEvent" {
		if evt.Payload.Action == "opened" {
			message = fmt.Sprintf("Opened pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
		} else if evt.Payload.Action == "edited" {
			message = fmt.Sprintf("Edited pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
		} else if evt.Payload.Action == "closed" {
			message = fmt.Sprintf("Closed pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
		} else if evt.Payload.Action == "reopened" {
			message = fmt.Sprintf("Reopened pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
		} else if evt.Payload.Action == "assigned" {
			message = fmt.Sprintf("Assigned pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
		} else if evt.Payload.Action == "unassigned" {
			message = fmt.Sprintf("Unassigned pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
		} else if evt.Payload.Action == "review_requested" {
			message = fmt.Sprintf("Review requested pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
		} else if evt.Payload.Action == "review_request_removed" {
			message = fmt.Sprintf("Review request removed pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
		} else if evt.Payload.Action == "labeled" {
			message = fmt.Sprintf("Labeled pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
		} else if evt.Payload.Action == "unlabeled" {
			message = fmt.Sprintf("Unlabeled pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
		} else {
			message = fmt.Sprintf("Synchronized pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
		}
	} else if evt.Type == "PullRequestReviewEvent" {
		message = fmt.Sprintf("Created pull request review in %s", evt.Repo.Name)
	} else if evt.Type == "PullRequestReviewCommentEvent" {
		message = fmt.Sprintf("Created pull request review comment in %s", evt.Repo.Name)
	} else if evt.Type == "PullRequestReviewThreadEvent" {
		if evt.Payload.Action == "resolved" {
			message = fmt.Sprintf("Resolved pull request review thread in %s", evt.Repo.Name)
		} else {
			message = fmt.Sprintf("Unresolved pull request review thread in %s", evt.Repo.Name)
		}
	} else if evt.Type == "ReleaseEvent" {
		message = fmt.Sprintf("Published release in %s", evt.Repo.Name)
	} else if evt.Type == "SponsorshipEvent" {
		message = fmt.Sprintf("Created sponsorship in %s", evt.Repo.Name)
	} else if evt.Type == "WatchEvent" {
		message = fmt.Sprintf("Starred %s", evt.Repo.Name)
	}

	act.Message = message

	return act, nil
}

func printEvents(activities []Activity, formatOutput string) {
	if formatOutput == "json" {
		printPrettyStruct(activities)
	} else if formatOutput == "table" {
		printAsTable(activities)
	} else {
		printAsRows(activities)
	}

}

func printAsRows(activities []Activity) {
	for i := 0; i < len(activities); i++ {
		act := activities[i]

		fmt.Printf("- %s\n", act.Message)
	}
}

func printPrettyStruct(activities []Activity) {
	val, err := json.MarshalIndent(activities, "", "    ")
	if err != nil {
		log.Fatalf("failed to marshal. error: %s", err)
	}
	fmt.Println(val)
}

func printAsTable(actitivies []Activity) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Event", "Message"})
	for i := 0; i < len(actitivies); i++ {
		act := actitivies[i]
		row := table.Row{
			i + 1,
			act.Event,
			act.Message,
		}
		t.AppendRow(row)
	}
	t.Render()
}

func getEvents(username string) ([]ApiEvent, error) {
	events := []ApiEvent{}

	url := fmt.Sprintf("https://api.github.com/users/%s/events", username)

	resp, err := http.Get(url)
	if err != nil {
		return events, fmt.Errorf("failed on get events. error: %s", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return events, fmt.Errorf("user not found")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return events, fmt.Errorf("failed on read bufer. error: %s", err)
	}

	if err := json.Unmarshal(data, &events); err != nil {
		return events, fmt.Errorf("failed on unmarshal body. error: %s", err)
	}

	return events, nil
}
