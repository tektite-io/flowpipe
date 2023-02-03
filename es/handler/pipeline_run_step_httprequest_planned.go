package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunStepHTTPRequestPlanned EventHandler

func (h PipelineRunStepHTTPRequestPlanned) HandlerName() string {
	return "pipeline_run_step_httprequest_planned"
}

func (PipelineRunStepHTTPRequestPlanned) NewEvent() interface{} {
	return &event.PipelineRunStepHTTPRequestPlanned{}
}

func (h PipelineRunStepHTTPRequestPlanned) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineRunStepHTTPRequestPlanned)

	fmt.Printf("[handler] %s: %v\n", h.HandlerName(), e)

	// We have another step to run
	cmd := &event.PipelineRunStepHTTPRequestExecute{
		SpanID: e.SpanID,
		Input:  e.Input,
	}

	return h.CommandBus.Send(ctx, cmd)
}
