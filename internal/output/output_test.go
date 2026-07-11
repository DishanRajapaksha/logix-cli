package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestCSVStreamWritesHeaderOnce(t *testing.T) {
	var b bytes.Buffer
	stream, err := NewStream(&b, "csv", []string{"tag", "value"})
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.Write([]string{"A", "1"}, map[string]any{"tag": "A"}); err != nil {
		t.Fatal(err)
	}
	if err := stream.Write([]string{"A", "2"}, map[string]any{"tag": "A"}); err != nil {
		t.Fatal(err)
	}
	if strings.Count(b.String(), "tag,value") != 1 {
		t.Fatalf("header count: %q", b.String())
	}
}

func TestRejectJSONStream(t *testing.T) {
	if _, err := NewStream(&bytes.Buffer{}, "json", nil); err == nil {
		t.Fatal("expected error")
	}
}
