package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
)

// InitRouter 初始化文章路由
// 参考smart模块结构，提供统一的路由初始化接口
// 注意：由于文章路由主要由smart模块的sys_content.go管理，
// 此函数主要作为兼容性保留，确保模块结构一致性
func InitRouter(r *gin.Engine, authMiddleware *jwt.GinJWTMiddleware) *gin.Engine {
	if r == nil {
		return r
	}

	// 可根据业务需求来设置接口版本
	v1 := r.Group("/api/v1")

	// 注册无需认证的路由
	RegisterArticlePublicRoutes(v1)

	// 注册需要认证的路由（如果提供了认证中间件）
	if authMiddleware != nil {
		RegisterArticleAuthRoutes(v1, authMiddleware)
	}

	return r
}
