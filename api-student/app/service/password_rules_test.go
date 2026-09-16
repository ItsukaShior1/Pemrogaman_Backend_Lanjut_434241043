package service

import "testing"

func TestValidatePassword_Strong(t *testing.T) {
	pw := "Budi123!Mantap"
	errs := ValidatePassword(pw, DefaultPasswordPolicy)

	if len(errs) != 0 {
		t.Errorf("expected tidak ada error untuk password kuat, dapat %v", errs)
	}
}

func TestValidatePassword_TooShort(t *testing.T) {
	pw := "Ab1!"
	errs := ValidatePassword(pw, DefaultPasswordPolicy)

	if len(errs) == 0 {
		t.Fatal("expected error untuk password terlalu pendek")
	}
	if _, ok := errs["length"]; !ok {
		t.Errorf("expected error length, dapat %v", errs)
	}
}

func TestValidatePassword_MissingCategories(t *testing.T) {
	pw := "budi123"
	errs := ValidatePassword(pw, DefaultPasswordPolicy)

	missing := []string{}
	for _, k := range []string{"uppercase", "symbol"} {
		if _, ok := errs[k]; ok {
			missing = append(missing, k)
		}
	}
	if len(missing) != 2 {
		t.Errorf("expected minimal 2 error (uppercase & symbol), dapat %v", errs)
	}
}

func TestValidatePassword_AllLower(t *testing.T) {
	pw := "abcdefghi"
	errs := ValidatePassword(pw, DefaultPasswordPolicy)

	for _, k := range []string{"uppercase", "digit", "symbol"} {
		if _, ok := errs[k]; !ok {
			t.Errorf("expected error %s, dapat %v", k, errs)
		}
	}
}

func TestValidatePassword_CustomPolicy(t *testing.T) {
	p := PasswordPolicy{
		MinLength:     4,
		RequireDigit:  true,
	}
	pw := "abcd1"
	errs := ValidatePassword(pw, p)

	if len(errs) != 0 {
		t.Errorf("expected tidak ada error dengan policy longgar, dapat %v", errs)
	}
}
