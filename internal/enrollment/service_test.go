package enrollment_test

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"

	courseSdk "github.com/dasilro/go_course_sdk/course/mock"
	userSdk "github.com/dasilro/go_course_sdk/user/mock"
	"github.com/dasilro/gocourse_domain/domain"
	"github.com/dasilro/gocourse_enrollment/internal/enrollment"
	"github.com/stretchr/testify/assert"
)

func TestService_GetAll(t *testing.T) {
	l := log.New(io.Discard, "", log.LstdFlags|log.Lshortfile)
	t.Run("should return an error", func(t *testing.T) {
		want := errors.New("my error")
		var wantCounter int = 1
		var counter int = 0
		repo := &mockRepository{
			GetAllMock: func(ctx context.Context, filters enrollment.Filters, offset int, limit int) ([]enrollment.Enrollment, error) {
				counter++
				return nil, errors.New("my error")
			},
		}

		service := enrollment.NewService(l, nil, nil, repo)

		enrollments, err := service.GetAll(context.Background(), enrollment.Filters{}, 0, 10)

		assert.Error(t, err)
		assert.Nil(t, enrollments)
		assert.Equal(t, wantCounter, counter)
		assert.EqualError(t, want, err.Error())
	})

	t.Run("should return all enrollments", func(t *testing.T) {
		want := []enrollment.Enrollment{
			{
				ID:       "1",
				CourseID: "1",
				UserID:   "1",
				Status:   "P",
			},
		}
		var wantCounter int = 1
		var counter int = 0
		repo := &mockRepository{
			GetAllMock: func(ctx context.Context, filters enrollment.Filters, offset int, limit int) ([]enrollment.Enrollment, error) {
				counter++
				return []enrollment.Enrollment{
					{
						ID:       "1",
						CourseID: "1",
						UserID:   "1",
						Status:   "P",
					},
				}, nil
			},
		}

		service := enrollment.NewService(l, nil, nil, repo)

		enrollments, err := service.GetAll(context.Background(), enrollment.Filters{}, 0, 10)

		assert.Nil(t, err)
		assert.NotNil(t, enrollments)
		assert.Equal(t, wantCounter, counter)
		assert.Equal(t, want, enrollments)
	})
}

func TestService_Update(t *testing.T) {
	l := log.New(io.Discard, "", 0)

	t.Run("should return an error", func(t *testing.T) {
		var want error = errors.New("my error")
		var wantCounter int = 1
		var count int = 0

		repo := &mockRepository{
			UpdateMock: func(ctx context.Context, id string, status *string) error {
				count++
				return errors.New("my error")
			},
		}

		service := enrollment.NewService(l, nil, nil, repo)

		status := "A"
		err := service.Update(context.Background(), "11", &status)

		assert.NotNil(t, err)
		assert.Equal(t, wantCounter, count)
		assert.EqualError(t, want, err.Error())
	})

	t.Run("should update an enrollment", func(t *testing.T) {
		var wantCounter int = 1
		var count int = 0
		var wantStatus string = "A"
		var wantID string = "11"

		repo := &mockRepository{
			UpdateMock: func(ctx context.Context, id string, status *string) error {
				count++
				assert.Equal(t, wantID, id)
				assert.NotNil(t, status)
				assert.Equal(t, wantStatus, *status)
				return nil
			},
		}

		service := enrollment.NewService(l, nil, nil, repo)

		status := "A"
		err := service.Update(context.Background(), "11", &status)

		assert.Nil(t, err)
		assert.Equal(t, wantCounter, count)
	})
}

func TestService_Count(t *testing.T) {
	l := log.New(io.Discard, "", 0)

	t.Run("should return an error", func(t *testing.T) {
		var want error = errors.New("my error")
		var wantCounter int = 1
		var count int = 0

		repo := &mockRepository{
			CountMock: func(ctx context.Context, filters enrollment.Filters) (int, error) {
				count++
				return 0, errors.New("my error")
			},
		}

		service := enrollment.NewService(l, nil, nil, repo)
		_, err := service.Count(context.Background(), enrollment.Filters{})

		assert.NotNil(t, err)
		assert.Equal(t, wantCounter, count)
		assert.EqualError(t, want, err.Error())
	})

	t.Run("should return count", func(t *testing.T) {
		var wantCountResult int = 5
		var wantCounter int = 1
		var count int = 0

		repo := &mockRepository{
			CountMock: func(ctx context.Context, filters enrollment.Filters) (int, error) {
				count++
				return 5, nil
			},
		}

		service := enrollment.NewService(l, nil, nil, repo)
		countResult, err := service.Count(context.Background(), enrollment.Filters{})

		assert.Nil(t, err)
		assert.Equal(t, wantCounter, count)
		assert.Equal(t, wantCountResult, countResult)
	})
}

