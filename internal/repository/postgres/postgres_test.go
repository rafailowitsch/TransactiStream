package postgres

import (
	"TransactiStream/internal/domain"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"testing"
	"time"
)

func setupPostgres(t *testing.T) (*pgx.Conn, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_USER":     "user",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatal(err)
	}

	dsn := fmt.Sprintf("postgres://user:password@%s:%s/testdb?sslmode=disable", host, port.Port())
	db, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	err = CreateTables(ctx, *db)
	if err != nil {
		t.Fatal(err)
	}

	teardown := func() {
		db.Close(ctx)
		postgresContainer.Terminate(ctx)
	}

	return db, teardown
}

func TestPostgres_Create(t *testing.T) {
	db, teardown := setupPostgres(t)
	defer teardown()

	p := NewPostgres(db)

	trans := &domain.Transaction{
		UserID:    "user1",
		Amount:    100.0,
		Currency:  "BTC",
		Timestamp: time.Now().UTC(),
	}

	id, err := p.Create(context.Background(), trans)
	assert.NoError(t, err)
	log.Println("transaction id:", id)

	var insertedTransaction domain.Transaction

	err = db.QueryRow(context.Background(), `SELECT user_id, amount, currency, created_at FROM transactions WHERE id = $1`, id).
		Scan(&insertedTransaction.UserID, &insertedTransaction.Amount, &insertedTransaction.Currency, &insertedTransaction.Timestamp)
	assert.NoError(t, err)

	assert.Equal(t, trans.UserID, insertedTransaction.UserID)
	assert.Equal(t, trans.Amount, insertedTransaction.Amount)
	assert.Equal(t, trans.Currency, insertedTransaction.Currency)
	assert.WithinDuration(t, trans.Timestamp, insertedTransaction.Timestamp, time.Second)
}

func TestPostgres_Read(t *testing.T) {
	db, teardown := setupPostgres(t)
	defer teardown()

	p := NewPostgres(db)

	trans := &domain.Transaction{
		UserID:    "user1",
		Amount:    100.0,
		Currency:  "BTC",
		Timestamp: time.Now().UTC(),
	}

	err := p.db.QueryRow(context.Background(), `INSERT INTO transactions (user_id, amount, currency, created_at) VALUES ($1, $2, $3, $4) RETURNING id`,
		trans.UserID, trans.Amount, trans.Currency, trans.Timestamp).Scan(&trans.ID)
	assert.NoError(t, err)

	readTrans, err := p.Read(context.Background(), trans.ID)
	assert.NoError(t, err)

	assert.Equal(t, trans.UserID, readTrans.UserID)
	assert.Equal(t, trans.Amount, readTrans.Amount)
	assert.Equal(t, trans.Currency, readTrans.Currency)
	assert.WithinDuration(t, trans.Timestamp, readTrans.Timestamp, time.Second)
}

func TestPostgres_Update(t *testing.T) {
	db, teardown := setupPostgres(t)
	defer teardown()

	p := NewPostgres(db)

	trans := &domain.Transaction{
		UserID:    "user1",
		Amount:    100.0,
		Currency:  "BTC",
		Timestamp: time.Now().UTC(),
	}

	err := p.db.QueryRow(context.Background(), `INSERT INTO transactions (user_id, amount, currency, created_at) VALUES ($1, $2, $3, $4) RETURNING id`,
		trans.UserID, trans.Amount, trans.Currency, trans.Timestamp).Scan(&trans.ID)
	assert.NoError(t, err)

	trans = &domain.Transaction{
		ID:        trans.ID,
		UserID:    "user2",
		Amount:    200.0,
		Currency:  "ETH",
		Timestamp: time.Now().UTC(),
	}

	err = p.Update(context.Background(), trans)
	assert.NoError(t, err)

	var updatedTrans domain.Transaction
	err = p.db.QueryRow(context.Background(), `SELECT user_id, amount, currency, created_at FROM transactions WHERE id = $1`, trans.ID).
		Scan(&updatedTrans.UserID, &updatedTrans.Amount, &updatedTrans.Currency, &updatedTrans.Timestamp)
	assert.NoError(t, err)

	assert.Equal(t, trans.UserID, updatedTrans.UserID)
	assert.Equal(t, trans.Amount, updatedTrans.Amount)
	assert.Equal(t, trans.Currency, updatedTrans.Currency)
	assert.WithinDuration(t, trans.Timestamp, updatedTrans.Timestamp, time.Second)
}

func TestPostgres_ReadAll(t *testing.T) {
	db, teardown := setupPostgres(t)
	defer teardown()

	p := NewPostgres(db)

	trans1 := &domain.Transaction{
		UserID:    "user1",
		Amount:    100.0,
		Currency:  "BTC",
		Timestamp: time.Now().UTC(),
	}

	trans2 := &domain.Transaction{
		UserID:    "user2",
		Amount:    200.0,
		Currency:  "ETH",
		Timestamp: time.Now().UTC(),
	}

	err := p.db.QueryRow(context.Background(), `INSERT INTO transactions (user_id, amount, currency, created_at) VALUES ($1, $2, $3, $4) RETURNING id`,
		trans1.UserID, trans1.Amount, trans1.Currency, trans1.Timestamp).Scan(&trans1.ID)
	assert.NoError(t, err)

	err = p.db.QueryRow(context.Background(), `INSERT INTO transactions (user_id, amount, currency, created_at) VALUES ($1, $2, $3, $4) RETURNING id`,
		trans2.UserID, trans2.Amount, trans2.Currency, trans2.Timestamp).Scan(&trans2.ID)
	assert.NoError(t, err)

	transactions, err := p.ReadAll(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, 2, len(transactions))

	for _, trans := range transactions {
		if trans.ID == trans1.ID {
			assert.Equal(t, trans1.UserID, trans.UserID)
			assert.Equal(t, trans1.Amount, trans.Amount)
			assert.Equal(t, trans1.Currency, trans.Currency)
			assert.WithinDuration(t, trans1.Timestamp, trans.Timestamp, time.Second)
		} else if trans.ID == trans2.ID {
			assert.Equal(t, trans2.UserID, trans.UserID)
			assert.Equal(t, trans2.Amount, trans.Amount)
			assert.Equal(t, trans2.Currency, trans.Currency)
			assert.WithinDuration(t, trans2.Timestamp, trans.Timestamp, time.Second)
		} else {
			t.Fatalf("Unexpected transaction ID: %s", trans.ID)
		}
	}
}
