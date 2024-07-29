package postgres

import (
	"TransactiStream/internal/domain"
	"context"
	"github.com/jackc/pgx/v5"
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
	err := p.db.QueryRow(ctx, `INSERT INTO transactions (user_id, amount, currency, created_at) VALUES ($1, $2, $3, $4) RETURNING id`,
		trans.UserID, trans.Amount, trans.Currency, trans.Timestamp).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (p *Postgres) Read(ctx context.Context, id string) (*domain.Transaction, error) {
	trans := &domain.Transaction{}

	err := p.db.QueryRow(ctx, `SELECT user_id, amount, currency, created_at FROM transactions WHERE id = $1`, id).
		Scan(&trans.UserID, &trans.Amount, &trans.Currency, &trans.Timestamp)
	if err != nil {
		return nil, err
	}

	return trans, nil
}

func (p *Postgres) Update(ctx context.Context, trans *domain.Transaction) error {
	_, err := p.db.Exec(ctx, `UPDATE transactions SET user_id = $1, amount = $2, currency = $3, created_at = $4 WHERE id = $5`,
		trans.UserID, trans.Amount, trans.Currency, trans.Timestamp, trans.ID)
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
