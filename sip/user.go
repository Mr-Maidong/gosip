package sipapi

import (
	"github.com/panjjo/gosip/db"
)

// User 用户信息
type User struct {
	db.DBModel
	// Username 用户名（唯一）
	Username string `json:"username" gorm:"column:username;type:varchar(64);not null;uniqueIndex;comment:'用户名'"`
	// Password 密码（加密存储）
	Password string `json:"-" gorm:"column:password;type:varchar(128);not null;comment:'密码'"`
	// Name 姓名
	Name string `json:"name" gorm:"column:name;type:varchar(128);comment:'姓名'"`
	// Email 邮箱
	Email string `json:"email" gorm:"column:email;type:varchar(128);comment:'邮箱'"`
	// Phone 手机号
	Phone string `json:"phone" gorm:"column:phone;type:varchar(32);comment:'手机号'"`
	// Role 角色：admin-管理员，operator-操作员，viewer-观察者
	Role string `json:"role" gorm:"column:role;type:varchar(32);default:'viewer';comment:'角色'"`
	// Status 状态：0-禁用，1-启用
	Status int `json:"status" gorm:"column:status;type:int;default:1;comment:'状态：0-禁用，1-启用'"`
	// LastLoginTime 最后登录时间
	LastLoginTime int64 `json:"lastLoginTime" gorm:"column:last_login_time;type:bigint;comment:'最后登录时间戳'"`
	// LastLoginIP 最后登录 IP
	LastLoginIP string `json:"lastLoginIP" gorm:"column:last_login_ip;type:varchar(64);comment:'最后登录 IP'"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

// UserLoginResponse 用户登录响应
type UserLoginResponse struct {
	UserInfo User   `json:"userInfo"`
	Token    string `json:"token"`
}

// UserChangePasswordRequest 修改密码请求
type UserChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" form:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" form:"newPassword" binding:"required"`
}
