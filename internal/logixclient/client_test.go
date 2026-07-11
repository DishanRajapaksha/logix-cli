package logixclient

import "testing"

func TestParseValue(t *testing.T) {
	cases := []struct {
		typ, raw string
		want     any
	}{
		{"bool", "true", true}, {"int", "-12", int16(-12)}, {"dint", "42", int32(42)}, {"real", "1.5", float32(1.5)}, {"string", "hello", "hello"},
	}
	for _, tc := range cases {
		got, err := ParseValue(tc.typ, tc.raw)
		if err != nil {
			t.Fatalf("%s: %v", tc.typ, err)
		}
		if got != tc.want {
			t.Fatalf("%s: got %#v want %#v", tc.typ, got, tc.want)
		}
	}
}

func TestWriteRejectsAuto(t *testing.T) {
	if _, err := ParseValue("auto", "1"); err == nil {
		t.Fatal("expected error")
	}
}

func TestNormaliseStringValue(t *testing.T) {
	if got := normaliseStringValue([]byte("hello")); got != "hello" {
		t.Fatalf("got %#v", got)
	}
	got := normaliseStringValue([]any{[]byte("a"), []byte("b")})
	values, ok := got.([]string)
	if !ok || len(values) != 2 || values[0] != "a" || values[1] != "b" {
		t.Fatalf("got %#v", got)
	}
}
