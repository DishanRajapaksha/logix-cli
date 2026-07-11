package config

import "testing"

func TestStarterRoundTrip(t *testing.T) {
	cfg, err := Parse(Marshal(Starter()))
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	profile, name, err := cfg.Profile("")
	if err != nil {
		t.Fatal(err)
	}
	if name != "local" || profile.Port != 44818 || profile.Path != "1,0" {
		t.Fatalf("unexpected profile: %#v %q", profile, name)
	}
}

func TestRejectUnknownField(t *testing.T) {
	_, err := Parse("default_profile: local\nprofiles:\n  local:\n    address: 1.2.3.4\n    nonsense: true\n")
	if err == nil {
		t.Fatal("expected error")
	}
}
