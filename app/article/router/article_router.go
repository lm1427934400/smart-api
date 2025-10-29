// @Author sunwenbo
// 2024/7/12 20:28
package router

import (
	"smart-api/app/article/apis"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
)

// RegisterArticleAuthRoutes 注册需要认证的文章路由
// 注意：现在主要由smart模块的sys_content.go管理路由注册
// 此函数保留以便于在需要时直接使用
func RegisterArticleAuthRoutes(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	// 创建新的Gin路由组并添加认证中间件
	r := v1.Group("/article").Use(authMiddleware.MiddlewareFunc())
	{
		// 使用新的Article处理器处理POST请求（文章发布）
		articleHandler := apis.ArticleHandler{}
		r.POST("", func(c *gin.Context) {
			articleHandler.MakeContext(c)
			articleHandler.UploadMarkdownFile(c)
		})

		// 直接使用文章API实例处理其他请求
		contentAPI := apis.NewArticleContentAPI()
		r.POST("/update", contentAPI.UpdateArticle)
		r.POST("/delete", contentAPI.DeleteArticle)
	}
}

// RegisterArticlePublicRoutes 注册无需认证的文章路由
// 注意：现在主要由smart模块的sys_content.go管理路由注册
// 此函数保留以便于在需要时直接使用
func RegisterArticlePublicRoutes(v1 *gin.RouterGroup) {
	// 直接使用文章API实例处理请求
	contentAPI := apis.NewArticleContentAPI()
	v1.GET("/article", contentAPI.GetPage)                             // 文章列表
	v1.GET("/article/detail/:id", contentAPI.GetDetail)                // 文章详情
	v1.GET("/article/preview", contentAPI.PreviewMarkdown)             // Markdown预览
}
