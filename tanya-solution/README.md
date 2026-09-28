# currency-service: запуск и проверка

CLI валютного сервиса: задачи [001–004](../docs/tasks/README.md). Он умеет:
- считать сумму по заданному курсу;
- импортировать курсы пяти источников в CSV;
- выбирать лучшее банковское предложение.

Все команды ниже выполняются из каталога `tanya-solution/`.

## Требования

- Go 1.26 или новее (`go version`).
- Доступ в интернет — только для `import` без `--input`.

## Сборка

```bash
go build -o currency-service ./cmd/currency-service
```

```bash
go vet ./...
```

## Переменные окружения

| Переменная | По умолчанию | Назначение |
|---|---|---|
| `DATA_DIR` | `./data` | Корень данных: `archive/`, `backoff/` |
| `SOURCES` | `cbr-xml` | Источники через запятую: `cbr-xml`, `cbr-json-mirror`, `moex-iss`, `tbank-public`, `raiffeisen-public` |
| `MAX_RATE_AGE` | `30m` | Максимальный возраст банковского предложения для `convert` |
| `LOG_LEVEL` | `INFO` | `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `EXTRA_CA_FILE` | — | PEM с дополнительными корневыми сертификатами (нужен для T-Bank, см. ниже) |

Логи пишутся в stdout в формате JSON, по одной записи на строку. Результат `calculate` и `convert` выводится отдельной строкой. `LOG_LEVEL=ERROR` оставляет в выводе только результат и ошибки.

## 1. calculate (задачи 001–002)

```bash
./currency-service calculate --amount 12.50 --rate 80.40
```

Ожидается `1005.00`. Другие примеры: `100 × 80 → 8000.00`, `0.10 × 3 → 0.30`. Нулевые, отрицательные, пустые и нечисловые значения отклоняются с кодом 1:

```bash
./currency-service calculate --amount 0 --rate 80; echo "exit=$?"
```

## 2. import из сети (задачи 003–004)

Все пять источников за один запуск:

```bash
SOURCES=cbr-xml,cbr-json-mirror,moex-iss,tbank-public,raiffeisen-public ./currency-service import; echo "exit=$?"
```

Для каждого источника пишется одна запись в лог:
- `INFO source imported` с `records` и путём к файлу `data/archive/YYYY-MM-DD/<source>-<UTC ts>.csv`;
- или `ERROR source fetch failed` с причиной.

Отказ одного источника не останавливает остальные, но код завершения тогда 1. Ранее сохранённые снимки не трогаются.

Посмотреть результат:

```bash
ls data/archive/*/
```

```bash
grep -h ',USD,RUB,' data/archive/*/*.csv
```

### T-Bank и сертификат Минцифры

Сертификат `www.tbank.ru` выпущен **Russian Trusted Root CA** (Минцифры), которого нет в системных хранилищах macOS и Linux. Без него импорт падает с ошибкой `x509: certificate signed by unknown authority`, остальные источники при этом сохраняются.

Чтобы включить T-Bank:
1. Скачать корневой сертификат с официальной страницы Госуслуг <https://www.gosuslugi.ru/crt>.
2. Сверить его SHA-256.
3. Передать путь к файлу через `EXTRA_CA_FILE`. Сертификат добавляется только к клиенту этого процесса, системное хранилище не меняется.

```bash
openssl x509 -in russian_trusted_root_ca.pem -noout -subject -fingerprint -sha256
```

```bash
EXTRA_CA_FILE=./russian_trusted_root_ca.pem SOURCES=tbank-public ./currency-service import
```

### Отсрочка после 429

При ответе `429` в `data/backoff/<source>.txt` записывается `next_allowed_at`: по `Retry-After`, а без заголовка через 15 минут. До этого момента источник пропускается с `WARN source skipped: backoff`.

## 3. Демо-режим: `import --input` и синтетические банки

Демо-данные хранятся отдельно в `DATA_DIR=./data-demo`, все ID источников там начинаются с `demo-`. Эти записи не являются реальными предложениями банков.

Разобрать сохранённый ответ ЦБ тем же декодером, что и в сетевом режиме:

```bash
DATA_DIR=./data-demo ./currency-service import --input examples/cbr_daily.xml
```

С другим `DATA_DIR` запуск отклоняется:

```bash
./currency-service import --input examples/cbr_daily.xml; echo "exit=$?"
```

Создать два синтетических банковских снимка с текущим временем получения. Это USD/RUB, cash, Moscow: у A buy 80 / sell 90, у B buy 81 / sell 89.

```bash
DATA_DIR=./data-demo ./scripts/seed-demo-banks.sh
```

## 4. convert (задача 004)

```bash
DATA_DIR=./data-demo ./currency-service convert --from USD --to RUB --amount 100 --channel cash
```

Ожидается `8100.00 RUB | Demo Bank B (demo-bank-b) cash Moscow | rate 81.000000000000 | ... | offers 2`.

Обратное направление считается по курсу `1/sell_rate` (18 знаков, half-even):

```bash
DATA_DIR=./data-demo ./currency-service convert --from RUB --to USD --amount 8900 --channel cash
```

Ожидается `100.00 USD | Demo Bank B ...` по курсу `1/89`.

Флаги:
- `--city` (по умолчанию `Moscow`);
- `--bank` оставляет только указанный банк.

Предложения другого канала, `kind=reference`/`market_reference`, `city=unknown` и записи старше `MAX_RATE_AGE` исключаются. Если подходящих предложений нет, пишется `WARN no fresh offers` и код завершения 1.

### На реальных данных

Реальные банковские записи хранятся с `city=unknown`: применимость к Москве не подтверждена ([каталог источников](../docs/06-moscow-data-sources.md)). Поэтому московская выдача на них пустая. Посмотреть их можно явно, после `import` из сети:

```bash
./currency-service convert --from USD --to RUB --amount 100 --channel cash --city unknown
```

Здесь предложение Райффайзена, наличные.

```bash
EXTRA_CA_FILE=./russian_trusted_root_ca.pem SOURCES=tbank-public ./currency-service import && ./currency-service convert --from EUR --to RUB --amount 100 --channel account --city unknown
```

Здесь T-Bank, DepositPayments → account.

### Ошибки ввода (код 1)

```bash
./currency-service convert --from USD --to USD --amount 100 --channel cash; echo "exit=$?"
```

```bash
./currency-service convert --from USD --to RUB --amount 100 --channel atm; echo "exit=$?"
```

## Проверка источников 2026-09-28

| ID | Endpoint | HTTP | Разбор |
|---|---|---|---|
| `cbr-xml` | `https://www.cbr.ru/scripts/XML_daily.asp` | 200 XML (windows-1251) | 54 записи, USD 84.3414 |
| `cbr-json-mirror` | `https://www.cbr-xml-daily.ru/daily_json.js` | 200 JSON | 54 записи, совпадают с XML |
| `moex-iss` | `…/securities/USD000UTSTOM.json` | 200 JSON | 1 запись, WAPRICE режима CETS |
| `tbank-public` | `https://www.tbank.ru/api/v1/currency_rates/` | TLS-ошибка без `EXTRA_CA_FILE`; с корнем Минцифры 200 JSON | 9 записей DepositPayments, USD buy 81.85 / sell 88.25 |
| `raiffeisen-public` | `…/oapi/currency_rate/get/?…source=CASH…` | 200 JSON | 2 записи (USD, EUR), USD buy 79 / sell 97 |

