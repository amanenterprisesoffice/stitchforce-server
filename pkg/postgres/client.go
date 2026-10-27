package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"stitchforce-server/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Client struct {
	DB *sql.DB
}

func New(cfg config.PostgresConfig) (*Client, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	return &Client{DB: db}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.DB.PingContext(pingCtx)
}

func (c *Client) Close() error {
	return c.DB.Close()
}
