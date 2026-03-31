package dto




type SemesterItem struct {
	ID   string `json:"id"`
	Year int    `json:"year"`
	Term int    `json:"term"`
}

type SemestersResponse struct {
	Semesters []SemesterItem `json:"semesters"`
	Total     int            `json:"total"`
}