package data

import "testing"

func TestSecretRoundTrip(t *testing.T) {
	box, err := newSecretBox("dev-only-encrypt-key")
	if err != nil {
		t.Fatal(err)
	}
	cipher, err := box.encrypt("s3cret")
	if err != nil {
		t.Fatal(err)
	}
	if cipher == "" || cipher == "s3cret" {
		t.Fatalf("cipher = %q", cipher)
	}
	plain, err := box.decrypt(cipher)
	if err != nil {
		t.Fatal(err)
	}
	if plain != "s3cret" {
		t.Fatalf("plain = %q", plain)
	}
	empty, err := box.encrypt("")
	if err != nil || empty != "" {
		t.Fatalf("empty cipher = %q, err = %v", empty, err)
	}
}

func TestSecretRejectsEmptyKey(t *testing.T) {
	if _, err := newSecretBox("  "); err == nil {
		t.Fatal("empty key")
	}
}
