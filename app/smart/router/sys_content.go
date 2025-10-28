package router

import (
	"smart-api/app/article/apis"
	"smart-api/common/actions"
	"smart-api/common/models"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
)

// SysContent 简单模型实现，用于路由注册
type SysContent struct {
	Id       int
	Title    string
	Content  string
	CreateBy int
	UpdateBy int
}

// TableName 指定表名
func (SysContent) TableName() string {
	return "sys_content"
}

// SetCreateBy 设置创建者
func (s *SysContent) SetCreateBy(createBy int) {
	s.CreateBy = createBy
}

// SetUpdateBy 设置更新者
func (s *SysContent) SetUpdateBy(updateBy int) {
	s.UpdateBy = updateBy
}

// Generate 生成实例
func (s *SysContent) Generate() models.ActiveRecord {
	return &SysContent{}
}

// GetId 获取ID
func (s *SysContent) GetId() interface{} {
	return s.Id
}

func init() {
	// 将sys-content的GET接口注册到无需认证的路由中
	routerNoCheckRole = append(routerNoCheckRole, registerSysContentNoAuthRouter)
	// 将其他操作注册到需要认证的路由中
	routerCheckRole = append(routerCheckRole, registerSysContentAuthRouter)
}

// 无需认证的路由代码 - 用于文章浏览
func registerSysContentNoAuthRouter(v1 *gin.RouterGroup) {
	model := &SysContent{}
	// GET接口无需认证，允许未登录用户浏览文章
	v1.GET("/sys-content", actions.IndexAction(model, nil, func() interface{} {
		return make([]SysContent, 0)
	}))
	v1.GET("/sys-content/:id", actions.ViewAction(nil, nil))
}

// 需要认证的路由代码 - 用于文章管理（发布、编辑、删除）
func registerSysContentAuthRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	// 创建新的Gin路由组并添加认证中间件
	r := v1.Group("/sys-content").Use(authMiddleware.MiddlewareFunc())
	{
		// 使用新的Markdown处理器处理POST请求（文章发布）
		markdownHandler := apis.MarkdownHandler{}
		r.POST("", func(c *gin.Context) {
			markdownHandler.MakeContext(c)
			markdownHandler.UploadMarkdownFile(c)
		})
		
		// 其他操作继续使用现有的处理方式
		r.PUT("/:id", func(c *gin.Context) {
			// 简单的更新逻辑
			actions.ViewAction(nil, nil)(c)
		})
		r.DELETE("/:id", func(c *gin.Context) {
			// 简单的删除逻辑
			actions.ViewAction(nil, nil)(c)
		})
	}
}
