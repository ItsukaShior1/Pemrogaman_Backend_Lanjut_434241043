package service

import (
	"strings"

	"api-student/app/model"
)

func validateNim(nim string) (string, bool) {
	nim = strings.TrimSpace(nim)
	if nim == "" {
		return "wajib diisi", false
	}
	if len(nim) < 4 || len(nim) > 20 {
		return "panjang harus antara 4 dan 20 karakter", false
	}
	return "", true
}

func validateName(name string) (string, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "wajib diisi", false
	}
	if len(name) > 255 {
		return "panjang maksimal 255 karakter", false
	}
	return "", true
}

func validateGrade(grade float64) (string, bool) {
	if grade < 0 || grade > 100 {
		return "harus bernilai antara 0 dan 100", false
	}
	return "", true
}

func ValidateCreate(req model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}

	if msg, ok := validateNim(req.NIM); !ok {
		errs["nim"] = msg
	}
	if msg, ok := validateName(req.Name); !ok {
		errs["name"] = msg
	}
	if msg, ok := validateGrade(req.Grade); !ok {
		errs["grade"] = msg
	}

	return errs
}

func ValidateReplace(req model.ReplaceStudentRequest) map[string]string {
	errs := map[string]string{}

	if msg, ok := validateNim(req.NIM); !ok {
		errs["nim"] = msg + " pada PUT"
	}
	if msg, ok := validateName(req.Name); !ok {
		errs["name"] = msg + " pada PUT"
	}
	if msg, ok := validateGrade(req.Grade); !ok {
		errs["grade"] = msg
	}

	return errs
}

func ApplyPatch(current model.Student, req model.PatchStudentRequest) (model.Student, map[string]string) {
	errs := map[string]string{}

	if req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil {
		errs["_"] = "tidak ada field yang diubah"
		return current, errs
	}

	if req.NIM != nil {
		nim := strings.TrimSpace(*req.NIM)
		if msg, ok := validateNim(nim); !ok {
			errs["nim"] = msg
		} else {
			current.NIM = nim
		}
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if msg, ok := validateName(name); !ok {
			errs["name"] = msg
		} else {
			current.Name = name
		}
	}

	if req.Grade != nil {
		if msg, ok := validateGrade(*req.Grade); !ok {
			errs["grade"] = msg
		} else {
			current.Grade = *req.Grade
		}
	}

	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current, errs
}
