package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip29"
)

func GetGroup(h string) *nip29.Group {
	var group nip29.Group

	if h == "" {
		return nil
	}

	data := GetItem("group", h)

	if err := json.Unmarshal(data, &group); err != nil {
		return nil
	}

	return &group
}

func PutGroup(group *nip29.Group) {
	data, err := json.Marshal(group)
	if err != nil {
		log.Println(err)
	} else {
		PutItem("group", group.Address.ID, data)
	}
}

func DeleteGroup(h string) {
	DeleteItem("group", h)
}

func ListGroups() []*nip29.Group {
	var groups []*nip29.Group

	for _, item := range ListItems("group") {
		var group nip29.Group

		err := json.Unmarshal([]byte(item), &group)
		if err != nil {
			log.Printf("Failed to unmarshal group %v %s", err, item)
			continue
		}

		groups = append(groups, &group)
	}

	return groups
}

func MakeGroup(h string) *nip29.Group {
	qualifiedID := fmt.Sprintf("%s'%s", RELAY_URL, h)
	group, err := nip29.NewGroup(qualifiedID)
	if err != nil {
		log.Printf("Failed to create group with qualified ID %s", qualifiedID)
		return nil
	}

	return &group
}

func GetGroupIDFromEvent(event *nostr.Event) string {
	hTag := event.Tags.GetFirst([]string{"h"})
	if hTag == nil {
		return ""
	}

	return hTag.Value()
}

func GetGroupFromEvent(event *nostr.Event) *nip29.Group {
	return GetGroup(GetGroupIDFromEvent(event))
}

func IsGroupMember(ctx context.Context, h string, pubkey string) bool {
	filter := nostr.Filter{
		Kinds: []int{nostr.KindSimpleGroupPutUser, nostr.KindSimpleGroupRemoveUser},
		Tags: nostr.TagMap{
			"p": []string{pubkey},
			"h": []string{h},
		},
	}

	events, err := GetBackend().QueryEvents(ctx, filter)

	if err != nil {
		log.Println(err)
	}

	for event := range events {
		if event.Kind == nostr.KindSimpleGroupPutUser {
			return true
		}

		if event.Kind == nostr.KindSimpleGroupRemoveUser {
			return false
		}
	}

	return false
}

func HandleCreateGroup(event *nostr.Event) {
	group := MakeGroup(GetGroupIDFromEvent(event))

	if group != nil {
		PutGroup(group)
	}
}

func HandleEditMetadata(event *nostr.Event) {
	group := GetGroupFromEvent(event)

	if group == nil {
		group = MakeGroup(GetGroupIDFromEvent(event))
	}

	group.LastMetadataUpdate = event.CreatedAt
	group.Name = group.Address.ID

	if tag := event.Tags.GetFirst([]string{"name", ""}); tag != nil {
		group.Name = (*tag)[1]
	}
	if tag := event.Tags.GetFirst([]string{"about", ""}); tag != nil {
		group.About = (*tag)[1]
	}
	if tag := event.Tags.GetFirst([]string{"picture", ""}); tag != nil {
		group.Picture = (*tag)[1]
	}

	if tag := event.Tags.GetFirst([]string{"private"}); tag != nil {
		group.Private = true
	}
	if tag := event.Tags.GetFirst([]string{"closed"}); tag != nil {
		group.Closed = true
	}

	PutGroup(group)
}

func HandleDeleteGroup(event *nostr.Event) {
	ctx := context.Background()
	id := GetGroupIDFromEvent(event)

	DeleteGroup(id)

	hFilter := nostr.Filter{
		Tags: nostr.TagMap{
			"h": []string{id},
		},
	}

	hCh, err := GetBackend().QueryEvents(ctx, hFilter)
	if err != nil {
		log.Println(err)
	} else {
		for event := range hCh {
			DeleteEvent(ctx, event)
		}
	}

	dFilter := nostr.Filter{
		Tags: nostr.TagMap{
			"d": []string{id},
		},
	}

	dCh, err := GetBackend().QueryEvents(ctx, dFilter)
	if err != nil {
		log.Println(err)
	} else {
		for event := range dCh {
			DeleteEvent(ctx, event)
		}
	}
}

func GenerateGroupMetadataEvents(ctx context.Context, filter nostr.Filter) []*nostr.Event {
	result := make([]*nostr.Event, 0)

	for _, group := range ListGroups() {
		event := group.ToMetadataEvent()

		if !filter.Matches(event) {
			continue
		}

		if err := event.Sign(RELAY_SECRET); err != nil {
			log.Println("Failed to sign metadata event", err)
		} else {
			result = append(result, event)
		}
	}

	return result
}

func GenerateGroupAdminsEvents(ctx context.Context, filter nostr.Filter) []*nostr.Event {
	result := make([]*nostr.Event, 0)

	for _, group := range ListGroups() {
		event := nostr.Event{
			Kind:      nostr.KindSimpleGroupAdmins,
			CreatedAt: nostr.Now(),
			Tags: nostr.Tags{
				nostr.Tag{"d", group.Address.ID},
			},
		}

		for _, pubkey := range RELAY_ADMINS {
			event.Tags = append(event.Tags, nostr.Tag{"p", pubkey})
		}

		if !filter.Matches(&event) {
			continue
		}

		if err := event.Sign(RELAY_SECRET); err != nil {
			log.Println("Failed to sign admins event", err)
		} else {
			result = append(result, &event)
		}
	}

	return result
}

func MakePutUserEvent(event *nostr.Event) *nostr.Event {
	putUser := nostr.Event{
		Kind:      nostr.KindSimpleGroupPutUser,
		CreatedAt: nostr.Now(),
		Tags: nostr.Tags{
			nostr.Tag{"p", event.PubKey},
			nostr.Tag{"h", GetGroupIDFromEvent(event)},
		},
	}

	if err := putUser.Sign(RELAY_SECRET); err != nil {
		log.Println(err)
	}

	return &putUser
}

func MakeRemoveUserEvent(event *nostr.Event) *nostr.Event {
	removeUser := nostr.Event{
		Kind:      nostr.KindSimpleGroupRemoveUser,
		CreatedAt: nostr.Now(),
		Tags: nostr.Tags{
			nostr.Tag{"p", event.PubKey},
			nostr.Tag{"h", GetGroupIDFromEvent(event)},
		},
	}

	if err := removeUser.Sign(RELAY_SECRET); err != nil {
		log.Println(err)
	}

	return &removeUser
}
