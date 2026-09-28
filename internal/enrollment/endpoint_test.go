package enrollment_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"

	courseSdk "github.com/dasilro/go_course_sdk/course"
	mockCourseSdk "github.com/dasilro/go_course_sdk/course/mock"
	userSdk "github.com/dasilro/go_course_sdk/user"
	mockUserSdk "github.com/dasilro/go_course_sdk/user/mock"

	"github.com/dasilro/go_lib_response/response"
	"github.com/dasilro/gocourse_domain/domain"
	"github.com/dasilro/gocourse_enrollment/internal/enrollment"
	"github.com/stretchr/testify/assert"
)

func TestCreateEndpoint(t *testing.T) {
	l := log.New(io.Discard, "", 0)
	t.Run("should return bad request when course id is empty", func(t *testing.T) {
		endpoint := enrollment.MakeEndpoints(nil, "10")
		_, err := endpoint.Create(context.Background(), enrollment.CreateReq{})

		assert.EqualError(t, enrollment.ErrCourseIdRequired, err.(*response.ErrorResponse).Error())
		assert.Equal(t, http.StatusBadRequest, err.(*response.ErrorResponse).Status)
	})

	t.Run("should return bad request when user id is empty", func(t *testing.T) {
		endpoint := enrollment.MakeEndpoints(nil, "10")
		_, err := endpoint.Create(context.Background(), enrollment.CreateReq{CourseID: "123"})

		assert.EqualError(t, enrollment.ErrUserIdRequired, err.(*response.ErrorResponse).Error())
		assert.Equal(t, http.StatusBadRequest, err.(*response.ErrorResponse).Status)
	})

	obj := []struct {
		tag            string
		repositoryMock enrollment.Repository
		userSdkMock    userSdk.Transport
		courseSdkMock  courseSdk.Transport
		wantErr        error
		wantCode       int
		wantResponse   *enrollment.Enrollment
	}{
		{
			tag: "should return an error if user sdk returns an unexpected error",
			userSdkMock: &mockUserSdk.UserSdkMock{
				GetMock: func(id string) (*domain.User, error) {
					return nil, errors.New("unexpected error")
				},
			},
			courseSdkMock: &mockCourseSdk.CourseSdkMock{
				GetMock: func(id string) (*domain.Course, error) {
					return nil, errors.New("unexpected error")
				},
			},

			wantErr:  errors.New("unexpected error"),
			wantCode: http.StatusInternalServerError,
		},
		{
			tag: "should return an error if course does not exist",
			userSdkMock: &mockUserSdk.UserSdkMock{
				GetMock: func(id string) (*domain.User, error) {
					return nil, errors.New("unexpected error")
				},
			},
			courseSdkMock: &mockCourseSdk.CourseSdkMock{
				GetMock: func(id string) (*domain.Course, error) {
					return nil, courseSdk.ErrNotFound{Message: "course not found"}
				},
			},

			wantErr:  courseSdk.ErrNotFound{Message: "course not found"},
			wantCode: http.StatusNotFound,
		},
		{
			tag: "should return an error if user does not exist",
			userSdkMock: &mockUserSdk.UserSdkMock{
				GetMock: func(id string) (*domain.User, error) {
					return nil, userSdk.ErrNotFound{Message: "user not found"}
				},
			},
			courseSdkMock: &mockCourseSdk.CourseSdkMock{
				GetMock: func(id string) (*domain.Course, error) {
					return &domain.Course{}, nil
				},
			},
			wantErr:  userSdk.ErrNotFound{Message: "user not found"},
			wantCode: http.StatusNotFound,
		},
		{
			tag: "should return an error if repository returns an unexpected error",
			userSdkMock: &mockUserSdk.UserSdkMock{
				GetMock: func(id string) (*domain.User, error) {
					return &domain.User{}, nil
				},
			},
			courseSdkMock: &mockCourseSdk.CourseSdkMock{
				GetMock: func(id string) (*domain.Course, error) {
					return &domain.Course{}, nil
				},
			},
			repositoryMock: &mockRepository{
				CreateMock: func(ctx context.Context, enrollment *enrollment.Enrollment) error {
					return errors.New("unexpected error")
				},
			},
			wantErr:  userSdk.ErrNotFound{Message: "unexpected error"},
			wantCode: http.StatusInternalServerError,
		},
		{
			tag: "should return the enrollment",
			userSdkMock: &mockUserSdk.UserSdkMock{
				GetMock: func(id string) (*domain.User, error) {
					return &domain.User{}, nil
				},
			},
			courseSdkMock: &mockCourseSdk.CourseSdkMock{
				GetMock: func(id string) (*domain.Course, error) {
					return &domain.Course{}, nil
				},
			},
			repositoryMock: &mockRepository{
				CreateMock: func(ctx context.Context, enrollment *enrollment.Enrollment) error {
					return nil
				},
			},
			wantErr:  nil,
			wantCode: http.StatusCreated,
			wantResponse: &enrollment.Enrollment{
				UserID:   "123",
				CourseID: "123",
				Status:   "P",
			},
		},
	}

	for _, obj := range obj {
		t.Run(obj.tag, func(t *testing.T) {
			service := enrollment.NewService(l, obj.courseSdkMock, obj.userSdkMock, obj.repositoryMock)
			endpoint := enrollment.MakeEndpoints(service, "10")
			resp, err := endpoint.Create(context.Background(), enrollment.CreateReq{UserID: "123", CourseID: "123"})
			fmt.Println("call to endpoint done")

			if obj.wantErr != nil {
				assert.NotNil(t, err)
				assert.Nil(t, resp)

				r := err.(response.Response)
				assert.EqualError(t, obj.wantErr, r.Error())
				assert.Equal(t, obj.wantCode, r.StatusCode())
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, resp)

				resp := resp.(response.Response)
				assert.Equal(t, obj.wantCode, resp.StatusCode())
				enrollment := resp.GetData().(*enrollment.Enrollment)
				assert.NotNil(t, enrollment)
				assert.Equal(t, obj.wantResponse.CourseID, enrollment.CourseID)
				assert.Equal(t, obj.wantResponse.UserID, enrollment.UserID)
				assert.Equal(t, obj.wantResponse.Status, enrollment.Status)
			}
		})
	}
}

