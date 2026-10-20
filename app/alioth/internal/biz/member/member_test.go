package member

import "testing"

func TestInviteTokenIsOpaqueAndHashed(t *testing.T) {
	first, err := inviteToken()
	if err != nil {
		t.Fatal(err)
	}
	second, err := inviteToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(first) < 40 || first == second {
		t.Fatalf("invite token is not random enough: %q %q", first, second)
	}
	if hashToken(first) == first {
		t.Fatal("invite token is stored in plaintext")
	}
	if optionalTokenHash("") != "" {
		t.Fatal("empty token must not produce a digest")
	}
	if optionalTokenHash(" ") != "" {
		t.Fatal("blank token must not produce a digest")
	}
	if optionalTokenHash(first) != hashToken(first) {
		t.Fatal("token digest is not stable")
	}
}
