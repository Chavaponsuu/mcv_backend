package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"mcv_backend/models"
)

var (
	ErrUnauthorized        = errors.New("unauthorized")
	ErrServiceUnavailable  = errors.New("service unavailable")
)

const (
	myCourseVilleBaseURL = "https://www.mycourseville.com/api/v1"
	requestTimeout       = 10 * time.Second
)

// FetchUserCourses retrieves the list of courses for an authenticated user
func FetchUserCourses(accessToken string) ([]models.Course, error) {
	client := &http.Client{Timeout: requestTimeout}
	
	url := fmt.Sprintf("%s/user/courses", myCourseVilleBaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrServiceUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	if resp.StatusCode >= 500 {
		return nil, ErrServiceUnavailable
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse models.MyCourseVilleResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}
