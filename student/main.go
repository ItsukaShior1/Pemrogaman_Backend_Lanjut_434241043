package main

import "fmt"

type Student struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

func (s Student) GetInfo() string {
	status := "non-aktif"
	if s.IsActive {
		status = "aktif"
	}
	return fmt.Sprintf("ID=%d | Name=%s | Grade=%.2f | Status=%s",
		s.ID, s.Name, s.Grade, status)
}

func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

func (s *Student) Activate() {
	s.IsActive = true
}

func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	s1 := Student{ID: 1, Name: "Sari", Grade: 85.0}

	fmt.Println("Keadaan awal")
	fmt.Println(s1.GetInfo())

	s1.Activate()
	s1.UpdateGrade(92.5)
	fmt.Println("\nSetelah Activate() dan UpdateGrade(92.5)")
	fmt.Println(s1.GetInfo())

	s1.Deactivate()
	fmt.Println("\nSetelah Deactivate()")
	fmt.Println(s1.GetInfo())

	fmt.Println("\nDipanggil via pointer")
	fmt.Println(s1.GetInfo())
}