func TestService_Create(t *testing.T) {
	l := log.New(io.Discard, "", 0)

	t.Run("should return an error in course sdk", func(t *testing.T) {
		var want error = errors.New("my error")
		var wantCounter int = 1
		var count int = 0

		courseSdk := &courseSdk.CourseSdkMock{
			GetMock: func(id string) (*domain.Course, error) {
				count++
				return nil, errors.New("my error")
			},
		}

		service := enrollment.NewService(l, courseSdk, nil, nil)
		enrollment, err := service.Create(context.Background(), "1", "1")

		assert.NotNil(t, err)
		assert.Equal(t, wantCounter, count)
		assert.EqualError(t, want, err.Error())
		assert.Nil(t, enrollment)
	})

	t.Run("should return an error in user sdk", func(t *testing.T) {
		var want error = errors.New("my error")
		var wantCounter int = 2
		var count int = 0

		courseSdk := &courseSdk.CourseSdkMock{
			GetMock: func(id string) (*domain.Course, error) {
				count++
				return &domain.Course{}, nil
			},
		}

		userSdk := &userSdk.UserSdkMock{
			GetMock: func(id string) (*domain.User, error) {
				count++
				return nil, errors.New("my error")
			},
		}

		service := enrollment.NewService(l, courseSdk, userSdk, nil)
		enrollment, err := service.Create(context.Background(), "1", "1")

		assert.NotNil(t, err)
		assert.Equal(t, wantCounter, count)
		assert.EqualError(t, want, err.Error())
		assert.Nil(t, enrollment)
	})

	t.Run("should return an error in repository", func(t *testing.T) {
		var want error = errors.New("my error")
		var wantCounter int = 3
		var count int = 0

		courseSdk := &courseSdk.CourseSdkMock{
			GetMock: func(id string) (*domain.Course, error) {
				count++
				return &domain.Course{}, nil
			},
		}

		userSdk := &userSdk.UserSdkMock{
			GetMock: func(id string) (*domain.User, error) {
				count++
				return &domain.User{}, nil
			},
		}

		repo := &mockRepository{
			CreateMock: func(ctx context.Context, enrollment *enrollment.Enrollment) error {
				count++
				return errors.New("my error")
			},
		}

		service := enrollment.NewService(l, courseSdk, userSdk, repo)
		enrollment, err := service.Create(context.Background(), "1", "1")

		assert.NotNil(t, err)
		assert.Equal(t, wantCounter, count)
		assert.EqualError(t, want, err.Error())
		assert.Nil(t, enrollment)
	})

	t.Run("should return enrollment", func(t *testing.T) {
		var wantCounter int = 3
		var count int = 0

		courseSdk := &courseSdk.CourseSdkMock{
			GetMock: func(id string) (*domain.Course, error) {
				count++
				return &domain.Course{}, nil
			},
		}

		userSdk := &userSdk.UserSdkMock{
			GetMock: func(id string) (*domain.User, error) {
				count++
				return &domain.User{}, nil
			},
		}

		repo := &mockRepository{
			CreateMock: func(ctx context.Context, enrollment *enrollment.Enrollment) error {
				count++
				return nil
			},
		}

		service := enrollment.NewService(l, courseSdk, userSdk, repo)
		enrollment, err := service.Create(context.Background(), "1", "2")

		assert.Nil(t, err)
		assert.Equal(t, wantCounter, count)
		assert.NotNil(t, enrollment)
		assert.Equal(t, "1", enrollment.CourseID)
		assert.Equal(t, "2", enrollment.UserID)
	})
}
