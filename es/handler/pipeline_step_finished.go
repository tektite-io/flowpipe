package handler

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
)

type PipelineStepFinished EventHandler

func (h PipelineStepFinished) HandlerName() string {
	return "handler.pipeline_step_finished"
}

func (PipelineStepFinished) NewEvent() interface{} {
	return &event.PipelineStepFinished{}
}

func (h PipelineStepFinished) Handle(ctx context.Context, ei interface{}) error {
	e := ei.(*event.PipelineStepFinished)
	cmd, err := event.NewPipelinePlan(event.ForPipelineStepFinished(e))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(e, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}