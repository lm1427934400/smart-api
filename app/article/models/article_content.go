package models

import (
	"smart-api/common/models"
	"time"

	"gorm.io/gorm"
)

// ArticleContent 文章内容表
// 用于存储系统中的文章内容数据
// 包含文章标题、内容、创建信息等字段
// 实现了ActiveRecord接口，支持ORM操作

type ArticleContent struct {
	Id          int64          `json:"id" gorm:"primaryKey;autoIncrement;comment:文章ID"`
	Title       string         `json:"title" gorm:"size:255;not null;comment:文章标题"`
	Content     string         `json:"content" gorm:"type:longtext;not null;comment:Markdown内容"`
	HtmlContent string         `json:"html_content" gorm:"column:html_content;type:longtext;comment:HTML内容"`
	CreateBy    int64          `json:"create_by" gorm:"column:create_by;index;comment:创建人ID"`
	UpdateBy    int64          `json:"update_by" gorm:"column:update_by;comment:更新人ID"`
	CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at;index;autoCreateTime;comment:创建时间"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime;comment:更新时间"`
	Status      int            `json:"status" gorm:"size:4;default:1;comment:状态：1-正常，0-禁用"`
	Cover       string         `json:"cover" gorm:"-;comment:封面图"`       // 不映射到数据库
	Summary     string         `json:"summary" gorm:"-;comment:摘要"`      // 不映射到数据库
	ViewCount   int            `json:"view_count" gorm:"-;comment:浏览次数"` // 不映射到数据库
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index;comment:删除时间"`
}

// TableName 指定表名
func (*ArticleContent) TableName() string {
	return "article_contents"
}

// Generate 生成实例
func (e *ArticleContent) Generate() models.ActiveRecord {
	o := *e
	return &o
}

// GetId 获取ID
func (e *ArticleContent) GetId() interface{} {
	return e.Id
}

// SetCreateBy 设置创建者
func (e *ArticleContent) SetCreateBy(createBy int) {
	e.CreateBy = int64(createBy)
}

// SetUpdateBy 设置更新者
func (e *ArticleContent) SetUpdateBy(updateBy int) {
	e.UpdateBy = int64(updateBy)
}

// GetCreateBy 获取创建者
func (e *ArticleContent) GetCreateBy() int64 {
	return e.CreateBy
}

// GetUpdateBy 获取更新者
func (e *ArticleContent) GetUpdateBy() int64 {
	return e.UpdateBy
}
