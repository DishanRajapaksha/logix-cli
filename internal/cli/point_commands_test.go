package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestPointsJSON(t *testing.T) {
	var out, errOut bytes.Buffer
	app := NewAppWithFactory(&out, &errOut, fakeFactory{&fakeClient{}})
	code := app.Run([]string{"points", "--config", writeConfig(t), "--format", "json"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), `"name": "motor_speed"`) || !strings.Contains(out.String(), `"tag": "Motor.Speed"`) {
		t.Fatalf("output=%s", out.String())
	}
}

func TestReadPointUsesConfiguredTag(t *testing.T) {
	var out, errOut bytes.Buffer
	client := &fakeClient{}
	app := NewAppWithFactory(&out, &errOut, fakeFactory{client})
	code := app.Run([]string{"read-point", "motor_speed", "--config", writeConfig(t), "--format", "json"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if client.reads != 1 || !strings.Contains(out.String(), `"tag": "Motor.Speed"`) || !strings.Contains(out.String(), `"unit": "rpm"`) {
		t.Fatalf("reads=%d output=%s", client.reads, out.String())
	}
}

func TestWritePointRejectsReadOnlyPoint(t *testing.T) {
	var out, errOut bytes.Buffer
	app := NewAppWithFactory(&out, &errOut, fakeFactory{&fakeClient{}})
	code := app.Run([]string{"write-point", "motor_speed", "--config", writeConfig(t), "--value", "5", "--yes"})
	if code != exitWriteRejected {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
}

func TestWritePointIsDryRunByDefault(t *testing.T) {
	var out, errOut bytes.Buffer
	client := &fakeClient{}
	app := NewAppWithFactory(&out, &errOut, fakeFactory{client})
	code := app.Run([]string{"write-point", "motor_enabled", "--config", writeConfig(t), "--value", "true"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if len(client.writes) != 0 || !strings.Contains(out.String(), "dry-run") {
		t.Fatalf("writes=%d output=%s", len(client.writes), out.String())
	}
}

func TestWritePointYesTransmits(t *testing.T) {
	var out, errOut bytes.Buffer
	client := &fakeClient{}
	app := NewAppWithFactory(&out, &errOut, fakeFactory{client})
	code := app.Run([]string{"write-point", "motor_enabled", "--config", writeConfig(t), "--value", "true", "--yes"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if len(client.writes) != 1 {
		t.Fatalf("writes=%d", len(client.writes))
	}
}

func TestWatchPointUsesStreamContract(t *testing.T) {
	var out, errOut bytes.Buffer
	client := &fakeClient{}
	app := NewAppWithFactory(&out, &errOut, fakeFactory{client})
	code := app.Run([]string{"watch-point", "motor_speed", "--config", writeConfig(t), "--count", "1", "--format", "jsonl"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if client.reads != 1 || !strings.Contains(out.String(), `"point":"motor_speed"`) {
		t.Fatalf("reads=%d output=%s", client.reads, out.String())
	}
}
