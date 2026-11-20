package repository

import (
	"context"
	"time"

	"github.com/DimaChekashov/fuchs-budget-platform/services/ledger-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Reports struct{ Pool *pgxpool.Pool }

func (r Reports) Apply(ctx context.Context, entry domain.Operation) error {
	_, err := r.Pool.Exec(ctx, `
   WITH accepted AS (
      INSERT INTO processed_operations (operation_id, user_id) VALUES ($1, $2)
      ON CONFLICT DO NOTHING RETURNING operation_id
   )
   INSERT INTO daily_totals (user_id, wallet_id, currency, day, income, expenses)
   SELECT $2, $3, $4, ($5::timestamptz AT TIME ZONE 'UTC')::date,
       CASE WHEN $6::text = 'deposit' THEN $7::text::numeric ELSE 0 END,
       CASE WHEN $6::text = 'withdrawal' THEN $7::text::numeric ELSE 0 END
   FROM accepted
   ON CONFLICT (user_id, wallet_id, currency, day) DO UPDATE
   SET income = daily_totals.income + EXCLUDED.income,
       expenses = daily_totals.expenses + EXCLUDED.expenses`,
		entry.ID, entry.UserID, entry.WalletID, entry.Currency, entry.RecordedAt, entry.Type, entry.Amount)
	return err
}

func (r Reports) Reject(ctx context.Context, topic string, partition int32, offset int64, reason string) error {
	_, err := r.Pool.Exec(ctx, `
    INSERT INTO rejected_events (topic, partition, kafka_offset, reason) VALUES ($1, $2, $3, $4)
    ON CONFLICT DO NOTHING`, topic, partition, offset, reason)
	return err
}

func (r Reports) Summary(ctx context.Context, query domain.Query) (domain.Report, error) {
	report := domain.Report{From: query.From.Format(time.DateOnly), To: query.To.Format(time.DateOnly)}
	err := pgx.BeginTxFunc(ctx, r.Pool, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `SELECT max(processed_at) FROM processed_operations WHERE user_id = $1`, query.UserID).
			Scan(&report.UpdatedAt); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, `
   SELECT currency, sum(income)::text, sum(expenses)::text
   FROM daily_totals WHERE user_id = $1 AND day BETWEEN $2::text::date AND $3::text::date
     AND ($4::uuid IS NULL OR wallet_id = $4)
     AND ($5::text = '' OR currency = $5)
   GROUP BY currency ORDER BY currency`, query.UserID, report.From, report.To, query.WalletID, query.Currency)
		if err != nil {
			return err
		}
		report.Totals, err = pgx.CollectRows(rows, pgx.RowToStructByPos[domain.Total])
		return err
	})
	if report.Totals == nil {
		report.Totals = []domain.Total{}
	}
	return report, err
}
