package common

import (
	"github.com/fiatjaf/eventstore/badger"
	"log"
)

var backend *badger.BadgerBackend

func GetBackend() *badger.BadgerBackend {
	if backend == nil {
		backend = &badger.BadgerBackend{Path: GetDataDir("events")}
		if err := backend.Init(); err != nil {
			log.Fatal("Failed to initialize backend:", err)
		}
	}

	return backend
}
