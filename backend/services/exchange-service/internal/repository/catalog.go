package repository

import (
	"context"

	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/exchange-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Catalog struct{ Pool *pgxpool.Pool }

func (r Catalog) List(ctx context.Context, user uuid.UUID) ([]domain.Currency, error) {
	rows, err := r.Pool.Query(ctx, `
   SELECT code, precision FROM currencies
   UNION ALL
   SELECT code, precision FROM user_currencies u
   WHERE user_id = $1 AND NOT EXISTS (SELECT 1 FROM currencies c WHERE c.code = u.code)
   ORDER BY code`, user)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[domain.Currency])
}

func (r Catalog) Get(ctx context.Context, user uuid.UUID, code string) (domain.Currency, error) {
	var result domain.Currency
	err := r.Pool.QueryRow(ctx, `
   SELECT code, precision FROM (
      SELECT code, precision, 0 AS priority FROM currencies WHERE code = $2
      UNION ALL SELECT code, precision, 1 FROM user_currencies WHERE user_id = $1 AND code = $2
   ) AS candidates ORDER BY priority LIMIT 1`, user, code).Scan(&result.Code, &result.Precision)
	return result, err
}

func (r Catalog) Add(ctx context.Context, user uuid.UUID, currency domain.Currency) (bool, error) {
	tag, err := r.Pool.Exec(ctx, `
   INSERT INTO user_currencies (user_id, code, precision)
   SELECT $1::uuid, $2::text, $3::smallint WHERE NOT EXISTS (SELECT 1 FROM currencies WHERE code = $2::text)
   ON CONFLICT DO NOTHING`, user, currency.Code, currency.Precision)
	return tag.RowsAffected() == 1, err
}

func (r Catalog) Bases(ctx context.Context) ([]string, error) {
	rows, err := r.Pool.Query(ctx, `SELECT code FROM currencies ORDER BY code`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[string])
}

func (r Catalog) SaveRates(ctx context.Context, rates []domain.Rate) error {
	return pgx.BeginFunc(ctx, r.Pool, func(tx pgx.Tx) error {
		for _, rate := range rates {
			if _, err := tx.Exec(ctx, `
     INSERT INTO exchange_rates (base, quote, observed_on, rate, source)
     VALUES ($1, $2, $3::text::date, $4::text::numeric, $5)
     ON CONFLICT (base, quote, observed_on, source) DO UPDATE SET rate = EXCLUDED.rate`,
				rate.Base, rate.Quote, rate.ObservedOn, rate.Value, rate.Source); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r Catalog) LatestRates(ctx context.Context, base string) ([]domain.Rate, error) {
	rows, err := r.Pool.Query(ctx, `
   SELECT DISTINCT ON (quote) base, quote, to_char(observed_on, 'YYYY-MM-DD'), rate::text, source
   FROM exchange_rates WHERE base = $1 ORDER BY quote, observed_on DESC, source`, base)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[domain.Rate])
}
