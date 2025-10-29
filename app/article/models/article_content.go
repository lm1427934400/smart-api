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
	Id          int64          `json:"id" gorm:"primaryKey;autoIncrement;comment:ID"`
	Title       string         `json:"title" gorm:"size:255;not null;comment:标题"`
	Content     string         `json:"content" gorm:"type:longtext;comment:Markdown内容"`
	HtmlContent string         `json:"html_content" gorm:"column:html_content;type:longtext;comment:HTML内容"`
	CreateBy    int64          `json:"create_by" gorm:"column:create_by;index;comment:创建人ID"`
	UpdateBy    int64          `json:"update_by" gorm:"column:update_by;comment:更新人ID"`
	CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at;index;autoCreateTime;comment:创建时间"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime;comment:更新时间"`
	Status      int            `json:"status" gorm:"size:4;default:1;comment:状态：1-正常，0-禁用"`
	Cover       string         `json:"cover" gorm:"column:cover;size:500;comment:封面图"`
	Summary     string         `json:"summary" gorm:"column:summary;size:500;comment:摘要"`
	ViewCount   int            `json:"view_count" gorm:"column:view_count;default:0;comment:浏览次数"`
	ParentId    *int64         `json:"parent_id" gorm:"column:parent_id;index;comment:父节点ID"`
	Type        string         `json:"type" gorm:"column:type;size:20;default:'article';index;comment:类型：article-文章，directory-目录"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index;comment:删除时间"`
	// 虚拟字段，用于树形结构
	Children []ArticleContent `json:"children,omitempty" gorm:"-"`
	Level    int              `json:"level,omitempty" gorm:"-"`
	Expanded bool             `json:"expanded,omitempty" gorm:"-"`
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
