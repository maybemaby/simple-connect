package data

import (
	"context"
	"log/slog"
	"os"

	"simple-connect/api/internal"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxTracer struct {
	logger *slog.Logger
}

func (tracer *pgxTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	tracer.logger.Debug("Query start", slog.String("sql", data.SQL))
	return ctx
}

func (tracer *pgxTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {

}

func NewPool(ctx context.Context, trace bool) (*pgxpool.Pool, error) {
	connStr := os.Getenv("DATABASE_URL")
	cfg, err := pgxpool.ParseConfig(connStr)

	if err != nil {
		return nil, err
	}

	if trace {
		cfg.BeforeConnect = func(ctx context.Context, connCfg *pgx.ConnConfig) error {
			connCfg.Tracer = &pgxTracer{
				logger: internal.BootstrapLogger(slog.LevelDebug, internal.JSONFormat, true).WithGroup("pgx"),
			}

			return nil
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)

	return pool, err
}
