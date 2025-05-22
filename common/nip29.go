package common

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

const (
	KIND_CREATE_GROUP = 9007
	KIND_EDIT_GROUP   = 9008
)

type Group struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	About       string                 `json:"about"`
	Picture     string                 `json:"picture"`
	Private     bool                   `json:"private"`
	Closed      bool                   `json:"closed"`
	CreatedAt   int64                  `json:"created_at"`
	UpdatedAt   int64                  `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// GetGroup retrieves a group from the database
func GetGroup(groupID string) (*Group, error) {
	data := GetItem("groups", groupID)
	if data == "" {
		return nil, fmt.Errorf("group not found")
	}

	var group Group
	if err := json.Unmarshal([]byte(data), &group); err != nil {
		return nil, err
	}

	return &group, nil
}

// SaveGroup stores a group in the database
func SaveGroup(group *Group) error {
	data, err := json.Marshal(group)
	if err != nil {
		return err
	}

	PutItem("groups", group.ID, string(data))
	return nil
}

// DeleteGroup removes a group from the database
func DeleteGroup(groupID string) {
	DeleteItem("groups", groupID)
}

func GenerateGroupEvents(ctx context.Context, filter nostr.Filter) (chan *nostr.Event, error) {
}

// HandleCreateGroup processes a kind 9007 create-group event
func HandleCreateGroup(ctx context.Context, event *nostr.Event) error {
	// Only relay admin can create groups
	if event.PubKey != RELAY_ADMIN {
		return fmt.Errorf("restricted: only relay admin can create groups")
	}

	groupID := GetGroupIDFromEvent(event)

	if groupID == "" {
		return fmt.Errorf("invalid: missing h tag with group ID")
	}

	// Check if group already exists
	if _, err := GetGroup(groupID); err == nil {
		return fmt.Errorf("invalid: group already exists")
	}

	// Extract group metadata
	name := event.Tags.GetFirst([]string{"name"})
	about := event.Tags.GetFirst([]string{"about"})
	picture := event.Tags.GetFirst([]string{"picture"})
	private := event.Tags.GetFirst([]string{"private"})
	closed := event.Tags.GetFirst([]string{"closed"})

	group := &Group{
		ID:        groupID,
		CreatedAt: event.CreatedAt.Time().Unix(),
		UpdatedAt: event.CreatedAt.Time().Unix(),
		CreatedBy: event.PubKey,
		Metadata:  make(map[string]interface{}),
	}

	if name != nil {
		group.Name = name.Value()
	}
	if about != nil {
		group.About = about.Value()
	}
	if picture != nil {
		group.Picture = picture.Value()
	}
	if private != nil {
		group.Private = private.Value() == "true"
	}
	if closed != nil {
		group.Closed = closed.Value() == "true"
	}

	if err := SaveGroup(group); err != nil {
		return fmt.Errorf("error: failed to save group: %v", err)
	}

	return nil
}

func HandleEditGroup(ctx context.Context, event *nostr.Event) error {
	// Only relay admin can edit groups
	if event.PubKey != RELAY_ADMIN {
		return fmt.Errorf("restricted: only relay admin can edit groups")
	}

	groupID := GetGroupIDFromEvent(event)

	if groupID == "" {
		return fmt.Errorf("invalid: missing h tag with group ID")
	}

	// Get existing group
	group, err := GetGroup(groupID)
	if err != nil {
		return fmt.Errorf("invalid: group not found")
	}

	// Update group metadata
	name := event.Tags.GetFirst([]string{"name"})
	about := event.Tags.GetFirst([]string{"about"})
	picture := event.Tags.GetFirst([]string{"picture"})
	private := event.Tags.GetFirst([]string{"private"})
	closed := event.Tags.GetFirst([]string{"closed"})

	if name != nil {
		group.Name = name.Value()
	}
	if about != nil {
		group.About = about.Value()
	}
	if picture != nil {
		group.Picture = picture.Value()
	}
	if private != nil {
		group.Private = private.Value() == "true"
	}
	if closed != nil {
		group.Closed = closed.Value() == "true"
	}

	group.UpdatedAt = event.CreatedAt.Time().Unix()

	if err := SaveGroup(group); err != nil {
		return fmt.Errorf("error: failed to update group: %v", err)
	}

	return nil
}

// GetGroupIDFromEvent extracts the group ID from an event's h tag
func GetGroupIDFromEvent(event *nostr.Event) string {
	hTag := event.Tags.GetFirst([]string{"h"})
	if hTag == nil {
		return ""
	}
	return hTag.Value()
}
