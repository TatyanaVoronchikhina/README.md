# Валютный сервис на Go

Консольный сервис курсов. Он считает сумму по заданному курсу, сохраняет снимки источников в CSV и выбирает лучшее свежее банковское предложение.

Код лежит в [tanya-solution/](tanya-solution/). Логи пишутся в stdout в формате JSON, по одной записи на строку. Результат `calculate` и `convert` печатается отдельной строкой.

## Документы

| Документ | Содержание |
|---|---|
| [Сценарии и правила](docs/01-user-scenarios.md) | Термины, функции сервиса, формулы выбора курса |
| [Архитектура](docs/02-architecture.md) | Компоненты, C4 и потоки данных |
| [OpenAPI](docs/03-openapi.yaml) | HTTP-запросы, ответы и ошибки |
| [CSV и эксплуатация](docs/04-data-and-operations.md) | Форматы файлов, конфигурация, логирование |
| [Аналитика](docs/05-analytics.md) | Дневные средние и правила истории |
| [Источники](docs/06-moscow-data-sources.md) | Адреса, контракты и примеры ответов |
| [Варианты архитектуры](docs/architecture/00-decision.md) | Принятое решение и три рассмотренных варианта |

## Запуск

Нужен Go 1.26 или новее (`go version`). Интернет нужен только для `import` без `--input`.

Все команды ниже выполняются из каталога `tanya-solution/`.

### Сборка

Перейти в каталог с модулем Go.

```bash
cd tanya-solution
```

Собрать программу. `./cmd/currency-service` — это каталог с исходниками (`main.go`, `calculate.go`, `import.go`, `convert.go`). Флаг `-o currency-service` кладёт результат сборки в файл `currency-service` в текущем каталоге.

```bash
go build -o currency-service ./cmd/currency-service
```

После этой команды рядом с исходниками появляется один исполняемый файл. Убедиться, что он есть:

```bash
ls -l currency-service
```

`./currency-service` — это и есть этот файл. Точка со слэшем значит «запустить файл из текущего каталога»: без `./` оболочка ищет команду в системном `PATH` и локальный файл не находит. Само имя без `./` — имя файла, его задал флаг `-o`. В git этот файл не коммитится: он собирается заново на той машине, где запускают сервис.

Дальше через пробел пишется команда и её флаги: `calculate`, `import`, `convert` или `help`. Например, `./currency-service calculate --amount 100 --rate 80` запускает собранный файл, а не каталог исходников.

Проверить пакеты анализатором `go vet`. Эта команда файл `currency-service` не создаёт и не запускает.

```bash
go vet ./...
```

Показать список команд и флагов уже собранного файла.

```bash
./currency-service help
```

### Переменные окружения

Переменная окружения — это пара «имя = значение», которую процесс читает при старте. Сервис смотрит их один раз в начале запуска, до разбора команды. Так один и тот же файл `./currency-service` можно направить в другой каталог данных, на другие источники или на другой уровень логов, не пересобирая его. Если переменную не задать, берётся значение по умолчанию из таблицы.

Задать переменную только для одной команды: имя и значение пишутся перед `./currency-service`, в ту же строку. Следующая команда в терминале их уже не видит.

```bash
LOG_LEVEL=ERROR ./currency-service calculate --amount 100 --rate 80
```

Несколько переменных на одну команду пишутся через пробел, тоже перед именем файла.

```bash
DATA_DIR=./data-demo LOG_LEVEL=ERROR ./currency-service convert --from USD --to RUB --amount 100 --channel cash
```

Задать переменную на всю сессию терминала. Её подхватят все следующие команды в этом окне, пока сессию не закроют или переменную не снимут.

```bash
export DATA_DIR=./data-demo
export LOG_LEVEL=ERROR
./currency-service convert --from USD --to RUB --amount 100 --channel cash
```

Снять переменную с сессии и снова пользоваться значением по умолчанию.

```bash
unset DATA_DIR
unset LOG_LEVEL
```

