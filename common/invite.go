package common

import (
  "context"

	eventstore "github.com/fiatjaf/eventstore/badger"
	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

func GenerateInvite(author string) string {
	claim := randomString(8)

	PutItem("invite", claim, author)

	return claim
}

func ConsumeInvite(claim string) string {
	author := GetItem("invite", claim)

	if author != "" {
		DeleteItem("invite", claim)
	}

	return author
}

func GenerateInviteEvents(ctx context.Context, backend *eventstore.BadgerBackend, filter nostr.Filter) []*nostr.Event {
	result := make([]*nostr.Event, 0)
  pubkey := khatru.GetAuthed(ctx)
  claim := GenerateInvite(pubkey)
	event := nostr.Event{
		Kind:      AUTH_INVITE,
		CreatedAt: nostr.Now(),
		Tags: nostr.Tags{
			nostr.Tag{"claim", claim},
		},
	}

	event.Sign(RELAY_SECRET)

	return result
}
