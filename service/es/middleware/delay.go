package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

// Middleware to delay the PipelineStepStart command execution (for backoff purpose)
//
// TODO: make it generic?
func PipelineStepStartCommandDelayMiddlewareWithContext(ctx context.Context) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {

			logger := fplog.Logger(ctx)

			handlerName := message.HandlerNameFromCtx(msg.Context())
			if handlerName != "command.pipeline_step_start" {
				return h(msg)
			}

			var pe event.PayloadWithEvent
			err := json.Unmarshal(msg.Payload, &pe)
			if err != nil {
				logger.Error("invalid log payload", "error", err)
				return nil, err
			}

			executionID := pe.Event.ExecutionID
			if executionID == "" {
				return nil, fperr.InternalWithMessage("no execution_id found in payload")
			}

			var payload event.PipelineStepStart
			err = json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				logger.Error("invalid log payload", "error", err)
				return nil, err
			}

			logger.Info("CommandDelayMiddlewareWithContext", "handlerName", handlerName, "payload", payload)

			if payload.DelayMs == 0 {
				return h(msg)
			}

			fmt.Println("DDDD")
			fmt.Println("DDDD")
			fmt.Println("DDDD", payload.StepExecutionID)
			waitTime := time.Millisecond * time.Duration(payload.DelayMs)

			select {
			case <-ctx.Done():
				return h(msg)
			case <-time.After(waitTime):
				// go on
			}

			//time.Sleep(waitTime)
			fmt.Println("DDDD - END", payload.StepExecutionID)
			fmt.Println("DDDD - END")
			fmt.Println("DDDD - END")

			return h(msg)
		}
	}
}