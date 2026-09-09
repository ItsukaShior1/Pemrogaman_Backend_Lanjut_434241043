package service

import (
	"reflect"
	"testing"

	"api-student/app/model"
)

func TestValidateCreate_Success(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM:      "2201110001",
		Name:     "Budi Santoso",
		Grade:    85.5,
		IsActive: true,
	}

	errs := ValidateCreate(req)

	if len(errs) != 0 {
		t.Errorf("expected tidak ada error, dapat %v", errs)
	}
}

func TestValidateCreate_MultipleErrors(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM:      "",  // kosong
		Name:     "",  // kosong
		Grade:    150, // di luar range
		IsActive: false,
	}

	errs := ValidateCreate(req)

	if len(errs) != 3 {
		t.Errorf("expected 3 error, dapat %d: %v", len(errs), errs)
	}
	if errs["nim"] == "" {
		t.Error("expected error untuk nim")
	}
	if errs["name"] == "" {
		t.Error("expected error untuk name")
	}
	if errs["grade"] == "" {
		t.Error("expected error untuk grade")
	}
}
func TestValidateReplace_RequiresAllFields(t *testing.T) {
	req := model.ReplaceStudentRequest{
		NIM:      "",
		Name:     "Andi",
		Grade:    80,
		IsActive: true,
	}

	errs := ValidateReplace(req)

	if len(errs) == 0 {
		t.Fatal("expected minimal satu error, dapat 0")
	}
	if got := errs["nim"]; got == "" || !contains(got, "pada PUT") {
		t.Errorf("expected pesan error nim menyebut 'pada PUT', dapat %q", got)
	}
}

func TestApplyPatch_PartialUpdate(t *testing.T) {
	current := model.Student{
		ID:       1,
		NIM:      "2201110001",
		Name:     "Budi",
		Grade:    70,
		IsActive: true,
	}
	gradeBaru := 95.5
	req := model.PatchStudentRequest{Grade: &gradeBaru}

	hasil, errs := ApplyPatch(current, req)

	if len(errs) != 0 {
		t.Fatalf("expected tidak ada error, dapat %v", errs)
	}
	if hasil.Grade != 95.5 {
		t.Errorf("expected Grade 95.5, dapat %v", hasil.Grade)
	}
	if hasil.NIM != "2201110001" {
		t.Errorf("NIM berubah padahal tidak di-PATCH: %q", hasil.NIM)
	}
	if hasil.Name != "Budi" {
		t.Errorf("Name berubah padahal tidak di-PATCH: %q", hasil.Name)
	}
	if !hasil.IsActive {
		t.Error("IsActive berubah padahal tidak di-PATCH")
	}
}

func TestApplyPatch_EmptyRequest(t *testing.T) {
	current := model.Student{ID: 1, NIM: "2201110001", Name: "Budi", Grade: 80}
	req := model.PatchStudentRequest{}

	_, errs := ApplyPatch(current, req)

	if errs["_"] != "tidak ada field yang diubah" {
		t.Errorf("expected error _ = 'tidak ada field yang diubah', dapat %v", errs)
	}
}

func TestApplyPatch_InvalidField(t *testing.T) {
	current := model.Student{ID: 1, NIM: "LAMA", Name: "Budi", Grade: 80}
	nimKosong := ""
	req := model.PatchStudentRequest{NIM: &nimKosong}

	hasil, errs := ApplyPatch(current, req)

	if errs["nim"] == "" {
		t.Error("expected error untuk nim kosong")
	}
	if hasil.NIM != "LAMA" {
		t.Errorf("NIM termutasi walau invalid: dapat %q, seharusnya %q", hasil.NIM, "LAMA")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestApplyPatch_NilPointerSafety(t *testing.T) {
	current := model.Student{ID: 1, NIM: "X", Name: "Y", Grade: 50}
	req := model.PatchStudentRequest{}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ApplyPatch panic pada input kosong: %v", r)
		}
	}()

	_, errs := ApplyPatch(current, req)
	if len(errs) == 0 {
		t.Error("expected error _ untuk request kosong")
	}
}

func TestValidateCreate_Trim(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM:      "  2201110001  ",
		Name:     "   Budi   ",
		Grade:    80,
		IsActive: true,
	}

	errs := ValidateCreate(req)

	if len(errs) != 0 {
		t.Errorf("expected tidak ada error setelah trim, dapat %v", errs)
	}
}

func TestValidateCreate_GradeBoundary(t *testing.T) {
	cases := []struct {
		name  string
		grade float64
		valid bool
	}{
		{"nol", 0, true},
		{"seratus", 100, true},
		{"negatif", -0.01, false},
		{"di atas", 100.01, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := model.CreateStudentRequest{
				NIM:   "2201110001",
				Name:  "Budi",
				Grade: tc.grade,
			}
			errs := ValidateCreate(req)
			_, hasErr := errs["grade"]
			if tc.valid && hasErr {
				t.Errorf("grade %v seharusnya valid, dapat error %q", tc.grade, errs["grade"])
			}
			if !tc.valid && !hasErr {
				t.Errorf("grade %v seharusnya invalid", tc.grade)
			}
		})
	}
}

func TestApplyPatch_MultipleFields(t *testing.T) {
	current := model.Student{
		ID: 1, NIM: "LAMA", Name: "Lama", Grade: 60, IsActive: false,
	}
	nimBaru := "BARU"
	nameBaru := "Baru"
	isActiveBaru := true
	req := model.PatchStudentRequest{
		NIM:      &nimBaru,
		Name:     &nameBaru,
		IsActive: &isActiveBaru,
	}

	hasil, errs := ApplyPatch(current, req)

	if len(errs) != 0 {
		t.Fatalf("expected tidak ada error, dapat %v", errs)
	}
	if !reflect.DeepEqual(hasil, model.Student{
		ID: 1, NIM: "BARU", Name: "Baru", Grade: 60, IsActive: true,
	}) {
		t.Errorf("hasil mutasi tidak sesuai, dapat %+v", hasil)
	}
}
