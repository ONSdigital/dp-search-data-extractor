package event

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/log.go/v2/log"
)

// ContentPublishedHandler ...
type ContentPublishedHandler struct {
}

// Handle takes a single event.
func (h *ContentPublishedHandler) Handle(ctx context.Context, cfg *config.Config, event *ContentPublished) (err error) {
	logData := log.Data{
		"event": event,
	}
	log.Info(ctx, "event handler called", logData)

	greeting := fmt.Sprintf("URL:%s, DataType:%s, CollectionID:%s", event.URL, event.DataType, event.CollectionID)
	err = ioutil.WriteFile(cfg.OutputFilePath, []byte(greeting), 0644)
	if err != nil {
		return err
	}

	logData["greeting"] = greeting
	log.Info(ctx, "hello world example handler called successfully", logData)
	log.Info(ctx, "event successfully handled", logData)

	return nil
}
