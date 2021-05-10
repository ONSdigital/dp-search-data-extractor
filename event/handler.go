package event

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/log.go/log"
)

// TODO: remove hello called example handler
// HelloCalledHandler ...
type HelloCalledHandler struct {
}

// Handle takes a single event.
func (h *HelloCalledHandler) Handle(ctx context.Context, cfg *config.Config, event *HelloCalled) (err error) {
	logData := log.Data{
		"event": event,
	}
	log.Event(ctx, "event handler called", log.INFO, logData)

	greeting := fmt.Sprintf("Hello, %s!", event.RecipientName)
	err = ioutil.WriteFile(cfg.OutputFilePath, []byte(greeting), 0644)
	if err != nil {
		return err
	}

	logData["greeting"] = greeting
	log.Event(ctx, "hello world example handler called successfully", log.INFO, logData)
	log.Event(ctx, "event successfully handled", log.INFO, logData)

	return nil
}
