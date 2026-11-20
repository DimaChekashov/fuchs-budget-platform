package service

import "testing"

func TestDecodeOperation(t *testing.T) {
	event := []byte(`{
  "version":1,
  "operation_id":"6bee2de3-8ec3-47cb-985a-97d583e8d5f9",
  "user_id":"99579292-582a-4d53-9d95-8fdd38adfc56",
  "wallet_id":"11111111-1111-4111-8111-111111111111",
  "type":"deposit",
  "amount":"9999999999999999.99",
  "currency":"RUB",
  "recorded_at":"2026-05-10T11:30:24+00:00"
 }`)
	decoded, err := decode(event)
	if err != nil || decoded.Amount != "9999999999999999.99" || decoded.Currency != "RUB" {
		t.Fatalf("invalid decoding: %+v, %v", decoded, err)
	}
	if _, err := decode([]byte(`{"version":2}`)); err == nil {
		t.Fatal("unsupported event version was accepted")
	}
}
