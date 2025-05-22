package common

import (
	"context"
	"slices"

	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

func SaveEvent(ctx context.Context, event *nostr.Event) error {
  if event.Kind == KIND_CREATE_GROUP {
    return HandleCreateGroup(ctx, event)
  }

  if event.Kind == KIND_EDIT_GROUP {
    return HandleEditGroup(ctx, event)
  }

  return nil
}

func RejectEvent(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
	pubkey := khatru.GetAuthed(ctx)

	if pubkey == "" {
		return true, "auth-required: authentication is required for access"
	}

	if AUTH_RESTRICT_USER && !HasAccess(pubkey) {
		return true, "restricted: you are not a memeber of this relay"
	}

	if AUTH_RESTRICT_AUTHOR && !HasAccess(event.PubKey) {
		return true, "restricted: event author is not a member of this relay"
	}

	return false, ""
}

func RejectFilter(ctx context.Context, filter nostr.Filter) (reject bool, msg string) {
	if slices.Contains(filter.Kinds, AUTH_JOIN) {
		return true, "restricted: join events cannot be queried"
	}

	pubkey := khatru.GetAuthed(ctx)

	if pubkey == "" {
		return true, "auth-required: authentication is required for access"
	}

	if AUTH_RESTRICT_USER && !HasAccess(pubkey) {
		return true, "restricted: you are not a memeber of this relay"
	}

	return false, ""
}
