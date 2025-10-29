package middleware

import (
	"fmt"
	"net/http"

	"github.com/casbin/casbin/v2/util"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/response"
)

// AuthCheckRole 权限检查中间件
func AuthCheckRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := api.GetRequestLogger(c)
		data, _ := c.Get(jwtauth.JwtPayloadKey)
		v := data.(jwtauth.MapClaims)
		e := sdk.Runtime.GetCasbinKey(c.Request.Host)
		// 不用鉴权的接口信息 settings.go
		var res, casbinExclude bool
		var err error
		//检查权限 - 添加详细调试日志
		fmt.Println("权限中间件-开始检查权限...")
		// 从context中获取所有可能的role相关键值
		fmt.Println("权限中间件-context中的角色信息:")
		// 尝试从context中获取角色信息
		roleKeyCtx, roleKeyCtxExists := c.Get("rolekey")
		RoleKeyCtx, RoleKeyCtxExists := c.Get("RoleKey")
		roleNameCtx, roleNameCtxExists := c.Get("roleName")
		userIDCtx, userIDCtxExists := c.Get("userId")

		fmt.Println("  - context rolekey存在:", roleKeyCtxExists, "值:", roleKeyCtx)
		fmt.Println("  - context RoleKey存在:", RoleKeyCtxExists, "值:", RoleKeyCtx)
		fmt.Println("  - context roleName存在:", roleNameCtxExists, "值:", roleNameCtx)
		fmt.Println("  - context userId存在:", userIDCtxExists, "值:", userIDCtx)

		// 从jwt claims中获取rolekey
		if roleKey, exists := v["rolekey"]; exists {
			fmt.Println("权限中间件-jwt claims中的rolekey值:", roleKey)
			if roleKey == "admin" {
				fmt.Println("权限中间件-从jwt claims检测到admin角色，直接通过权限验证")
				res = true
				c.Next()
				return
			}
		} else {
			fmt.Println("权限中间件-jwt claims中未找到rolekey字段")
		}

		// 检查jwt claims中其他可能的角色字段
		if roleKey, exists := v["RoleKey"]; exists {
			fmt.Println("权限中间件-jwt claims中的RoleKey值:", roleKey)
			if roleKey == "admin" {
				fmt.Println("权限中间件-从jwt claims检测到admin角色(RoleKey)，直接通过权限验证")
				res = true
				c.Next()
				return
			}
		}

		if roleName, exists := v["roleName"]; exists {
			fmt.Println("权限中间件-jwt claims中的roleName值:", roleName)
			if roleName == "admin" {
				fmt.Println("权限中间件-从jwt claims检测到admin角色(roleName)，直接通过权限验证")
				res = true
				c.Next()
				return
			}
		}

		// 输出所有jwt claims键值，用于调试
		fmt.Println("权限中间件-jwt claims键值列表:")
		for key, value := range v {
			fmt.Println("jwt claim键值:", key, "=", value)
		}
		// 不做验证的接口信息
		for _, i := range CasbinExclude {
			if util.KeyMatch2(c.Request.URL.Path, i.Url) && c.Request.Method == i.Method {
				casbinExclude = true
				break
			}
		}
		if casbinExclude {
			log.Infof("Casbin exclusion, no validation method:%s path:%s", c.Request.Method, c.Request.URL.Path)
			c.Next()
			return
		}
		res, err = e.Enforce(v["rolekey"], c.Request.URL.Path, c.Request.Method)
		if err != nil {
			log.Errorf("AuthCheckRole error:%s method:%s path:%s", err, c.Request.Method, c.Request.URL.Path)
			response.Error(c, 500, err, "")
			return
		}

		if res {
			log.Infof("isTrue: %v role: %s method: %s path: %s", res, v["rolekey"], c.Request.Method, c.Request.URL.Path)
			c.Next()
		} else {
			log.Warnf("isTrue: %v role: %s method: %s path: %s message: %s", res, v["rolekey"], c.Request.Method, c.Request.URL.Path, "当前request无权限，请管理员确认！")
			c.JSON(http.StatusOK, gin.H{
				"code": 403,
				"msg":  "对不起，您没有该接口访问权限，请联系管理员",
			})
			c.Abort()
			return
		}

	}
}
