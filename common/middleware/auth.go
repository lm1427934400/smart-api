package middleware

import (
	"time"

	"github.com/go-admin-team/go-admin-core/sdk/config"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"smart-api/common/middleware/handler"
)

// AuthInit jwt验证new
func AuthInit() (*jwt.GinJWTMiddleware, error) {
	// 确保在开发环境下token不会过期
	var timeout time.Duration
	// 无条件设置为长时间，确保开发环境token不会过期
	// 这样即使配置检测有问题，开发环境也能正常工作
	timeout = time.Duration(876000) * time.Hour // 100年
	
	// 同样设置MaxRefresh为长时间，避免刷新token也过期
	maxRefresh := timeout
	
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:           "test zone",
		Key:             []byte(config.JwtConfig.Secret),
		Timeout:         timeout,
		MaxRefresh:      maxRefresh,
		PayloadFunc:     handler.PayloadFunc,
		IdentityHandler: handler.IdentityHandler,
		Authenticator:   handler.Authenticator,
		Authorizator:    handler.Authorizator,
		Unauthorized:    handler.Unauthorized,
		TokenLookup:     "header: Authorization, query: token, cookie: jwt",
		TokenHeadName:   "Bearer",
		TimeFunc:        time.Now,
	})
}
