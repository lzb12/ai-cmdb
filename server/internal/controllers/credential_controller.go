package controllers

import (
	"errors"
	"net/http"

	"cmdb-server/internal/services"
	"cmdb-server/internal/utils"
	"github.com/gin-gonic/gin"
)

type CredentialController struct {
	credentialService *services.CredentialService
}

type upsertCredentialRequest struct {
	Name        string `json:"name"`
	AuthType    string `json:"authType"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	PrivateKey  string `json:"privateKey"`
	Passphrase  string `json:"passphrase"`
	Description string `json:"description"`
}

func NewCredentialController(credentialService *services.CredentialService) *CredentialController {
	return &CredentialController{credentialService: credentialService}
}

func (ctl *CredentialController) List(c *gin.Context) {
	items, err := ctl.credentialService.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询凭据失败"})
		return
	}

	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, toCredentialSafe(&item))
	}
	c.JSON(http.StatusOK, gin.H{"items": result})
}

func (ctl *CredentialController) Create(c *gin.Context) {
	var req upsertCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	input := services.UpsertCredentialInput{
		Name:        req.Name,
		AuthType:    req.AuthType,
		Username:    req.Username,
		Password:    req.Password,
		PrivateKey:  req.PrivateKey,
		Passphrase:  req.Passphrase,
		Description: req.Description,
	}

	item, err := ctl.credentialService.Create(input)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentialInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "凭据字段不完整或格式错误"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "保存凭据失败，名称可能重复"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": item.ID})
}

func (ctl *CredentialController) Update(c *gin.Context) {
	id, err := utils.ParseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效凭据 ID"})
		return
	}

	var req upsertCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	input := services.UpsertCredentialInput{
		Name:        req.Name,
		AuthType:    req.AuthType,
		Username:    req.Username,
		Password:    req.Password,
		PrivateKey:  req.PrivateKey,
		Passphrase:  req.Passphrase,
		Description: req.Description,
	}

	if err := ctl.credentialService.Update(id, input); err != nil {
		if errors.Is(err, services.ErrInvalidCredentialInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "更新字段不完整或格式错误"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "更新凭据失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func (ctl *CredentialController) Delete(c *gin.Context) {
	id, err := utils.ParseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效凭据 ID"})
		return
	}

	err = ctl.credentialService.Delete(id)
	if err != nil {
		if errors.Is(err, services.ErrCredentialInUse) {
			c.JSON(http.StatusConflict, gin.H{"error": "凭据仍被资产绑定，无法删除"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除凭据失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
