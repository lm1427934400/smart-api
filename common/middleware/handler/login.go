package handler

import (
	"errors"

	log "github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk/pkg"
	"gorm.io/gorm"
)

type Login struct {
	Username string `form:"UserName" json:"username" binding:"required"`
	Password string `form:"Password" json:"password" binding:"required"`
	Source   string `json:"source" comment:"用户来源" vd:"len($)>0" default:"1"`
	Code     string `form:"Code" json:"code" binding:"required"`
	UUID     string `form:"UUID" json:"uuid" binding:"required"`
}

func (u *Login) GetUser(tx *gorm.DB) (user SysUser, role SysRole, err error) {
	// 只查询需要的字段，而不是SELECT *
	err = tx.Table("sys_user").Select("user_id, username, password, role_id, status").Where("username = ? AND status = '2'", u.Username).First(&user).Error
	if err != nil {
		log.Errorf("get user error, %s", err.Error())
		return
	}
	// 预先检查密码长度，避免不必要的比较操作
	if len(u.Password) == 0 || len(user.Password) < 60 { // bcrypt哈希密码通常至少60个字符
		err = errors.New("invalid password length")
		log.Errorf("user login error, invalid password length")
		return
	}
	_, err = pkg.CompareHashAndPassword(user.Password, u.Password)
	if err != nil {
		log.Errorf("user login error, %s", err.Error())
		return
	}
	// 角色信息也只查询需要的字段，必须包含role_key！
	err = tx.Table("sys_role").Select("role_id, role_name, role_key, data_scope").Where("role_id = ?", user.RoleId).First(&role).Error
	if err != nil {
		log.Errorf("get role error, %s", err.Error())
		return
	}
	log.Infof("用户 %s 登录成功，角色信息: roleId=%d, roleName=%s, roleKey=%s", u.Username, role.RoleId, role.RoleName, role.RoleKey)
	return
}
