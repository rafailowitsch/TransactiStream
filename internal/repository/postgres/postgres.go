package postgres

import (
	"TransactiStream/internal/domain"
	"TransactiStream/internal/logger"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"
)

type Postgres struct {
	db *pgx.Conn
}

func NewPostgres(db *pgx.Conn) *Postgres {
	return &Postgres{
		db: db,
	}
}

func (p *Postgres) Create(ctx context.Context, trans *domain.Transaction) (string, error) {
	var id string

	if trans.Timestamp.IsZero() {
		trans.Timestamp = time.Now()
	}

	err := p.db.QueryRow(ctx, `INSERT INTO transactions (user_id, amount, currency, created_at) VALUES ($1, $2, $3, $4) RETURNING id`,
		trans.UserID, trans.Amount, trans.Currency, trans.Timestamp).Scan(&id)
	if err != nil {
		return "", err
	}

	trans.ID = id
	logger.Infof("Repo: transaction created: %v", *trans)

	return id, nil
}

func (p *Postgres) Read(ctx context.Context, id string) (*domain.Transaction, error) {
	trans := &domain.Transaction{}

	err := p.db.QueryRow(ctx, `SELECT user_id, amount, currency, created_at FROM transactions WHERE id = $1`, id).
		Scan(&trans.UserID, &trans.Amount, &trans.Currency, &trans.Timestamp)
	if err != nil {
		return nil, err
	}

	logger.Infof("Repo: transaction readed: %v", *trans)

	return trans, nil
}

func (p *Postgres) Update(ctx context.Context, trans *domain.Transaction) error {
	_, err := p.db.Exec(ctx, `UPDATE transactions SET user_id = $1, amount = $2, currency = $3, done=$4, created_at = $5 WHERE id = $6`,
		trans.UserID, trans.Amount, trans.Currency, trans.Done, trans.Timestamp, trans.ID)
	if err != nil {
		return err
	}

	logger.Infof("Repo: transaction updated: %v", *trans)

	return nil
}

func (p *Postgres) SetProcessedAt(ctx context.Context, id string) error {
	processedAt := time.Now()

	query := `
    UPDATE transactions 
    SET 
        processed_at = $1, 
        processing_time = $1 - created_at 
    WHERE id = $2`

	_, err := p.db.Exec(ctx, query, processedAt, id)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) ReadAll(ctx context.Context) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction

	rows, err := p.db.Query(ctx, `SELECT id, user_id, amount, currency, created_at FROM transactions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		trans := &domain.Transaction{}
		err = rows.Scan(&trans.ID, &trans.UserID, &trans.Amount, &trans.Currency, &trans.Timestamp)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, trans)
	}

	return transactions, nil
}

func (p *Postgres) GetStatistics(ctx context.Context) (*domain.Statistics, error) {
	stats := &domain.Statistics{}

	query := `SELECT COUNT(*) FROM transactions`
	if err := p.db.QueryRow(ctx, query).Scan(&stats.TotalTransactions); err != nil {
		return nil, fmt.Errorf("failed to get total transactions: %w", err)
	}

	// number of failed transactions (done == false)
	query = `SELECT COUNT(*) FROM transactions WHERE done = FALSE`
	if err := p.db.QueryRow(ctx, query).Scan(&stats.FailedTransactions); err != nil {
		return nil, fmt.Errorf("failed to get failed transactions: %w", err)
	}

	// number of users
	query = `SELECT COUNT(DISTINCT user_id) FROM transactions`
	if err := p.db.QueryRow(ctx, query).Scan(&stats.TotalUsers); err != nil {
		return nil, fmt.Errorf("failed to get total users: %w", err)
	}

	query = `SELECT AVG(EXTRACT(EPOCH FROM processing_time)) FROM transactions WHERE processing_time IS NOT NULL`
	var avgProcessingTime float64
	if err := p.db.QueryRow(ctx, query).Scan(&avgProcessingTime); err != nil {
		return nil, fmt.Errorf("failed to get average processing time: %w", err)
	}
	stats.AverageProcessingTime = avgProcessingTime

	// list of unique currencies (currency)
	query = `SELECT DISTINCT currency FROM transactions`
	rows, err := p.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get unique currencies: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var currency string
		if err := rows.Scan(&currency); err != nil {
			return nil, fmt.Errorf("failed to scan currency: %w", err)
		}
		stats.Currencies = append(stats.Currencies, currency)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return stats, nil
}
