package common

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"log"
	"strings"
	"sync"
)

var (
	database     *badger.DB
	databaseOnce sync.Once
)

func GetDatabase() *badger.DB {
	databaseOnce.Do(func() {
		var err error
		database, err = badger.Open(badger.DefaultOptions(GetDataDir("frith")))
		if err != nil {
			log.Fatal("Failed to open badger db:", err)
		}
	})
	return database
}

func PutItem(tbl string, key string, value []byte) {
	if err := GetDatabase().Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(tbl+":"+key), value)
	}); err != nil {
		fmt.Println(err)
	}
}

func GetItem(tbl string, key string) []byte {
	var result []byte
	err := GetDatabase().View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(tbl + ":" + key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			result = val
			return nil
		})
	})

	if err != nil && err != badger.ErrKeyNotFound {
		fmt.Println(err)
	}

	return result
}

func HasItem(tbl string, key string) bool {
	err := GetDatabase().View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(tbl + ":" + key))
		return err
	})
	return err == nil
}

func DeleteItem(tbl string, key string) {
	err := GetDatabase().Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(tbl + ":" + key))
	})
	if err != nil {
		fmt.Println(err)
	}
}

func ListItems(tbl string) map[string]string {
	result := make(map[string]string)

	GetDatabase().View(func(txn *badger.Txn) error {
		prefix := tbl + ":"
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			key := strings.TrimPrefix(string(item.Key()), prefix)
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			result[key] = string(val)
		}
		return nil
	})

	return result
}
