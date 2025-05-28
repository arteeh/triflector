package common

import (
	"github.com/fiatjaf/khatru"
	"sync"
)

var (
	relay     *khatru.Relay
	relayOnce sync.Once
)

func GetRelay() *khatru.Relay {
	relayOnce.Do(func() {
		relay = khatru.NewRelay()
		relay.Info.Name = RELAY_NAME
		relay.Info.Icon = RELAY_ICON
		// relay.Info.Self = RELAY_SELF
		relay.Info.PubKey = First(RELAY_ADMINS)
		relay.Info.Description = RELAY_DESCRIPTION
		relay.Info.SupportedNIPs = append(relay.Info.SupportedNIPs, 29)

		relay.OnConnect = append(relay.OnConnect, khatru.RequestAuth)
		relay.RejectFilter = append(relay.RejectFilter, RejectFilter)
		relay.QueryEvents = append(relay.QueryEvents, QueryEvents)
		relay.DeleteEvent = append(relay.DeleteEvent, DeleteEvent)
		relay.RejectEvent = append(relay.RejectEvent, RejectEvent)
		relay.StoreEvent = append(relay.StoreEvent, SaveEvent)
		relay.OnEventSaved = append(relay.OnEventSaved, OnEventSaved)
	})

	return relay
}
