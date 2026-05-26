package controllers

import (
	"errors"
	"net/http"

	"cmdb-server/internal/services"
	"cmdb-server/internal/utils"
	"github.com/gin-gonic/gin"
)

type AssetController struct {
	assetService *services.AssetService
}

type upsertAssetRequest struct {
	Hostname     string   `json:"hostname"`
	Address      string   `json:"address"`
	Port         int      `json:"port"`
	OS           string   `json:"os"`
	Environment  string   `json:"environment"`
	Owner        string   `json:"owner"`
	CredentialID *uint    `json:"credentialId"`
	Tags         []string `json:"tags"`
}

func NewAssetController(assetService *services.AssetService) *AssetController {
	return &AssetController{assetService: assetService}
}

func (ctl *AssetController) List(c *gin.Context) {
	items, err := ctl.assetService.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询资产失败"})
		return
	}

	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, gin.H{
			"id":           item.ID,
			"hostname":     item.Hostname,
			"address":      item.Address,
			"port":         item.Port,
			"os":           item.OS,
			"environment":  item.Environment,
			"owner":        item.Owner,
			"credentialId": item.CredentialID,
			"tags":         utils.SplitTags(item.TagsText),
			"createdAt":    item.CreatedAt,
			"updatedAt":    item.UpdatedAt,
			"credential":   toCredentialSafe(item.Credential),
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": result})
}

func (ctl *AssetController) Create(c *gin.Context) {
	var req upsertAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	input := services.UpsertAssetInput{
		Hostname:     req.Hostname,
		Address:      req.Address,
		Port:         req.Port,
		OS:           req.OS,
		Environment:  req.Environment,
		Owner:        req.Owner,
		CredentialID: req.CredentialID,
		Tags:         req.Tags,
	}
	asset, err := ctl.assetService.Create(input)
	if err != nil {
		if errors.Is(err, services.ErrInvalidAssetInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "hostname 和 address 必填"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "创建资产失败，主机名可能重复"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": asset.ID})
}

func (ctl *AssetController) Update(c *gin.Context) {
	id, err := utils.ParseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效资产 ID"})
		return
	}

	var req upsertAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	input := services.UpsertAssetInput{
		Hostname:     req.Hostname,
		Address:      req.Address,
		Port:         req.Port,
		OS:           req.OS,
		Environment:  req.Environment,
		Owner:        req.Owner,
		CredentialID: req.CredentialID,
		Tags:         req.Tags,
	}

	if err := ctl.assetService.Update(id, input); err != nil {
		if errors.Is(err, services.ErrInvalidAssetInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "hostname 和 address 必填"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "更新资产失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func (ctl *AssetController) Delete(c *gin.Context) {
	id, err := utils.ParseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效资产 ID"})
		return
	}

	if err := ctl.assetService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
