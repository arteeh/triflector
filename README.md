# Frith

This is a relay based on [Khatru](https://github.com/fiatjaf/khatru) which implements a range of access controls. It's designed to be used with [Flotilla](https://flotilla.social) as a community relay (complete with NIP 29 support), but it can also be used outside of a community context (unlike other implementations).

## Environment

The following environment variables are optional:

- `PORT` - the port to run on
- `DATA_DIR` - the directory where you would like to store database files and media. Defaults to `./data`, and is set to `/tmp/data` when containerized.
- `RELAY_URL` - the url of your relay
- `RELAY_NAME` - the name of your relay
- `RELAY_ICON` - an icon for your relay
- `RELAY_PUBKEY` - the public key of your relay
- `RELAY_DESCRIPTION` - your relay's description
- `RELAY_CLAIMS` - a comma-separated list of claims to auto-approve for relay access
- `RELAY_AUTH_BACKEND` - a url to delegate authorization to
- `RELAY_WHITELIST` - a comma-separate list of pubkeys to allow access for
- `RELAY_RESTRICT_USER` - whether to only accept events published by authenticated users. Defaults to `true`. If `false`, no AUTH challenge will be sent.
- `RELAY_RESTRICT_AUTHOR` - whether to only accept events signed by authorized users. Defaults to `false`.
- `RELAY_GENERATE_CLAIMS` - whether to allows relay members to generate invite codes. Defaults to `false`.
- `GROUP_AUTO_JOIN` - whether relay members can join `open` groups without approval. Defaults to `false`.
- `GROUP_AUTO_LEAVE` - whether relay members can leave groups without approval. Defaults to `true`.

## Access control

Several different policies are available for granting access, described below. If _any_ of these checks passes, access will be granted via NIP 42 AUTH for both read and write.

### Pubkey whitelist

To allow a static list of pubkeys, set the `RELAY_WHITELIST` env variable to a comma-separated list of pubkeys.

### Arbitrary policy

You can dynamically allow/deny pubkey access by setting the `RELAY_AUTH_BACKEND` env variable to a URL.

The pubkey in question will be appended to this URL and a GET request will be made against it. A 200 means the key is allowed to read and write to the relay; any other status code will deny access.

For example, providing `RELAY_AUTH_BACKEND=http://example.com/check-auth?pubkey=` will result in a GET request being made to `http://example.com/check-auth?pubkey=<pubkey>`.

### Relay claims

A user may send a `kind 28934` claim event to this relay. If the `claim` tag is in the `RELAY_CLAIMS` list, the pubkey which signed the event will be granted access to the relay.

## Development

Run `go run .` to run the project. Be sure to run `go fmt .` before committing.

## Deployment

Frith can be run using an OCI container:

```sh
podman run -it \
  -p 3334:3334 \
  -v ./data:/tmp/data \
  --env-file .env \
  ghcr.io/coracle-social/frith
```
