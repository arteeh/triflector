package common

import (
	"context"
	"slices"

	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

// RejectFilter

func RejectFilter(ctx context.Context, filter nostr.Filter) (reject bool, msg string) {
	pubkey := khatru.GetAuthed(ctx)

	if pubkey == "" {
		return true, "auth-required: authentication is required for access"
	}

	if AUTH_RESTRICT_USER && !HasAccess(pubkey) {
		return true, "restricted: you are not a memeber of this relay"
	}

	return false, ""
}

// QueryEvents

func QueryEvents(ctx context.Context, filter nostr.Filter) (chan *nostr.Event, error) {
  ch := make(chan *nostr.Event)

	go func() {
		defer close(ch)

		if slices.Contains(filter.Kinds, nostr.KindSimpleGroupMetadata) {
			for _, evt := range GenerateGroupMetadataEvents(ctx, backend, filter) {
				ch <- evt
			}
		}

		if GENERATE_CLAIMS && slices.Contains(filter.Kinds, AUTH_INVITE) {
			for _, evt := range GenerateInviteEvents(ctx, backend, filter) {
				ch <- evt
			}
		}

  	if upstream, err := backend.QueryEvents(ctx, filter); err != nil {
    	for evt := range upstream {
      	ch <- evt
    	}
  	}
	}()

	return ch, nil
}

// RejectEvent

func RejectEvent(ctx context.Context, evt *nostr.Event) (reject bool, msg string) {
	pubkey := khatru.GetAuthed(ctx)

	if pubkey == "" {
		return true, "auth-required: authentication is required for access"
	}

  // Process relay-level join requests first
	if evt.Kind == AUTH_JOIN && evt.PubKey == pubkey {
  	tag := evt.Tags.GetFirst([]string{"claim"})

  	if tag != nil {
    	claim := tag.Value()

    	if IsValidClaim(claim) || HasAccess(ConsumeInvite(claim)) {
    		AddUserClaim(evt.PubKey, claim)
    	}

  		if !HasAccess(pubkey) {
  			return true, "restricted: failed to validate invite code"
  		}
  	}
	}

  // Restrict based on auth user
	if AUTH_RESTRICT_USER && !HasAccess(pubkey) {
		return true, "restricted: you are not a memeber of this relay"
	}

  // Restrict based on event author
	if AUTH_RESTRICT_AUTHOR && !HasAccess(evt.PubKey) {
		return true, "restricted: event author is not a member of this relay"
	}

  // Group administration kinds are restricted
  groupAdminKinds := []int{
  	nostr.KindSimpleGroupPutUser,
  	nostr.KindSimpleGroupRemoveUser,
  	nostr.KindSimpleGroupEditMetadata,
  	nostr.KindSimpleGroupDeleteEvent,
  	nostr.KindSimpleGroupCreateGroup,
  	nostr.KindSimpleGroupDeleteGroup,
  }

  if slices.Contains(groupAdminKinds, evt.Kind) {
  	if !slices.Contains(RELAY_ADMINS, evt.PubKey) {
  		return true, "restricted: only relay admin can administer groups"
  	}

  	if GetGroupIDFromEvent(evt) == "" {
  		return true, "invalid: missing h tag"
  	}
  }

  // Generated events can't be published directly
  groupMetaKinds := []int{
    nostr.KindSimpleGroupMetadata,
  	nostr.KindSimpleGroupAdmins,
  	nostr.KindSimpleGroupMembers,
  	nostr.KindSimpleGroupRoles,
  }

  if slices.Contains(groupMetaKinds, evt.Kind) {
		return true, "invalid: group metadata cannot be set directly"
  }

	return false, ""
}

// SaveEvent

func SaveEvent(ctx context.Context, evt *nostr.Event) error {
  return GetBackend().SaveEvent(ctx, evt)
}

// DeleteEvent

func DeleteEvent(ctx context.Context, evt *nostr.Event) error {
  return GetBackend().DeleteEvent(ctx, evt)
}