func TestGetAllEndpoint(t *testing.T) {
	l := log.New(io.Discard, "", 0)

	t.Run("should return an error if count returns an unexpected error", func(t *testing.T) {
		wantErr := errors.New("unexpected error")
		service := enrollment.NewService(l, nil, nil, &mockRepository{
			CountMock: func(ctx context.Context, filters enrollment.Filters) (int, error) {
				return 0, wantErr
			},
		})
		endpoint := enrollment.MakeEndpoints(service, "10")
		_, err := endpoint.GetAll(context.Background(), enrollment.GetAllReq{})

		assert.Error(t, err)
		resp := err.(response.Response)
		assert.EqualError(t, wantErr, resp.Error())
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode())
	})

	t.Run("should return an error if meta returns an error", func(t *testing.T) {
		wantErr := errors.New("strconv.Atoi: parsing \"invalid number\": invalid syntax")
		service := enrollment.NewService(l, nil, nil, &mockRepository{
			CountMock: func(ctx context.Context, filters enrollment.Filters) (int, error) {
				return 3, nil
			},
		})
		endpoint := enrollment.MakeEndpoints(service, "invalid number")
		_, err := endpoint.GetAll(context.Background(), enrollment.GetAllReq{})

		assert.Error(t, err)
		resp := err.(response.Response)
		assert.EqualError(t, wantErr, resp.Error())
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode())
	})

	t.Run("should return an error if GetAll repository returns an error", func(t *testing.T) {
		wantErr := errors.New("unexpected error")
		service := enrollment.NewService(l, nil, nil, &mockRepository{
			CountMock: func(ctx context.Context, filters enrollment.Filters) (int, error) {
				return 3, nil
			},
			GetAllMock: func(ctx context.Context, filters enrollment.Filters, offset int, limit int) ([]enrollment.Enrollment, error) {
				return nil, wantErr
			},
		})
		endpoint := enrollment.MakeEndpoints(service, "10")
		_, err := endpoint.GetAll(context.Background(), enrollment.GetAllReq{})

		assert.Error(t, err)
		resp := err.(response.Response)
		assert.EqualError(t, wantErr, resp.Error())
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode())
	})

	t.Run("should return the enrollments", func(t *testing.T) {
		wantEnrollments := []enrollment.Enrollment{
			{UserID: "11", CourseID: "12", Status: "P"},
			{UserID: "21", CourseID: "22", Status: "P"},
			{UserID: "31", CourseID: "32", Status: "P"},
		}
		service := enrollment.NewService(l, nil, nil, &mockRepository{
			CountMock: func(ctx context.Context, filters enrollment.Filters) (int, error) {
				return 3, nil
			},
			GetAllMock: func(ctx context.Context, filters enrollment.Filters, offset int, limit int) ([]enrollment.Enrollment, error) {
				return []enrollment.Enrollment{
					{UserID: "11", CourseID: "12", Status: "P"},
					{UserID: "21", CourseID: "22", Status: "P"},
					{UserID: "31", CourseID: "32", Status: "P"},
				}, nil
			},
		})
		endpoint := enrollment.MakeEndpoints(service, "10")
		resp, err := endpoint.GetAll(context.Background(), enrollment.GetAllReq{})

		assert.Nil(t, err)
		successResp := resp.(*response.SuccessResponse)
		assert.Empty(t, successResp.Error())
		assert.Equal(t, http.StatusOK, successResp.StatusCode())
		enrollments := successResp.GetData().([]enrollment.Enrollment)
		assert.Equal(t, wantEnrollments, enrollments)
	})
}

