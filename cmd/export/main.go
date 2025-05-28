package main

import (
	"context"
	"fmt"
	"log"

	eventstore "github.com/fiatjaf/eventstore/badger"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nbd-wtf/go-nostr"
)

func main() {
	ctx := context.Background()

	events, err := common.GetBackend().QueryEvents(ctx, nostr.Filter{})
	if err != nil {
		log.Fatal("Failed to query events:", err)
	}

	for evt := range events {
		fmt.Println(evt)
	}
}
