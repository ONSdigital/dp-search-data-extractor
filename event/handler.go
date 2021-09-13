package event

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/log.go/v2/log"
)

//go:generate moq -out mock/zebedee.go -pkg mock . ZebedeeClient

// ContentPublishedHandler ...
type ContentPublishedHandler struct {
	ZebedeeCli ZebedeeClient
}

// Handle takes a single event.
func (h *ContentPublishedHandler) Handle(ctx context.Context, cfg *config.Config, event *ContentPublished) (err error) {
	logData := log.Data{
		"event": event,
	}
	log.Info(ctx, "event handler called", logData)

	greeting := fmt.Sprintf("URL:%s, DataType:%s, CollectionID:%s", event.URL, event.DataType, event.CollectionID)

	contentPublished, err := h.ZebedeeCli.GetPublishedData(ctx, event.URL)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(cfg.OutputFilePath, contentPublished, 0644)
	if err != nil {
		return err
	}

	logData["greeting"] = greeting
	log.Info(ctx, "handler called successfully", logData)
	log.Info(ctx, "event successfully handled", logData)

	return nil
}

// ZebedeeClient defines the zebedee client
type ZebedeeClient interface {
	GetPublishedData(ctx context.Context, uriString string) ([]byte, error)
}
