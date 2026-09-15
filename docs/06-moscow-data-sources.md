# Источники курсов для Москвы: инвентарь и контракты

Проверено **2026-09-09** обычными HTTPS `GET` без авторизации с хоста разработки. «Доступен» означает, что endpoint ответил `200` и возвращает машиночитаемые данные в момент проверки; это не является обещанием стабильности и не заменяет проверку условий использования перед production.

## Кандидаты среди крупных розничных банков Москвы

Это рабочий охват десяти заметных банков, а не рейтинг по активам. Таблица намеренно различает публичную страницу с курсами и API, пригодный для автоматизации.

| Банк | Публичная страница | Автоматизируемый источник | Решение для MVP |
|---|---|---|---|
| Сбербанк | https://www.sberbank.ru/ru/quotes/currencies | Документация SberTech описывает интеграционный JSON-RPC, но публичный production endpoint не найден | Не подключать без партнёрского доступа |
| ВТБ | https://www.vtb.ru/personal/platezhi-i-perevody/obmen-valyuty/ | Стабильный публичный контракт не найден | Ручная проверка; не скрейпить |
| Газпромбанк | https://www.gazprombank.ru/personal/increase/currency-exchange/ | Проверенный кандидат `.../api/v1/currency/rates` ответил `403` | Не подключать |
| Альфа-Банк | https://alfabank.ru/currency/ | Публичный документированный API курсов не найден | Не подключать без согласования |
| Совкомбанк | https://sovcombank.ru/cards/currency | Публичный документированный API курсов не найден | Не подключать без согласования |
| T-Bank | https://www.tbank.ru/about/exchange/ | `https://www.tbank.ru/api/v1/currency_rates/` ответил `200 JSON` | Адаптер включается через SOURCES |
| МКБ | https://mkb.ru/personal/currency | Публичный документированный API курсов не найден | Не подключать без согласования |
| Райффайзен Банк | https://www.raiffeisen.ru/currency_rates/ | `https://www.raiffeisen.ru/oapi/currency_rate/get/` ответил `200 JSON` | Адаптер включается через SOURCES |
| Банк ДОМ.РФ | https://domrfbank.ru/currency/ | Публичный документированный API курсов не найден | Не подключать без согласования |
| Банк Русский Стандарт | https://www.rsb.ru/currency/ | Публичный документированный API курсов не найден | Не подключать без согласования |

T-Bank и Райффайзен здесь описаны по наблюдавшимся ответам, без гарантии стабильного developer-контракта. Включение задаётся списком SOURCES. При отказе или изменении схемы importer пишет ERROR в консоль, пропускает текущий ответ и продолжает другие источники. Следующая попытка — по расписанию, с учётом Retry-After.

## Проверенные машиночитаемые источники

| ID | Источник | Endpoint | Результат | Назначение |
|---|---|---|---|---|
| `cbr-xml` | ЦБ РФ | `https://www.cbr.ru/scripts/XML_daily.asp` | `200 XML` | Официальный дневной ориентир RUB |
| `cbr-json-mirror` | CBR XML Daily | `https://www.cbr-xml-daily.ru/daily_json.js` | `200 JSON` | Удобный контрольный канал; не первичный источник |
| `moex-iss` | Московская биржа ISS | `https://iss.moex.com/iss/engines/currency/markets/selt/securities/USD000UTSTOM.json` | `200 JSON` | Рыночный reference-rate, не наличный курс банка |
| `tbank-public` | T-Bank | `https://www.tbank.ru/api/v1/currency_rates/` | `200 JSON` | Карточные/платёжные курсы, не обязательно наличные |
| `raiffeisen-public` | Райффайзен Банк | `https://www.raiffeisen.ru/oapi/currency_rate/get/?counterCurrency=RUB&source=CASH&currencies=USD,EUR` | `200 JSON` | Офисные buy/sell курсы; география в ответе не задана |
| `banks-rates` | BanksRates | `https://api.banks-rates.com/v1/banks` | `200 JSON` | Справочный источник центробанков; не московские розничные предложения |
| `er-api` | ExchangeRate-API | `https://open.er-api.com/v6/latest/USD` | `200 JSON` | Внешний контроль консистентности, не источник банковских курсов |

Банковские buy/sell в этом наборе есть у T-Bank и Райффайзен; каналы и применимость к Москве требуют отдельного подтверждения. Это не подтверждённый охват десяти московских обменников. ЦБ, MOEX и справочные API архивируются отдельно; автоматическая сверка отклонений отложена. До подтверждения географии банковская запись получает city=unknown.

## Как использовать каталог

Ответы ниже сокращены для чтения; это не полные сохранённые HTTP-ответы. Результаты проверок относятся к указанной выше дате. Не все страницы из таблицы банков проверялись напрямую. В задаче 003 подключается ЦБ XML, в задаче 004 — отдельно зеркало ЦБ JSON, MOEX, T-Bank и Райффайзен. BanksRates и ExchangeRate-API остаются справочными примерами, реализация в шесть обязательных задач не входит.

