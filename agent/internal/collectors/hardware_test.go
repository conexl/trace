package collectors

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectTemperatures(t *testing.T) {
	root := t.TempDir()
	zone := filepath.Join(root, "class/thermal/thermal_zone0")
	if err := os.MkdirAll(zone, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(zone, "type"), []byte("cpu\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(zone, "temp"), []byte("42500\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	sensors := collectTemperatures(root)
	if len(sensors) != 1 {
		t.Fatalf("sensors = %#v", sensors)
	}
	if sensors[0].Name != "cpu" || sensors[0].Temperature != 42.5 {
		t.Fatalf("sensor = %#v", sensors[0])
	}
}

func TestParseSMARTHealth(t *testing.T) {
	healthy, err := parseSMARTHealth("SMART overall-health self-assessment test result: PASSED")
	if err != nil || !healthy {
		t.Fatalf("parse passed = %v, %v", healthy, err)
	}
	healthy, err = parseSMARTHealth("SMART overall-health self-assessment test result: FAILED!")
	if err != nil || healthy {
		t.Fatalf("parse failed = %v, %v", healthy, err)
	}
}
