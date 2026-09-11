package auth

import "testing"

func TestValidatePasswordStrength(t *testing.T) {
	if err := ValidatePassword("short"); err == nil {
		t.Fatal("expected short password to fail validation")
	}

	if err := ValidatePassword("StrongPassword123!"); err != nil {
		t.Fatalf("expected strong password to pass validation: %v", err)
	}
}
