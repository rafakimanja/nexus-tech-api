	// Command seed creates the database schema and inserts development data.
// Run it independently with: go run ./cmd/seed
package main

import (
	"fmt"
	"log"

	"api/internal/database"
	"api/internal/domain/applicant"
	"api/internal/domain/areas"
	"api/internal/domain/areas_teacher"
	"api/internal/domain/demand"
	"api/internal/domain/teacher"
	"api/internal/domain/uauth"
	"api/internal/domain/user"

	"gorm.io/gorm"
)

func main() {
	dbService := database.New()
	defer func() {
		if err := dbService.Close(); err != nil {
			log.Printf("could not close database connection: %v", err)
		}
	}()

	db := dbService.DB()
	models := []any{
		&user.User{}, &uauth.UserAuth{}, &areas.Area{}, &areas_teacher.ProfessorArea{},
		&teacher.Professor{}, &applicant.Demandante{}, &demand.Demanda{},
	}
	if err := db.AutoMigrate(models...); err != nil {
		log.Fatalf("could not migrate database: %v", err)
	}
	// Older schema versions may have attached profile-to-user foreign keys to
	// users instead of their profile tables. Those reversed constraints prevent
	// creating users whose IDs do not already exist in the profile tables.
	for _, constraint := range []string{"fk_professors_user", "fk_demandantes_user"} {
		if err := db.Exec("ALTER TABLE users DROP CONSTRAINT IF EXISTS " + constraint).Error; err != nil {
			log.Fatalf("could not repair users foreign key %s: %v", constraint, err)
		}
	}

	if err := seedAreas(db); err != nil {
		log.Fatalf("could not seed areas: %v", err)
	}
	if err := seedDemoRecords(db); err != nil {
		log.Fatalf("could not seed demo records: %v", err)
	}

	fmt.Println("Banco preparado com áreas e registros de demonstração.")
}

func seedAreas(db *gorm.DB) error {
	names := []string{
		"Desenvolvimento de Software",
		"Inteligência Artificial",
		"Banco de Dados",
		"Redes de Computadores",
	}
	for _, name := range names {
		area := areas.Area{Name: name}
		if err := db.Where("name = ?", name).FirstOrCreate(&area).Error; err != nil {
			return fmt.Errorf("area %q: %w", name, err)
		}
	}
	return nil
}

