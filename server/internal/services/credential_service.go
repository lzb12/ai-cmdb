package services

import (
	"errors"
	"strings"

	"cmdb-server/internal/global"
	"cmdb-server/internal/models"
)

var ErrInvalidCredentialInput = errors.New("invalid credential input")
var ErrCredentialInUse = errors.New("credential in use")

type CredentialService struct{}

type UpsertCredentialInput struct {
	Name        string
	AuthType    string
	Username    string
	Password    string
	PrivateKey  string
	Passphrase  string
	Description string
}

func NewCredentialService() *CredentialService {
	return &CredentialService{}
}

func (s *CredentialService) List() ([]models.Credential, error) {
	var items []models.Credential
	err := global.DB.Order("updated_at desc").Find(&items).Error
	return items, err
}

func (s *CredentialService) Create(input UpsertCredentialInput) (*models.Credential, error) {
	next, err := s.normalizeInput(input)
	if err != nil {
		return nil, err
	}
	if err := global.DB.Create(next).Error; err != nil {
		return nil, err
	}
	return next, nil
}

func (s *CredentialService) Update(id uint, input UpsertCredentialInput) error {
	var existing models.Credential
	if err := global.DB.First(&existing, id).Error; err != nil {
		return err
	}

	name := strings.TrimSpace(input.Name)
	username := strings.TrimSpace(input.Username)
	description := strings.TrimSpace(input.Description)
	if name == "" || username == "" {
		return ErrInvalidCredentialInput
	}

	authType := strings.TrimSpace(input.AuthType)
	if authType == "" {
		authType = existing.AuthType
	}
	if authType != "password" && authType != "key" {
		return ErrInvalidCredentialInput
	}

	existing.Name = name
	existing.Username = username
	existing.Description = description
	existing.AuthType = authType

	if authType == "password" {
		if input.Password != "" {
			enc, err := global.ENCRYPTOR.Encrypt(input.Password)
			if err != nil {
				return err
			}
			existing.PasswordEnc = enc
		}
		existing.PrivateKeyEnc = ""
		existing.PassphraseEnc = ""
		if existing.PasswordEnc == "" {
			return ErrInvalidCredentialInput
		}
	} else {
		if strings.TrimSpace(input.PrivateKey) != "" {
			enc, err := global.ENCRYPTOR.Encrypt(input.PrivateKey)
			if err != nil {
				return err
			}
			existing.PrivateKeyEnc = enc
		}
		if input.Passphrase != "" {
			enc, err := global.ENCRYPTOR.Encrypt(input.Passphrase)
			if err != nil {
				return err
			}
			existing.PassphraseEnc = enc
		}
		existing.PasswordEnc = ""
		if existing.PrivateKeyEnc == "" {
			return ErrInvalidCredentialInput
		}
	}

	return global.DB.Save(&existing).Error
}

func (s *CredentialService) Delete(id uint) error {
	var assetCount int64
	if err := global.DB.Model(&models.HostAsset{}).Where("credential_id = ?", id).Count(&assetCount).Error; err != nil {
		return err
	}
	if assetCount > 0 {
		return ErrCredentialInUse
	}
	return global.DB.Delete(&models.Credential{}, id).Error
}

func (s *CredentialService) FindByID(id uint) (*models.Credential, error) {
	var item models.Credential
	if err := global.DB.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *CredentialService) normalizeInput(input UpsertCredentialInput) (*models.Credential, error) {
	name := strings.TrimSpace(input.Name)
	authType := strings.TrimSpace(input.AuthType)
	username := strings.TrimSpace(input.Username)
	description := strings.TrimSpace(input.Description)

	if name == "" || username == "" {
		return nil, ErrInvalidCredentialInput
	}
	if authType != "password" && authType != "key" {
		return nil, ErrInvalidCredentialInput
	}

	next := &models.Credential{
		Name:        name,
		AuthType:    authType,
		Username:    username,
		Description: description,
	}

	if authType == "password" {
		if input.Password == "" {
			return nil, ErrInvalidCredentialInput
		}
		enc, err := global.ENCRYPTOR.Encrypt(input.Password)
		if err != nil {
			return nil, err
		}
		next.PasswordEnc = enc
	} else {
		privateKey := strings.TrimSpace(input.PrivateKey)
		if privateKey == "" {
			return nil, ErrInvalidCredentialInput
		}
		enc, err := global.ENCRYPTOR.Encrypt(privateKey)
		if err != nil {
			return nil, err
		}
		next.PrivateKeyEnc = enc
		if input.Passphrase != "" {
			passEnc, err := global.ENCRYPTOR.Encrypt(input.Passphrase)
			if err != nil {
				return nil, err
			}
			next.PassphraseEnc = passEnc
		}
	}

	return next, nil
}
