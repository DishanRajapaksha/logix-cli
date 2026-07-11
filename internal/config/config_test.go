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
	point, err := cfg.Point("MOTOR_SPEED")
	if err != nil {
		t.Fatal(err)
	}
	if point.Tag != "Motor.Speed" || point.Type != "real" || point.Elements != 1 || point.Unit != "rpm" {
		t.Fatalf("unexpected point: %#v", point)
	}
}

func TestRejectUnknownField(t *testing.T) {
	_, err := Parse("default_profile: local\nprofiles:\n  local:\n    address: 1.2.3.4\n    nonsense: true\n")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRejectDuplicatePointNamesCaseInsensitively(t *testing.T) {
	cfg := Starter()
	cfg.Points = append(cfg.Points, Point{Name: "MOTOR_SPEED", Tag: "Other", Type: "real"})
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected duplicate point error")
	}
}

func TestWritablePointRequiresExplicitType(t *testing.T) {
	cfg := Starter()
	cfg.Points = []Point{{Name: "unsafe", Tag: "Motor.Enable", Writable: true}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected writable point type error")
	}
}

func TestPointDefaultsTypeAndElements(t *testing.T) {
	cfg := Starter()
	cfg.Points = []Point{{Name: "counter", Tag: "Counter"}}
	cfg.Groups = nil
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	point, err := cfg.Point("counter")
	if err != nil {
		t.Fatal(err)
	}
	if point.Type != "auto" || point.Elements != 1 {
		t.Fatalf("unexpected defaults: %#v", point)
	}
}
