package event

import (
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/flowpipe/internal/util"
)

// PipelineQueue commands a pipeline to be queued for execution.
type PipelineQueue struct {
	Event *Event          `json:"event"`
	Name string          `json:"name"`
	Args resources.Input `json:"args"`

	// The name of the mod including its version number. May be blank if not required,
	// for example top level mod or 1st level children. Since the 1st level children must have
	// unique names, we don't need ModFullVersion
	ModFullVersion string `json:"mod_full_version"`

	// Pipeline execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`

	// If this is a child pipeline then set the parent pipeline execution ID
	ParentStepExecutionID string `json:"parent_step_execution_id,omitempty"`
	ParentExecutionID     string `json:"parent_execution_id,omitempty"`

	// If pipeline is triggered by a trigger, this is the trigger name
	Trigger string

	// If pipeline is triggered by query trigger, this is the capture name
	TriggerCapture string
}

func (e *PipelineQueue) GetEvent() *Event {
	return e.Event
}

func (e *PipelineQueue) HandlerName() string {
	return CommandPipelineQueue
}

func (e *PipelineQueue) GetName() string {
	return e.Name
}

func (e *PipelineQueue) GetType() string {
	return "pipeline"
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineQueueOption func(*PipelineQueue) error

// NewPipelineQueue creates a new PipelineQueue event.
func NewPipelineQueue(opts ...PipelineQueueOption) (*PipelineQueue, error) {
	// Defaults
	e := &PipelineQueue{
		PipelineExecutionID: util.NewPipelineExecutionId(),
	}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

// ForPipelineQueue returns a PipelineQueueOption that sets the fields of the
// PipelineQueue event from a PipelineQueue command.
func ForPipelineStepStartedToPipelineQueue(e *StepPipelineStarted) PipelineQueueOption {
	return func(cmd *PipelineQueue) error {
		cmd.Event = NewChildEvent(e.Event)
		cmd.PipelineExecutionID = e.ChildPipelineExecutionID

		cmd.ParentStepExecutionID = e.StepExecutionID
		cmd.ParentExecutionID = e.PipelineExecutionID

		cmd.Name = e.ChildPipelineName
		cmd.Args = e.ChildPipelineArgs
		cmd.ModFullVersion = e.ChildPipelineModFullVersion
		return nil
	}
}