func seedDemoRecords(db *gorm.DB) error {
	requesterSpecs := []struct {
		name, email, phone, cpf, address string
	}{
		{"Camila Ribeiro", "camila.ribeiro@demo.example", "+55 (51) 98472-1635", "MOCK-CPF-741852", "Rua das Acácias, 142, Centro, Porto Alegre/RS"},
		{"Marina Oliveira", "marina.oliveira@demo.example", "+55 (51) 99731-5084", "MOCK-CPF-638419", "Av. Bento Gonçalves, 860, Partenon, Porto Alegre/RS"},
		{"Lucas Martins", "lucas.martins@demo.example", "+55 (51) 99126-7348", "MOCK-CPF-295763", "Rua Sete de Setembro, 57, Centro, Osório/RS"},
	}
	requesters := make([]applicant.Demandante, 0, len(requesterSpecs))
	for _, spec := range requesterSpecs {
		account := user.User{Name: spec.name, Email: spec.email}
		if err := db.Where("email = ?", spec.email).FirstOrCreate(&account).Error; err != nil {
			return fmt.Errorf("demandante user %s: %w", spec.email, err)
		}
		if err := db.Model(&account).Update("role", "demandante").Error; err != nil {
			return fmt.Errorf("demandante role %s: %w", spec.email, err)
		}
		profile := applicant.Demandante{
			ID: account.ID, Phone: spec.phone, Type: "Pessoa física",
			CpfCnpj: spec.cpf, Adress: spec.address, Strikes: 0,
		}
		if err := db.Where("id = ?", account.ID).FirstOrCreate(&profile).Error; err != nil {
			return fmt.Errorf("demandante profile %s: %w", spec.email, err)
		}
		requesters = append(requesters, profile)
	}

	professorSpecs := []struct {
		name, email string
		siape       uint
		areas       []string
	}{
		{"Prof. Rafael Almeida", "rafael.almeida@demo.example", 4837216, []string{"Desenvolvimento de Software", "Banco de Dados"}},
		{"Profa. Ana Souza", "ana.souza@demo.example", 5271943, []string{"Inteligência Artificial"}},
		{"Prof. Bruno Lima", "bruno.lima@demo.example", 6193852, []string{"Redes de Computadores", "Desenvolvimento de Software"}},
	}
	professors := make([]teacher.Professor, 0, len(professorSpecs))
	for _, spec := range professorSpecs {
		account := user.User{Name: spec.name, Email: spec.email}
		if err := db.Where("email = ?", spec.email).FirstOrCreate(&account).Error; err != nil {
			return fmt.Errorf("professor user %s: %w", spec.email, err)
		}
		if err := db.Model(&account).Update("role", "professor").Error; err != nil {
			return fmt.Errorf("professor role %s: %w", spec.email, err)
		}
		profile := teacher.Professor{ID: account.ID, Siape: spec.siape, Is_active: true}
		if err := db.Where("id = ?", account.ID).FirstOrCreate(&profile).Error; err != nil {
			return fmt.Errorf("professor profile %s: %w", spec.email, err)
		}
		for _, areaName := range spec.areas {
			var area areas.Area
			if err := db.Where("name = ?", areaName).First(&area).Error; err != nil {
				return fmt.Errorf("find area %q: %w", areaName, err)
			}
			association := areas_teacher.ProfessorArea{
				ProfessorID: profile.ID, AreasID: area.ID,
				Type: areas_teacher.AreaTypeTechnical,
			}
			if err := db.Where("professor_id = ? AND areas_id = ?", profile.ID, area.ID).
				FirstOrCreate(&association).Error; err != nil {
				return fmt.Errorf("professor area %s: %w", spec.email, err)
			}
		}
		professors = append(professors, profile)
	}

	areaIDs := make(map[string]uint)
	for _, name := range []string{"Desenvolvimento de Software", "Inteligência Artificial", "Banco de Dados", "Redes de Computadores"} {
		var area areas.Area
		if err := db.Where("name = ?", name).First(&area).Error; err != nil {
			return fmt.Errorf("find area %q: %w", name, err)
		}
		areaIDs[name] = area.ID
	}
	// Stable titles make the sample requests safe to create on every run.
	demandSpecs := []struct {
		title, description, status, area string
		requester, professor             int
	}{
		{"Sistema de agendamento de salas (Demo)", "Criar uma aplicação web simples para consultar e reservar salas disponíveis.", "Aberta", "Desenvolvimento de Software", 0, 0},
		{"Análise de dados de evasão (Demo)", "Explorar dados acadêmicos anonimizados para identificar fatores associados à evasão.", "Em andamento", "Inteligência Artificial", 1, 1},
		{"Organização do acervo digital (Demo)", "Modelar um banco de dados para catalogar materiais didáticos e facilitar buscas.", "Aberta", "Banco de Dados", 2, 0},
		{"Melhoria da rede do laboratório (Demo)", "Avaliar a cobertura e sugerir melhorias para a rede do laboratório de informática.", "Aberta", "Redes de Computadores", 0, 2},
	}
	for _, spec := range demandSpecs {
		entry := demand.Demanda{
			DemandanteID: requesters[spec.requester].ID,
			ProfessorID:  professors[spec.professor].ID,
			AreaID:       areaIDs[spec.area],
			Title:        spec.title, Description: spec.description,
			Status: spec.status, IsCancelled: false,
		}
		if err := db.Where("title = ?", spec.title).FirstOrCreate(&entry).Error; err != nil {
			return fmt.Errorf("demand %q: %w", spec.title, err)
		}
	}
	return nil
}
