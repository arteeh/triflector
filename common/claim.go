package common

import (
  "context"
	"slices"
	"strings"
	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

func IsValidClaim(claim string) bool {
	return slices.Contains(RELAY_CLAIMS, claim)
}

func GetUserClaims(pubkey string) []string {
  return split(GetItem("claim", pubkey), ",")
}

func AddUserClaim(pubkey string, claim string) {
	claims := GetUserClaims(pubkey)

	if !slices.Contains(claims, claim) {
		claims = append(claims, claim)

		PutItem("claim", pubkey, strings.Join(claims, ","))
	}
}

func RejectAccessRequest(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
	pubkey := khatru.GetAuthed(ctx)

	if event.Kind == AUTH_JOIN && event.PubKey == pubkey {
  	tag := event.Tags.GetFirst([]string{"claim"})

  	if tag == nil {
  		return
  	}

  	claim := tag.Value()

  	if IsValidClaim(claim) || HasAccess(ConsumeInvite(claim)) {
  		AddUserClaim(event.PubKey, claim)
  	}

		if !HasAccess(pubkey) {
			return true, "restricted: failed to validate invite code"
		}
	}

	return false, ""
}
