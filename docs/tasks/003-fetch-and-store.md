# 003 — Первый источник и CSV

## Цель

Реализовать получение официальных курсов ЦБ РФ и сохранение нормализованных записей в CSV.

## Зависимости и контракты

Зависит от задачи 002. [ЦБ XML: запрос и ответ](../06-moscow-data-sources.md#1-цб-рф--xml), [CSV, конфигурация и ошибки](../04-data-and-operations.md). Библиотеки: `net/http`, `encoding/xml`, `encoding/csv`, `io`, `os`; `golang.org/x/text/encoding/charmap`.

## Объём работ

1. Добавить config и команду import. Получать ЦБ XML одним GET с таймаутом 5 секунд.
2. Декодировать Windows-1251; извлекать CharCode, Nominal, Value. Десятичную запятую заменить точкой, Value разделить на Nominal один раз.
3. Добавить RateRecord по CSV-контракту. ЦБ сохранять как kind=reference без банковских buy/sell.
4. Реализовать запись снимка через временный файл, Flush, Sync, Close, rename и обратное чтение.
5. Добавить import --input <path> для разбора сохранённого XML через тот же io.Reader-декодер. В этом режиме использовать DATA_DIR=./data-demo, source=demo-cbr; при другом DATA_DIR отклонять запуск.
6. На ошибке запроса, разбора или записи возвращать ошибку в importer; записывать один ERROR и сохранять предыдущие снимки.

## Критерии готовности

- [ ] Успешный HTTP-импорт создаёт CSV с временем получения, источником и нормализованными курсами USD/EUR; чтение возвращает сохранённые значения.
- [ ] Пример Value=8500 и Nominal=100 даёт 85 за единицу; нулевой Nominal отклоняется.
- [ ] Запуск --input создаёт только demo-cbr в отдельном каталоге.
- [ ] Недоступность источника или ошибка записи даёт ERROR и код 1; предыдущий CSV сохраняется.

## Целевая древовидная архитектура проекта

```text
.
├── cmd/currency-service/main.go
├── internal/domain/
├── internal/logging/
├── internal/config/config.go
├── internal/source/cbr_xml.go
├── internal/archive/csv.go
├── internal/importer/importer.go
└── data/archive/
```