Старый адрес `https://cbr.ru` из решения задачи 003 отдаёт HTML-страницу, а не XML; из Go-клиента он возвращал `403`. В коде используется `https://www.cbr.ru/scripts/XML_daily.asp`.

## Структура

```text
cmd/currency-service/     main.go (конфигурация, логгер, выбор команды) + calculate.go, import.go, convert.go
internal/domain/          RateRecord, Snapshot, инварианты, выбор предложений, decimal-арифметика
internal/config/          переменные окружения
internal/logging/         JSON-логгер slog
internal/source/          Source, реестр SOURCES, транспорт, адаптеры
internal/archive/         CSV-снимки (атомарная запись, обратное чтение), backoff
internal/importer/        обход источников, логирование результата, отсрочки
internal/application/     convert и best поверх архива (для CLI и будущего HTTP)
scripts/                  seed-demo-banks.sh
examples/                 cbr_daily.xml — сохранённый ответ ЦБ (2026-09-28) для import --input
```

Правила зависимостей описаны в [docs/architecture/00-decision.md](../docs/architecture/00-decision.md).

Новый источник добавляется так:
1. Файл `internal/source/<name>.go` с функцией `parse<Name>(io.Reader, Meta)`.
2. Строка в `registry` в `internal/source/source.go`.
