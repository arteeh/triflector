package common

import (
	"context"
	"log"
	"slices"

	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

// RejectFilter

func RejectFilter(ctx context.Context, filter nostr.Filter) (reject bool, msg string) {
	if RELAY_RESTRICT_USER {
		pubkey := khatru.GetAuthed(ctx)

		if pubkey == "" {
			return true, "auth-required: authentication is required for access"
		}

		if !HasAccess(pubkey) {
			return true, "restricted: you are not a member of this relay"
		}
	}

	return false, ""
}

// QueryEvents

func QueryEvents(ctx context.Context, filter nostr.Filter) (chan *nostr.Event, error) {
	ch := make(chan *nostr.Event)
	pubkey := khatru.GetAuthed(ctx)

	stripSignature := func(event *nostr.Event) *nostr.Event {
		if RELAY_STRIP_SIGNATURES && !slices.Contains(RELAY_ADMINS, pubkey) {
			event.Sig = ""
		}

		return event
	}

	go func() {
		defer close(ch)

		if RELAY_ENABLE_GROUPS && slices.Contains(filter.Kinds, nostr.KindSimpleGroupMetadata) {
			for _, event := range GenerateGroupMetadataEvents(ctx, filter) {
				ch <- stripSignature(event)
			}
		}

		if RELAY_ENABLE_GROUPS && slices.Contains(filter.Kinds, nostr.KindSimpleGroupAdmins) {
			for _, event := range GenerateGroupAdminsEvents(ctx, filter) {
				ch <- stripSignature(event)
			}
		}

		if RELAY_GENERATE_CLAIMS && slices.Contains(filter.Kinds, AUTH_INVITE) {
			for _, event := range GenerateInviteEvents(ctx, filter) {
				ch <- stripSignature(event)
			}
		}

		upstream, err := GetBackend().QueryEvents(ctx, filter)

		if err != nil {
			log.Println(err)
		}

		for event := range upstream {
			g := GetGroupFromEvent(event)

			if g == nil || !g.Private || IsGroupMember(ctx, g.Address.ID, pubkey) {
				ch <- stripSignature(event)
			}
		}
	}()

	return ch, nil
}

// RejectEvent

func RejectEvent(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
	pubkey := khatru.GetAuthed(ctx)

	recipientAuthKinds := []int{
		nostr.KindZap,
		1059,
	}

	// For zap receipts and gift wraps, authorize the recipient. For everything else, make sure the
	// authenticated user is the same as the event author
	if slices.Contains(recipientAuthKinds, event.Kind) {
		recipientTag := event.Tags.GetFirst([]string{"p"})

		if recipientTag != nil {
			pubkey = recipientTag.Value()
		}
	} else if pubkey != event.PubKey {
		return true, "restricted: you cannot publish events on behalf of others"
	}

	// Auth is always required to publish events
	if pubkey == "" {
		return true, "auth-required: authentication is required for access"
	}

	// Check both restrict settings since they're the same here
	if (RELAY_RESTRICT_USER || RELAY_RESTRICT_AUTHOR) && !HasAccess(pubkey) {
		return true, "restricted: you are not a member of this relay"
	}

	// Group-level access

	h := GetGroupIDFromEvent(event)
	g := GetGroup(h)

	groupMetaKinds := []int{
		nostr.KindSimpleGroupMetadata,
		nostr.KindSimpleGroupAdmins,
		nostr.KindSimpleGroupMembers,
		nostr.KindSimpleGroupRoles,
	}

	groupAdminKinds := []int{
		nostr.KindSimpleGroupPutUser,
		nostr.KindSimpleGroupRemoveUser,
		nostr.KindSimpleGroupEditMetadata,
		nostr.KindSimpleGroupDeleteEvent,
		nostr.KindSimpleGroupCreateGroup,
		nostr.KindSimpleGroupDeleteGroup,
	}

	groupRequestKinds := []int{
		nostr.KindSimpleGroupJoinRequest,
		nostr.KindSimpleGroupLeaveRequest,
	}

	groupKinds := slices.Concat(groupAdminKinds, groupRequestKinds)

	if slices.Contains(groupMetaKinds, event.Kind) {
		return true, "invalid: group metadata cannot be set directly"
	}

	if slices.Contains(groupAdminKinds, event.Kind) {
		if !RELAY_ENABLE_GROUPS {
			return true, "invalid: group events not accepted on this relay"
		}

		if !slices.Contains(RELAY_ADMINS, pubkey) {
			return true, "restricted: only relay admins can manage groups"
		}
	}

	if event.Kind == nostr.KindSimpleGroupJoinRequest {
		if !RELAY_ENABLE_GROUPS {
			return true, "invalid: group events not accepted on this relay"
		}

		if IsGroupMember(ctx, h, pubkey) {
			return true, "duplicate: already a member"
		}
	}

	if event.Kind == nostr.KindSimpleGroupLeaveRequest {
		if !RELAY_ENABLE_GROUPS {
			return true, "invalid: group events not accepted on this relay"
		}

		if !IsGroupMember(ctx, h, pubkey) {
			return true, "duplicate: not currently a member"
		}
	}

	if event.Kind == nostr.KindSimpleGroupCreateGroup {
		if !RELAY_ENABLE_GROUPS {
			return true, "invalid: group events not accepted on this relay"
		}

		if h == "" {
			return true, "invalid: invalid group ID"
		}

		if g != nil {
			return true, "invalid: that group already exists"
		}
	} else if slices.Contains(groupKinds, event.Kind) || h != "" {
		if !RELAY_ENABLE_GROUPS {
			return true, "invalid: group events not accepted on this relay"
		}

		if g == nil {
			return true, "invalid: unknown group"
		}

		if !slices.Contains(groupRequestKinds, event.Kind) && g.Closed && !IsGroupMember(ctx, h, pubkey) {
			return true, "restricted: you are not a member of this group"
		}
	}

	return false, ""
}

// SaveEvent

func SaveEvent(ctx context.Context, event *nostr.Event) error {
	return GetBackend().SaveEvent(ctx, event)
}

// OnEventSaved

func OnEventSaved(ctx context.Context, event *nostr.Event) {
	if event.Kind == nostr.KindSimpleGroupJoinRequest && GROUP_AUTO_JOIN {
		putUserEvent := MakePutUserEvent(event)

		if err := GetBackend().SaveEvent(ctx, putUserEvent); err != nil {
			log.Println(err)
		} else {
			GetRelay().BroadcastEvent(putUserEvent)
		}
	}

	if event.Kind == nostr.KindSimpleGroupLeaveRequest && GROUP_AUTO_LEAVE {
		removeUserEvent := MakeRemoveUserEvent(event)

		if err := GetBackend().SaveEvent(ctx, removeUserEvent); err != nil {
			log.Println(err)
		} else {
			GetRelay().BroadcastEvent(removeUserEvent)
		}
	}

	if event.Kind == nostr.KindSimpleGroupCreateGroup {
		HandleCreateGroup(event)
	}

	if event.Kind == nostr.KindSimpleGroupEditMetadata {
		HandleEditMetadata(event)
	}

	if event.Kind == nostr.KindSimpleGroupDeleteGroup {
		HandleDeleteGroup(event)
	}
}

// DeleteEvent

func DeleteEvent(ctx context.Context, event *nostr.Event) error {
	return GetBackend().DeleteEvent(ctx, event)
}
