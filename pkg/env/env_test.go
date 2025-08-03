package env

import (
	"os"
	"testing"
	"time"
)

func TestGetEnvString(t *testing.T) {
	os.Setenv("TEST_STRING", "value")
	defer os.Clearenv()

	if value := GetEnvString("TEST_STRING", "default"); value != "value" {
		t.Errorf("Expected value, got %v", value)
	}

	if value := GetEnvString("NON_EXISTENT", "default"); value != "default" {
		t.Errorf("Expected default, got %v", value)
	}
}

func TestGetEnvInt(t *testing.T) {
	os.Setenv("TEST_INT", "5")
	defer os.Clearenv()

	if value := GetEnvInt("TEST_INT", 1); value != 5 {
		t.Errorf("Expected 5, got %v", value)
	}

	if value := GetEnvInt("NON_EXISTENT", 1); value != 1 {
		t.Errorf("Expected 1, got %v", value)
	}

	os.Setenv("BAD_INT", "bad")
	if value := GetEnvInt("BAD_INT", 42); value != 42 {
		t.Errorf("Expected default 42, got %v", value)
	}
}

func TestGetEnvFloat(t *testing.T) {
	os.Setenv("TEST_FLOAT", "5.5")
	defer os.Clearenv()

	if value := GetEnvFloat("TEST_FLOAT", 1.1); value != 5.5 {
		t.Errorf("Expected 5.5, got %v", value)
	}

	if value := GetEnvFloat("NON_EXISTENT", 1.1); value != 1.1 {
		t.Errorf("Expected 1.1, got %v", value)
	}

	os.Setenv("BAD_FLOAT", "bad")
	if value := GetEnvFloat("BAD_FLOAT", 3.14); value != 3.14 {
		t.Errorf("Expected default 3.14, got %v", value)
	}
}

func TestGetEnvBool(t *testing.T) {
	os.Setenv("TEST_BOOL", "true")
	defer os.Clearenv()

	if value := GetEnvBool("TEST_BOOL", false); value != true {
		t.Errorf("Expected true, got %v", value)
	}

	if value := GetEnvBool("NON_EXISTENT", false); value != false {
		t.Errorf("Expected false, got %v", value)
	}

	os.Setenv("BAD_BOOL", "bad")
	if value := GetEnvBool("BAD_BOOL", true); value != true {
		t.Errorf("Expected default true, got %v", value)
	}
}

func TestGetEnvDuration(t *testing.T) {
	os.Setenv("TEST_DURATION", "5s")
	defer os.Clearenv()

	if value := GetEnvDuration("TEST_DURATION", 1*time.Second); value != 5*time.Second {
		t.Errorf("Expected 5s, got %v", value)
	}

	if value := GetEnvDuration("NON_EXISTENT", 2*time.Second); value != 2*time.Second {
		t.Errorf("Expected 2s, got %v", value)
	}

	os.Setenv("BAD_DURATION", "bad")
	if value := GetEnvDuration("BAD_DURATION", 3*time.Second); value != 3*time.Second {
		t.Errorf("Expected default 3s, got %v", value)
	}
}
