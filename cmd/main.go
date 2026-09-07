package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	courseSdk "github.com/dasilro/go_course_sdk/course"
	userSdk "github.com/dasilro/go_course_sdk/user"
	"github.com/dasilro/gocourse_enrollment/internal/enrollment"
	"github.com/dasilro/gocourse_enrollment/pkg/bootstrap"
	"github.com/dasilro/gocourse_enrollment/pkg/handler"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	l := bootstrap.InitLogger()

	db, err := bootstrap.DBConnection()
	if err != nil {
		l.Fatal(err)
	}

	courseTrans := courseSdk.NewHttpClient(os.Getenv("API_COURSE_URL"), "")
	userTrans := userSdk.NewHttpClient(os.Getenv("API_USER_URL"), "")

	ctx := context.Background()
	enrollmentRepo := enrollment.NewRepo(l, db)
	enrollmentSrv := enrollment.NewService(l, courseTrans, userTrans, enrollmentRepo)

	h := handler.NewEnrollmentHTTPServer(ctx, enrollment.MakeEndpoints(enrollmentSrv, os.Getenv("PAGINATOR_LIMIT_DEFAULT")))

	port := os.Getenv("PORT")
	address := fmt.Sprintf("127.0.0.1:%s", port)
	srv := &http.Server{
		Handler:      accessControl(h),
		Addr:         address,
		WriteTimeout: time.Second * 5,
		ReadTimeout:  time.Second * 5,
	}

	errCh := make(chan error)
	go func() {
		l.Println("listen in ", address)
		errCh <- srv.ListenAndServe()
	}()

	err = <-errCh
	if err != nil {
		log.Fatal(err)
	}
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, PUT, OPTIONS, HEAD")
		w.Header().Set("Access-Control-Allow-Headers", "Accept,Authorization,Cache-Control,Content-Type")

		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)
	})

}