Общий контракт нормализованной записи и политика ошибок — в [документе 04](04-data-and-operations.md). Для всех адаптеров действуют одинаковые правила логирования из [задачи 002](tasks/002-console-logger.md).

## Контракты и примеры

### 1. ЦБ РФ — XML

```http
GET /scripts/XML_daily.asp HTTP/1.1
Host: www.cbr.ru
Accept: application/xml
```

```xml
<ValCurs Date="10.09.2026" name="Foreign Currency Market">
  <Valute ID="R01235">
    <NumCode>840</NumCode><CharCode>USD</CharCode>
    <Nominal>1</Nominal><Value>85,4594</Value>
  </Valute>
</ValCurs>
```

Маппинг: `CharCode → base`, `RUB → quote`, `Value / Nominal → rate`. Десятичную запятую заменить на точку до `decimal.Decimal`. Это официальный курс, не `buy/sell` банка.

### 2. CBR XML Daily — JSON mirror

```http
GET /daily_json.js HTTP/1.1
Host: www.cbr-xml-daily.ru
Accept: application/json
```

```json
{"Date":"2026-09-10T11:30:00+03:00","Valute":{"USD":{"CharCode":"USD","Nominal":1,"Value":85.4594,"Previous":86.473}}}
```

Маппинг: `Valute.*.CharCode`, `Nominal`, `Value`. Адаптер не должен быть единственным источником: это сторонний mirror официальных данных.

### 3. MOEX ISS — USD/RUB TOM

```http
GET /iss/engines/currency/markets/selt/securities/USD000UTSTOM.json HTTP/1.1
Host: iss.moex.com
Accept: application/json
```

```json
{"securities":{"columns":["SECID","LOTSIZE","FACEUNIT","PREVPRICE"],"data":[["USD000UTSTOM",1000,"USD",86.48]]},"marketdata":{"columns":["HIGH","LOW","OPEN","LAST","WAPRICE","BID","OFFER","UPDATETIME"],"data":[[86.28,85.1675,86.28,85.1675,85.5995,null,null,"20:08:03"]]}}
```

Не полагаться на порядковые индексы без `columns`: ISS отдаёт таблицы `columns + data`. Для baseline использовать `WAPRICE` или `LAST` только когда они не `null`; не превращать `BID/OFFER` биржи в банковские buy/sell.

### 4. T-Bank — публичный JSON endpoint

```http
GET /api/v1/currency_rates/ HTTP/1.1
Host: www.tbank.ru
Accept: application/json
```

```json
{"resultCode":"OK","payload":{"lastUpdate":{"milliseconds":1788974435882},"rates":[{"category":"DepositPayments","fromCurrency":{"code":840,"name":"USD"},"toCurrency":{"code":643,"name":"RUB"},"buy":80.85,"sell":93.15}]}}
```

Сохранять `category` как тип операции. Маппинг `fromCurrency.name → base`, `toCurrency.name → quote`, `buy/sell → bank rates`; семантику направления подтвердить ручной сверкой и не смешивать категории. Включение: `SOURCES=cbr-xml,tbank-public`. В MVP используется только DepositPayments → channel=account.

### 5. Райффайзен Банк — публичный JSON endpoint

```http
GET /oapi/currency_rate/get/?counterCurrency=RUB&source=CASH&currencies=USD,EUR HTTP/1.1
Host: www.raiffeisen.ru
Accept: application/json
```

```json
{"success":true,"data":{"rates":[{"code":"RUB","updatedAt":1788966000,"exchange":[{"code":"USD","rates":{"sell":{"value":97.9,"multiplier":1},"buy":{"value":79.7,"multiplier":1}}}]}]}}
```

В этом направлении `counterCurrency=RUB`, поэтому `exchange[].code` — иностранная валюта. Сохранять `source=CASH` как канал и `updatedAt` как время снимка. Поля `multiplier` учитывать до нормализации. Включение: `SOURCES=cbr-xml,raiffeisen-public`, channel=cash.

### 6–7. Контрольные JSON-источники

`GET https://api.banks-rates.com/v1/banks` возвращает список национальных банков и метаданные; `GET https://open.er-api.com/v6/latest/USD` возвращает `{base_code, time_last_update_utc, rates}`. BanksRates пишет в консоль число элементов каталога; ExchangeRate-API сохраняет справочные курсы. Ни один не участвует в ранжировании банков.

## Единые правила интеграций

Один запрос за запуск, таймаут 5 секунд. На ошибке вернуть её импортёру: он пишет одну запись в консоль и продолжает остальные источники. Никаких немедленных повторов и постоянного отключения. Для 429 действует отсрочка из Retry-After по [документу 04](04-data-and-operations.md).

Проверять обязательные поля, типы, положительность курса и номинала. Сохранять источник, канал, URL, время получения и публикации. Дальнейшее подключение выполняется в [задачах 003–004](tasks/README.md) с отдельной ручной приёмкой каждого адаптера.
