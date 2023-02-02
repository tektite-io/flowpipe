package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type Failed EventHandler

func (h Failed) HandlerName() string {
	return "handler.failed"
}

func (Failed) NewEvent() interface{} {
	return &event.Failed{}
}

func (h Failed) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.Failed)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)

	return nil

}
