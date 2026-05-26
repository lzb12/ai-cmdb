package controllers

import (
	"cmdb-server/internal/models"
	"github.com/gin-gonic/gin"
)

func toCredentialSafe(item *models.Credential) gin.H {
	if item == nil {
		return nil
	}
	return gin.H{
		"id":          item.ID,
		"name":        item.Name,
		"authType":    item.AuthType,
		"username":    item.Username,
		"description": item.Description,
		"createdAt":   item.CreatedAt,
		"updatedAt":   item.UpdatedAt,
	}
}
