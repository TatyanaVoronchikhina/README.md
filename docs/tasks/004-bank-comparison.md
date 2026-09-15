# 004 — Несколько источников и лучший курс

## Цель

Подключить пять источников и реализовать выбор лучшего сопоставимого банковского предложения.

## Зависимости и контракты

Зависит от задачи 003. [Каталог контрактов](../06-moscow-data-sources.md), [правила сравнения](../01-user-scenarios.md), [отказы и Retry-After](../04-data-and-operations.md). Библиотеки: `encoding/json`, `net/url`, `sort`, `time`; decimal.

Термины задачи: **`Source`** — интерфейс адаптера источника: метод `ID() string` возвращает идентификатор из каталога (например `cbr-xml`), метод `Fetch(ctx) ([]RateRecord, error)` выполняет один запрос и возвращает нормализованные записи (`RateRecord` определён в задаче 003). **Ключ записи** — набор полей `source + bank + city + channel + base + quote`; «последняя запись каждого ключа» означает, что для каждого такого набора берётся запись с максимальным `observed_at`, а не один последний файл на весь архив.

## Объём работ

1. Выделить Source с ID и Fetch; сохранить адаптер ЦБ XML. import перебирает SOURCES последовательно и продолжает обработку после отказа одного источника.
2. Добавить зеркало ЦБ JSON: Valute.Value / Nominal, отдельный source, kind=reference.
3. Добавить MOEX: искать значения по columns, брать WAPRICE или LAST; оба null — ошибка. kind=market_reference.
4. Добавить T-Bank: проверять resultCode, выбирать DepositPayments → account, извлекать buy/sell и lastUpdate.
5. Добавить Райффайзен: source=CASH → cash, проверять success, multiplier и updatedAt; неизвестную семантику отклонять, неподтверждённую географию отмечать unknown.
6. Добавить `convert --from USD --to RUB --amount 100 --channel cash` с city=Moscow по умолчанию. Выбирать последние записи каждого ключа, фильтровать bank, канал, город и свежесть; сортировать и рассчитывать сумму.
7. Для проверки сравнения подготовить два синтетических банковских CSV одного канала в DATA_DIR=./data-demo с source=demo-bank-a/b. Не смешивать account и cash.

## Критерии готовности

- [ ] Все пять адаптеров подключаются независимо через SOURCES. Для каждого указан результат запроса и разбора; внешняя недоступность зафиксирована отдельно.
- [ ] Для 100 USD курсы покупки 80 и 81 дают победителя с 8100.00 RUB.
- [ ] Для обратного направления используется 1/sell; устаревшие записи, другой канал, kind=reference и city=unknown исключаются из Moscow-выдачи.
- [ ] Один отказ даёт одну ERROR-запись, успешные источники сохраняются. Retry-After соблюдается по общим правилам.

## Целевая древовидная архитектура проекта

```text
.
├── cmd/currency-service/main.go
├── internal/domain/
├── internal/logging/
├── internal/config/
├── internal/source/
│   ├── source.go
│   ├── cbr_xml.go
│   ├── cbr_json.go
│   ├── moex.go
│   ├── tbank.go
│   └── raiffeisen.go
├── internal/archive/csv.go
├── internal/importer/importer.go
├── internal/application/convert.go
└── data/archive/
```
