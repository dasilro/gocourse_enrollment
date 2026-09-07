package enrollment

import (
	"errors"
	"fmt"
)

var ErrCourseIdRequired = errors.New("Course id is required")
var ErrUserIdRequired = errors.New("User id is required")
var ErrStatusRequired = errors.New("Status is required")

type ErrNotFound struct {
	EnrollmentID string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("enrollment %s does not exist", e.EnrollmentID)
}

type ErrInvalidStatus struct {
	Status string
}

func (e ErrInvalidStatus) Error() string {
	return fmt.Sprintf("invalid '%s' status", e.Status)
}
