package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/segmentio/kafka-go"
)

type Transaction struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Done      bool      `json:"done"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	brokers := []string{"kafka:9092"}
	newTransactionsTopic := "new_transactions"
	processedTransactionsTopic := "processed_transactions"

	ctx := context.Background()

	log.Println("mocktstream started")

	go func() {
		if err := processTransactions(ctx, brokers, newTransactionsTopic, processedTransactionsTopic); err != nil {
			log.Fatalf("Failed to process transactions: %v", err)
		}
	}()

	select {}

}

func processTransactions(ctx context.Context, brokers []string, readTopic, writeTopic string) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    readTopic,
		GroupID:  "transaction-processor-group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	writer := kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    writeTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("failed to read message: %w", err)
		}

		var trans Transaction
		err = json.Unmarshal(m.Value, &trans)
		if err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		log.Printf("Received transaction: %+v\n", trans)

		// Симуляция обработки от 1 до 5 секунд
		processingTime := time.Duration(rand.Intn(5)+1) * time.Second
		time.Sleep(processingTime)

		// Случайная ошибка в 20% случаев
		if rand.Float32() < 0.2 {
			trans.Done = false
			log.Printf("Simulated error for transaction: %+v\n", trans)
		} else {
			trans.Done = true
		}

		message, err := json.Marshal(trans)
		if err != nil {
			log.Printf("Failed to marshal transaction: %v", err)
			continue
		}

		err = writer.WriteMessages(ctx,
			kafka.Message{
				Key:   []byte(trans.ID),
				Value: message,
			},
		)
		if err != nil {
			log.Printf("Failed to write message: %v", err)
			continue
		}

		log.Printf("Processed transaction: %+v\n", trans)
	}
}