| Переменная | По умолчанию | Зачем |
|---|---|---|
| `DATA_DIR` | `./data` | Куда писать CSV-снимки (`archive/`) и файлы отсрочки после ответа 429 (`backoff/`). Путь считается от каталога, из которого запущена команда. Для учебных снимков ставят `./data-demo`, чтобы они не смешались с курсами из сети. `import --input` принимает только каталог с именем `data-demo`. |
| `SOURCES` | `cbr-xml` | Какие источники качать в `import`, через запятую без пробелов: `cbr-xml`, `cbr-json-mirror`, `moex-iss`, `tbank-public`, `raiffeisen-public`. По умолчанию качается только ЦБ XML, чтобы один запуск не ходил сразу во все API. Пустой список — ошибка запуска. На `calculate` и `convert` не влияет: конвертация читает уже сохранённые файлы. |
| `MAX_RATE_AGE` | `30m` | Насколько старым может быть банковское предложение в `convert`. Формат как у длительности Go: `30m`, `90s`, `1h`. Ноль и неразобранная строка — ошибка запуска. Предложение старше этого порога в выбор не попадает, поэтому вчерашний CSV не выдаётся за текущий курс банка. |
| `LOG_LEVEL` | `INFO` | Какие JSON-логи печатать: `DEBUG`, `INFO`, `WARN`, `ERROR`. `ERROR` оставляет сообщения об ошибках; строка с результатом `calculate` и `convert` печатается в любом случае. Нужен, чтобы при проверке видеть только сумму или, наоборот, подробности запроса. Неизвестное значение — ошибка запуска. |
| `EXTRA_CA_FILE` | пусто | Путь к PEM-файлу с дополнительными корневыми сертификатами. Нужен для `www.tbank.ru`: его сертификат подписан Russian Trusted Root CA, которого нет в хранилище macOS и Linux. Пустое значение значит «доверять только системным корням»; остальные источники тогда работают как обычно. |

### Расчёт по заданному курсу

Посчитать `12.50 × 80.40`. Ожидается `1005.00`.

```bash
./currency-service calculate --amount 12.50 --rate 80.40
```

Посчитать `100 × 80`. Ожидается `8000.00`.

```bash
./currency-service calculate --amount 100 --rate 80
```

Посчитать `0.10 × 3`. Ожидается `0.30`. Запятая в числе допустима наравне с точкой.

```bash
./currency-service calculate --amount 0.10 --rate 3
```

Проверить отказ на нулевой сумме. Код завершения — 1. Так же отклоняются отрицательные, пустые и нечисловые значения, а у суммы — больше двух знаков после разделителя.

```bash
./currency-service calculate --amount 0 --rate 80; echo "exit=$?"
```

### Импорт курсов из сети

Скачать дневной курс ЦБ (XML) и записать снимок в `data/archive/ГГГГ-ММ-ДД/`. Это источник по умолчанию.

```bash
./currency-service import
```

Скачать четыре источника, которые открываются без дополнительного сертификата: ЦБ XML, зеркало ЦБ JSON, MOEX и Райффайзен. Отказ одного источника не останавливает остальные; код завершения тогда 1. Уже записанные снимки не перезаписываются.

```bash
SOURCES=cbr-xml,cbr-json-mirror,moex-iss,raiffeisen-public ./currency-service import
```

Показать файлы снимков за все дни.

```bash
ls data/archive/*/
```

Показать строки USD/RUB во всех сохранённых CSV.

```bash
grep -h ',USD,RUB,' data/archive/*/*.csv
```

Сертификат `www.tbank.ru` выпущен Russian Trusted Root CA (Минцифры). Его нет в системных хранилищах macOS и Linux, поэтому T-Bank без дополнительного корня завершается ошибкой `x509: certificate signed by unknown authority`. Корневой сертификат скачивается со страницы Госуслуг <https://www.gosuslugi.ru/crt> и передаётся только этому процессу.

Сверить субъект и SHA-256 скачанного PEM.

```bash
openssl x509 -in russian_trusted_root_ca.pem -noout -subject -fingerprint -sha256
```

