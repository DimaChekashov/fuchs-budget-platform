# Fuchs Budget Platform

[Архитектура бэкенда, фронтенда и монорепозитория](docs/architecture.md)

## Локальный бэкенд

Понадобятся Go 1.26.4, Goose с драйвером PostgreSQL и Docker Compose. Команды ниже выполняются из корня репозитория. Запускайте каждый сервис в отдельном терминале.

```powershell
docker compose up -d --wait postgres-identity postgres-wallet postgres-exchange redis

Copy-Item services/identity-service/.env.example services/identity-service/.env
Copy-Item services/wallet-service/.env.example services/wallet-service/.env
Copy-Item services/exchange-service/.env.example services/exchange-service/.env

goose -dir services/identity-service/migrations postgres 'postgres://fuchs:secret@localhost:5433/identity_db?sslmode=disable' up
goose -dir services/wallet-service/migrations postgres 'postgres://fuchs:secret@localhost:5434/wallet_db?sslmode=disable' up
goose -dir services/exchange-service/migrations postgres 'postgres://fuchs:secret@localhost:5436/exchange_db?sslmode=disable' up
./scripts/import-legacy-currencies.ps1
```

Для повторной настройки используйте существующие `.env`. У Identity, Wallet и Exchange должен совпадать `JWT_SECRET`. Примеры конфигурации предназначены для локального запуска. Укажите `EXCHANGE_SERVICE_URL=http://localhost:8083` в существующем `.env` Wallet после запуска Exchange. До включения этой настройки Wallet сохраняет прежний режим создания кошельков с двумя дробными знаками.

```powershell
# Терминал Identity, http://localhost:8080
Set-Location services/identity-service
go run .

# Терминал Wallet, http://localhost:8081
Set-Location services/wallet-service
go run .

# Терминал Exchange, http://localhost:8083
Set-Location services/exchange-service
go run .
```

Контракты HTTP API v1 описаны в [Identity OpenAPI](services/identity-service/api/openapi.yaml), [Wallet OpenAPI](services/wallet-service/api/openapi.yaml) и [Exchange OpenAPI](services/exchange-service/api/openapi.yaml).

## Денежные операции в API v1

База хранит баланс и суммы как `NUMERIC(18,2)`. API по-прежнему принимает и возвращает JSON-числа. Суммы должны быть положительными и точно представимыми с точностью валюты кошелька. Запрос, требующий округления, получает `400`. При создании кошелька Wallet сохраняет точность из Exchange. Миграция сохраняет два знака у существующих кошельков. Существующие балансы и операции остаются без пересчёта.

Для безопасного повтора пополнения или списания передайте `Idempotency-Key`. Один ключ на кошелёк закрепляется за типом операции, числовой суммой и описанием после успешной записи. Повтор возвращает тот же результат, а изменение параметров с тем же ключом получает `409`. Клиенты v1 могут вызывать операции без этого заголовка. Откат миграции `00002` удаляет сохранённые ключи повторов, поэтому после отката прежний ключ не защитит от повторного проведения операции.

Exchange каждые шесть часов загружает справочные курсы стандартных валют из [Frankfurter v2](https://frankfurter.dev/). При сбое поставщика API возвращает последнее сохранённое наблюдение и его дату. При недоступности Redis чтение идёт из PostgreSQL. Пользовательские коды доступны только владельцу и могут не иметь справочного курса. Скрипт импорта переносит поддерживаемые по формату коды существующих кошельков в персональные справочники, а старые кошельки сохраняют свои исходные коды.

## Проверки бэкенда

Запустите PostgreSQL для Wallet, как указано выше. Интеграционные тесты Wallet сами применяют миграции к временной схеме, затем удаляют её. Без `WALLET_TEST_DATABASE_URL` они пропускаются.

```powershell
$env:WALLET_TEST_DATABASE_URL = 'postgres://fuchs:secret@localhost:5434/wallet_db?sslmode=disable'
go test -count=1 ./services/identity-service/... ./services/wallet-service/... ./services/exchange-service/... ./pkg/jwtmiddleware/...
```

## Веб-приложение

Понадобится Node.js 22 или новее. Запустите Identity, Wallet и Exchange по инструкции выше.

```powershell
Set-Location apps/web
npm ci
npm run dev
```

Откройте `http://localhost:3000`. Если API слушают другие адреса, скопируйте `.env.example` в `.env.local` и задайте их там.

```powershell
# Из apps/web, Go-сервисы должны быть запущены для сквозных тестов
npm run generate:api
npm run build
npm run typecheck
npx playwright install chromium
npm run test:e2e
```

Команда `generate:api` обновляет типы клиента из OpenAPI. Playwright проверяет сценарии на телефоне, планшете и десктопе.
