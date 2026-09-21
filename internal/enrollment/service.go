package enrollment

import (
	"context"
	"log"

	courseSdk "github.com/dasilro/go_course_sdk/course"
	userSdk "github.com/dasilro/go_course_sdk/user"
)

type (
	Filters struct {
		CourseID string
		UserID   string
	}

	Service interface {
		Create(ctx context.Context, courseId, userId string) (*Enrollment, error)
		GetAll(ctx context.Context, filters Filters, offset, limit int) ([]Enrollment, error)
		Update(ctx context.Context, id string, status *string) error
		Count(ctx context.Context, filters Filters) (int, error)
	}

	service struct {
		log         *log.Logger
		courseTrans courseSdk.Transport
		userTrans   userSdk.Transport
		repo        Repository
	}
)

func NewService(log *log.Logger, courseTrans courseSdk.Transport, userTrans userSdk.Transport, repo Repository) Service {
	return &service{
		log:         log,
		courseTrans: courseTrans,
		userTrans:   userTrans,
		repo:        repo,
	}
}

func (s service) Create(ctx context.Context, courseId, userId string) (*Enrollment, error) {
	enrollment := Enrollment{
		CourseID: courseId,
		UserID:   userId,
		Status:   Pending,
	}

	if _, err := s.courseTrans.Get(courseId); err != nil {
		return nil, err
	}

	if _, err := s.userTrans.Get(userId); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, &enrollment); err != nil {
		s.log.Println(err.Error())
		return nil, err
	}

	s.log.Println("[SUCCESS] Service - Create - enrollments")
	return &enrollment, nil
}

func (s service) GetAll(ctx context.Context, filters Filters, offset int, limit int) ([]Enrollment, error) {
	s.log.Println("[SUCCESS] Service - GetAll - enrollments")
	return s.repo.GetAll(ctx, filters, offset, limit)
}

func (s service) Update(ctx context.Context, id string, status *string) error {

	if status != nil {
		switch EnrollStatus(*status) {
		case Pending, Active, Studying, Inactive:
		default:
			return ErrInvalidStatus{Status: *status}
		}
	}

	s.log.Println("[SUCCESS] Service - Update - enrollments")
	return s.repo.Update(ctx, id, status)
}

func (s service) Count(ctx context.Context, filters Filters) (int, error) {
	return s.repo.Count(ctx, filters)
}
