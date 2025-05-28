package common

import (
  "log"
	"github.com/fiatjaf/eventstore/badger"
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
