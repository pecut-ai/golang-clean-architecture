package messaging

import (
	"golang-clean-architecture/internal/logging"
	"golang-clean-architecture/internal/model"

	"github.com/IBM/sarama"
)

type AddressProducer struct {
	Producer[*model.AddressEvent]
}

func NewAddressProducer(producer sarama.SyncProducer, log *logging.Logger) *AddressProducer {
	return &AddressProducer{
		Producer: Producer[*model.AddressEvent]{
			Producer: producer,
			Topic:    "addresses",
			Log:      log,
		},
	}
}
