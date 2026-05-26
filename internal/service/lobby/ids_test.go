package lobby

import "testing"

func TestNewJoinCode_LengthAndCharset(t *testing.T) {
	code, err := newJoinCode()
	if err != nil {
		t.Fatalf("newJoinCode() error: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("expected length 6, got %d (%q)", len(code), code)
	}

	for _, ch := range code {
		ok := (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
		if !ok {
			t.Fatalf("invalid char %q in join code %q", ch, code)
		}
	}
}

func TestNewGameID_NonEmpty(t *testing.T) {
	id, err := newGameID()
	if err != nil {
		t.Fatalf("newGameID() error: %v", err)
	}
	if id == "" {
		t.Fatalf("expected non-empty id")
	}
}

func TestNewPlayerID_NonEmpty(t *testing.T) {
	id, err := newPlayerID()
	if err != nil {
		t.Fatalf("newPlayerID() error: %v", err)
	}
	if id == "" {
		t.Fatalf("expected non-empty id")
	}
}
