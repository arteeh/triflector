package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip29"
)

func GetGroup(h string) (*nip29.Group, error) {
	var group nip29.Group

	err := json.Unmarshal(GetBytes("group", h), &group)

	return &group, err
}

func PutGroup(group *nip29.Group) {
	data, err := json.Marshal(group)
	if err != nil {
		log.Println(err)
	}

	PutBytes("group", group.Address.ID, data)
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
			log.Println(err)
			continue
		}

		groups = append(groups, &group)
	}

	return groups
}

func GetGroupIDFromEvent(event *nostr.Event) string {
	hTag := event.Tags.GetFirst([]string{"h"})
	if hTag == nil {
		return ""
	}

	return hTag.Value()
}

func GetGroupFromEvent(event *nostr.Event) (*nip29.Group, error) {
	return GetGroup(GetGroupIDFromEvent(event))
}

func IsGroupMember(ctx context.Context, h string, pubkey string) bool {
	filter := nostr.Filter{
		Limit: 1,
		Kinds: []int{nostr.KindSimpleGroupPutUser, nostr.KindSimpleGroupRemoveUser},
		Tags: nostr.TagMap{
			"p": []string{pubkey},
		},
	}

	events, err := backend.QueryEvents(ctx, filter)

	if err != nil {
		log.Println(err)
	}

	for evt := range events {
		if evt.Kind == nostr.KindSimpleGroupPutUser {
			return true
		}
	}

	return false
}

func HandleCreateGroup(evt *nostr.Event) {
	group, err := nip29.NewGroup(fmt.Sprintf("%s'%s", RELAY_URL, GetGroupIDFromEvent(evt)))

	if err != nil {
		PutGroup(&group)
	}
}

func HandleEditMetadata(evt *nostr.Event) {
	group, err := GetGroupFromEvent(evt)
	if err != nil {
		log.Println(err)
		return
	}

	group.LastMetadataUpdate = evt.CreatedAt
	group.Name = group.Address.ID

	if tag := evt.Tags.GetFirst([]string{"name", ""}); tag != nil {
		group.Name = (*tag)[1]
	}
	if tag := evt.Tags.GetFirst([]string{"about", ""}); tag != nil {
		group.About = (*tag)[1]
	}
	if tag := evt.Tags.GetFirst([]string{"picture", ""}); tag != nil {
		group.Picture = (*tag)[1]
	}

	if tag := evt.Tags.GetFirst([]string{"private"}); tag != nil {
		group.Private = true
	}
	if tag := evt.Tags.GetFirst([]string{"closed"}); tag != nil {
		group.Closed = true
	}

	PutGroup(group)
}

func HandleDeleteGroup(evt *nostr.Event) {
	DeleteGroup(GetGroupIDFromEvent(evt))
}

func GenerateGroupMetadataEvents(ctx context.Context, filter nostr.Filter) []*nostr.Event {
	result := make([]*nostr.Event, 0)

	for _, group := range ListGroups() {
		event := group.ToMetadataEvent()

		if filter.Matches(event) {
			if err := event.Sign(RELAY_SECRET); err != nil {
				log.Println(err)
			}

			result = append(result, event)
		}
	}

	return result
}

func MakePutUserEvent(evt *nostr.Event) *nostr.Event {
	event := nostr.Event{
		Kind:      nostr.KindSimpleGroupPutUser,
		CreatedAt: nostr.Now(),
		Tags: nostr.Tags{
			nostr.Tag{"p", evt.PubKey},
			nostr.Tag{"h", GetGroupIDFromEvent(evt)},
		},
	}

	if err := event.Sign(RELAY_SECRET); err != nil {
		log.Println(err)
	}

	return &event
}

func MakeRemoveUserEvent(evt *nostr.Event) *nostr.Event {
	event := nostr.Event{
		Kind:      nostr.KindSimpleGroupRemoveUser,
		CreatedAt: nostr.Now(),
		Tags: nostr.Tags{
			nostr.Tag{"p", evt.PubKey},
			nostr.Tag{"h", GetGroupIDFromEvent(evt)},
		},
	}

	if err := event.Sign(RELAY_SECRET); err != nil {
		log.Println(err)
	}

	return &event
}
