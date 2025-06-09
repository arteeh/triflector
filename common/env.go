package common

import (
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nbd-wtf/go-nostr"
	"log"
	"os"
	"strings"
)

var PORT string
var DATA_DIR string
var RELAY_URL string
var RELAY_NAME string
var RELAY_ICON string
var RELAY_ADMINS []string
var RELAY_SECRET string
var RELAY_SELF string
var RELAY_DESCRIPTION string
var RELAY_CLAIMS []string
var RELAY_AUTH_BACKEND string
var RELAY_WHITELIST []string
var RELAY_RESTRICT_USER bool
var RELAY_RESTRICT_AUTHOR bool
var RELAY_GENERATE_CLAIMS bool
var RELAY_ENABLE_GROUPS bool
var GROUP_AUTO_JOIN bool
var GROUP_AUTO_LEAVE bool

func SetupEnvironment() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	var env = make(map[string]string)

	for _, item := range os.Environ() {
		parts := strings.SplitN(item, "=", 2)
		env[parts[0]] = parts[1]
	}

	getEnv := func(k string, fallback ...string) (v string) {
		v = env[k]

		if v == "" && len(fallback) > 0 {
			v = fallback[0]
		}

		return v
	}

	PORT = getEnv("PORT", "3334")
	DATA_DIR = getEnv("DATA_DIR", "./data")
	RELAY_URL = getEnv("RELAY_URL", "localhost:3334")
	RELAY_NAME = getEnv("RELAY_NAME", "Frith")
	RELAY_ICON = getEnv("RELAY_ICON", "https://hbr.coracle.social/fd73de98153b615f516d316d663b413205fd2e6e53d2c6064030ab57a7685bbd.jpg")
	RELAY_ADMINS = Split(getEnv("RELAY_ADMINS", ""), ",")
	RELAY_SECRET = getEnv("RELAY_SECRET", nostr.GeneratePrivateKey())
	RELAY_SELF, _ = nostr.GetPublicKey(RELAY_SECRET)
	RELAY_DESCRIPTION = getEnv("RELAY_DESCRIPTION", "A nostr relay for hosting groups.")
	RELAY_CLAIMS = Split(getEnv("RELAY_CLAIMS", ""), ",")
	RELAY_AUTH_BACKEND = getEnv("RELAY_AUTH_BACKEND", "")
	RELAY_WHITELIST = Split(getEnv("RELAY_WHITELIST", ""), ",")
	RELAY_RESTRICT_USER = getEnv("RELAY_RESTRICT_USER", "true") == "true"
	RELAY_RESTRICT_AUTHOR = getEnv("RELAY_RESTRICT_AUTHOR", "false") == "true"
	RELAY_GENERATE_CLAIMS = getEnv("RELAY_GENERATE_CLAIMS", "false") == "true"
	RELAY_ENABLE_GROUPS = getEnv("RELAY_ENABLE_GROUPS", "false") == "true"
	GROUP_AUTO_JOIN = getEnv("GROUP_AUTO_JOIN", "false") == "true"
	GROUP_AUTO_LEAVE = getEnv("GROUP_AUTO_LEAVE", "true") == "true"
}

func GetDataDir(dir string) string {
	return fmt.Sprintf("%s/%s", DATA_DIR, dir)
}
