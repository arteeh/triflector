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
	pubkey := khatru.GetAuthed(ctx)

	if pubkey == "" {
		return true, "auth-required: authentication is required for access"
	}

	if RELAY_RESTRICT_USER && !HasAccess(pubkey) {
		return true, "restricted: you are not a memeber of this relay"
	}

	return false, ""
}

// QueryEvents

func QueryEvents(ctx context.Context, filter nostr.Filter) (chan *nostr.Event, error) {
	ch := make(chan *nostr.Event)
	pubkey := khatru.GetAuthed(ctx)

	go func() {
		defer close(ch)

		if slices.Contains(filter.Kinds, nostr.KindSimpleGroupMetadata) {
			for _, event := range GenerateGroupMetadataEvents(ctx, filter) {
				ch <- event
			}
		}

		if RELAY_GENERATE_CLAIMS && slices.Contains(filter.Kinds, AUTH_INVITE) {
			for _, event := range GenerateInviteEvents(ctx, filter) {
				ch <- event
			}
		}

		upstream, err := GetBackend().QueryEvents(ctx, filter)

		if err != nil {
			log.Println(err)
		}

		for event := range upstream {
			g := GetGroupFromEvent(event)

			if g == nil || !g.Private || IsGroupMember(ctx, g.Address.ID, pubkey) {
				ch <- event
			}
		}
	}()

	return ch, nil
}

// RejectEvent

func RejectEvent(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
	pubkey := khatru.GetAuthed(ctx)

	if pubkey == "" {
		return true, "auth-required: authentication is required for access"
	}

	// Process relay-level join requests before anything else
	if event.Kind == AUTH_JOIN && event.PubKey == pubkey {
		tag := event.Tags.GetFirst([]string{"claim"})

		if tag != nil {
			claim := tag.Value()

			if IsValidClaim(claim) || HasAccess(ConsumeInvite(claim)) {
				AddUserClaim(event.PubKey, claim)
			}

			if !HasAccess(pubkey) {
				return true, "restricted: failed to validate invite code"
			}
		}
	}

	// Relay-level access

	if RELAY_RESTRICT_USER && !HasAccess(pubkey) {
		return true, "restricted: you are not a member of this relay"
	}

	if RELAY_RESTRICT_AUTHOR && !HasAccess(event.PubKey) {
		return true, "restricted: event author is not a member of this relay"
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

	if slices.Contains(groupAdminKinds, event.Kind) && !slices.Contains(RELAY_ADMINS, event.PubKey) {
		return true, "restricted: only relay admin can manage groups"
	}

	if event.Kind == nostr.KindSimpleGroupJoinRequest && IsGroupMember(ctx, h, event.PubKey) {
		return true, "duplicate: already a member"
	}

	if event.Kind == nostr.KindSimpleGroupLeaveRequest && !IsGroupMember(ctx, h, event.PubKey) {
		return true, "duplicate: not currently a member"
	}

	if event.Kind == nostr.KindSimpleGroupCreateGroup {
		if h == "" {
			return true, "invalid: invalid group ID"
		}

		if g != nil {
			return true, "invalid: that group already exists"
		}
	} else if slices.Contains(groupKinds, event.Kind) || h != "" {
		if g == nil {
			return true, "invalid: unknown group"
		}

		if !slices.Contains(groupRequestKinds, event.Kind) && g.Closed && !IsGroupMember(ctx, h, event.PubKey) {
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
