package profile

import (
	"api/internal/domain/areas_teacher"
	"context"
	"errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"testing"
)

func TestProfilePersistencePostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not configured")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(1)
	for _, q := range []string{
		`CREATE TEMP TABLE users (id bigint PRIMARY KEY, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz, name text, role text, email text, foto_url text)`,
		`CREATE TEMP TABLE professors (id bigint PRIMARY KEY REFERENCES users(id), siape text, is_active boolean)`,
		`CREATE TEMP TABLE demandantes (id bigint PRIMARY KEY REFERENCES users(id), phone text, type text, cpf_cnpj text, adress text, strikes integer)`,
		`CREATE TEMP TABLE areas (id bigint PRIMARY KEY, name text)`,
		`CREATE TEMP TABLE professor_areas (id bigserial PRIMARY KEY, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz, professor_id bigint REFERENCES professors(id), areas_id bigint REFERENCES areas(id), type text)`,
		`INSERT INTO users (id,name,role,email) VALUES (1,'Old','professor','teacher@example.com'),(2,'Applicant','demandante','applicant@example.com')`,
		`INSERT INTO areas VALUES (1,'Technical'),(2,'Interest')`,
	} {
		if err := db.Exec(q).Error; err != nil {
			t.Fatal(err)
		}
	}
	r := NewRepository(db)
	ctx := context.Background()
	input := TeacherInput{Name: "New", Siape: "0123456", Areas: []areas_teacher.AssociationInput{{AreasID: 1, Type: "técnica"}, {AreasID: 2, Type: "interesse"}}}
	if err := r.SaveTeacher(ctx, 1, input, true); err != nil {
		t.Fatal(err)
	}
	var professor struct {
		Siape    string
		IsActive bool
	}
	if err := db.Table("professors").Where("id = 1").Take(&professor).Error; err != nil {
		t.Fatal(err)
	}
	if professor.Siape != "0123456" || !professor.IsActive {
		t.Fatalf("unexpected professor: %+v", professor)
	}
	if err := r.SaveTeacher(ctx, 1, input, true); !errors.Is(err, ErrExists) {
		t.Fatalf("duplicate create: %v", err)
	}
	input.Name = "Must rollback"
	input.Areas[1].AreasID = 999
	if err := r.SaveTeacher(ctx, 1, input, false); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("invalid area: %v", err)
	}
	var name string
	if err := db.Table("users").Where("id=1").Pluck("name", &name).Error; err != nil {
		t.Fatal(err)
	}
	if name != "New" {
		t.Fatal("account changes were not rolled back")
	}
	if err := db.Exec("UPDATE professors SET is_active=false WHERE id=1").Error; err != nil {
		t.Fatal(err)
	}
	input.Areas[1].AreasID = 2
	if err := r.SaveTeacher(ctx, 1, input, false); !errors.Is(err, ErrForbidden) {
		t.Fatalf("inactive teacher updated: %v", err)
	}
	app := ApplicantInput{Name: "Applicant", Phone: "51999999999", Type: "pessoaFisica", CpfCnpj: "12345678900", Address: "Address"}
	if err := r.SaveApplicant(ctx, 2, app, true); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("UPDATE demandantes SET strikes=2 WHERE id=2").Error; err != nil {
		t.Fatal(err)
	}
	app.Address = "Changed"
	if err := r.SaveApplicant(ctx, 2, app, false); err != nil {
		t.Fatal(err)
	}
	var strikes int
	if err := db.Table("demandantes").Where("id=2").Pluck("strikes", &strikes).Error; err != nil {
		t.Fatal(err)
	}
	if strikes != 2 {
		t.Fatalf("strikes changed to %d", strikes)
	}
}
