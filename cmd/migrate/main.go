package main

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/nbd-wtf/go-nostr"

	"frith/common"
)

func main() {
	ctx := context.Background()

	common.SetupEnvironment()

	defer common.GetDatabase().Close()

	log.Println("Starting group migration...")

	var ids []string

	ch, err := common.GetBackend().QueryEvents(ctx, nostr.Filter{})
	if err != nil {
		log.Fatal("failed to query events", err)
	}

	for event := range ch {
		for _, tag := range event.Tags {
			if len(tag) >= 2 && tag[0] == "h" && tag[1] != "" && !slices.Contains(ids, tag[1]) && common.GetGroup(tag[1]) != nil {
				ids = append(ids, tag[1])
			}
		}
	}

	log.Printf("Found %d group ids", len(ids))

	for _, id := range ids {
		log.Printf("Processing group %s", id)
		err := processGroup(ctx, id)
		if err != nil {
			log.Printf("Error processing group %s: %v", id, err)
		}
	}

	log.Println("Group migration completed")
}

func processGroup(ctx context.Context, id string) error {
	createEvent := &nostr.Event{
		Kind:      nostr.KindSimpleGroupCreateGroup,
		CreatedAt: nostr.Now(),
		Tags: nostr.Tags{
			nostr.Tag{"h", id},
		},
	}

	if err := createEvent.Sign(common.RELAY_SECRET); err != nil {
		return fmt.Errorf("failed to sign create group event: %w", err)
	}

	if err := common.SaveEvent(ctx, createEvent); err != nil {
		return fmt.Errorf("failed to save create group event: %w", err)
	}

	common.OnEventSaved(ctx, createEvent)

	name := getGroupName(ctx, id)
	if name != "" {
		editEvent := &nostr.Event{
			Kind:      nostr.KindSimpleGroupEditMetadata,
			CreatedAt: nostr.Now(),
			Tags: nostr.Tags{
				nostr.Tag{"h", id},
				nostr.Tag{"name", name},
			},
		}

		if err := editEvent.Sign(common.RELAY_SECRET); err != nil {
			return fmt.Errorf("failed to sign edit metadata event: %w", err)
		}

		if err := common.SaveEvent(ctx, editEvent); err != nil {
			return fmt.Errorf("failed to save edit metadata event: %w", err)
		}

		common.OnEventSaved(ctx, editEvent)
	}

	return nil
}

func getGroupName(ctx context.Context, id string) string {
	ch, err := common.GetBackend().QueryEvents(ctx, nostr.Filter{
		Kinds: []int{nostr.KindSimpleGroupList},
	})
	if err != nil {
		log.Fatal("failed to query events", err)
	}

	for event := range ch {
		for _, tag := range event.Tags {
			if len(tag) >= 4 && tag[0] == "group" && tag[1] == id && tag[3] != "" {
				return tag[3]
			}
		}
	}

	return ""
}
