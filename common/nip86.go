package common

import (
	"context"
	"fmt"
	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip86"
	"slices"
)

func enableManaagementApi(relay *khatru.Relay) {
	relay.RejectFilter = append(
		relay.RejectFilter,
		func(ctx context.Context, filter nostr.Filter) (reject bool, msg string) {
			if GetString("bannedpubkey", khatru.GetAuthed(ctx)) == "" {
				return true, "restricted: you have been banned from this relay"
			}

			return false, ""
		},
	)

	relay.RejectEvent = append(
		relay.RejectEvent,
		func(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
			if GetString("bannedpubkey", khatru.GetAuthed(ctx)) == "" {
				return true, "restricted: you have been banned from this relay"
			}

			if GetString("bannedpubkey", event.PubKey) == "" {
				return true, "restricted: event author has been banned from this relay"
			}

			if GetString("bannedevent", event.ID) == "" {
				return true, "restricted: event has been banned from this relay"
			}

			return false, ""
		},
	)

	relay.ManagementAPI.RejectAPICall = append(
		relay.ManagementAPI.RejectAPICall,
		func(ctx context.Context, mp nip86.MethodParams) (reject bool, msg string) {
			if !slices.Contains(RELAY_ADMINS, khatru.GetAuthed(ctx)) {
				return true, "blocked: only relay admins can manage this relay."
			}

			return false, ""
		},
	)

	relay.ManagementAPI.BanPubKey = func(ctx context.Context, pubkey string, reason string) error {
		PutString("bannedpubkey", pubkey, reason)
		return nil
	}

	relay.ManagementAPI.ListBannedPubKeys = func(ctx context.Context) ([]nip86.PubKeyReason, error) {
		items := ListItems("bannedpubkey")
		reasons := make([]nip86.PubKeyReason, len(items))

		for pubkey, reason := range items {
			reasons = append(
				reasons,
				nip86.PubKeyReason{
					PubKey: pubkey,
					Reason: reason,
				},
			)
		}

		return reasons, nil
	}

	relay.ManagementAPI.BanEvent = func(ctx context.Context, id string, reason string) error {
		filter := nostr.Filter{
			IDs: []string{id},
		}

		ch, err := GetBackend().QueryEvents(ctx, filter)
		if err != nil {
			return fmt.Errorf("internal error: failed to query events")
		}

		for event := range ch {
			DeleteEvent(ctx, event)
		}

		PutString("bannedevent", id, reason)

		return nil
	}

	relay.ManagementAPI.ListBannedEvents = func(ctx context.Context) ([]nip86.IDReason, error) {
		items := ListItems("bannedevent")
		reasons := make([]nip86.IDReason, len(items))

		for id, reason := range items {
			reasons = append(
				reasons,
				nip86.IDReason{
					ID:     id,
					Reason: reason,
				},
			)
		}

		return reasons, nil
	}
}
