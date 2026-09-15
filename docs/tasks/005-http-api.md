# 005 — Конвертация через HTTP

## Цель

Предоставить конвертацию и сравнение через HTTP API на базе существующей прикладной логики.

## Зависимости и контракты

Зависит от задачи 004. [OpenAPI](../03-openapi.yaml): convert, best, healthz. Библиотеки: `net/http`, `encoding/json`, `log/slog`.

## Объём работ

1. Добавить serve на http.ServeMux с адресом HTTP_ADDR.
2. Подключить GET /v1/convert и /v1/rates/best к application; использовать тот же расчёт, что в CLI.
3. Валидировать query-параметры по OpenAPI, decimal сериализовать строкой.
4. GET /healthz возвращает 200 {"status":"ok"} без чтения CSV и проверки источников.
5. Возвращать 400 при неверном вводе, 404 convert без предложения, 200 best с пустым массивом. Ошибка CSV → 500 с общим сообщением и одной подробной ERROR-записью в консоль.

## Критерии готовности

- [ ] Запрос `curl 'http://localhost:8080/v1/convert?from=USD&to=RUB&amount=100&channel=cash'` на данных задачи 004 возвращает 8100.00 у лучшего банка.
- [ ] best с limit=1 возвращает одну запись; неверный channel даёт 400.
- [ ] Проверены отсутствие предложений и 200 healthz при недоступном источнике.
- [ ] HTTP 500 не содержит локальных путей и внутренних подробностей. История реализуется в задаче 006.

## Целевая древовидная архитектура проекта

```text
.
├── cmd/currency-service/main.go
├── internal/domain/
├── internal/logging/
├── internal/config/
├── internal/source/
├── internal/archive/
├── internal/importer/
├── internal/application/convert.go
├── internal/httpapi/server.go
└── data/archive/
```
