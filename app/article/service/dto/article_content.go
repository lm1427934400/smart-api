package dto

import (
	"smart-api/app/article/models"
	"time"
)

// ArticleContentQuery 文章查询参数
type ArticleContentQuery struct {
	Title     string `json:"title" form:"title" binding:"omitempty"`
	Status    int    `json:"status" form:"status" binding:"omitempty"`
	CreateBy  int64  `json:"createBy" form:"createBy" binding:"omitempty"`
	PageIndex int    `json:"pageIndex" form:"pageIndex" binding:"omitempty,min=1"`
	PageSize  int    `json:"pageSize" form:"pageSize" binding:"omitempty,min=1,max=100"`
	ParentId  *int64 `json:"parentId" form:"parentId" binding:"omitempty"`
	Type      string `json:"type" form:"type" binding:"omitempty,oneof=article directory"`
	Level     int    `json:"level" form:"level" binding:"omitempty,min=1"`
}

// ArticleContentInsertReq 文章新增请求
type ArticleContentInsertReq struct {
	Title       string `json:"title" binding:"required,max=255"`
	Content     string `json:"content" binding:"omitempty"` // 对于目录，内容可以为空
	HtmlContent string `json:"html_content" binding:"omitempty"`
	Status      int    `json:"status" binding:"omitempty,oneof=0 1"`
	Cover       string `json:"cover" binding:"omitempty,max=500"`
	Summary     string `json:"summary" binding:"omitempty,max=500"`
	ParentId    *int64 `json:"parent_id" binding:"omitempty"`
	Type        string `json:"type" binding:"omitempty,oneof=article directory"`
}

// Generate 生成模型
func (s *ArticleContentInsertReq) Generate(model *models.ArticleContent) {
	model.Title = s.Title
	model.Content = s.Content
	model.HtmlContent = s.HtmlContent
	model.Status = s.Status
	if model.Status == 0 {
		model.Status = 1 // 默认正常状态
	}
	model.Cover = s.Cover
	model.Summary = s.Summary
	model.ParentId = s.ParentId
	// 设置默认类型为article
	model.Type = s.Type
	if model.Type == "" {
		model.Type = "article"
	}
}

// ArticleContentUpdateReq 文章更新请求
type ArticleContentUpdateReq struct {
	Id          int64  `json:"id" binding:"required"`
	Title       string `json:"title" binding:"required,max=255"`
	Content     string `json:"content" binding:"omitempty"` // 对于目录，内容可以为空
	HtmlContent string `json:"html_content" binding:"omitempty"`
	Status      int    `json:"status" binding:"omitempty,oneof=0 1"`
	Cover       string `json:"cover" binding:"omitempty,max=500"`
	Summary     string `json:"summary" binding:"omitempty,max=500"`
	ParentId    *int64 `json:"parent_id" binding:"omitempty"`
	Type        string `json:"type" binding:"omitempty,oneof=article directory"`
}

// Generate 生成模型
func (s *ArticleContentUpdateReq) Generate(model *models.ArticleContent) {
	model.Id = s.Id
	model.Title = s.Title
	model.Content = s.Content
	model.HtmlContent = s.HtmlContent
	model.Status = s.Status
	model.Cover = s.Cover
	model.Summary = s.Summary
	model.ParentId = s.ParentId
	model.Type = s.Type
}

// ArticleContentResponse 文章响应数据
type ArticleContentResponse struct {
	Id          int64                    `json:"id"`
	Title       string                   `json:"title"`
	Content     string                   `json:"content"`
	HtmlContent string                   `json:"html_content"`
	CreateBy    int64                    `json:"create_by"`
	UpdateBy    int64                    `json:"update_by"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	Status      int                      `json:"status"`
	Cover       string                   `json:"cover"`
	Summary     string                   `json:"summary"`
	ViewCount   int                      `json:"view_count"`
	ParentId    *int64                   `json:"parent_id,omitempty"`
	Type        string                   `json:"type"`
	Children    []ArticleContentResponse `json:"children,omitempty"`
	Level       int                      `json:"level,omitempty"`
	Expanded    bool                     `json:"expanded,omitempty"`
	// 可能需要关联用户信息
	CreateByName string `json:"create_by_name,omitempty"`
	UpdateByName string `json:"update_by_name,omitempty"`
}

// ToResponse 转换为响应数据
func ToResponse(model *models.ArticleContent) *ArticleContentResponse {
	response := &ArticleContentResponse{
		Id:          model.Id,
		Title:       model.Title,
		Content:     model.Content,
		HtmlContent: model.HtmlContent,
		CreateBy:    model.CreateBy,
		UpdateBy:    model.UpdateBy,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		Status:      model.Status,
		Cover:       model.Cover,
		Summary:     model.Summary,
		ViewCount:   model.ViewCount,
		ParentId:    model.ParentId,
		Type:        model.Type,
		Level:       model.Level,
		Expanded:    model.Expanded,
	}

	// 递归转换子节点
	if len(model.Children) > 0 {
		response.Children = make([]ArticleContentResponse, len(model.Children))
		for i, child := range model.Children {
			childResponse := ToResponse(&child)
			response.Children[i] = *childResponse
		}
	}

	return response
}

// ToResponseList 转换为响应数据列表
func ToResponseList(models []models.ArticleContent) []ArticleContentResponse {
	responses := make([]ArticleContentResponse, 0, len(models))
	for i := range models {
		responses = append(responses, *ToResponse(&models[i]))
	}
	return responses
}

// ArticleContentDeleteReq 文章删除请求
type ArticleContentDeleteReq struct {
	Ids []int64 `json:"ids" binding:"required"`
}

// GetId 获取ID
func (s *ArticleContentDeleteReq) GetId() interface{} {
	return s.Ids
}
