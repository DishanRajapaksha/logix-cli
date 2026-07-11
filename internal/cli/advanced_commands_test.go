package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/DishanRajapaksha/logix-cli/internal/logixclient"
)

func (f *fakeClient) Identity() (logixclient.Identity, error) {
	return logixclient.Identity{
		VendorID:     1,
		DeviceType:   14,
		ProductCode:  1756,
		Revision:     "35.11",
		Status:       0x1234,
		SerialNumber: 0x01020304,
		ProductName:  "1756-L85E",
	}, nil
}

func TestStatusReadsIdentity(t *testing.T) {
	var out, errOut bytes.Buffer
	app := NewAppWithFactory(&out, &errOut, fakeFactory{&fakeClient{}})
	code := app.Run([]string{"status", "--config", writeConfig(t), "--format", "json"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), `"product_name": "1756-L85E"`) || !strings.Contains(out.String(), `"status_hex": "0x1234"`) {
		t.Fatalf("output=%s", out.String())
	}
}

func TestIdentifySnapshot(t *testing.T) {
	var out, errOut bytes.Buffer
	app := NewAppWithFactory(&out, &errOut, fakeFactory{&fakeClient{}})
	code := app.Run([]string{"identify", "--config", writeConfig(t), "--format", "csv"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "VENDOR_ID,DEVICE_TYPE,PRODUCT_CODE") || !strings.Contains(out.String(), "1756-L85E") {
		t.Fatalf("output=%s", out.String())
	}
}

func TestReadMultiUsesOneClientAndReadsEveryItem(t *testing.T) {
	var out, errOut bytes.Buffer
	client := &fakeClient{}
	app := NewAppWithFactory(&out, &errOut, fakeFactory{client})
	code := app.Run([]string{
		"read-multi", "--config", writeConfig(t), "--format", "json",
		"--item", "Motor.Speed=real", "--item", "Program:Main.Counter=dint",
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if client.reads != 2 {
		t.Fatalf("reads=%d", client.reads)
	}
	if !strings.Contains(out.String(), `"tag": "Motor.Speed"`) || !strings.Contains(out.String(), `"tag": "Program:Main.Counter"`) {
		t.Fatalf("output=%s", out.String())
	}
}

func TestWriteMultiIsDryRunByDefault(t *testing.T) {
	var out, errOut bytes.Buffer
	client := &fakeClient{}
	app := NewAppWithFactory(&out, &errOut, fakeFactory{client})
	code := app.Run([]string{
		"write-multi", "--config", writeConfig(t),
		"--set", "Motor.Enable=bool:true", "--set", "Recipe=dint:12",
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if len(client.writes) != 0 {
		t.Fatalf("writes=%d", len(client.writes))
	}
	if strings.Count(out.String(), "dry-run") != 2 {
		t.Fatalf("output=%s", out.String())
	}
}

func TestWriteMultiYesTransmitsEveryWrite(t *testing.T) {
	var out, errOut bytes.Buffer
	client := &fakeClient{}
	app := NewAppWithFactory(&out, &errOut, fakeFactory{client})
	code := app.Run([]string{
		"write-multi", "--config", writeConfig(t), "--yes",
		"--set", "Motor.Enable=bool:true", "--set", "Recipe=dint:12",
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if len(client.writes) != 2 {
		t.Fatalf("writes=%d", len(client.writes))
	}
}

func TestWatchMultiCountsCycles(t *testing.T) {
	var out, errOut bytes.Buffer
	client := &fakeClient{}
	app := NewAppWithFactory(&out, &errOut, fakeFactory{client})
	code := app.Run([]string{
		"watch-multi", "--config", writeConfig(t), "--format", "jsonl", "--count", "1",
		"--item", "Motor.Speed=real", "--item", "Counter=dint",
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if client.reads != 2 {
		t.Fatalf("reads=%d", client.reads)
	}
	if strings.Count(strings.TrimSpace(out.String()), "\n") != 1 {
		t.Fatalf("output=%q", out.String())
	}
}

func TestWriteSpecPreservesStringPunctuation(t *testing.T) {
	spec, err := parseWriteSpec("Program:Main.Message=string:a=b:c")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Tag != "Program:Main.Message" || spec.Value != "a=b:c" {
		t.Fatalf("spec=%#v", spec)
	}
}

func TestReadSpecSupportsProgramScopedTag(t *testing.T) {
	spec, err := parseReadSpec("Program:Main.Counter=dint:4")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Tag != "Program:Main.Counter" || spec.Type != "dint" || spec.Elements != 4 {
		t.Fatalf("spec=%#v", spec)
	}
}
