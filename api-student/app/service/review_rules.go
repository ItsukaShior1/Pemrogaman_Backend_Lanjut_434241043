package service

import "strings"

// NormalizeTags merapikan tag sebelum disimpan: huruf kecil, tanpa spasi di
// tepi, tanpa duplikat, urutan pertama dipertahankan.
//
// Tanpa normalisasi, "Tugas-Banyak", "tugas-banyak", dan " tugas-banyak "
// tercatat sebagai tiga tag berbeda, dan ringkasan tag terpopuler menjadi
// tidak bermakna.
func NormalizeTags(tags []string) []string {
	result := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			continue
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		result = append(result, t)
	}
	return result
}

func validateRating(rating int) (string, bool) {
	if rating < 1 || rating > 5 {
		return "harus bernilai antara 1 dan 5", false
	}
	return "", true
}

func validateComment(comment string) (string, bool) {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return "wajib diisi", false
	}
	if len(comment) < 10 {
		return "minimal 10 karakter", false
	}
	if len(comment) > 1000 {
		return "maksimal 1000 karakter", false
	}
	return "", true
}

func validateTags(tags []string) (string, bool) {
	if len(tags) > 5 {
		return "maksimal 5 item", false
	}
	for _, t := range tags {
		if len(t) < 2 {
			return "tiap tag minimal 2 karakter", false
		}
		if len(t) > 30 {
			return "tiap tag maksimal 30 karakter", false
		}
	}
	return "", true
}

func validateReplyBody(body string) (string, bool) {
	body = strings.TrimSpace(body)
	if body == "" {
		return "wajib diisi", false
	}
	if len(body) < 2 {
		return "minimal 2 karakter", false
	}
	if len(body) > 500 {
		return "maksimal 500 karakter", false
	}
	return "", true
}
