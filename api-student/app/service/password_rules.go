package service

import "strings"

type PasswordPolicy struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSymbol  bool
}

var DefaultPasswordPolicy = PasswordPolicy{
	MinLength:     8,
	RequireUpper:  true,
	RequireLower:  true,
	RequireDigit:  true,
	RequireSymbol: true,
}

func ValidatePassword(pw string, p PasswordPolicy) map[string]string {
	errs := map[string]string{}

	if len(pw) < p.MinLength {
		errs["length"] = "panjang minimal " + intToStr(p.MinLength) + " karakter"
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSymbol := false

	for _, r := range pw {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		default:
			if r >= 33 && r <= 126 {
				hasSymbol = true
			}
		}
	}

	if p.RequireUpper && !hasUpper {
		errs["uppercase"] = "minimal satu huruf besar"
	}
	if p.RequireLower && !hasLower {
		errs["lowercase"] = "minimal satu huruf kecil"
	}
	if p.RequireDigit && !hasDigit {
		errs["digit"] = "minimal satu angka"
	}
	if p.RequireSymbol && !hasSymbol {
		errs["symbol"] = "minimal satu simbol"
	}

	return errs
}

func ValidateUsername(u string) map[string]string {
	errs := map[string]string{}
	u = strings.TrimSpace(u)
	if u == "" {
		errs["required"] = "wajib diisi"
		return errs
	}
	if len(u) < 3 {
		errs["length"] = "panjang minimal 3 karakter"
	}
	if len(u) > 64 {
		errs["length"] = "panjang maksimal 64 karakter"
	}
	for _, r := range u {
		isLower := r >= 'a' && r <= 'z'
		isUpper := r >= 'A' && r <= 'Z'
		isDigit := r >= '0' && r <= '9'
		isAllowed := isLower || isUpper || isDigit || r == '_' || r == '.' || r == '-'
		if !isAllowed {
			errs["chars"] = "hanya huruf, angka, underscore, titik, atau strip"
			break
		}
	}
	return errs
}

func ValidateEmail(e string) map[string]string {
	errs := map[string]string{}
	e = strings.TrimSpace(e)
	if e == "" {
		errs["required"] = "wajib diisi"
		return errs
	}
	at := strings.Index(e, "@")
	if at <= 0 || at == len(e)-1 {
		errs["format"] = "format email tidak valid"
	}
	if strings.Count(e, "@") != 1 {
		errs["format"] = "format email tidak valid"
	}
	if !strings.Contains(e[at+1:], ".") {
		errs["format"] = "format email tidak valid"
	}
	return errs
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
