package kafkaService

import (
	"TransactiStream/internal/domain"
	"TransactiStream/internal/logger"
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
)

type Repository interface {
	Update(ctx context.Context, trans *domain.Transaction) error
	SetProcessedAt(ctx context.Context, id string) error
}

type KafkaService struct {
	writer *kafka.Writer
	reader *kafka.Reader
	repo   Repository
}

func NewKafka(brokers []string, writeTopic, readTopic, groupID string, repo Repository) *KafkaService {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    writeTopic,
		Balancer: &kafka.LeastBytes{},
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    readTopic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &KafkaService{
		writer: writer,
		reader: reader,
		repo:   repo,
	}
}

func (k *KafkaService) SendMessage(ctx context.Context, trans *domain.Transaction) error {
	message, err := json.Marshal(trans)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	err = k.writer.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(trans.ID),
			Value: message,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	logger.Infof("Message sent successfully: %s", trans.ID)
	return nil
}

func (k *KafkaService) ReceiveMessages(ctx context.Context) error {
	for {
		m, err := k.reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("failed to read message: %w", err)
		}

		var trans domain.Transaction
		err = json.Unmarshal(m.Value, &trans)
		if err != nil {
			logger.Errorf("failed to unmarshal message: %v", err)
			continue
		}

		logger.Infof("Received transaction: %+v", trans)

		if err := k.repo.Update(ctx, &trans); err != nil {
			logger.Errorf("failed to update transaction in repository: %v", err)
			continue
		}

		if err := k.repo.SetProcessedAt(ctx, trans.ID); err != nil {
			logger.Errorf("failed to set processed at time: %v", err)
			continue
		}

	}
}

func CreateTopic(brokers []string, topic string) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	var connController *kafka.Conn
	connController, err = kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return err
	}
	defer connController.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = connController.CreateTopics(topicConfigs...)
	if err != nil {
		return err
	}

	return nil
}
