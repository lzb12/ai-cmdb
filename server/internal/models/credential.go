package models

import "time"

type Credential struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:120;not null;uniqueIndex" json:"name"`
	AuthType      string    `gorm:"size:20;not null" json:"authType"`
	Username      string    `gorm:"size:120;not null" json:"username"`
	PasswordEnc   string    `gorm:"type:text" json:"-"`
	PrivateKeyEnc string    `gorm:"type:longtext" json:"-"`
	PassphraseEnc string    `gorm:"type:text" json:"-"`
	Description   string    `gorm:"size:300" json:"description"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (Credential) TableName() string {
	return "credentials"
}
