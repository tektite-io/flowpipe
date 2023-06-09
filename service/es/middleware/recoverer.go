package middleware

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

// Recover from Go panic middleware. Based on Watermill Recoverer middleware.
// The panic will be wrapped in a Flowpipe Error and set as a fatal error (non-retryable, this is the default setting).
type Recoverer struct {
	Ctx context.Context
}

// Holds the recovered panic's error along with the stacktrace.
type RecoveredPanicError struct {
	V          interface{}
	Stacktrace string
}

func (p RecoveredPanicError) Error() string {
	return fmt.Sprintf("panic occurred: %#v, stacktrace: \n%s", p.V, p.Stacktrace)
}

func (rec Recoverer) Middleware(h message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) (messages []*message.Message, err error) {
		panicked := true

		defer func() {
			if r := recover(); r != nil || panicked {
				logger := fplog.Logger(rec.Ctx)
				logger.Error("Recovered from panic", "error", err)
				recoveredPanicErr := RecoveredPanicError{V: r, Stacktrace: string(debug.Stack())}

				// Flowpipe error by default is not retryable
				internalErr := fperr.Internal(recoveredPanicErr)
				err = internalErr
			}
		}()

		messages, err = h(msg)
		panicked = false
		return messages, err
	}
}