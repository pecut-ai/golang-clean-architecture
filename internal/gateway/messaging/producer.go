package messaging

import (
	"context"
	"encoding/json"
	"golang-clean-architecture/internal/logging"
	"golang-clean-architecture/internal/model"

	"github.com/IBM/sarama"
)

type Producer[T model.Event] struct {
	Producer sarama.SyncProducer
	Topic    string
	Log      *logging.Logger
}

func (p *Producer[T]) GetTopic() *string {
	return &p.Topic
}

func (p *Producer[T]) Send(ctx context.Context, event T) error {
	log := p.Log.FromContext(ctx)
	value, err := json.Marshal(event)
	if err != nil {
		log.WithError(err).Error("failed to marshal event")
		return err
	}

	message := &sarama.ProducerMessage{
		Topic: p.Topic,
		Key:   sarama.StringEncoder(event.GetId()),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := p.Producer.SendMessage(message)
	if err != nil {
		log.WithError(err).Error("failed to produce message")
		return err
	}

	log.Debugf("Message sent to topic %s, partition %d, offset %d", p.Topic, partition, offset)
	return nil
}
