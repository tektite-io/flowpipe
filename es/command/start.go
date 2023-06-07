package command

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

type StartHandler CommandHandler

func (h StartHandler) HandlerName() string {
	return "command.start"
}

func (h StartHandler) NewCommand() interface{} {
	return &event.Start{}
}

func (h StartHandler) Handle(ctx context.Context, c interface{}) error {

	cmd, ok := c.(*event.Start)
	if !ok {
		fplog.Logger(ctx).Error("invalid command type", "expected", "*event.Start", "actual", c)
		return fperr.BadRequestWithMessage("invalid command type expected *event.Start")
	}

	fplog.Logger(ctx).Info("(13) start command handler", "executionID", cmd.Event.ExecutionID)

	e := event.Started{
		Event: event.NewFlowEvent(cmd.Event),
	}

	return h.EventBus.Publish(ctx, &e)
}
