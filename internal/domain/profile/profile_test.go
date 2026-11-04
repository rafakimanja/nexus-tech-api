package profile

import (
	"api/internal/domain/areas_teacher"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeRepo struct {
	calls     int
	id        uint
	teacher   TeacherInput
	applicant ApplicantInput
}

func (r *fakeRepo) SaveTeacher(_ context.Context, id uint, input TeacherInput, _ bool) error {
	r.calls++
	r.id = id
	r.teacher = input
	return nil
}
func (r *fakeRepo) SaveApplicant(_ context.Context, id uint, input ApplicantInput, _ bool) error {
	r.calls++
	r.id = id
	r.applicant = input
	return nil
}

func TestProfileRejectsProtectedFieldsAndForeignOwners(t *testing.T) {
	for _, tc := range []struct {
		name, role, method, path, body string
		status                         int
	}{
		{"activation", "professor", "POST", "/teachers/", `{"name":"Nome","siape":"0123456","is_active":true}`, 400},
		{"owner input", "professor", "POST", "/teachers/", `{"id":2}`, 400},
		{"foreign owner", "professor", "PUT", "/teachers/2", `{}`, 403},
		{"wrong role", "demandante", "POST", "/teachers/", `{}`, 403},
		{"strikes", "demandante", "POST", "/applicants/", `{"strikes":0}`, 400},
		{"role", "demandante", "POST", "/applicants/", `{"role":"admin"}`, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{}
			handler := NewHandler(NewService(repo), nil)
			r := gin.New()
			r.Use(func(c *gin.Context) { c.Set("userID", uint(1)); c.Set("role", tc.role); c.Next() })
			r.POST("/teachers/", handler.SaveTeacher)
			r.PUT("/teachers/:id", handler.SaveTeacher)
			r.POST("/applicants/", handler.SaveApplicant)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
			if w.Code != tc.status || repo.calls != 0 {
				t.Fatalf("status=%d calls=%d body=%s", w.Code, repo.calls, w.Body.String())
			}
		})
	}
}
func TestTeacherValidationAndLeadingZero(t *testing.T) {
	r := &fakeRepo{}
	s := NewService(r)
	input := TeacherInput{Name: " Nome ", Siape: "0123456", Areas: []areas_teacher.AssociationInput{{AreasID: 1, Type: "técnica"}, {AreasID: 2, Type: "interesse"}}}
	if err := s.SaveTeacher(context.Background(), 7, input, true); err != nil {
		t.Fatal(err)
	}
	if r.id != 7 || r.teacher.Siape != "0123456" || r.teacher.Name != "Nome" {
		t.Fatalf("unexpected input: %+v", r.teacher)
	}
	input.Areas = input.Areas[:1]
	if err := s.SaveTeacher(context.Background(), 7, input, true); !errors.Is(err, ErrInvalid) {
		t.Fatalf("missing interest accepted: %v", err)
	}
}
