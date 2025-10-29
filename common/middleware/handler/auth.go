package handler

import (
	"github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/captcha"
	"net/http"
	"smart-api/app/system/service"
	"smart-api/app/system/service/dto"
	"smart-api/common"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/config"
	"github.com/go-admin-team/go-admin-core/sdk/pkg"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/response"
	"github.com/mssola/user_agent"
	"smart-api/common/global"
)

func PayloadFunc(data interface{}) jwt.MapClaims {
	if v, ok := data.(map[string]interface{}); ok {
		u, _ := v["user"].(SysUser)
		r, _ := v["role"].(SysRole)
		return jwt.MapClaims{
			jwt.IdentityKey:  u.UserId,
			jwt.RoleIdKey:    r.RoleId,
			jwt.RoleKey:      r.RoleKey,
			"rolekey":        r.RoleKey, // 添加小写的rolekey，确保权限检查中间件能识别
			jwt.NiceKey:      u.Username,
			jwt.DataScopeKey: r.DataScope,
			jwt.RoleNameKey:  r.RoleName,
		}
	}
	return jwt.MapClaims{}
}

func IdentityHandler(c *gin.Context) interface{} {
	claims := jwt.ExtractClaims(c)
	return map[string]interface{}{
		"IdentityKey": claims["identity"],
		"UserName":    claims["nice"],
		"RoleKey":     claims["rolekey"],
		"rolekey":     claims["rolekey"], // 添加小写的rolekey，确保权限中间件能识别
		"UserId":      claims["identity"],
		"RoleIds":     claims["roleid"],
		"DataScope":   claims["datascope"],
	}
}

// Authenticator 获取token
func Authenticator(c *gin.Context) (interface{}, error) {
	log := api.GetRequestLogger(c)
	db, err := pkg.GetOrm(c)
	if err != nil {
		log.Errorf("get db error, %s", err.Error())
		response.Error(c, 500, err, "数据库连接获取失败")
		return nil, jwt.ErrFailedAuthentication
	}
	var loginVals Login
	var status = "2"
	var msg = "登录成功"
	var username = ""
	defer func() {
		LoginLogToDB(c, status, msg, username)
	}()
	if err = c.ShouldBind(&loginVals); err != nil {
		username = loginVals.Username
		msg = "数据解析失败"
		status = "1"
		return nil, jwt.ErrMissingLoginValues
	}
	if config.ApplicationConfig.Mode != "dev" {
		if !captcha.Verify(loginVals.UUID, loginVals.Code, true) {
			username = loginVals.Username
			msg = "验证码错误"
			status = "1"
			return nil, jwt.ErrInvalidVerificationode
		}
	}
	
	// 先尝试系统登录，这是最常用的方式
	username = loginVals.Username
	if sysUser, role, e := loginVals.GetUser(db); e == nil {
		// 如果系统登录成功，直接返回
		return map[string]interface{}{"user": sysUser, "role": role}, nil
	}
	
	// 系统登录失败，如果是LDAP登录请求，则尝试LDAP登录
	if loginVals.Source == "LDAP" {
		log.Info("尝试LDAP登录:", loginVals.Username)
		if err = handleLDAPLogin(c, loginVals); err == nil {
			// LDAP登录成功后，再次尝试获取用户信息
			if sysUser, role, e := loginVals.GetUser(db); e == nil {
				return map[string]interface{}{"user": sysUser, "role": role}, nil
			}
		}
	}
	
	// 所有登录方式都失败
	msg = "登录失败"
	status = "1"
	log.Warnf("%s login failed!", loginVals.Username)
	return nil, jwt.ErrFailedAuthentication
}

func handleLDAPLogin(c *gin.Context, loginVals Login) error {

	// 初始化 gorm.DB
	orm, err := pkg.GetOrm(c)
	if err != nil {
		return err
	}

	// 获取 logger 实例
	log := logger.NewLogger() // 根据实际方法调整

	logHelper := logger.NewHelper(log)

	// 初始化 SysLdap 实例，并设置 Orm 和 Log
	ldapService := service.SysLdap{
		Orm: orm,
		Log: logHelper,
	}

	ldapUser := dto.LdapUserInsertReq{
		Username: loginVals.Username,
		Password: loginVals.Password,
		Source:   loginVals.Source,
	}

	// 调用 LdapAuth 方法进行 LDAP 认证和用户注册
	return ldapService.LdapAuth(&ldapUser)
}

// LoginLogToDB Write log to database
func LoginLogToDB(c *gin.Context, status string, msg string, username string) {
	if !config.LoggerConfig.EnabledDB {
		return
	}
	log := api.GetRequestLogger(c)
	l := make(map[string]interface{})

	ua := user_agent.New(c.Request.UserAgent())
	l["ipaddr"] = common.GetClientIP(c)
	l["loginLocation"] = "" // pkg.GetLocation(common.GetClientIP(c),gaConfig.ExtConfig.AMap.Key)
	l["loginTime"] = pkg.GetCurrentTime()
	l["status"] = status
	l["remark"] = c.Request.UserAgent()
	browserName, browserVersion := ua.Browser()
	l["browser"] = browserName + " " + browserVersion
	l["os"] = ua.OS()
	l["platform"] = ua.Platform()
	l["username"] = username
	l["msg"] = msg

	q := sdk.Runtime.GetMemoryQueue(c.Request.Host)
	message, err := sdk.Runtime.GetStreamMessage("", global.LoginLog, l)
	if err != nil {
		log.Errorf("GetStreamMessage error, %s", err.Error())
		//日志报错错误，不中断请求
	} else {
		err = q.Append(message)
		if err != nil {
			log.Errorf("Append message error, %s", err.Error())
		}
	}
}

// LogOut
// @Summary 退出登录
// @Description 获取token
// LoginHandler can be used by clients to get a jwt token.
// Reply will be of the form {"token": "TOKEN"}.
// @Accept  application/json
// @Product application/json
// @Success 200 {string} string "{"code": 200, "msg": "成功退出系统" }"
// @Router /logout [post]
// @Security Bearer
func LogOut(c *gin.Context) {
	LoginLogToDB(c, "2", "退出成功", user.GetUserName(c))
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "退出成功",
	})

}

func Authorizator(data interface{}, c *gin.Context) bool {
	// 正确处理IdentityHandler返回的数据结构
	if v, ok := data.(map[string]interface{}); ok {
		// 处理角色信息
		if roleKey, ok := v["RoleKey"].(string); ok {
			c.Set("role", roleKey)
			c.Set("roleName", roleKey)
			c.Set("rolekey", roleKey)
		}
		// 尝试从rolekey字段获取角色
		if roleKey, ok := v["rolekey"].(string); ok {
			c.Set("role", roleKey)
			c.Set("roleName", roleKey)
			c.Set("rolekey", roleKey)
		}
		// 处理其他用户信息
		if roleIds, ok := v["RoleIds"]; ok {
			c.Set("roleIds", roleIds)
		}
		if userId, ok := v["UserId"]; ok {
			c.Set("userId", userId)
			c.Set("identity", userId)
		}
		if userName, ok := v["UserName"].(string); ok {
			c.Set("userName", userName)
			c.Set("nice", userName)
		}
		if dataScope, ok := v["DataScope"]; ok {
			c.Set("dataScope", dataScope)
		}
		// 只要能获取到用户ID，就认为认证成功
		_, hasUserId := v["UserId"]
		return hasUserId
	}
	return false
}

func Unauthorized(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  message,
	})
}
