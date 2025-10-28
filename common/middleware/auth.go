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
	if config.ApplicationConfig.Mode == "dev" {
		// 开发环境：设置为100年
		timeout = time.Duration(876000) * time.Hour
	} else {
		// 生产环境：从配置读取
		if config.JwtConfig.Timeout != 0 {
			timeout = time.Duration(config.JwtConfig.Timeout) * time.Second
		} else {
			timeout = time.Hour
		}
	}
	
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
