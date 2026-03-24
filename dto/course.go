package dto


type CourseItem struct {
	ID         string `json:"id"`
	CourseCode string `json:"course_code"`
	Title      string `json:"title"`
	Credits    int    `json:"credits"`
}

type CoursesResponse struct {
	Courses []CourseItem `json:"courses"`
	Total   int          `json:"total"`
}