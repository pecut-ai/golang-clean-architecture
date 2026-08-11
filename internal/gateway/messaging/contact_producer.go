package messaging

import (
	"golang-clean-architecture/internal/logging"
	"golang-clean-architecture/internal/model"

	"github.com/IBM/sarama"
)

type ContactProducer struct {
	Producer[*model.ContactEvent]
}

func NewContactProducer(producer sarama.SyncProducer, log *logging.Logger) *ContactProducer {
	return &ContactProducer{
		Producer: Producer[*model.ContactEvent]{
			Producer: producer,
			Topic:    "contacts",
			Log:      log,
		},
	}
}
