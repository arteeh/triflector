package common

import (
	"context"
	"slices"

	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

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

  groupAdminKinds := []int{
  	nostr.KindSimpleGroupPutUser,
  	nostr.KindSimpleGroupRemoveUser,
  	nostr.KindSimpleGroupEditMetadata,
  	nostr.KindSimpleGroupDeleteEvent,
  	nostr.KindSimpleGroupCreateGroup,
  	nostr.KindSimpleGroupDeleteGroup,
  }

  if slices.Contains(groupAdminKinds, event.Kind) {
  	if event.PubKey != RELAY_ADMIN {
  		return true, "restricted: only relay admin can create groups"
  	}

  	if GetGroupIDFromEvent(event) == "" {
  		return true, "invalid: missing h tag"
  	}
  }

  groupMetaKinds := []int{
    nostr.KindSimpleGroupMetadata,
  	nostr.KindSimpleGroupAdmins,
  	nostr.KindSimpleGroupMembers,
  	nostr.KindSimpleGroupRoles,
  }

  if slices.Contains(groupMetaKinds, event.Kind) {
		return true, "invalid: group metadata cannot be set directly"
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
