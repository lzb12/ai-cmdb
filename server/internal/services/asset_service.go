package services

import (
	"errors"
	"strings"

	"cmdb-server/internal/global"
	"cmdb-server/internal/models"
	"cmdb-server/internal/utils"
)

var ErrInvalidAssetInput = errors.New("invalid asset input")

type AssetService struct{}

type UpsertAssetInput struct {
	Hostname     string
	Address      string
	Port         int
	OS           string
	Environment  string
	Owner        string
	CredentialID *uint
	Tags         []string
}

func NewAssetService() *AssetService {
	return &AssetService{}
}

func (s *AssetService) List() ([]models.HostAsset, error) {
	var items []models.HostAsset
	err := global.DB.Preload("Credential").Order("updated_at desc").Find(&items).Error
	return items, err
}

func (s *AssetService) Create(input UpsertAssetInput) (*models.HostAsset, error) {
	asset, err := normalizeAssetInput(input)
	if err != nil {
		return nil, err
	}
	if err := global.DB.Create(asset).Error; err != nil {
		return nil, err
	}
	return asset, nil
}

func (s *AssetService) Update(id uint, input UpsertAssetInput) error {
	asset, err := normalizeAssetInput(input)
	if err != nil {
		return err
	}

	var existing models.HostAsset
	if err := global.DB.First(&existing, id).Error; err != nil {
		return err
	}

	existing.Hostname = asset.Hostname
	existing.Address = asset.Address
	existing.Port = asset.Port
	existing.OS = asset.OS
	existing.Environment = asset.Environment
	existing.Owner = asset.Owner
	existing.CredentialID = asset.CredentialID
	existing.TagsText = asset.TagsText

	return global.DB.Save(&existing).Error
}

func (s *AssetService) Delete(id uint) error {
	return global.DB.Delete(&models.HostAsset{}, id).Error
}

func normalizeAssetInput(input UpsertAssetInput) (*models.HostAsset, error) {
	hostname := strings.TrimSpace(input.Hostname)
	address := strings.TrimSpace(input.Address)
	if hostname == "" || address == "" {
		return nil, ErrInvalidAssetInput
	}

	port := input.Port
	if port <= 0 {
		port = 22
	}
	env := strings.TrimSpace(input.Environment)
	if env == "" {
		env = "prod"
	}
	owner := strings.TrimSpace(input.Owner)
	if owner == "" {
		owner = "未指定"
	}

	return &models.HostAsset{
		Hostname:     hostname,
		Address:      address,
		Port:         port,
		OS:           strings.TrimSpace(input.OS),
		Environment:  env,
		Owner:        owner,
		CredentialID: input.CredentialID,
		TagsText:     utils.JoinTags(input.Tags),
	}, nil
}
