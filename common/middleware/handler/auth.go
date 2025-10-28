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
		"UserId":      claims["identity"],
		"RoleIds":     claims["roleid"],
		"DataScope":   claims["datascope"],
	}
}

// Authenticator 获取token
// @Summary 登陆
// @Description 获取token
// @Description LoginHandler can be used by clients to get a jwt token.
// @Description Payload needs to be json in the form of {"username": "USERNAME", "password": "PASSWORD"}.
// @Description Reply will be of the form {"token": "TOKEN"}.
// @Description dev mode：It should be noted that all fields cannot be empty, and a value of 0 can be passed in addition to the account password
// @Description 注意：开发模式：需要注意全部字段不能为空，账号密码外可以传入0值
// @Tags 登陆
// @Accept  application/json
// @Product application/json
// @Param account body Login  true "account"
// @Success 200 {string} string "{"code": 200, "expire": "2019-08-07T12:45:48+08:00", "token": ".eyJleHAiOjE1NjUxNTMxNDgsImlkIjoiYWRtaW4iLCJvcmlnX2lhdCI6MTU2NTE0OTU0OH0.-zvzHvbg0A" }"
// @Router /api/v1/login [post]
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
	if loginVals.Source == "LDAP" {
		// LDAP 登录逻辑
		if err = handleLDAPLogin(c, loginVals); err != nil {
			msg = "LDAP 登录失败"
			status = "1"
			log.Warnf("%s login failed!", loginVals.Username)
			return nil, err
		}
	} else if loginVals.Source == "SYSTEM" {
		// 系统登录逻辑
		if sysUser, role, e := loginVals.GetUser(db); e == nil {
			username = loginVals.Username
			return map[string]interface{}{"user": sysUser, "role": role}, nil
		} else {
			msg = "登录失败"
			status = "1"
			log.Warnf("%s login failed!", loginVals.Username)
			return nil, jwt.ErrFailedAuthentication
		}
	} else {
		msg = "登录失败"
		status = "1"
		log.Warnf("%s login failed!", loginVals.Username)
		return nil, jwt.ErrFailedAuthentication
	}
	// 如果 LDAP 登录成功但未找到用户信息
	if sysUser, role, e := loginVals.GetUser(db); e == nil {
		username = loginVals.Username
		return map[string]interface{}{"user": sysUser, "role": role}, nil
	}

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
		// 从v中直接获取字段，而不是尝试类型断言为models.SysUser和models.SysRole
		if roleKey, ok := v["RoleKey"].(string); ok {
			c.Set("role", roleKey)
			// 同时设置roleName，确保GetRoleName(c)能获取到角色名称
			c.Set("roleName", roleKey)
			// 也设置rolekey，这可能是user.GetRoleName(c)实际使用的键
			c.Set("rolekey", roleKey)
		}
		if roleIds, ok := v["RoleIds"]; ok {
			c.Set("roleIds", roleIds)
		}
		if userId, ok := v["UserId"]; ok {
			c.Set("userId", userId)
			// 同时设置identity，确保GetUserId(c)能获取到用户ID
			c.Set("identity", userId)
		}
		if userName, ok := v["UserName"].(string); ok {
			c.Set("userName", userName)
			// 同时设置nice，确保GetUserName(c)能获取到用户名
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
