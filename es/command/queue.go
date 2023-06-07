package command

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

type QueueHandler CommandHandler

func (h QueueHandler) HandlerName() string {
	return "command.queue"
}

func (h QueueHandler) NewCommand() interface{} {
	return &event.Queue{}
}

func (h QueueHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.Queue)
	if !ok {
		fplog.Logger(ctx).Error("invalid command type", "expected", "*event.Queue", "actual", c)
		return fperr.BadRequestWithMessage("invalid command type expected *event.Queue")
	}

	fplog.Logger(ctx).Info("(12) queue command handler", "executionID", cmd.Event.ExecutionID, "cmd", cmd)

	e := event.Queued{
		Event: event.NewFlowEvent(cmd.Event),
	}

	return h.EventBus.Publish(ctx, &e)
}
