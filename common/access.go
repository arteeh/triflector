package common

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

// Constants for relay-level access

const (
	AUTH_JOIN   = 28934
	AUTH_INVITE = 28935
)

// Claims defined statically and consumed by user

func IsValidClaim(claim string) bool {
	return slices.Contains(RELAY_CLAIMS, claim)
}

func GetUserClaims(pubkey string) []string {
	return Split(GetItem("claim", pubkey), ",")
}

func AddUserClaim(pubkey string, claim string) {
	claims := GetUserClaims(pubkey)

	if !slices.Contains(claims, claim) {
		claims = append(claims, claim)

		PutItem("claim", pubkey, strings.Join(claims, ","))
	}
}

// Invites issued by members

func GenerateInvite(author string) string {
	claim := RandomString(8)

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

func GenerateInviteEvents(ctx context.Context, filter nostr.Filter) []*nostr.Event {
	result := make([]*nostr.Event, 0)
	pubkey := khatru.GetAuthed(ctx)
	claim := GenerateInvite(pubkey)
	event := nostr.Event{
		Kind:      AUTH_INVITE,
		CreatedAt: nostr.Now(),
		Tags: nostr.Tags{
			nostr.Tag{"claim", claim},
		},
	}

	event.Sign(RELAY_SECRET)

	return result
}

// Access policies

func HasAccess(pubkey string) bool {
	return slices.Contains(RELAY_ADMINS, pubkey) ||
		HasAccessUsingWhitelist(pubkey) ||
		HasAccessUsingClaim(pubkey) ||
		HasAccessUsingBackend(pubkey)
}

func HasAccessUsingWhitelist(pubkey string) bool {
	return slices.Contains(RELAY_WHITELIST, pubkey)
}

func HasAccessUsingClaim(pubkey string) bool {
	return len(GetUserClaims(pubkey)) > 0
}

type BackendAccess struct {
	granted bool
	expires time.Time
}

var backend_acl = make(map[string]BackendAccess)
var backend_acl_mu sync.Mutex

func HasAccessUsingBackend(pubkey string) bool {
	backend_acl_mu.Lock()
	defer backend_acl_mu.Unlock()

	// If we don't have a backend, we're done
	if RELAY_AUTH_BACKEND == "" {
		return false
	}

	// If we have an un-expired entry, use it
	if access, ok := backend_acl[pubkey]; ok && access.expires.After(time.Now()) {
		return access.granted
	}

	// Fetch the url
	res, err := http.Get(fmt.Sprintf("%s%s", RELAY_AUTH_BACKEND, pubkey))

	// If we get a 200, consider it good
	if err == nil {
		expire_after, _ := time.ParseDuration("1m")

		backend_acl[pubkey] = BackendAccess{
			granted: res.StatusCode == 200,
			expires: time.Now().Add(expire_after),
		}
	} else {
		fmt.Println(err)
	}

	return backend_acl[pubkey].granted
}