Скачать курсы T-Bank с этим корнем и записать снимок `tbank-public`.

```bash
EXTRA_CA_FILE=./russian_trusted_root_ca.pem SOURCES=tbank-public ./currency-service import
```

При ответе `429` в `data/backoff/<source>.txt` записывается `next_allowed_at`: по заголовку `Retry-After`, а без него — через 15 минут. До этого момента источник пропускается с `WARN source skipped: backoff`.

### Демо-данные и конвертация

Демо-снимки лежат в `DATA_DIR=./data-demo`. Их идентификаторы начинаются с `demo-`. Это не курсы реальных банков.

Разобрать сохранённый ответ ЦБ тем же декодером, что и сетевой импорт, и записать снимок `demo-cbr`. Справочный курс в выборе банка не участвует.

```bash
DATA_DIR=./data-demo ./currency-service import --input examples/cbr_daily.xml
```

Убедиться, что `import --input` с каталогом данных по умолчанию отклоняется. Код завершения — 1.

```bash
./currency-service import --input examples/cbr_daily.xml; echo "exit=$?"
```

Записать два синтетических снимка USD/RUB, cash, Moscow, с текущим временем: Bank A — buy 80 / sell 90, Bank B — buy 81 / sell 89.

```bash
DATA_DIR=./data-demo ./scripts/seed-demo-banks.sh
```

Выбрать лучшее предложение и посчитать 100 USD → RUB. Ожидается `8100.00 RUB` от Demo Bank B, курс 81, два предложения.

```bash
DATA_DIR=./data-demo ./currency-service convert --from USD --to RUB --amount 100 --channel cash
```

Посчитать обратное направление 8900 RUB → USD по курсу `1/sell` (18 знаков, half-even). Ожидается `100.00 USD` от Demo Bank B, курс `1/89`.

```bash
DATA_DIR=./data-demo ./currency-service convert --from RUB --to USD --amount 8900 --channel cash
```

Оставить только Demo Bank A. Ожидается `8000.00 RUB`, курс 80, одно предложение. Имя банка сравнивается без учёта регистра.

```bash
DATA_DIR=./data-demo ./currency-service convert --from USD --to RUB --amount 100 --channel cash --bank "Demo Bank A"
```

Флаги `convert`: `--from`, `--to`, `--amount`, `--channel` (`cash`, `card`, `account`), `--city` (по умолчанию `Moscow`), `--bank`. В выбор попадают банковские записи того же канала и города не старше `MAX_RATE_AGE`. Если таких нет, пишется `WARN no fresh offers` и код завершения 1.

### Конвертация по живым банковским курсам

Скачать наличные курсы Райффайзена в `data/`.

```bash
SOURCES=raiffeisen-public ./currency-service import
```

Посчитать 100 USD → RUB по этим записям. Город в снимке — `unknown`, поэтому его нужно указать явно. Сумма зависит от текущего курса банка.

```bash
./currency-service convert --from USD --to RUB --amount 100 --channel cash --city unknown
```

Повторить ту же конвертацию с городом `Moscow` по умолчанию. Подходящих записей нет, код завершения — 1.

```bash
./currency-service convert --from USD --to RUB --amount 100 --channel cash; echo "exit=$?"
```

Скачать T-Bank и посчитать 100 EUR → RUB по каналу `account` (в ответе банка это DepositPayments). Нужен PEM из набора импорта выше.

```bash
EXTRA_CA_FILE=./russian_trusted_root_ca.pem SOURCES=tbank-public ./currency-service import
```

```bash
./currency-service convert --from EUR --to RUB --amount 100 --channel account --city unknown
```

### Ошибочный ввод конвертации

Отклонить одинаковые валюты. Код завершения — 1.

```bash
./currency-service convert --from USD --to USD --amount 100 --channel cash; echo "exit=$?"
```

Отклонить неизвестный канал.

```bash
./currency-service convert --from USD --to RUB --amount 100 --channel atm; echo "exit=$?"
```
