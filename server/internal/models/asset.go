package models

import "time"

type HostAsset struct {
	ID           uint        `gorm:"primaryKey" json:"id"`
	Hostname     string      `gorm:"size:120;not null;uniqueIndex" json:"hostname"`
	Address      string      `gorm:"size:120;not null" json:"address"`
	Port         int         `gorm:"not null;default:22" json:"port"`
	OS           string      `gorm:"size:80" json:"os"`
	Environment  string      `gorm:"size:40;not null;default:prod" json:"environment"`
	Owner        string      `gorm:"size:120" json:"owner"`
	CredentialID *uint       `gorm:"index" json:"credentialId"`
	Credential   *Credential `gorm:"foreignKey:CredentialID" json:"credential,omitempty"`
	TagsText     string      `gorm:"column:tags_text;type:text" json:"-"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}

func (HostAsset) TableName() string {
	return "host_assets"
}
