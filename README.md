# Fuchs Budget Platform

[Архитектура бэкенда, фронтенда и монорепозитория](docs/architecture.md)

## Локальный бэкенд

Понадобятся Go 1.26.4, Goose с драйвером PostgreSQL и Docker Compose. Команды ниже выполняются из корня репозитория. Для запуска сервисов потребуются два терминала.

```powershell
docker compose up -d --wait postgres-identity postgres-wallet

Copy-Item services/identity-service/.env.example services/identity-service/.env
Copy-Item services/wallet-service/.env.example services/wallet-service/.env

goose -dir services/identity-service/migrations postgres 'postgres://fuchs:secret@localhost:5433/identity_db?sslmode=disable' up
goose -dir services/wallet-service/migrations postgres 'postgres://fuchs:secret@localhost:5434/wallet_db?sslmode=disable' up
```

Для повторной настройки используйте существующие `.env`. У Identity и Wallet должен совпадать `JWT_SECRET`. Примеры конфигурации предназначены для локального запуска.

```powershell
# Терминал Identity, http://localhost:8080
Set-Location services/identity-service
go run .

# Терминал Wallet, http://localhost:8081
Set-Location services/wallet-service
go run .
```

Контракты HTTP API v1 описаны в [Identity OpenAPI](services/identity-service/api/openapi.yaml) и [Wallet OpenAPI](services/wallet-service/api/openapi.yaml).

## Денежные операции в API v1

База хранит баланс и суммы как `NUMERIC(18,2)`. API по-прежнему принимает и возвращает JSON-числа. Суммы в валюте кошелька должны быть положительными и точно представимыми до сотых. Запрос, требующий округления, получает `400`. Клиентам следует округлять по своим бизнес-правилам до запроса и при необходимости хранить суммы в десятичном типе. Существующие балансы и операции остаются без пересчёта. Валютный справочник и индивидуальная точность валют относятся к этапу реализации Exchange.

Для безопасного повтора пополнения или списания передайте `Idempotency-Key`. Один ключ на кошелёк закрепляется за типом операции, числовой суммой и описанием после успешной записи. Повтор возвращает тот же результат, а изменение параметров с тем же ключом получает `409`. Клиенты v1 могут вызывать операции без этого заголовка. Откат миграции `00002` удаляет сохранённые ключи повторов, поэтому после отката прежний ключ не защитит от повторного проведения операции.

## Проверки бэкенда

Запустите PostgreSQL для Wallet, как указано выше. Интеграционные тесты Wallet сами применяют миграции к временной схеме, затем удаляют её. Без `WALLET_TEST_DATABASE_URL` они пропускаются.

```powershell
$env:WALLET_TEST_DATABASE_URL = 'postgres://fuchs:secret@localhost:5434/wallet_db?sslmode=disable'
go test -count=1 ./services/identity-service/... ./services/wallet-service/... ./pkg/jwtmiddleware/...
```

## Веб-приложение

Понадобится Node.js 22 или новее. Запустите Identity и Wallet по инструкции выше.

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
