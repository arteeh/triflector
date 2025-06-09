package common

import (
	"github.com/fiatjaf/eventstore/badger"
	"log"
	"sync"
)

var (
	backend     *badger.BadgerBackend
	backendOnce sync.Once
)

func GetBackend() *badger.BadgerBackend {
	backendOnce.Do(func() {
		backend = &badger.BadgerBackend{Path: GetDataDir("events")}
		if err := backend.Init(); err != nil {
			log.Fatal(err)
		}
	})

	return backend
}
