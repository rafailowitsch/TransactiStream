package http

import (
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
}

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{
		repo: repo,
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

	w.WriteHeader(http.StatusCreated)
}
