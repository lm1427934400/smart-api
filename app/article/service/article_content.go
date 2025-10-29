package service

import (
	"fmt"
	"smart-api/app/article/models"
	"smart-api/app/article/service/dto"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"
)

// ArticleContentService 文章服务
type ArticleContentService struct {
	service.Service
}

// NewArticleContentService 创建服务实例
func NewArticleContentService(s *service.Service) *ArticleContentService {
	return &ArticleContentService{Service: *s}
}

// GetPage 获取分页列表
func (s *ArticleContentService) GetPage(c *gin.Context, req *dto.ArticleContentQuery) (interface{}, error) {
	// 安全检查
	if s == nil {
		return nil, fmt.Errorf("服务实例为空")
	}
	if req == nil {
		return nil, fmt.Errorf("请求参数为空")
	}
	if s.Orm == nil {
		return nil, fmt.Errorf("数据库连接为空")
	}

	// 初始化响应结构
	pageInfo := struct {
		List      interface{} `json:"list"`
		Count     int64       `json:"count"`
		PageIndex int         `json:"pageIndex"`
		PageSize  int         `json:"pageSize"`
	}{
		PageIndex: req.PageIndex,
		PageSize:  req.PageSize,
	}

	// 设置默认分页参数
	if pageInfo.PageIndex == 0 {
		pageInfo.PageIndex = 1
	}
	if pageInfo.PageSize == 0 {
		pageInfo.PageSize = 10
	}

	// 直接查询原始数据，避免转换问题
	var list []models.ArticleContent
	// 明确指定表名，使用Unscoped()禁用软删除过滤
	db := s.Orm.Model(&models.ArticleContent{}).Table("article_contents").Unscoped()

	// 添加用户过滤条件
	if req.CreateBy > 0 {
		db = db.Where("create_by = ?", req.CreateBy)
	}

	// 计算总数
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("查询总数失败: %v", err)
	}
	pageInfo.Count = count

	// 查询数据列表
	offset := (pageInfo.PageIndex - 1) * pageInfo.PageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageInfo.PageSize).Find(&list).Error; err != nil {
		return nil, fmt.Errorf("查询数据失败: %v", err)
	}

	// 直接返回原始数据，不进行转换
	pageInfo.List = list

	return pageInfo, nil
}

// Get 获取单个文章
func (s *ArticleContentService) Get(id int64) (*dto.ArticleContentResponse, error) {
	var data models.ArticleContent
	// 明确指定表名，使用Unscoped()禁用软删除过滤
	err := s.Orm.Table("article_contents").Unscoped().First(&data, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("文章不存在")
		}
		return nil, err
	}

	// 增加浏览次数
	go func() {
		s.Orm.Table("article_contents").Model(&data).Update("view_count", gorm.Expr("view_count + ?", 1))
	}()

	return dto.ToResponse(&data), nil
}

// Insert 新增文章
func (s *ArticleContentService) Insert(c *gin.Context, req *dto.ArticleContentInsertReq) error {
	// 获取当前用户ID并转换为int类型
	userId := int(user.GetUserId(c))

	// 创建模型
	var data models.ArticleContent
	req.Generate(&data)

	// 设置创建者和更新者
	data.SetCreateBy(userId)
	data.SetUpdateBy(userId)

	// 保存到数据库，明确指定表名，使用Unscoped()禁用软删除字段
	return s.Orm.Table("article_contents").Unscoped().Create(&data).Error
}

// Update 更新文章
func (s *ArticleContentService) Update(c *gin.Context, req *dto.ArticleContentUpdateReq) error {
	// 获取当前用户ID并转换为int类型
	userId := int(user.GetUserId(c))

	// 查询现有记录
	var data models.ArticleContent
	err := s.Orm.Table("article_contents").Unscoped().First(&data, req.Id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("文章不存在")
		}
		return err
	}

	// 更新数据
	req.Generate(&data)
	data.SetUpdateBy(userId)

	// 保存到数据库，使用Unscoped()禁用软删除字段
	return s.Orm.Table("article_contents").Unscoped().Save(&data).Error
}

// Delete 删除文章
func (s *ArticleContentService) Delete(ids []int64) error {
	return s.Orm.Table("article_contents").Unscoped().Delete(&models.ArticleContent{}, ids).Error
}

// CreateArticle 创建文章（用于Markdown处理器）
func (s *ArticleContentService) CreateArticle(title, content, htmlContent string, userId int64) (int64, error) {
	// 创建文章记录
	article := &models.ArticleContent{
		Title:       title,
		Content:     content,
		HtmlContent: htmlContent,
		Status:      1, // 正常状态
	}

	// 设置创建者和更新者
	article.SetCreateBy(int(userId))
	article.SetUpdateBy(int(userId))

	// 保存到数据库，明确指定表名，使用Unscoped()禁用软删除字段
	err := s.Orm.Table("article_contents").Unscoped().Create(article).Error
	if err != nil {
		return 0, err
	}

	return article.Id, nil
}

// UpdateArticle 更新文章
func (s *ArticleContentService) UpdateArticle(id int64, title, content, htmlContent string, userId int64) error {
	var article models.ArticleContent
	err := s.Orm.Table("article_contents").Unscoped().First(&article, id).Error
	if err != nil {
		return err
	}

	// 更新字段
	article.Title = title
	article.Content = content
	article.HtmlContent = htmlContent
	article.SetUpdateBy(int(userId))

	return s.Orm.Table("article_contents").Unscoped().Save(&article).Error
}
