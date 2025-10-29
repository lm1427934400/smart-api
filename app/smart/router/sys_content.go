package router

import (
	"smart-api/app/article/apis"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
)

// 初始化函数，注册所有路由
func init() {
	// 将认证路由添加到需要检查的角色列表
	routerCheckRole = append(routerCheckRole, registerSysContentAuthRouter)
	// 将公开路由添加到无需认证的路由列表
	routerNoCheckRole = append(routerNoCheckRole, registerSysContentPublicRouter)
}

// 需要认证的路由代码 - 用于文章管理（发布、编辑、删除、查看当前用户文章）
func registerSysContentAuthRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	// 创建新的Gin路由组并添加认证中间件
	r := v1.Group("/article").Use(authMiddleware.MiddlewareFunc())
	{
		// 使用新的Article处理器处理POST请求（文章发布）
		articleHandler := apis.ArticleHandler{}
		r.POST("", func(c *gin.Context) {
			articleHandler.MakeContext(c)
			articleHandler.UploadMarkdownFile(c)
		})
		
		// 直接使用文章API实例处理其他请求，避免空指针错误
		contentAPI := apis.NewArticleContentAPI()
		r.POST("/update", contentAPI.UpdateArticle)
		r.POST("/delete", contentAPI.DeleteArticle)
		// 当前用户文章路由需要认证
		r.GET("/current-user", contentAPI.GetCurrentUserArticles) // 当前用户文章
	}
}

// 注册无需认证的文章路由
func registerSysContentPublicRouter(v1 *gin.RouterGroup) {
	// 直接使用文章API实例处理请求
	contentAPI := apis.NewArticleContentAPI()
	// 确保文章列表路由无需认证即可访问
	v1.GET("/article", contentAPI.GetPage)                             // 文章列表
	// 确保文章详情路由无需认证即可访问
	v1.GET("/article/detail/:id", contentAPI.GetDetail)                // 文章详情
	// Markdown预览路由
	v1.GET("/article/preview", contentAPI.PreviewMarkdown)             // Markdown预览
}
