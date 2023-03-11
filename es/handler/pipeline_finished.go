package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/execution"
)

type PipelineFinished EventHandler

func (h PipelineFinished) HandlerName() string {
	return "handler.pipeline_finished"
}

func (PipelineFinished) NewEvent() interface{} {
	return &event.PipelineFinished{}
}

func (h PipelineFinished) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineFinished)

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	parentStepExecution, err := ex.ParentStepExecution(e.PipelineExecutionID)
	if err != nil {
		return err
	}
	if parentStepExecution != nil {
		cmd, err := event.NewPipelineStepFinish(
			event.ForPipelineFinished(e),
			event.WithPipelineExecutionID(parentStepExecution.PipelineExecutionID),
			event.WithStepExecutionID(parentStepExecution.ID))
		if err != nil {
			return err
		}
		return h.CommandBus.Send(ctx, &cmd)
	} else {
		// Dump the final execution state
		jsonStr, _ := json.MarshalIndent(ex, "", "  ")
		fmt.Println(string(jsonStr))

		// Dump step outputs
		stepOutputs, err := ex.PipelineStepOutputs(e.PipelineExecutionID)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(stepOutputs)
		}
	}

	return nil
}
