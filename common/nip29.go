package common

import (
	"context"
	"fmt"

	eventstore "github.com/fiatjaf/eventstore/badger"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip29"
)

// GetGroupIDFromEvent extracts the group ID from an event's h tag
func GetGroupIDFromEvent(event *nostr.Event) string {
	hTag := event.Tags.GetFirst([]string{"h"})
	if hTag == nil {
		return ""
	}

	return hTag.Value()
}

func GenerateGroupMetadataEvents(ctx context.Context, backend *eventstore.BadgerBackend, filter nostr.Filter) []*nostr.Event {
  groups := make(map[string]*nip29.Group)
	result := make([]*nostr.Event, 0)
	thisFilter := nostr.Filter{
		Kinds: []int{nostr.KindSimpleGroupCreateGroup, nostr.KindSimpleGroupDeleteGroup, nostr.KindSimpleGroupEditMetadata},
		Tags:  nostr.TagMap{},
	}

	if filter.Tags["d"] != nil {
		thisFilter.Tags["h"] = filter.Tags["d"]
	}

	events, err := backend.QueryEvents(ctx, thisFilter)
	if err != nil {
		return result
	}

	for event := range events {
		id := GetGroupIDFromEvent(event)

		if event.Kind == nostr.KindSimpleGroupDeleteGroup {
  		delete(groups, id)
		} else if _, ok := groups[id]; !ok {
  		group, err := nip29.NewGroup(fmt.Sprintf("%s'%s", RELAY_URL, id))

  		if err != nil {
    		continue
  		}

  		groups[id] = &group
		}

		if event.Kind == nostr.KindSimpleGroupEditMetadata {
  		EditMetadata(groups[id], event)
		}
	}

	for _, group := range groups {
  	event := group.ToMetadataEvent()
		event.Sign(RELAY_SECRET)

		result = append(result, event)
	}

	return result
}

func EditMetadata(group *nip29.Group, evt *nostr.Event) {
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
}
