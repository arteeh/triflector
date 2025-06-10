package common

import (
	"context"
	"fmt"
	"github.com/fiatjaf/khatru"
	"log"
	"slices"
	"sync"

	"github.com/nbd-wtf/go-nostr"
)

var (
	relay     *khatru.Relay
	relayOnce sync.Once
)

func GetRelay() *khatru.Relay {
	relayOnce.Do(func() {
		relay = khatru.NewRelay()
		relay.Info.Name = RELAY_NAME
		relay.Info.Icon = RELAY_ICON
		// relay.Info.Self = RELAY_SELF
		relay.Info.PubKey = First(RELAY_ADMINS)
		relay.Info.Description = RELAY_DESCRIPTION
		relay.Info.Software = "https://github.com/coracle-social/frith"
		relay.Info.Version = "v0.1.0"

		if RELAY_ENABLE_GROUPS {
			relay.Info.SupportedNIPs = append(relay.Info.SupportedNIPs, 29)
		}

		relay.OnConnect = append(relay.OnConnect, khatru.RequestAuth)
		relay.RejectFilter = append(relay.RejectFilter, RejectFilter)
		relay.QueryEvents = append(relay.QueryEvents, QueryEvents)
		relay.DeleteEvent = append(relay.DeleteEvent, DeleteEvent)
		relay.RejectEvent = append(relay.RejectEvent, RejectEvent)
		relay.StoreEvent = append(relay.StoreEvent, SaveEvent)
		relay.OnEventSaved = append(relay.OnEventSaved, OnEventSaved)

		enableManaagementApi(relay)
	})

	migrateGroups()

	return relay
}

func migrateGroups() {
	ctx := context.Background()

	log.Println("Starting group migration...")

	var ids []string

	ch, err := GetBackend().QueryEvents(ctx, nostr.Filter{
		Limit: 1000,
		Kinds: []int{nostr.KindSimpleGroupChatMessage},
	})
	if err != nil {
		log.Fatal("failed to query events", err)
	}

	for event := range ch {
		for _, tag := range event.Tags {
			if len(tag) >= 2 && tag[0] == "h" && tag[1] != "" && !slices.Contains(ids, tag[1]) {
				ids = append(ids, tag[1])
			}
		}
	}

	log.Printf("Found %d group ids", len(ids))

	for _, id := range ids {
		if GetGroup(id) == nil {
			log.Printf("Migrating group %s", id)
			err := migrateGroup(ctx, id)
			if err != nil {
				log.Printf("Error migrating group %s: %v", id, err)
			}
		} else {
			log.Printf("Skipping group %s", id)
		}
	}

	log.Println("Group migration completed")
}

func migrateGroup(ctx context.Context, id string) error {
	createEvent := &nostr.Event{
		Kind:      nostr.KindSimpleGroupCreateGroup,
		CreatedAt: nostr.Now(),
		Tags: nostr.Tags{
			nostr.Tag{"h", id},
		},
	}

	if err := createEvent.Sign(RELAY_SECRET); err != nil {
		return fmt.Errorf("failed to sign create group event: %w", err)
	}

	if err := SaveEvent(ctx, createEvent); err != nil {
		return fmt.Errorf("failed to save create group event: %w", err)
	}

	OnEventSaved(ctx, createEvent)

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

		if err := editEvent.Sign(RELAY_SECRET); err != nil {
			return fmt.Errorf("failed to sign edit metadata event: %w", err)
		}

		if err := SaveEvent(ctx, editEvent); err != nil {
			return fmt.Errorf("failed to save edit metadata event: %w", err)
		}

		OnEventSaved(ctx, editEvent)
	}

	return nil
}

func getGroupName(ctx context.Context, id string) string {
	ch, err := GetBackend().QueryEvents(ctx, nostr.Filter{
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
