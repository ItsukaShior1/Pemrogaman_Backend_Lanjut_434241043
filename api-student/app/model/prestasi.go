package model

import "time"

type Prestasi struct {
	IDPrestasi int       `json:"id_prestasi"`
	StudentID  int       `json:"student_id"`
	Title      string    `json:"title"`
	Champion   int       `json:"champion"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreatePrestasiRequest struct {
	StudentID int    `json:"student_id"`
	Title     string `json:"title"`
	Champion  int    `json:"champion"`
}

type ReplacePrestasiRequest struct {
	StudentID int    `json:"student_id"`
	Title     string `json:"title"`
	Champion  int    `json:"champion"`
}

type PatchPrestasiRequest struct {
	StudentID *int    `json:"student_id,omitempty"`
	Title     *string `json:"title,omitempty"`
	Champion  *int    `json:"champion,omitempty"`
}
type WebResponsePrestasi struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
