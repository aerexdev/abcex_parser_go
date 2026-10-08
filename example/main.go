package main

import (
	"context"
	"fmt"
	"log"

	abcex "github.com/aerexdev/abcex_parser_go"
)

func main() {
	client := abcex.New()

	tick, err := client.FetchTick(context.Background())
	if err != nil {
		log.Fatalf("fetch error: %v", err)
	}

	fmt.Printf("USDT/RUB ask (Green): %.6f\n", float64(tick.Green)/1_000_000)
	fmt.Printf("USDT/RUB bid (Red):   %.6f\n", float64(tick.Red)/1_000_000)
	fmt.Printf("Mid price:            %.6f\n", float64(tick.Mid())/1_000_000)
}
