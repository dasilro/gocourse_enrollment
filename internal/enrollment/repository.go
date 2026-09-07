package enrollment

import (
	"context"
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, enrollment *Enrollment) error
	GetAll(ctx context.Context, filters Filters, offset, limit int) ([]Enrollment, error)
	Update(ctx context.Context, id string, status *string) error
	Count(ctx context.Context, filters Filters) (int, error)
}

type repo struct {
	log *log.Logger
	db  *gorm.DB
}

func NewRepo(log *log.Logger, db *gorm.DB) Repository {
	return &repo{log: log, db: db}
}

func (repo *repo) Create(ctx context.Context, enrollment *Enrollment) error {
	if err := repo.db.WithContext(ctx).Create(enrollment).Error; err != nil {
		repo.log.Println(err)
		return err
	}
	repo.log.Println("Create enrollment with id: ", enrollment.ID)
	return nil
}

func (repo *repo) GetAll(ctx context.Context, filters Filters, offset int, limit int) ([]Enrollment, error) {
	var u []Enrollment
	tx := applyFilters(repo.db.WithContext(ctx).Model(&u), filters)
	tx = tx.Limit(limit).Offset(offset)
	err := tx.Order("created_at desc").Find(&u).Error

	if err != nil {
		repo.log.Println(err)
		return nil, err
	}

	return u, nil
}

func applyFilters(tx *gorm.DB, filters Filters) *gorm.DB {

	if filters.CourseID != "" {
		filters.CourseID = fmt.Sprintf("%s", strings.ToLower(filters.CourseID))
		tx = tx.Where("lower(course_id) = ?", filters.CourseID)
	}

	if filters.UserID != "" {
		filters.UserID = fmt.Sprintf("%s", strings.ToLower(filters.UserID))
		tx = tx.Where("lower(user_id) = ?", filters.UserID)
	}

	return tx
}

func (repo *repo) Count(ctx context.Context, filters Filters) (int, error) {
	var count int64
	tx := repo.db.WithContext(ctx).Model(&Enrollment{})
	tx = applyFilters(tx, filters)
	if err := tx.Count(&count).Error; err != nil {
		repo.log.Println(err)
		return 0, err
	}
	return int(count), nil
}

func (repo *repo) Update(ctx context.Context, id string, status *string) error {
	values := make(map[string]interface{})

	if status != nil {
		values["status"] = *status
	}

	result := repo.db.WithContext(ctx).Model(&Enrollment{}).Where("id = ?", id).Updates(values)
	if result.Error != nil {
		repo.log.Println(result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		repo.log.Printf("user %s doesn't exist", id)
		return ErrNotFound{EnrollmentID: id}
	}
	return nil
}
