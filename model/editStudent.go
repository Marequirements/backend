package model

type EditStudentRequest struct {
	StudentID  string     `json:"studentId"`
	NewStudent NewStudent `json:"newStudent"`
}
