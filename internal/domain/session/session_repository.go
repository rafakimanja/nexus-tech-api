package session

import (
	"context"
	"errors"
	"strings"

	"api/internal/domain/applicant"
	"api/internal/domain/teacher"
	"api/internal/domain/uauth"
	"api/internal/domain/user"
	"gorm.io/gorm"
)

var ErrAccountConflict = errors.New("account requires identity linking")

type Identity struct{ Subject, Email, Name, Picture string }

type Repository interface {
	FindOrCreate(context.Context, Identity) (user.User, error)
	Current(context.Context, uint) (Profile, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) FindOrCreate(ctx context.Context, identity Identity) (user.User, error) {
	var account user.User
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Serialize creation for the same email or Google subject, including first login.
		for _, key := range []string{"login:email:" + identity.Email, "login:google:" + identity.Subject} {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", key).Error; err != nil {
				return err
			}
		}
		var link uauth.UserAuth
		err := tx.Unscoped().Where("provider = ? AND provider_user_id = ?", "google", identity.Subject).First(&link).Error
		if err == nil {
			if link.DeletedAt.Valid {
				return ErrAccountConflict
			}
			return tx.First(&account, link.UserID).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var count int64
		if err := tx.Unscoped().Model(&user.User{}).Where("lower(email) = ?", identity.Email).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrAccountConflict
		}
		account = user.User{Name: identity.Name, Email: identity.Email, Role: "demandante", Foto_url: &identity.Picture}
		if err := tx.Create(&account).Error; err != nil {
			return err
		}
		return tx.Create(&uauth.UserAuth{UserID: account.ID, Provider: "google", ProviderUserID: identity.Subject}).Error
	})
	return account, err
}

type Profile struct {
	ID              uint    `json:"id"`
	Name            string  `json:"name"`
	Email           string  `json:"email"`
	Role            string  `json:"role"`
	PhotoURL        *string `json:"photo_url"`
	ProfileComplete bool    `json:"profile_complete"`
	Blocked         bool    `json:"blocked"`
	BlockReason     string  `json:"block_reason,omitempty"`
}

func (r *repository) Current(ctx context.Context, id uint) (Profile, error) {
	db := r.db.WithContext(ctx)
	var account user.User
	if err := db.First(&account, id).Error; err != nil {
		return Profile{}, err
	}
	p := Profile{ID: account.ID, Name: account.Name, Email: account.Email, Role: account.Role, PhotoURL: account.Foto_url}
	switch account.Role {
	case "admin":
		p.ProfileComplete = true
	case "professor":
		var teacher teacher.Professor
		err := db.First(&teacher, id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return p, nil
		}
		if err != nil {
			return p, err
		}
		p.ProfileComplete = teacher.Siape != 0
		p.Blocked = !teacher.Is_active
		if p.Blocked {
			p.BlockReason = "professor_inactive"
		}
	case "demandante":
		var applicant applicant.Demandante
		err := db.First(&applicant, id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return p, nil
		}
		if err != nil {
			return p, err
		}
		p.ProfileComplete = strings.TrimSpace(applicant.Phone) != "" && strings.TrimSpace(applicant.Type) != "" && strings.TrimSpace(applicant.CpfCnpj) != "" && strings.TrimSpace(applicant.Adress) != ""
		p.Blocked = applicant.Strikes >= 3
		if p.Blocked {
			p.BlockReason = "strike_limit"
		}
	default:
		p.Blocked = true
		p.BlockReason = "invalid_role"
	}
	return p, nil
}
