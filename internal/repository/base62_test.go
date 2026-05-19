package repository

import "testing"

func TestDecodeMaxUint64(t *testing.T) {
	want := ^uint64(0)
	encoded := Encode(want)

	got, ok := Decode(encoded)
	if !ok {
		t.Fatalf("expected Decode(%q) to succeed", encoded)
	}
	if got != want {
		t.Fatalf("expected %d, got %d", want, got)
	}
}

func TestDecodeRejectsOverflow(t *testing.T) {
	overflow := Encode(^uint64(0)) + "0"

	got, ok := Decode(overflow)
	if ok {
		t.Fatalf("expected Decode(%q) to reject overflow, got %d", overflow, got)
	}
}
