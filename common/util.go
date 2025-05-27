package common

import (
  "strings"
	"math/rand"
)


const (
  AUTH_JOIN = 28934
  AUTH_INVITE = 28935
 //  PUT_USER = 9000
 //  REMOVE_USER = 9001
 //  EDIT_META = 9002
 //  DELETE_EVENT = 9005
 //  CREATE_GROUP = 9007
 //  DELETE_GROUP = 9008
 //  CREATE_INVITE = 9009
 //  GROUP_JOIN = 9021
 //  GROUP_LEAVE = 9022
	// GROUP_META = 39000
	// GROUP_ADMINS = 39001
	// GROUP_MEMBERS = 39002
	// GROUP_ROLES = 39003
)

func keys[K comparable, V any](m map[K]V) []K {
	ks := make([]K, len(m))

	i := 0
	for k := range m {
		ks[i] = k
		i++
	}

	return ks
}

func filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}

	return
}

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func split(s string, delim string) []string {
  if s == "" {
    return []string{}
  } else {
  	return strings.Split(s, delim)
  }
}
