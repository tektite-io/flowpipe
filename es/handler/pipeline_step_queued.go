package handler

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

type PipelineStepQueued EventHandler

func (h PipelineStepQueued) HandlerName() string {
	return "handler.pipeline_step_queued"
}

func (PipelineStepQueued) NewEvent() interface{} {
	return &event.PipelineStepQueued{}
}

func (h PipelineStepQueued) Handle(ctx context.Context, ei interface{}) error {

	logger := fplog.Logger(ctx)
	e, ok := ei.(*event.PipelineStepQueued)

	if !ok {
		logger.Error("invalid event type", "expected", "*event.PipelineStepQueued", "actual", ei)
		return fperr.BadRequestWithMessage("invalid event type expected *event.PipelineStepQueued")
	}

	logger.Info("[10] pipeline_step_queued event handler", "executionID", e.Event.ExecutionID, "e", e)

	// Step has been queued (but not yet started), so here we just need to start the step
	// the code should be the same as the pipeline_planned event handler
	cmd, err := event.NewPipelineStepStart(event.ForPipelineStepQueued(e), event.WithStep(e.StepName, e.StepInput, e.ForEach))
	if err != nil {
		err := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepQueuedToPipelineFail(e, err)))
		if err != nil {
			fplog.Logger(ctx).Error("Error publishing event", "error", err)
		}

		return nil
	}

	logger.Info("[10] pipeline_step_queued event handler #3.A - sending pipeline step START command", "command", cmd)
	if err := h.CommandBus.Send(ctx, &cmd); err != nil {
		err := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepQueuedToPipelineFail(e, err)))
		if err != nil {
			fplog.Logger(ctx).Error("Error publishing event", "error", err)
		}
		return nil
	}

	return nil
}