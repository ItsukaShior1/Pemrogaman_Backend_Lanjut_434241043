package service

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"api-student/app/model"
	"api-student/app/repository"
)

// NormalizeTechnologies merapikan daftar teknologi sebelum disimpan: huruf
// kecil, tanpa spasi di tepi, tanpa duplikat, urutan pertama dipertahankan.
//
// Tanpa normalisasi, "Go", "go", dan " GO " akan dihitung sebagai tiga
// teknologi berbeda pada statistik, padahal maksudnya sama.
func NormalizeTechnologies(tech []string) []string {
	result := make([]string, 0, len(tech))
	seen := make(map[string]struct{}, len(tech))
	for _, t := range tech {
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

func validateHeadline(headline string) (string, bool) {
	headline = strings.TrimSpace(headline)
	if headline == "" {
		return "wajib diisi", false
	}
	if len(headline) > 120 {
		return "maksimal 120 karakter", false
	}
	return "", true
}

func validateBio(bio string) (string, bool) {
	bio = strings.TrimSpace(bio)
	if bio == "" {
		return "wajib diisi", false
	}
	if len(bio) > 2000 {
		return "maksimal 2000 karakter", false
	}
	return "", true
}

func validateProjectTitle(title string) (string, bool) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "wajib diisi", false
	}
	if len(title) > 150 {
		return "maksimal 150 karakter", false
	}
	return "", true
}

func validateTechnologies(tech []string) (string, bool) {
	if len(tech) == 0 {
		return "minimal 1 teknologi", false
	}
	if len(tech) > 10 {
		return "maksimal 10 teknologi", false
	}
	for _, t := range tech {
		if len(t) > 40 {
			return "tiap teknologi maksimal 40 karakter", false
		}
	}
	return "", true
}

func validateProjectURL(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", true
	}
	if len(raw) > 500 {
		return "maksimal 500 karakter", false
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return "tautan harus berupa URL yang valid", false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "tautan harus diawali http:// atau https://", false
	}
	return "", true
}

func validateYear(year int) (string, bool) {
	if year < 2000 || year > time.Now().Year()+1 {
		return "harus berada antara 2000 dan tahun depan", false
	}
	return "", true
}

func validateCertificateName(name string) (string, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "wajib diisi", false
	}
	if len(name) > 150 {
		return "maksimal 150 karakter", false
	}
	return "", true
}

func validateCertificateIssuer(issuer string) (string, bool) {
	issuer = strings.TrimSpace(issuer)
	if issuer == "" {
		return "wajib diisi", false
	}
	if len(issuer) > 150 {
		return "maksimal 150 karakter", false
	}
	return "", true
}

// validateCertificateDate menerima format YYYY-MM-DD.
func validateCertificateDate(raw string) (string, time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "wajib diisi", time.Time{}, false
	}
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return "format harus YYYY-MM-DD", time.Time{}, false
	}
	return "", parsed, true
}

func validateProjectRequest(p model.ProjectRequest) map[string]string {
	errs := map[string]string{}
	if msg, ok := validateProjectTitle(p.Title); !ok {
		errs["title"] = msg
	}
	if msg, ok := validateTechnologies(p.Technologies); !ok {
		errs["technologies"] = msg
	}
	if msg, ok := validateProjectURL(p.URL); !ok {
		errs["url"] = msg
	}
	if msg, ok := validateYear(p.Year); !ok {
		errs["year"] = msg
	}
	return errs
}

// ValidateReplacePortfolio memeriksa isi PUT /students/:id/portfolio.
func ValidateReplacePortfolio(req model.ReplacePortfolioRequest) map[string]string {
	errs := map[string]string{}

	if msg, ok := validateHeadline(req.Headline); !ok {
		errs["headline"] = msg
	}
	if msg, ok := validateBio(req.Bio); !ok {
		errs["bio"] = msg
	}

	if len(req.Projects) > repository.MaxProjectsPerPortfolio {
		errs["projects"] = "maksimal 20 proyek"
	}
	for i, p := range req.Projects {
		for field, msg := range validateProjectRequest(p) {
			errs["projects."+strconv.Itoa(i)+"."+field] = msg
		}
	}

	if len(req.Certificates) > 20 {
		errs["certificates"] = "maksimal 20 sertifikat"
	}
	for i, c := range req.Certificates {
		if msg, ok := validateCertificateName(c.Name); !ok {
			errs["certificates."+strconv.Itoa(i)+".name"] = msg
		}
		if msg, ok := validateCertificateIssuer(c.Issuer); !ok {
			errs["certificates."+strconv.Itoa(i)+".issuer"] = msg
		}
		if msg, _, ok := validateCertificateDate(c.Date); !ok {
			errs["certificates."+strconv.Itoa(i)+".date"] = msg
		}
	}

	return errs
}

// ValidateAddProject memeriksa isi POST /students/:id/portfolio/projects.
func ValidateAddProject(req model.AddProjectRequest) map[string]string {
	return validateProjectRequest(model.ProjectRequest{
		Title:        req.Title,
		Technologies: req.Technologies,
		URL:          req.URL,
		Year:         req.Year,
	})
}
