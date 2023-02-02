package state

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type EventLogEntry struct {
	EventType string          `json:"event_type"`
	Timestamp *time.Time      `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

type StackEntry struct {
	PipelineName string `json:"pipeline_name"`
	StepIndex    int    `json:"step_index"`
}

type Stack map[string]StackEntry

type State struct {
	IdentityID   string                 `json:"identity_id"`
	WorkspaceID  string                 `json:"workspace_id"`
	PipelineName string                 `json:"pipeline_name"`
	Input        map[string]interface{} `json:"input"`
	RunID        string                 `json:"run_id"`
	Stack        Stack                  `json:"stack"`
}

func NewState(runID string) (*State, error) {
	s := &State{}
	s.Stack = Stack{}
	err := s.Load(runID)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *State) Load(runID string) error {

	logFile := fmt.Sprintf("logs/%s.jsonl", runID)

	// Open the event log
	f, err := os.Open(logFile)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// TODO - by default this has a max line size of 64K, see https://stackoverflow.com/a/16615559
	for scanner.Scan() {

		ba := scanner.Bytes()

		// Get the run ID from the payload
		var e EventLogEntry
		err := json.Unmarshal(ba, &e)
		if err != nil {
			return err
		}

		switch e.EventType {

		case "event.Queue":
			// Get the run ID from the payload
			var queue event.Queue
			err := json.Unmarshal(e.Payload, &queue)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			s.IdentityID = queue.IdentityID
			s.WorkspaceID = queue.WorkspaceID
			s.PipelineName = queue.PipelineName
			s.Input = queue.Input
			s.RunID = queue.RunID

		case "event.Execute":
			var execute event.Execute
			err := json.Unmarshal(e.Payload, &execute)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			s.Stack[execute.StackID] = StackEntry{PipelineName: execute.PipelineName, StepIndex: execute.StepIndex}

		default:
			// Ignore unknown types while loading
		}

	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if s.RunID == "" {
		return errors.New(fmt.Sprintf("load_failed: %s", logFile))
	}

	return nil

}
