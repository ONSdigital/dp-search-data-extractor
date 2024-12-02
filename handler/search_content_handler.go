package handler

import (
	"context"
	"fmt"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	ReleaseDataType = "release"
)

type SearchContentHandler struct {
	Cfg      *config.Config
	Producer kafka.IProducer
}

// Handle processes the search-content-updated event and generates messages.
func (h *SearchContentHandler) Handle(ctx context.Context, _ int, msg kafka.Message) error {
	e := &models.SearchContentUpdate{}
	if err := schema.SearchContentUpdateEvent.Unmarshal(msg.GetData(), e); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	logData := log.Data{"event": e}
	log.Info(ctx, "search content event received", logData)

	if err := h.sendSearchDataImported(ctx, *e); err != nil {
		return err
	}

	log.Info(ctx, "search content event successfully handled", logData)
	return nil
}

func (h *SearchContentHandler) sendSearchDataImported(ctx context.Context, resource models.SearchContentUpdate) error {
	searchDataImport := models.SearchDataImport{
		URI:             resource.URI,
		Title:           resource.Title,
		DataType:        resource.ContentType,
		Summary:         resource.Summary,
		Survey:          resource.Survey,
		MetaDescription: resource.MetaDescription,
		Topics:          resource.Topics,
		ReleaseDate:     resource.ReleaseDate,
		Language:        resource.Language,
		Edition:         resource.Edition,
		DatasetID:       resource.DatasetID,
		CDID:            resource.CDID,
		CanonicalTopic:  resource.CanonicalTopic,
	}

	if resource.ContentType == ReleaseDataType {
		searchDataImport.Cancelled = resource.Release.Cancelled
		searchDataImport.Finalised = resource.Release.Finalised
		searchDataImport.Published = resource.Release.Published
		searchDataImport.ProvisionalDate = resource.Release.ProvisionalDate
	}

	data, err := schema.SearchDataImportEvent.Marshal(searchDataImport)
	if err != nil {
		return fmt.Errorf("failed to marshal search-content-update event: %w", err)
	}

	// Marshall Avro and sending message
	if err := h.Producer.Send(schema.SearchDataImportEvent, data); err != nil {
		log.Error(ctx, "error while attempting to send SearchDataImport event to producer", err)
		return fmt.Errorf("failed to send search data import event: %w", err)
	}

	log.Info(ctx, "search-data-imported event sent", log.Data{"uri": searchDataImport.URI})
	return nil
}