func TestUpdateEndpoint(t *testing.T) {
	l := log.New(io.Discard, "", 0)

	t.Run("should return an error if status is empty", func(t *testing.T) {
		endpoint := enrollment.MakeEndpoints(nil, "10")
		status := ""
		_, err := endpoint.Update(context.Background(), enrollment.UpdateReq{Status: &status})

		assert.Error(t, err)
		resp := err.(response.Response)
		assert.EqualError(t, enrollment.ErrStatusRequired, resp.Error())
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
	})

	t.Run("should return an error if repository returns a not found error", func(t *testing.T) {
		service := enrollment.NewService(l, nil, nil, &mockRepository{
			UpdateMock: func(ctx context.Context, id string, status *string) error {
				return enrollment.ErrNotFound{EnrollmentID: id}
			},
		})
		endpoint := enrollment.MakeEndpoints(service, "10")
		status := "A"
		_, err := endpoint.Update(context.Background(), enrollment.UpdateReq{ID: "20", Status: &status})

		assert.Error(t, err)
		resp := err.(response.Response)
		assert.EqualError(t, enrollment.ErrNotFound{EnrollmentID: "20"}, resp.Error())
		assert.Equal(t, http.StatusNotFound, resp.StatusCode())
	})

	t.Run("should return an error if repository returns an unexpected error", func(t *testing.T) {
		wantErr := errors.New("unexpected error")
		service := enrollment.NewService(l, nil, nil, &mockRepository{
			UpdateMock: func(ctx context.Context, id string, status *string) error {
				return errors.New("unexpected error")
			},
		})
		endpoint := enrollment.MakeEndpoints(service, "10")
		status := "A"
		_, err := endpoint.Update(context.Background(), enrollment.UpdateReq{ID: "20", Status: &status})

		assert.Error(t, err)
		resp := err.(response.Response)
		assert.EqualError(t, wantErr, resp.Error())
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode())
	})

	t.Run("should return an error if repository returns an unexpected error", func(t *testing.T) {
		service := enrollment.NewService(l, nil, nil, &mockRepository{
			UpdateMock: func(ctx context.Context, id string, status *string) error {
				return nil
			},
		})
		endpoint := enrollment.MakeEndpoints(service, "10")
		status := "A"
		resp, err := endpoint.Update(context.Background(), enrollment.UpdateReq{ID: "20", Status: &status})

		assert.Nil(t, err)
		r := resp.(response.Response)
		assert.Empty(t, r.Error())
		assert.Equal(t, http.StatusOK, r.StatusCode())
	})
}
