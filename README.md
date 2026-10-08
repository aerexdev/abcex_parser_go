# abcex-rate-parser

Go-клиент для получения курсов USDT/RUB из стакана биржи [ABCEX](https://abcex.io).

Разработан как альтернатива парсерам Rapira после недоступности их API (октябрь 2026).

## Установка

```bash
go get github.com/aerexdev/abcex_parser_go
```

## Использование

```go
import abcex "github.com/aerexdev/abcex_parser_go"

client := abcex.New()
tick, err := client.FetchTick(context.Background())
if err != nil {
    log.Fatal(err)
}

// Цены в micro-units (×10⁶): 87_660_000 = 87.66 RUB/USDT
fmt.Printf("Ask (Green): %.2f RUB\n", float64(tick.Green)/1_000_000)
fmt.Printf("Bid (Red):   %.2f RUB\n", float64(tick.Red)/1_000_000)
fmt.Printf("Mid:         %.2f RUB\n", float64(tick.Mid())/1_000_000)
```

## API

### `Tick`

| Поле    | Описание                                    |
|---------|---------------------------------------------|
| `Green` | Best ask (лучшее предложение продажи USDT)  |
| `Red`   | Best bid (лучшее предложение покупки USDT)  |

Значения хранятся в **micro-units** (умноженные на 10⁶) для целочисленной арифметики без потери точности.

`tick.Mid()` — среднее между ask и bid.

### Endpoint

```
GET https://hub.abcex.io/api/v2/exchange/public/orderbook/depth?instrumentCode=USDTRUB
```

Аутентификация не требуется.

## Зависимости

Только стандартная библиотека Go. Нет внешних зависимостей.

## Лицензия

MIT
