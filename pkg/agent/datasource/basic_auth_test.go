package datasource

import (
	"testing"
)

func TestNewBasicAuthWithString(t *testing.T) {
	basicAuth := NewBasicAuth("username", "password")
	authString := basicAuth.String()
	t.Log(authString)
	bsAuth := NewBasicAuthWithString(authString)
	t.Log(bsAuth.Password, " ", bsAuth.Username)
	bsAuth = NewBasicAuthWithString("dYz8G8yaYyylVTb9jsZj0JzCi4P2Pq6S72UCBucill4=")
	t.Log(bsAuth.Password, " ", bsAuth.Username)
}
