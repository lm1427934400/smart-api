// @Author auto-generated
// 2024/7/12 20:42
package article

import (
	"smart-api/app/article/router"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
)

// InitRouter 初始化文章模块路由
// 参考smart模块结构，提供统一的路由初始化入口
func InitRouter(r *gin.Engine, authMiddleware *jwt.GinJWTMiddleware) *gin.Engine {
	return router.InitRouter(r, authMiddleware)
}
