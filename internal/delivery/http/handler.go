package http

import (
	kafkaService "TransactiStream/internal/delivery/kafka"
	"TransactiStream/internal/domain"
	"TransactiStream/internal/logger"
	"context"
	"encoding/json"
	"net/http"
)

type Repository interface {
	Create(ctx context.Context, trans *domain.Transaction) (string, error)
	Read(ctx context.Context, id string) (*domain.Transaction, error)
	Update(ctx context.Context, trans *domain.Transaction) error
	ReadAll(ctx context.Context) ([]*domain.Transaction, error)
	GetStatistics(ctx context.Context) (*domain.Statistics, error)
}

type Handler struct {
	repo     Repository
	kafkaSrv *kafkaService.KafkaService
}

func NewHandler(repo Repository, kafka *kafkaService.KafkaService) *Handler {
	return &Handler{
		repo:     repo,
		kafkaSrv: kafka,
	}
}

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var (
		trans *domain.Transaction
		err   error
		ctx   = r.Context()
	)

	if err = json.NewDecoder(r.Body).Decode(&trans); err != nil {
		logger.Errorf("Error decoding request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if trans.UserID == "" || trans.Currency == "" || trans.Amount == 0 {
		logger.Error("One of the fields is empty")
		http.Error(w, "One of the fields is empty", http.StatusBadRequest)
		return
	}

	if trans.ID, err = h.repo.Create(ctx, trans); err != nil {
		logger.Errorf("Error creating transaction: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = h.kafkaSrv.SendMessage(ctx, trans); err != nil {
		logger.Errorf("Error sending message to Kafka: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetAllTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	transactions, err := h.repo.ReadAll(ctx)
	if err != nil {
		logger.Errorf("Error reading all transactions: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(transactions); err != nil {
		logger.Errorf("Error encoding transactions: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	var (
		stats *domain.Statistics
		ctx   = r.Context()
		err   error
	)

	if stats, err = h.repo.GetStatistics(ctx); err != nil {
		logger.Errorf("Error getting statistics: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(stats); err != nil {
		logger.Errorf("Error encoding statistics: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
