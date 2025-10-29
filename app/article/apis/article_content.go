package apis

import (
	"fmt"
	"smart-api/app/article/models"
	"smart-api/app/article/service"
	"smart-api/app/article/service/dto"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// ArticleContentAPI 文章API处理器
type ArticleContentAPI struct {
	api.Api
}

// NewArticleContentAPI 创建API处理器实例
func NewArticleContentAPI() *ArticleContentAPI {
	return &ArticleContentAPI{}
}

// RegisterRoutes 注册路由
func (e *ArticleContentAPI) RegisterRoutes(router *gin.RouterGroup) {
	content := router.Group("/article")
	{
		content.GET("/", e.GetPage)                            // 文章列表
		content.GET("/detail/:id", e.GetDetail)                // 文章详情
		content.POST("/", e.AddArticle)                        // 添加文章
		content.POST("/update", e.UpdateArticle)               // 更新文章
		content.POST("/delete", e.DeleteArticle)               // 删除文章
		content.GET("/current-user", e.GetCurrentUserArticles) // 当前用户文章
		content.GET("/preview", e.PreviewMarkdown)             // Markdown预览
	}
}

// GetPage 获取文章分页列表
func (e *ArticleContentAPI) GetPage(c *gin.Context) {
	// 初始化API上下文和数据库连接
	e.MakeContext(c)
	e.MakeOrm()

	// 检查初始化是否成功
	if e.Errors != nil {
		fmt.Println("初始化错误:", e.Errors)
		e.Error(500, e.Errors, "初始化失败")
		return
	}

	// 检查Orm是否初始化成功
	if e.Orm == nil {
		fmt.Println("Orm is nil after initialization")
		e.Error(500, nil, "数据库连接未初始化")
		return
	}

	fmt.Println("Orm initialized successfully")

	// 创建服务实例
	s := &service.ArticleContentService{}
	// 手动设置Orm
	s.Orm = e.Orm

	req := dto.ArticleContentQuery{}
	// 绑定请求参数
	e.Bind(&req)
	if e.Errors != nil {
		fmt.Println("Bind error:", e.Errors)
		e.Error(500, e.Errors, "参数错误")
		return
	}

	result, err := s.GetPage(c, &req)
	if err != nil {
		fmt.Println("GetPage error:", err)
		e.Error(500, err, "查询失败")
		return
	}

	e.OK(result, "查询成功")
}

// GetDetail 获取文章详情
func (e *ArticleContentAPI) GetDetail(c *gin.Context) {
	// 初始化API上下文和数据库连接
	e.MakeContext(c)
	e.MakeOrm()

	// 检查初始化是否成功
	if e.Errors != nil {
		fmt.Println("初始化错误:", e.Errors)
		e.Error(500, e.Errors, "初始化失败")
		return
	}

	// 检查Orm是否初始化成功
	if e.Orm == nil {
		fmt.Println("Orm is nil after initialization")
		e.Error(500, nil, "数据库连接未初始化")
		return
	}

	// 创建服务实例
	s := &service.ArticleContentService{}
	// 手动设置Orm
	s.Orm = e.Orm

	id := c.Param("id")
	if id == "" {
		e.Error(400, nil, "ID不能为空")
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		e.Error(400, err, "ID格式错误")
		return
	}

	result, err := s.Get(idInt)
	if err != nil {
		e.Error(500, err, "查询失败")
		return
	}

	e.OK(result, "查询成功")
}

// AddArticle 新增文章
func (e *ArticleContentAPI) AddArticle(c *gin.Context) {
	// 初始化API上下文和数据库连接
	e.MakeContext(c)
	e.MakeOrm()

	// 检查初始化是否成功
	if e.Errors != nil {
		fmt.Println("初始化错误:", e.Errors)
		e.Error(500, e.Errors, "初始化失败")
		return
	}

	// 检查Orm是否初始化成功
	if e.Orm == nil {
		fmt.Println("Orm is nil after initialization")
		e.Error(500, nil, "数据库连接未初始化")
		return
	}

	// 创建服务实例
	s := &service.ArticleContentService{}
	// 手动设置Orm
	s.Orm = e.Orm

	req := dto.ArticleContentInsertReq{}
	// 绑定请求参数
	e.Bind(&req)
	if e.Errors != nil {
		fmt.Println("Bind error:", e.Errors)
		e.Error(500, e.Errors, "参数错误")
		return
	}

	if err := s.Insert(c, &req); err != nil {
		fmt.Println("Insert error:", err)
		e.Error(500, err, "创建失败")
		return
	}

	e.OK(nil, "创建成功")
}

// UpdateArticle 更新文章
func (e *ArticleContentAPI) UpdateArticle(c *gin.Context) {
	// 初始化API上下文和数据库连接
	e.MakeContext(c)
	e.MakeOrm()

	// 检查初始化是否成功
	if e.Errors != nil {
		fmt.Println("初始化错误:", e.Errors)
		e.Error(500, e.Errors, "初始化失败")
		return
	}

	// 检查Orm是否初始化成功
	if e.Orm == nil {
		fmt.Println("Orm is nil after initialization")
		e.Error(500, nil, "数据库连接未初始化")
		return
	}

	// 创建服务实例
	s := &service.ArticleContentService{}
	// 手动设置Orm
	s.Orm = e.Orm

	req := dto.ArticleContentUpdateReq{}
	// 绑定请求参数
	e.Bind(&req)
	if e.Errors != nil {
		fmt.Println("Bind error:", e.Errors)
		e.Error(500, e.Errors, "参数错误")
		return
	}

	if err := s.Update(c, &req); err != nil {
		fmt.Println("Update error:", err)
		e.Error(500, err, "更新失败")
		return
	}

	e.OK(nil, "更新成功")
}

// DeleteArticle 删除文章
func (e *ArticleContentAPI) DeleteArticle(c *gin.Context) {
	// 初始化API上下文和数据库连接
	e.MakeContext(c)
	e.MakeOrm()

	// 检查初始化是否成功
	if e.Errors != nil {
		fmt.Println("初始化错误:", e.Errors)
		e.Error(500, e.Errors, "初始化失败")
		return
	}

	// 检查Orm是否初始化成功
	if e.Orm == nil {
		fmt.Println("Orm is nil after initialization")
		e.Error(500, nil, "数据库连接未初始化")
		return
	}

	// 创建服务实例
	s := &service.ArticleContentService{}
	// 手动设置Orm
	s.Orm = e.Orm

	var ids []int64
	err := c.ShouldBindJSON(&ids)
	if err != nil {
		e.Error(400, err, "参数错误")
		return
	}

	if err := s.Delete(ids); err != nil {
		fmt.Println("Delete error:", err)
		e.Error(500, err, "删除失败")
		return
	}

	e.OK(nil, "删除成功")
}

// GetCurrentUserArticles 获取当前用户的文章列表（支持树形结构）
func (e *ArticleContentAPI) GetCurrentUserArticles(c *gin.Context) {
	// 初始化API上下文和数据库连接
	e.MakeContext(c)
	e.MakeOrm()

	// 检查初始化是否成功
	if e.Errors != nil {
		fmt.Println("初始化错误:", e.Errors)
		e.Error(500, e.Errors, "初始化失败")
		return
	}

	// 检查Orm是否初始化成功
	if e.Orm == nil {
		fmt.Println("Orm is nil after initialization")
		e.Error(500, nil, "数据库连接未初始化")
		return
	}

	// 创建服务实例
	s := &service.ArticleContentService{}
	// 手动设置Orm
	s.Orm = e.Orm

	// 获取当前用户ID
	userID := user.GetUserId(c)
	if userID == 0 {
		e.Error(401, nil, "未登录")
		return
	}

	// 设置查询参数，只查询当前用户的文章
	// 注意：不设置Type字段，因为数据库中不存在该字段
	req := dto.ArticleContentQuery{}
	req.CreateBy = int64(userID)
	// 默认获取树形结构
	req.Level = 0 // 0表示不限制层级

	// 直接使用数据库查询，避免通过service层的type字段过滤
	var articles []models.ArticleContent
	err := e.Orm.Table("article_contents").
		Where("create_by = ?", userID).
		Find(&articles).Error

	if err != nil {
		fmt.Println("GetCurrentUserArticles error:", err)
		e.Error(500, err, "查询失败")
		return
	}

	// 构建响应数据
	response := struct {
		List      interface{} `json:"list"`
		Count     int64       `json:"count"`
		PageIndex int         `json:"pageIndex"`
		PageSize  int         `json:"pageSize"`
	}{
		List:      dto.ToResponseList(articles),
		Count:     int64(len(articles)),
		PageIndex: 1,
		PageSize:  10,
	}

	e.OK(response, "查询成功")
}

// StringToInt64 字符串转int64
func StringToInt64(s string) (int64, error) {
	var i int64
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

// PreviewMarkdown Markdown预览
func (e *ArticleContentAPI) PreviewMarkdown(c *gin.Context) {
	content := c.Query("content")
	if content == "" {
		e.Error(400, nil, "内容不能为空")
		return
	}

	// 创建Markdown解析器和HTML渲染器
	parser := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs | parser.FencedCode)
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	// 解析Markdown为HTML
	markdownBytes := []byte(content)
	htmlBytes := markdown.ToHTML(markdownBytes, parser, renderer)

	result := map[string]interface{}{
		"html": string(htmlBytes),
	}
	e.OK(result, "预览成功")
}
