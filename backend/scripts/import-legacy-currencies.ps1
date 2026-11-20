# Run after migrating Wallet and Exchange in local Docker Compose.
$ErrorActionPreference = 'Stop'
$walletQuery = @'
SELECT COALESCE(json_agg(currencies), '[]'::json)
FROM (
  SELECT DISTINCT user_id, upper(btrim(currency)) AS code, currency_precision AS precision
  FROM wallets WHERE upper(btrim(currency)) ~ '^[A-Z][A-Z0-9]{2,9}$'
) AS currencies
'@
$rows = docker compose exec -T postgres-wallet psql -v ON_ERROR_STOP=1 -U fuchs -d wallet_db -Atc $walletQuery
if ($LASTEXITCODE -ne 0) { throw 'Could not export existing wallet currencies.' }
$payload = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes(($rows -join "`n")))
$exchangeQuery = @"
INSERT INTO user_currencies (user_id, code, precision)
SELECT input.user_id, input.code, MAX(input.precision)
FROM json_to_recordset(convert_from(decode('$payload', 'base64'), 'UTF8')::json)
  AS input(user_id UUID, code VARCHAR(10), precision SMALLINT)
WHERE NOT EXISTS (SELECT 1 FROM currencies WHERE currencies.code = input.code)
GROUP BY input.user_id, input.code
ON CONFLICT (user_id, code) DO NOTHING
"@
$exchangeQuery | docker compose exec -T postgres-exchange psql -v ON_ERROR_STOP=1 -U fuchs -d exchange_db
if ($LASTEXITCODE -ne 0) { throw 'Could not import existing wallet currencies.' }
