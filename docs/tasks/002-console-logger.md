# 002 — Структурный логгер в консоль

## Цель

Реализовать единый консольный логгер для команд и последующих HTTP-обработчиков.

## Зависимости и контракты

Зависит от задачи 001. [Формат логов](../04-data-and-operations.md#структурный-логгер). Библиотеки: стандартные `log/slog`, `os`.

## Объём работ

1. Создать logging.New(level) на slog.NewJSONHandler(os.Stdout, ...); уровень из LOG_LEVEL, значения DEBUG/INFO/WARN/ERROR.
2. В main создать один *slog.Logger с service=currency-service и передавать его в операции.
3. Добавлять operation и поля контекста: error, source, records, duration_ms. Нижние слои возвращают ошибки; граница операции пишет их один раз.
4. Заменить вывод ошибок calculate на ERROR-запись. Результат расчёта оставлять отдельной строкой stdout.
5. При некорректном LOG_LEVEL вывести JSON-ошибку через запасной INFO-логгер и завершиться с кодом 1.

## Критерии готовности

- [ ] Каждая строка логгера — JSON с time, level, msg, service и operation.
- [ ] Ошибка amount вызывает одну ERROR-запись без дублей.
- [ ] LOG_LEVEL=ERROR скрывает INFO; успешный результат calculate остаётся видимым.
- [ ] Успешный расчёт содержит INFO с duration_ms; неизвестный уровень завершает процесс с кодом 1.

## Целевая древовидная архитектура проекта

```text
.
├── cmd/currency-service/main.go
├── internal/domain/conversion.go
├── internal/logging/logger.go
└── go.mod
```
