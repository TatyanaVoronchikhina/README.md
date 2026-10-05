# currency-service

Сборка и команды запуска описаны в [README репозитория](../README.md#запуск). Их выполняют из этого каталога.

## Структура

```text
cmd/currency-service/     разбор команд calculate, import, convert
internal/domain/          записи курсов, инварианты, выбор предложения, decimal-арифметика
internal/config/          переменные окружения
internal/logging/         JSON-логгер slog
internal/source/          реестр источников, транспорт, адаптеры
internal/archive/         CSV-снимки и отсрочка после 429
internal/importer/        обход источников и запись снимков
internal/application/     convert поверх архива
scripts/                  seed-demo-banks.sh
examples/                 cbr_daily.xml — сохранённый ответ ЦБ для import --input
```

Правила зависимостей описаны в [docs/architecture/00-decision.md](../docs/architecture/00-decision.md).

Новый источник добавляется двумя шагами: файл `internal/source/<name>.go` с функцией `parse<Name>(io.Reader, Meta)` и строка в `registry` в `internal/source/source.go`.
