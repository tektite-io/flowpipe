package handler

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
)

type PipelineQueued EventHandler

func (h PipelineQueued) HandlerName() string {
	return "handler.pipeline_queued"
}

func (PipelineQueued) NewEvent() interface{} {
	return &event.PipelineQueued{}
}

func (h PipelineQueued) Handle(ctx context.Context, ei interface{}) error {
	e := ei.(*event.PipelineQueued)
	cmd, err := event.NewPipelineLoad(event.ForPipelineQueued(e))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineQueuedToPipelineFail(e, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}