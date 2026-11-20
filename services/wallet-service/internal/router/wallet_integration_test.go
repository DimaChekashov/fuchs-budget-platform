package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/config"
	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMoneyOperations(t *testing.T) {
	dsn := os.Getenv("WALLET_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("WALLET_TEST_DATABASE_URL is required for PostgreSQL integration tests")
	}
	pool := migratedPool(t, dsn)
	const secret = "test-secret"
	router := New(pool, &config.Config{JWTSecret: secret})
	userID := uuid.New()
	token := signToken(t, secret, userID)

	newWallet := func(t *testing.T) uuid.UUID {
		t.Helper()
		response := request(router, token, "/api/v1/wallets", `{"name":"Cash"}`, "")
		var wallet domain.Wallet
		if response.Code != http.StatusCreated || json.Unmarshal(response.Body.Bytes(), &wallet) != nil {
			t.Fatalf("create wallet: %d %s", response.Code, response.Body)
		}
		return wallet.ID
	}
	operation := func(route, auth string, walletID uuid.UUID, amount, description, key string) *httptest.ResponseRecorder {
		body := fmt.Sprintf(`{"wallet_id":%q,"amount":%s,"description":%q}`, walletID.String(), amount, description)
		return request(router, auth, "/api/v1/wallets/"+route, body, key)
	}
	checkState := func(t *testing.T, walletID uuid.UUID, balance string, count int) {
		t.Helper()
		var gotBalance string
		var gotCount int
		err := pool.QueryRow(t.Context(), `
   SELECT balance::text, (SELECT count(*) FROM transactions WHERE wallet_id = $1)
   FROM wallets WHERE id = $1`, walletID).Scan(&gotBalance, &gotCount)
		if err != nil || gotBalance != balance || gotCount != count {
			t.Fatalf("balance = %q, operations = %d, error = %v; want %q and %d", gotBalance, gotCount, err, balance, count)
		}
	}
	requireStatus := func(t *testing.T, response *httptest.ResponseRecorder, want int) {
		t.Helper()
		if response.Code != want {
			t.Fatalf("got %d %s; want %d", response.Code, response.Body, want)
		}
	}

	t.Run("rolls back balance if journal insert fails", func(t *testing.T) {
		wallet := newWallet(t)
		response := operation("deposit", token, wallet, "12.34", strings.Repeat("a", 256), "failed")
		requireStatus(t, response, http.StatusInternalServerError)
		checkState(t, wallet, "0.00", 0)
		requireStatus(t, operation("deposit", token, wallet, "12.34", "", "failed"), http.StatusOK)
		checkState(t, wallet, "12.34", 1)
	})
	t.Run("concurrent withdrawals cannot overspend", func(t *testing.T) {
		wallet := newWallet(t)
		requireStatus(t, operation("deposit", token, wallet, "10", "", ""), http.StatusOK)
		start := make(chan struct{})
		results := make(chan int, 2)
		for range 2 {
			go func() {
				<-start
				results <- operation("withdraw", token, wallet, "7", "", "").Code
			}()
		}
		close(start)
		first, second := <-results, <-results
		if first > second {
			first, second = second, first
		}
		if first != http.StatusOK || second != http.StatusBadRequest {
			t.Fatalf("withdrawals returned %d and %d", first, second)
		}
		checkState(t, wallet, "3.00", 2)
	})
	t.Run("replays an operation and rejects key reuse", func(t *testing.T) {
		wallet := newWallet(t)
		start := make(chan struct{})
		results := make(chan *httptest.ResponseRecorder, 2)
		for _, amount := range []string{"12.30", "123e-1"} {
			go func() {
				<-start
				results <- operation("deposit", token, wallet, amount, "income", "deposit-1")
			}()
		}
		close(start)
		first, replay := <-results, <-results
		requireStatus(t, first, http.StatusOK)
		requireStatus(t, replay, http.StatusOK)
		if first.Body.String() != replay.Body.String() {
			t.Fatalf("replay changed the response: %s / %s", first.Body, replay.Body)
		}
		requireStatus(t, operation("withdraw", token, wallet, "12.3", "income", "deposit-1"), http.StatusConflict)
		checkState(t, wallet, "12.30", 1)
	})
	t.Run("enforces ownership and amount precision", func(t *testing.T) {
		wallet := newWallet(t)
		stranger := signToken(t, secret, uuid.New())
		requireStatus(t, operation("deposit", stranger, wallet, "1", "", ""), http.StatusForbidden)
		requireStatus(t, operation("deposit", token, uuid.New(), "1", "", ""), http.StatusNotFound)
		requireStatus(t, operation("deposit", token, wallet, "0.005", "", ""), http.StatusBadRequest)
		checkState(t, wallet, "0.00", 0)
		requireStatus(t, operation("deposit", token, wallet, "9999999999999999.99", "", ""), http.StatusOK)
		requireStatus(t, operation("deposit", token, wallet, "0.01", "", ""), http.StatusBadRequest)
		checkState(t, wallet, "9999999999999999.99", 1)
	})
}

func migratedPool(t *testing.T, dsn string) *pgxpool.Pool {
	t.Helper()
	admin, err := pgxpool.New(t.Context(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(t.Context(), "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Error(err)
		}
	})
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	pool, err := pgxpool.NewWithConfig(t.Context(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	files, err := filepath.Glob("../../migrations/*.sql")
	if err != nil || len(files) == 0 {
		t.Fatalf("find wallet migrations: %v", err)
	}
	for _, name := range files {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		up, _, found := strings.Cut(string(data), "-- +goose Down")
		if !found {
			t.Fatalf("missing Down marker in %s", name)
		}
		if _, err := pool.Exec(t.Context(), up); err != nil {
			t.Fatalf("apply %s: %v", name, err)
		}
	}
	return pool
}

func signToken(t *testing.T, secret string, user uuid.UUID) string {
	t.Helper()
	claims := jwt.MapClaims{"sub": user.String(), "exp": time.Now().Add(time.Hour).Unix()}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func request(router http.Handler, token, route, body, key string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, route, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}
