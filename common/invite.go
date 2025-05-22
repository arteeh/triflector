package common

import (
  "log"
  "context"
  "slices"

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

func GenerateInviteEvents(ctx context.Context, filter nostr.Filter) (chan *nostr.Event, error) {
	ch := make(chan *nostr.Event)
  pubkey := khatru.GetAuthed(ctx)

  go func() {
    if GENERATE_CLAIMS && slices.Contains(filter.Kinds, AUTH_INVITE) && HasAccess(pubkey){
      claim := GenerateInvite(pubkey)
    	event := nostr.Event{
    		Kind:      AUTH_INVITE,
    		CreatedAt: nostr.Now(),
    		Tags: nostr.Tags{
    			nostr.Tag{"claim", claim},
    		},
    	}

    	if err := event.Sign(RELAY_SECRET); err != nil {
    		log.Fatal("Failed to sign event:", err)
    	} else {
    		ch <- &event
    	}

    	close(ch)
    }
  }()

  return ch, nil
}
