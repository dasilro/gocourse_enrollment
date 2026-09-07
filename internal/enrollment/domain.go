package enrollment

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Enrollment struct {
	ID        string       `json:"id" gorm:"type:char(36);not null;primary_key;unique_index"`
	CourseID  string       `json:"course_id,omitempty" gorm:"type:char(36)"`
	UserID    string       `json:"user_id,omitempty" gorm:"type:char(36)"`
	Status    EnrollStatus `json:"status" gorm:"type:char(2)"`
	CreatedAt *time.Time   `json:"_"`
	UpdatedAt *time.Time   `json:"-"`
}

type EnrollStatus string

const (
	Pending  EnrollStatus = "P"
	Active   EnrollStatus = "A"
	Studying EnrollStatus = "S"
	Inactive EnrollStatus = "I"
)

func (e *Enrollment) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return
}
