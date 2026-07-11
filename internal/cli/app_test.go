package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DishanRajapaksha/logix-cli/internal/config"
	"github.com/DishanRajapaksha/logix-cli/internal/logixclient"
)

type fakeFactory struct{ client *fakeClient }

func (f fakeFactory) New(logixclient.Options) (logixclient.Client, error) { return f.client, nil }

type fakeClient struct {
	writes []any
	reads  int
}

func (f *fakeClient) Connect() error    { return nil }
func (f *fakeClient) Disconnect() error { return nil }
func (f *fakeClient) Programs() ([]logixclient.Program, error) {
	return []logixclient.Program{{Name: "MainProgram", ID: 1}}, nil
}
func (f *fakeClient) Tags() ([]logixclient.Tag, error) {
	return []logixclient.Tag{{Name: "Motor.Speed", Type: "REAL", Instance: 2}}, nil
}
func (f *fakeClient) Read(tag, typ string, elements uint16) (any, string, error) {
	f.reads++
	return float32(12.5), "real", nil
}
func (f *fakeClient) Write(tag string, value any) error {
	f.writes = append(f.writes, value)
	return nil
}

func writeConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(config.Marshal(config.Starter())), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestReadJSON(t *testing.T) {
	var out, errOut bytes.Buffer
	client := &fakeClient{}
	app := NewAppWithFactory(&out, &errOut, fakeFactory{client})
	code := app.Run([]string{"read", "--config", writeConfig(t), "--format", "json", "--type", "real", "Motor.Speed"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), `"tag": "Motor.Speed"`) {
		t.Fatalf("output=%s", out.String())
	}
}

func TestWriteIsDryRunByDefault(t *testing.T) {
	var out, errOut bytes.Buffer
	client := &fakeClient{}
	app := NewAppWithFactory(&out, &errOut, fakeFactory{client})
	code := app.Run([]string{"write", "--config", writeConfig(t), "--type", "dint", "--value", "42", "Counter"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if len(client.writes) != 0 {
		t.Fatal("dry-run transmitted a write")
	}
	if !strings.Contains(out.String(), "dry-run") {
		t.Fatalf("output=%s", out.String())
	}
}

func TestWriteYesTransmits(t *testing.T) {
	var out, errOut bytes.Buffer
	client := &fakeClient{}
	app := NewAppWithFactory(&out, &errOut, fakeFactory{client})
	code := app.Run([]string{"write", "--config", writeConfig(t), "--type", "dint", "--value", "42", "--yes", "Counter"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if len(client.writes) != 1 {
		t.Fatalf("writes=%d", len(client.writes))
	}
}

func TestWatchRejectsJSON(t *testing.T) {
	var out, errOut bytes.Buffer
	app := NewAppWithFactory(&out, &errOut, fakeFactory{&fakeClient{}})
	code := app.Run([]string{"watch", "--config", writeConfig(t), "--format", "json", "--count", "1", "Tag"})
	if code != exitOutputError {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
}

func TestGlobalFlagsBeforeCommand(t *testing.T) {
	var out, errOut bytes.Buffer
	app := NewAppWithFactory(&out, &errOut, fakeFactory{&fakeClient{}})
	code := app.Run([]string{"--config", writeConfig(t), "--format", "json", "programs"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
}

func TestTagMayPrecedeFlags(t *testing.T) {
	var out, errOut bytes.Buffer
	app := NewAppWithFactory(&out, &errOut, fakeFactory{&fakeClient{}})
	code := app.Run([]string{"read", "Motor.Speed", "--config", writeConfig(t), "--format", "json", "--type", "real"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
}

func TestWriteRequiresValueEvenForString(t *testing.T) {
	var out, errOut bytes.Buffer
	app := NewAppWithFactory(&out, &errOut, fakeFactory{&fakeClient{}})
	code := app.Run([]string{"write", "Message", "--config", writeConfig(t), "--type", "string"})
	if code != exitConfigError {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
}

func TestGlobalFlagsBeforeTagCommand(t *testing.T) {
	var out, errOut bytes.Buffer
	app := NewAppWithFactory(&out, &errOut, fakeFactory{&fakeClient{}})
	code := app.Run([]string{"--config", writeConfig(t), "--format", "json", "read", "Motor.Speed", "--type", "real"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
}
