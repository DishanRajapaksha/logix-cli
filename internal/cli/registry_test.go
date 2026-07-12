package cli

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/DishanRajapaksha/industrial-cli-kit/command"
	"github.com/DishanRajapaksha/industrial-cli-kit/completion"
)

func TestRegistryMatchesDispatcher(t *testing.T) {
	dispatched := []string{
		"init-config", "validate-config", "test-connection", "status", "identify", "programs",
		"tags", "points", "groups", "read", "read-multi", "read-point", "read-group",
		"write", "write-multi", "write-point", "write-group", "watch", "watch-multi",
		"watch-point", "watch-group", "completions", "help", "version",
	}
	registered := map[string]bool{}
	for _, registeredCommand := range cliRegistry.Commands {
		if registered[registeredCommand.Name] {
			t.Fatalf("duplicate registry command %q", registeredCommand.Name)
		}
		registered[registeredCommand.Name] = true
	}
	for _, name := range dispatched {
		if !registered[name] {
			t.Errorf("dispatcher command %q is not registered", name)
		}
	}
	for name := range registered {
		found := false
		for _, candidate := range dispatched {
			if candidate == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("registered command %q is not dispatched", name)
		}
	}
}

func TestRegistryGlobalFlagsMatchNormalizer(t *testing.T) {
	for _, global := range cliRegistry.GlobalFlags {
		value := "value"
		if global.Name == "path" {
			value = ""
			if !global.AllowEmpty {
				t.Error("registry path flag does not allow an explicit empty value")
			}
		}
		args := []string{"--" + global.Name}
		if global.TakesValue {
			args = append(args, value)
		}
		args = append(args, "status")
		normalised, err := normaliseGlobalFlags(args)
		if err != nil {
			t.Errorf("registered global flag --%s is rejected: %v", global.Name, err)
			continue
		}
		if len(normalised) == 0 || normalised[0] != "status" {
			t.Errorf("normalising --%s produced %v", global.Name, normalised)
		}
	}

	got, err := normaliseGlobalFlags([]string{"--path=", "status"})
	if err != nil {
		t.Fatalf("inline empty path rejected: %v", err)
	}
	want := []string{"status", "--path", ""}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normaliseGlobalFlags(--path=) = %#v, want %#v", got, want)
	}
}

func TestRegistryNormalizerPreservesTagsPointsAndGroups(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "tag",
			args: []string{"--address", "192.0.2.10", "read", "Motor.Speed", "--type", "real"},
			want: []string{"read", "Motor.Speed", "--address", "192.0.2.10", "--type", "real"},
		},
		{
			name: "named point",
			args: []string{"--profile", "local", "write-point", "motor_enabled", "--value", "true", "--yes"},
			want: []string{"write-point", "motor_enabled", "--profile", "local", "--value", "true", "--yes"},
		},
		{
			name: "group",
			args: []string{"--format", "jsonl", "watch-group", "motor", "--duration", "30s"},
			want: []string{"watch-group", "motor", "--format", "jsonl", "--duration", "30s"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normaliseGlobalFlags(test.args)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("normaliseGlobalFlags() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestRegistryLeadingArgumentsAndSafetyFlags(t *testing.T) {
	for _, name := range []string{
		"read", "read-point", "read-group", "write", "write-point", "write-group",
		"watch", "watch-point", "watch-group", "completions",
	} {
		if got := registryCommand(t, name).LeadingArgs; got != 1 {
			t.Errorf("%s LeadingArgs=%d, want 1", name, got)
		}
	}

	for _, name := range []string{"write", "write-multi", "write-point", "write-group"} {
		registered := registryCommand(t, name)
		assertFlag(t, registered.Flags, "yes", false)
		assertFlag(t, registered.Flags, "dry-run", false)
	}
}

func TestGeneratedCompletionsContainWriteAndWatchFlags(t *testing.T) {
	var out bytes.Buffer
	if err := completion.Write(&out, "bash", cliRegistry); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"write-multi", "watch-group", "--yes", "--dry-run", "--interval", "--duration",
		"complete -F _logix_cli_completion logix-cli",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("completion output missing %q", want)
		}
	}
}

func registryCommand(t *testing.T, name string) command.Command {
	t.Helper()
	for _, registered := range cliRegistry.Commands {
		if registered.Name == name {
			return registered
		}
	}
	t.Fatalf("registry command %q not found", name)
	return command.Command{}
}

func assertFlag(t *testing.T, flags []command.Flag, name string, takesValue bool) {
	t.Helper()
	for _, flag := range flags {
		if flag.Name == name {
			if flag.TakesValue != takesValue {
				t.Fatalf("flag --%s TakesValue=%v, want %v", name, flag.TakesValue, takesValue)
			}
			return
		}
	}
	t.Fatalf("flag --%s not found", name)
}
