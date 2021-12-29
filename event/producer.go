package event

import (
	"context"
	"fmt"

	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/log.go/v2/log"
)

//go:generate moq -out ./mock/producer.go -pkg mock . Marshaller

// Marshaller defines a type for marshalling the requested object into the required format.
type Marshaller interface {
	Marshal(s interface{}) ([]byte, error)
}

// SearchDataImportProducer produces kafka messages for instances which have been successfully processed.
type SearchDataImportProducer struct {
	Marshaller Marshaller
	Producer   kafka.IProducer
}

// SearchDataImport produce a kafka message for an instance which has been successfully processed.
func (p SearchDataImportProducer) SearchDataImport(ctx context.Context, event models.SearchDataImport) error {
	bytes, err := p.Marshaller.Marshal(event)
	if err != nil {
		log.Fatal(ctx, "Marshaller.Marshal", err)
		return fmt.Errorf(fmt.Sprintf("Marshaller.Marshal returned an error: event=%v: %%w", event), err)
	}
	p.Producer.Channels().Output <- bytes
	log.Info(ctx, "++++++++++++completed successfully", log.Data{"event": event, "package": "event.SearchDataImport"})
	return nil
}

// SearchDatasetVersionImport produce a kafka message for an instance which has been successfully processed.
func (p SearchDataImportProducer) SearchDatasetVersionMetadataImport(ctx context.Context, event models.SearchDataVersionMetadataImport) error {

	bytes, err := p.Marshaller.Marshal(&models.SearchDataVersionMetadataImport{
		CollectionId: event.CollectionId,
		Edition:      event.Edition,
		ID:           event.ID,
		DatasetId:    event.DatasetId,
		Version:      event.Version,
		ReleaseDate:  event.ReleaseDate,
	})

	if err != nil {
		log.Fatal(ctx, "Marshaller.Marshal", err)
		return fmt.Errorf(fmt.Sprintf("Marshaller.Marshal returned an error: event=%v: %%w", event), err)
	}

	p.Producer.Channels().Output <- bytes
	log.Info(ctx, "***************completed successfully", log.Data{"event": event, "package": "event.DatasetVersionSearchDataImport"})
	return nil
}
