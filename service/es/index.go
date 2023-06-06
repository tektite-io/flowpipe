package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/garsue/watermillzap"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/es/command"
	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/es/handler"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
	"github.com/turbot/flowpipe/pipeline"
	"github.com/turbot/flowpipe/util"

	//nolint:depguard // TODO temporary to get things going
	"go.uber.org/zap"
)

type ESService struct {
	Ctx context.Context

	RunID      string
	CommandBus *cqrs.CommandBus

	Status    string     `json:"status"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`
}

func NewESService(ctx context.Context) (*ESService, error) {
	// Defaults
	es := &ESService{
		Ctx:    ctx,
		Status: "initialized",
	}
	return es, nil
}

func (es *ESService) Start() error {
	// Convenience
	logger := fplog.Logger(es.Ctx)

	logger.Debug("ES starting")
	defer logger.Debug("ES started")

	pipelineDir := viper.GetString("pipeline.dir")

	logger.Debug("Pipeline dir", "dir", pipelineDir)

	_, err := pipeline.LoadPipelines(es.Ctx, pipelineDir)
	if err != nil {
		return err
	}

	cqrsMarshaler := cqrs.JSONMarshaler{}

	goChannelConfig := gochannel.Config{
		// TODO - I really don't understand this and I'm not sure it's necessary.
		//OutputChannelBuffer: 10000,
		//Persistent:          true,
	}
	wLogger := watermillzap.NewLogger(logger.Zap)

	commandsPubSub := gochannel.NewGoChannel(goChannelConfig, wLogger)
	eventsPubSub := gochannel.NewGoChannel(goChannelConfig, wLogger)

	// CQRS is built on messages router. Detailed documentation: https://watermill.io/docs/messages-router/
	router, err := message.NewRouter(message.RouterConfig{}, wLogger)
	if err != nil {
		return err
	}

	// Simple middleware which will recover panics from event or command handlers.
	// More about router middlewares you can find in the documentation:
	// https://watermill.io/docs/messages-router/#middleware
	//
	// List of available middlewares you can find in message/router/middleware.
	//router.AddMiddleware(middleware.RandomFail(0.5))
	//router.AddMiddleware(middleware.Recoverer)

	// Log to file for creation of state
	router.AddMiddleware(LogEventMiddlewareWithContext(es.Ctx))

	// Dump the state of the event sourcing log with every event
	//router.AddMiddleware(DumpState(ctx))

	// Throttle, if required
	//router.AddMiddleware(middleware.NewThrottle(4, time.Second).Middleware)

	// Retry, if required
	/*
		retry := middleware.Retry{
			MaxRetries: 3,
		}
		router.AddMiddleware(retry.Middleware)
	*/

	// cqrs.Facade is facade for Command and Event buses and processors.
	// You can use facade, or create buses and processors manually (you can inspire with cqrs.NewFacade)
	cqrsFacade, err := cqrs.NewFacade(cqrs.FacadeConfig{
		GenerateCommandsTopic: func(commandName string) string {
			// we are using queue RabbitMQ config, so we need to have topic per command type
			return commandName
		},
		CommandHandlers: func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.CommandHandler {
			return []cqrs.CommandHandler{
				command.PipelineCancelHandler{EventBus: eb},
				command.PipelineFailHandler{EventBus: eb},
				command.PipelineFinishHandler{EventBus: eb},
				command.PipelineLoadHandler{EventBus: eb},
				command.PipelinePlanHandler{EventBus: eb},
				command.PipelineQueueHandler{EventBus: eb},
				command.PipelineStartHandler{EventBus: eb},
				command.PipelineStepFinishHandler{EventBus: eb},
				command.PipelineStepStartHandler{EventBus: eb},
				command.QueueHandler{EventBus: eb},
				command.StartHandler{EventBus: eb},
				command.StopHandler{EventBus: eb},
			}
		},
		CommandsPublisher: commandsPubSub,
		CommandsSubscriberConstructor: func(handlerName string) (message.Subscriber, error) {
			// we can reuse subscriber, because all commands have separated topics
			return commandsPubSub, nil
		},
		GenerateEventsTopic: func(eventName string) string {
			return eventName
		},
		EventHandlers: func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.EventHandler {
			return []cqrs.EventHandler{
				handler.Failed{CommandBus: cb},
				handler.Loaded{CommandBus: cb},
				handler.PipelineCanceled{CommandBus: cb},
				handler.PipelineFailed{CommandBus: cb},
				handler.PipelineFinished{CommandBus: cb},
				handler.PipelineLoaded{CommandBus: cb},
				handler.PipelinePlanned{CommandBus: cb},
				handler.PipelineQueued{CommandBus: cb},
				handler.PipelineStarted{CommandBus: cb},
				handler.PipelineStepFinished{CommandBus: cb},
				handler.PipelineStepStarted{CommandBus: cb},
				handler.Queued{CommandBus: cb},
				handler.Started{CommandBus: cb},
				handler.Stopped{CommandBus: cb},
			}
		},
		EventsPublisher: eventsPubSub,
		EventsSubscriberConstructor: func(handlerName string) (message.Subscriber, error) {
			// we can reuse subscriber, because all commands have separated topics
			return eventsPubSub, nil
		},
		/*
			EventsSubscriberConstructor: func(handlerName string) (message.Subscriber, error) {
				config := amqp.NewDurablePubSubConfig(
					amqpAddress,
					amqp.GenerateQueueNameTopicNameWithSuffix(handlerName),
				)
				return amqp.NewSubscriber(config, logger)
			},
		*/
		Router:                router,
		CommandEventMarshaler: cqrsMarshaler,
		Logger:                wLogger,
	})
	if err != nil {
		panic(err)
	}

	if cqrsFacade == nil {
		panic(fperr.InternalWithMessage("cqrsFacade is nil"))
	}

	runID := util.NewProcessID()

	es.RunID = runID
	es.CommandBus = cqrsFacade.CommandBus()

	// processors are based on router, so they will work when router will start
	go func() {
		err := router.Run(es.Ctx)
		if err != nil {
			panic(err)
		}
	}()

	return nil
}

func LogEventMiddlewareWithContext(ctx context.Context) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {

			var pe event.PayloadWithEvent
			err := json.Unmarshal(msg.Payload, &pe)
			if err != nil {
				panic("TODO - invalid log payload, log me?")
			}

			executionID := pe.Event.ExecutionID
			if executionID == "" {
				m := fmt.Sprintf("SHOULD NOT HAPPEN - No execution_id found in payload: %s", msg.Payload)
				return nil, errors.New(m)
			}

			var payload map[string]interface{}
			err = json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				panic("TODO - invalid log payload, log me?")
			}

			payloadWithoutEvent := make(map[string]interface{})
			for key, value := range payload {
				if key == "event" {
					continue
				}
				payloadWithoutEvent[key] = value
			}
			//nolint:forbidigo // TODO temporary to get things going
			fmt.Printf("%s %-30s %s\n", pe.Event.CreatedAt.Format("15:04:05.000"), message.HandlerNameFromCtx(msg.Context()), payloadWithoutEvent)

			logger := fplog.ExecutionLogger(ctx, executionID)
			defer logger.Sync() //nolint:errcheck // TODO temporary to get things going
			logger.Info("es",
				// Structured context as strongly typed Field values.
				zap.String("event_type", message.HandlerNameFromCtx(msg.Context())),
				// zap adds ts field automatically, so don't need zap.Time("created_at", time.Now()),
				zap.Any("payload", payload),
			)

			return h(msg)

		}
	}
}