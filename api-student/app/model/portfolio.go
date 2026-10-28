package model

import "time"


type Portfolio struct {
	ID           string        `json:"id"`
	StudentID    int           `json:"student_id"`
	Headline     string        `json:"headline"`
	Bio          string        `json:"bio"`
	Projects     []Project     `json:"projects"`
	Certificates []Certificate `json:"certificates"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}


type Project struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	URL          string   `json:"url"`
	Year         int      `json:"year"`
}

type Certificate struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Issuer string    `json:"issuer"`
	Date   time.Time `json:"date"`
}

type ReplacePortfolioRequest struct {
	Headline     string               `json:"headline"`
	Bio          string               `json:"bio"`
	Projects     []ProjectRequest     `json:"projects"`
	Certificates []CertificateRequest `json:"certificates"`
}

type ProjectRequest struct {
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	URL          string   `json:"url"`
	Year         int      `json:"year"`
}

type CertificateRequest struct {
	Name   string `json:"name"`
	Issuer string `json:"issuer"`
	Date   string `json:"date"`
}

type AddProjectRequest struct {
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	URL          string   `json:"url"`
	Year         int      `json:"year"`
}

type TechnologyStat struct {
	Technology string `json:"technology"`
	Total      int    `json:"total"`
}
