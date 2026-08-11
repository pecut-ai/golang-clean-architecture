package config

import (
	"fmt"

	"github.com/IBM/sarama"
)

func OpenKafkaProducer(cfg KafkaConfig) (sarama.SyncProducer, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 3
	producer, err := sarama.NewSyncProducer(cfg.BootstrapServers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("open Kafka producer: %w", err)
	}
	return producer, nil
}

func OpenKafkaConsumerGroup(cfg KafkaConfig) (sarama.ConsumerGroup, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Return.Errors = true
	if cfg.AutoOffsetReset == "earliest" {
		saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	} else {
		saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	}
	consumer, err := sarama.NewConsumerGroup(cfg.BootstrapServers, cfg.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("open Kafka consumer group: %w", err)
	}
	return consumer, nil
}
