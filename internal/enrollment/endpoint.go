package enrollment

import (
	"context"
	"errors"

	"github.com/dasilro/go_lib_response/response"
	"github.com/dasilro/gocourse_meta/meta"

	courseSdk "github.com/dasilro/go_course_sdk/course"
	userSdk "github.com/dasilro/go_course_sdk/user"
)

type (
	Controller func(ctx context.Context, request interface{}) (interface{}, error)

	Endpoints struct {
		Create Controller
		GetAll Controller
		Update Controller
	}

	CreateReq struct {
		CourseID string `json:"course_id"`
		UserID   string `json:"user_id"`
	}

	GetAllReq struct {
		CourseID string `json:"course_id"`
		UserID   string `json:"user_id"`
		Limit    int
		Page     int
	}

	UpdateReq struct {
		ID     string  `json:"id"`
		Status *string `json:"status"`
	}
)

func MakeEndpoints(s Service, paginatorLimitDefault string) Endpoints {
	return Endpoints{
		Create: makeCreateEndpoint(s),
		GetAll: makeGetAllEndpoint(s, paginatorLimitDefault),
		Update: makeUpdateEndpoint(s),
	}
}

func makeCreateEndpoint(s Service) Controller {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreateReq)

		if req.CourseID == "" {
			return nil, response.BadRequest(ErrCourseIdRequired.Error())
		}

		if req.UserID == "" {

			return nil, response.BadRequest(ErrUserIdRequired.Error())
		}

		enrollment, err := s.Create(ctx, req.CourseID, req.UserID)
		if err != nil {
			if errors.As(err, &courseSdk.ErrNotFound{}) ||
				errors.As(err, &userSdk.ErrNotFound{}) {
				return nil, response.NotFound(err.Error())
			}
			return nil, response.InternalServerError(err.Error())
		}
		return response.Created("success", enrollment, nil), nil
	}
}

func makeGetAllEndpoint(s Service, paginatorLimitDefault string) Controller {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetAllReq)
		filters := Filters{
			CourseID: req.CourseID,
			UserID:   req.UserID,
		}

		count, err := s.Count(ctx, filters)
		if err != nil {
			return nil, response.InternalServerError(err.Error())
		}

		meta, err := meta.New(req.Page, req.Limit, count, paginatorLimitDefault)
		if err != nil {
			return nil, response.InternalServerError(err.Error())
		}

		users, err := s.GetAll(ctx, filters, meta.Offset(), meta.Limit())
		if err != nil {
			return nil, response.InternalServerError(err.Error())
		}
		return response.OK("success", users, meta), nil
	}
}

func makeUpdateEndpoint(s Service) Controller {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateReq)

		if req.Status != nil && *req.Status == "" {
			return nil, response.BadRequest(ErrStatusRequired.Error())
		}

		err := s.Update(ctx, req.ID, req.Status)
		if err != nil {
			if errors.As(err, &ErrInvalidStatus{}) {
				return nil, response.BadRequest(err.Error())
			}

			if errors.As(err, &ErrNotFound{}) {
				return nil, response.NotFound(err.Error())
			}
			return nil, response.InternalServerError(err.Error())
		}
		return response.OK("success", nil, nil), nil
	}
}
