package event

import (
	"context"
	"io/ioutil"

	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/log.go/v2/log"
)

//go:generate moq -out mock/zebedee.go -pkg mock . ZebedeeClient

// ContentPublishedHandler ...
type ContentPublishedHandler struct {
	ZebedeeCli ZebedeeClient
}

// ZebedeeClient defines the zebedee client
type ZebedeeClient interface {
	GetPublishedData(ctx context.Context, uriString string) ([]byte, error)
}

// Handle takes a single event.
func (h *ContentPublishedHandler) Handle(ctx context.Context, cfg *config.Config, event *ContentPublished) (err error) {
	logData := log.Data{
		"event": event,
	}
	log.Info(ctx, "event handler called", logData)

	contentPublished, err := h.ZebedeeCli.GetPublishedData(ctx, event.URL)
	if err != nil {
		log.Fatal(ctx, "failed to retrieve content published from zebedee", err)
		return err
	}

	// TODO : we are simply writing the retrieved content to a file for now but in future 
	// this will be passed on to the dp-search-data-importer via an as yet unwritten kafka producer.
	err = ioutil.WriteFile(cfg.OutputFilePath, contentPublished, 0644)
	if err != nil {
		return err
	}

	log.Info(ctx, "event successfully handled", logData)

	return nil
}
