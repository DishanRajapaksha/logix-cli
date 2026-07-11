package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestGroupsJSON(t *testing.T) {
	var out, errOut bytes.Buffer
	app := NewAppWithFactory(&out, &errOut, fakeFactory{&fakeClient{}})
	code := app.Run([]string{"groups", "--config", writeConfig(t), "--format", "json"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), `"name": "motor"`) || !strings.Contains(out.String(), `"motor_speed"`) {
		t.Fatalf("output=%s", out.String())
	}
}

func TestReadGroupUsesOneConnectionAndAllPoints(t *testing.T) {
	var out, errOut bytes.Buffer
	client := &fakeClient{}
	app := NewAppWithFactory(&out, &errOut, fakeFactory{client})
	code := app.Run([]string{"read-group", "motor", "--config", writeConfig(t), "--format", "json"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if client.reads != 2 || !strings.Contains(out.String(), `"group": "motor"`) {
		t.Fatalf("reads=%d output=%s", client.reads, out.String())
	}
}

func TestWriteGroupIsDryRunByDefault(t *testing.T) {
	var out, errOut bytes.Buffer
	client := &fakeClient{}
	app := NewAppWithFactory(&out, &errOut, fakeFactory{client})
	code := app.Run([]string{"write-group", "motor", "--config", writeConfig(t), "--set", "motor_enabled=true"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if len(client.writes) != 0 || !strings.Contains(out.String(), "dry-run") {
		t.Fatalf("writes=%d output=%s", len(client.writes), out.String())
	}
}

func TestWriteGroupRejectsPointOutsideGroup(t *testing.T) {
	var out, errOut bytes.Buffer
	app := NewAppWithFactory(&out, &errOut, fakeFactory{&fakeClient{}})
	code := app.Run([]string{"write-group", "motor", "--config", writeConfig(t), "--set", "other=true"})
	if code != exitConfigError {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
}

func TestWriteGroupRejectsReadOnlyPoint(t *testing.T) {
	var out, errOut bytes.Buffer
	app := NewAppWithFactory(&out, &errOut, fakeFactory{&fakeClient{}})
	code := app.Run([]string{"write-group", "motor", "--config", writeConfig(t), "--set", "motor_speed=5", "--yes"})
	if code != exitWriteRejected {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
}

func TestWatchGroupCountsCycles(t *testing.T) {
	var out, errOut bytes.Buffer
	client := &fakeClient{}
	app := NewAppWithFactory(&out, &errOut, fakeFactory{client})
	code := app.Run([]string{"watch-group", "motor", "--config", writeConfig(t), "--count", "1", "--format", "jsonl"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if client.reads != 2 || strings.Count(strings.TrimSpace(out.String()), "\n") != 1 {
		t.Fatalf("reads=%d output=%q", client.reads, out.String())
	}
}

func TestYesAndDryRunConflict(t *testing.T) {
	var out, errOut bytes.Buffer
	app := NewAppWithFactory(&out, &errOut, fakeFactory{&fakeClient{}})
	code := app.Run([]string{"write-point", "motor_enabled", "--config", writeConfig(t), "--value", "true", "--yes", "--dry-run"})
	if code != exitConfigError {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
}

func TestWatchDurationStopsAfterFirstCycle(t *testing.T) {
	var out, errOut bytes.Buffer
	client := &fakeClient{}
	app := NewAppWithFactory(&out, &errOut, fakeFactory{client})
	code := app.Run([]string{"watch-group", "motor", "--config", writeConfig(t), "--duration", "1ns", "--format", "jsonl"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if client.reads != 2 {
		t.Fatalf("reads=%d", client.reads)
	}
}
