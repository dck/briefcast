package middleware

import (
	"strings"
	"testing"
)

func TestComputeHMAC(t *testing.T) {
	t.Parallel()

	sig1 := computeHMAC("token-1", "secret")
	sig2 := computeHMAC("token-1", "secret")
	if sig1 != sig2 {
		t.Fatalf("expected deterministic signature")
	}
	if len(sig1) != 64 {
		t.Fatalf("expected hex sha256 length 64, got %d", len(sig1))
	}

	sig3 := computeHMAC("token-2", "secret")
	if sig1 == sig3 {
		t.Fatalf("expected different signatures for different messages")
	}
}

func TestGenerateUUID(t *testing.T) {
	t.Parallel()

	id1, err := generateUUID()
	if err != nil {
		t.Fatalf("generateUUID() error: %v", err)
	}
	id2, err := generateUUID()
	if err != nil {
		t.Fatalf("generateUUID() second error: %v", err)
	}
	if id1 == id2 {
		t.Fatalf("expected unique UUIDs")
	}

	parts := strings.Split(id1, "-")
	if len(parts) != 5 {
		t.Fatalf("unexpected UUID format: %q", id1)
	}
	expected := []int{8, 4, 4, 4, 12}
	for i, p := range parts {
		if len(p) != expected[i] {
			t.Fatalf("unexpected UUID part length at index %d: %q", i, p)
		}
	}
}
