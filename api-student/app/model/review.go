package model

import "time"

// Review adalah bentuk ulasan kelas yang dikenal seluruh aplikasi.
//
// Tidak ada satu pun tipe dari driver MongoDB di sini. ID disimpan sebagai
// string, bukan bson.ObjectID, supaya package model tetap bebas dari driver.
type Review struct {
	ID           string    `json:"id"`
	CourseID     int       `json:"course_id"`
	UserID       int       `json:"user_id"`
	Username     string    `json:"username"`
	Rating       int       `json:"rating"`
	Comment      string    `json:"comment"`
	Tags         []string  `json:"tags"`
	Replies      []Reply   `json:"replies"`
	HelpfulCount int       `json:"helpful_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Reply disimpan DI DALAM dokumen Review, bukan di collection terpisah.
// Balasan tidak pernah dibaca tanpa ulasan induknya.
type Reply struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateReviewRequest struct {
	Rating  int      `json:"rating"`
	Comment string   `json:"comment"`
	Tags    []string `json:"tags"`
}

// UpdateReviewRequest memakai pointer: nil berarti "tidak dikirim".
type UpdateReviewRequest struct {
	Rating  *int      `json:"rating,omitempty"`
	Comment *string   `json:"comment,omitempty"`
	Tags    *[]string `json:"tags,omitempty"`
}

type ReplyRequest struct {
	Body string `json:"body"`
}

// ReviewQuery adalah parameter daftar ulasan satu kelas.
type ReviewQuery struct {
	CourseID int
	Tag      string
	Before   string
	Limit    int
}

type ReviewSummary struct {
	CourseID      int            `json:"course_id"`
	Count         int            `json:"count"`
	AverageRating float64        `json:"average_rating"`
	Distribution  map[string]int `json:"distribution"`
	TopTags       []TagCount     `json:"top_tags"`
}

type TagCount struct {
	Tag   string `json:"tag"`
	Total int    `json:"total"`
}

type Course struct {
	ID        int       `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Semester  string    `json:"semester"`
	CreatedAt time.Time `json:"created_at"`
}
