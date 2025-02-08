package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	Ref string `json:"ref"`
	RefType string `json:"ref_type"`
	Commits []ApiCommit `json:"commits"`
	Forkee ApiForkee `json:"forkee"`
	Action string `json:"action"`
	Member string `json:"member"`
	Number int `json:"number"`
}

type ApiEvent struct {
	ID string `json:"id"`
	Type string `json:"type"`
	Repo ApiRepo `json:"repo"`
	Payload ApiPayload `json:"payload"`
	CreatedAt string `json:"created_at"`
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		os.Exit(1)
	}

	username := flag.Arg(0)

	events, err := getEvents(username)
	if err != nil {
		log.Fatal(err)
	}

	activities, err := formatEvent(events)
	if err != nil {
		log.Fatal(err)
	}

	printEvents(activities)
}

func formatEvent(events []ApiEvent) ([]string, error) {
	activities := []string{}
	for i := 0; i < len(events); i++ {
		evt := events[i]
		act := ""
		if evt.Type == "PushEvent" {
			act = fmt.Sprintf("Pushed %d commits to %s", len(evt.Payload.Commits), evt.Repo.Name)
		} else if evt.Type == "CreateEvent" {
			if evt.Payload.RefType == "repository" {
				act = fmt.Sprintf("Created a new repository called %s", evt.Repo.Name)
			} else if evt.Payload.RefType == "branch" {
				act = fmt.Sprintf("Created a new branch %s in %s", evt.Payload.Ref, evt.Repo.Name)
			} else {
				act = fmt.Sprintf("Created a new tag %s in %s", evt.Payload.Ref, evt.Repo.Name)
			}
		} else if evt.Type == "DeleteEvent" {
			if evt.Payload.RefType == "branch" {
				act = fmt.Sprintf("Created a new branch %s in %s", evt.Payload.Ref, evt.Repo.Name)
			} else {
				act = fmt.Sprintf("Created a new tag %s in %s", evt.Payload.Ref, evt.Repo.Name)
			}
		} else if evt.Type == "ForkEvent" {
			act = fmt.Sprintf("Forked repository to %s", evt.Payload.Forkee.FullName)
		} else if evt.Type == "GollumEvent" {
			act = fmt.Sprintf("Created page in wiki to %s", evt.Repo.Name)
		} else if evt.Type == "IssueCommentEvent" {
			if evt.Payload.Action == "created" {
				act = fmt.Sprintf("Created a new comment in %s", evt.Repo.Name)
			} else if evt.Payload.Action == "edited" {
				act = fmt.Sprintf("Edited a comment in %s", evt.Repo.Name)
			} else {
				act = fmt.Sprintf("Deleted a comment in %s", evt.Repo.Name)
			}
		} else if evt.Type == "IssuesEvent" {
			if evt.Payload.Action == "opened" {
				act = fmt.Sprintf("Opened a new issue in %s", evt.Repo.Name)
			} else if evt.Payload.Action == "edited" {
				act = fmt.Sprintf("Edited a issue in %s", evt.Repo.Name)
			} else if evt.Payload.Action == "closed" {
				act = fmt.Sprintf("Closed a issue in %s", evt.Repo.Name)
			} else if evt.Payload.Action == "reopened" {
				act = fmt.Sprintf("Reopened a issue in %s", evt.Repo.Name)
			} else if evt.Payload.Action == "assigned" {
				act = fmt.Sprintf("Assigned a issue in %s", evt.Repo.Name)
			} else if evt.Payload.Action == "unassigned" {
				act = fmt.Sprintf("Unassigned a issue in %s", evt.Repo.Name)
			} else if evt.Payload.Action == "labeled" {
				act = fmt.Sprintf("Labeled a issue in %s", evt.Repo.Name)
			} else if evt.Payload.Action == "unlabeled" {
				act = fmt.Sprintf("Unlabeled a issue in %s", evt.Repo.Name)
			}
		} else if evt.Type == "MemberEvent" {
			act = fmt.Sprintf("Added a member %s to %s", evt.Payload.Member, evt.Repo.Name)
		} else if evt.Type == "PublicEvent" {
			act = fmt.Sprintf("The repository %s is public", evt.Repo.Name)
		} else if evt.Type == "PullRequestEvent" {
			if evt.Payload.Action == "opened" {
				act = fmt.Sprintf("Opened pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
			} else if evt.Payload.Action == "edited" {
				act = fmt.Sprintf("Edited pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
			} else if evt.Payload.Action == "closed" {
				act = fmt.Sprintf("Closed pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
			} else if evt.Payload.Action == "reopened" {
				act = fmt.Sprintf("Reopened pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
			} else if evt.Payload.Action == "assigned" {
				act = fmt.Sprintf("Assigned pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
			} else if evt.Payload.Action == "unassigned" {
				act = fmt.Sprintf("Unassigned pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
			} else if evt.Payload.Action == "review_requested" {
				act = fmt.Sprintf("Review requested pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
			} else if evt.Payload.Action == "review_request_removed" {
				act = fmt.Sprintf("Review request removed pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
			} else if evt.Payload.Action == "labeled" {
				act = fmt.Sprintf("Labeled pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
			} else if evt.Payload.Action == "unlabeled" {
				act = fmt.Sprintf("Unlabeled pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
			} else {
				act = fmt.Sprintf("Synchronized pull request #%d in %s", evt.Payload.Number, evt.Repo.Name)
			}
		} else if evt.Type == "PullRequestReviewEvent" {
			act = fmt.Sprintf("Created pull request review in %s", evt.Repo.Name)
		} else if evt.Type == "PullRequestReviewCommentEvent" {
			act = fmt.Sprintf("Created pull request review comment in %s", evt.Repo.Name)
		} else if evt.Type == "PullRequestReviewThreadEvent" {
			if evt.Payload.Action == "resolved" {
				act = fmt.Sprintf("Resolved pull request review thread in %s", evt.Repo.Name)
			} else {
				act = fmt.Sprintf("Unresolved pull request review thread in %s", evt.Repo.Name)
			}
		} else if evt.Type == "ReleaseEvent" {
			act = fmt.Sprintf("Published release in %s", evt.Repo.Name)
		} else if evt.Type == "SponsorshipEvent" {
			act = fmt.Sprintf("Created sponsorship in %s", evt.Repo.Name)
		} else if evt.Type == "WatchEvent" {
			act = fmt.Sprintf("Starred %s", evt.Repo.Name)
		}

		if len(act) > 0 {
			activities = append(activities, act)
		}
	}

	return activities, nil
}

func printEvents(activities []string) {
	for i := 0; i < len(activities); i++ {
		act := activities[i]

		fmt.Printf("- %s\n", act)
	}
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