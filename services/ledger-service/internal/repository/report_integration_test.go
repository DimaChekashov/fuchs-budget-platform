package repository

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DimaChekashov/fuchs-budget-platform/services/ledger-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestReportProjection(t *testing.T) {
	dsn := os.Getenv("LEDGER_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("LEDGER_TEST_DATABASE_URL is required")
	}
	admin, err := pgxpool.New(t.Context(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(t.Context(), "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := admin.Exec(cleanup, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Error(err)
		}
	}()
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	pool, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	migration, err := os.ReadFile("../../migrations/00001_create_reports.sql")
	if err != nil {
		t.Fatal(err)
	}
	up, _, found := strings.Cut(string(migration), "-- +goose Down")
	if !found {
		t.Fatal("missing migration Down marker")
	}
	if _, err := pool.Exec(t.Context(), up); err != nil {
		t.Fatal(err)
	}

	repo := Reports{Pool: pool}
	owner, other := uuid.New(), uuid.New()
	firstWallet, secondWallet := uuid.New(), uuid.New()
	day, _ := time.Parse(time.DateOnly, "2026-05-10")
	deposit := domain.Operation{ID: uuid.New(), UserID: owner, WalletID: firstWallet, Currency: "RUB",
		Type: "deposit", Amount: "12.30", RecordedAt: day}
	withdrawal := domain.Operation{ID: uuid.New(), UserID: owner, WalletID: firstWallet, Currency: "RUB",
		Type: "withdrawal", Amount: "2.10", RecordedAt: day}
	foreignCurrency := domain.Operation{ID: uuid.New(), UserID: owner, WalletID: secondWallet, Currency: "USD",
		Type: "deposit", Amount: "3.00", RecordedAt: day}
	differentUser := deposit
	differentUser.ID, differentUser.UserID = uuid.New(), other
	previousDay := deposit
	previousDay.ID, previousDay.RecordedAt = uuid.New(), day.AddDate(0, 0, -1)
	for _, event := range []domain.Operation{deposit, withdrawal, foreignCurrency, deposit, differentUser, previousDay} {
		if err := repo.Apply(t.Context(), event); err != nil {
			t.Fatal(err)
		}
	}

	q := domain.Query{UserID: owner, From: day, To: day}
	report, err := repo.Summary(t.Context(), q)
	if err != nil || report.UpdatedAt == nil || len(report.Totals) != 2 {
		t.Fatalf("incorrect report: %+v, %v", report, err)
	}
	if got := report.Totals[0]; got.Currency != "RUB" || got.Income != "12.30" || got.Expenses != "2.10" {
		t.Fatalf("duplicate or other user's operations affected RUB totals: %+v", got)
	}
	q.WalletID = &secondWallet
	report, err = repo.Summary(t.Context(), q)
	if err != nil || len(report.Totals) != 1 || report.Totals[0].Currency != "USD" {
		t.Fatalf("wallet filter returned %+v, %v", report, err)
	}
	q.Currency = "RUB"
	report, err = repo.Summary(t.Context(), q)
	if err != nil || len(report.Totals) != 0 {
		t.Fatalf("currency filter returned %+v, %v", report, err)
	}
}
