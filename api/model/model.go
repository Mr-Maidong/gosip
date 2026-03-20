package model

import (
	"github.com/panjjo/gosip/db"
)

// ==================== 角色模型 ====================

// Role 角色信息
type Role struct {
	db.DBModel
	// Name 角色名称
	Name string `json:"name" gorm:"column:name;type:varchar(128);not null;comment:'角色名称'"`
	// Code 角色编码（唯一标识）
	Code string `json:"code" gorm:"column:code;type:varchar(64);not null;uniqueIndex;comment:'角色编码'"`
	// Description 描述
	Description string `json:"description" gorm:"column:description;type:varchar(256);comment:'描述'"`
	// Status 状态：0-禁用，1-启用
	Status int `json:"status" gorm:"column:status;type:int;default:1;comment:'状态：0-禁用，1-启用'"`
	// Sort 排序
	Sort int `json:"sort" gorm:"column:sort;type:int;default:0;comment:'排序'"`
}

// TableName 表名
func (Role) TableName() string {
	return "roles"
}

// ==================== 权限模型 ====================

// Permission 权限信息
type Permission struct {
	db.DBModel
	// Name 权限名称
	Name string `json:"name" gorm:"column:name;type:varchar(128);not null;comment:'权限名称'"`
	// Code 权限编码（唯一标识）
	Code string `json:"code" gorm:"column:code;type:varchar(128);not null;uniqueIndex;comment:'权限编码'"`
	// Type 权限类型：menu-菜单，button-按钮，api-接口
	Type string `json:"type" gorm:"column:type;type:varchar(32);default:'api';comment:'权限类型'"`
	// Path 路径（API 接口路径或前端路由）
	Path string `json:"path" gorm:"column:path;type:varchar(256);comment:'路径'"`
	// Method HTTP 方法（GET/POST/PUT/DELETE）
	Method string `json:"method" gorm:"column:method;type:varchar(16);comment:'HTTP 方法'"`
	// ParentID 父权限 ID
	ParentID uint `json:"parentId" gorm:"column:parent_id;type:int;default:0;comment:'父权限 ID'"`
	// Description 描述
	Description string `json:"description" gorm:"column:description;type:varchar(256);comment:'描述'"`
}

// TableName 表名
func (Permission) TableName() string {
	return "permissions"
}

// ==================== 关联模型 ====================

// UserRole 用户角色关联
type UserRole struct {
	db.DBModel
	// UserID 用户 ID
	UserID uint `json:"userId" gorm:"column:user_id;type:int;not null;uniqueIndex:idx_user_role;comment:'用户 ID'"`
	// RoleID 角色 ID
	RoleID uint `json:"roleId" gorm:"column:role_id;type:int;not null;uniqueIndex:idx_user_role;comment:'角色 ID'"`
}

// TableName 表名
func (UserRole) TableName() string {
	return "user_role"
}

// RolePermission 角色权限关联
type RolePermission struct {
	db.DBModel
	// RoleID 角色 ID
	RoleID uint `json:"roleId" gorm:"column:role_id;type:int;not null;uniqueIndex:idx_role_perm;comment:'角色 ID'"`
	// PermissionID 权限 ID
	PermissionID uint `json:"permissionId" gorm:"column:permission_id;type:int;not null;uniqueIndex:idx_role_perm;comment:'权限 ID'"`
}

// TableName 表名
func (RolePermission) TableName() string {
	return "role_permission"
}

// ==================== 请求响应结构 ====================

// RoleWithPermissions 角色带权限列表
type RoleWithPermissions struct {
	Role
	Permissions []Permission `json:"permissions"`
}

// UserWithRoles 用户带角色列表
type UserWithRoles struct {
	User
	Roles []Role `json:"roles"`
}

// PermissionTree 权限树节点
type PermissionTree struct {
	Permission
	Children []PermissionTree `json:"children"`
}

// AssignPermissionsRequest 分配权限请求
type AssignPermissionsRequest struct {
	PermissionIDs []uint `json:"permissionIds" binding:"required"`
}

// AssignRolesRequest 分配角色请求
type AssignRolesRequest struct {
	RoleIDs []uint `json:"roleIds" binding:"required"`
}

// CheckPermissionResponse 权限检查响应
type CheckPermissionResponse struct {
	HasPermission bool `json:"hasPermission"`
}

// UserPermissionsResponse 用户权限响应
type UserPermissionsResponse struct {
	Permissions []Permission `json:"permissions"`
	Codes       []string     `json:"codes"`
}

// User 用户信息（从 sip 包移过来）
type User struct {
	db.DBModel
	// Username 用户名
	Username string `json:"username" gorm:"column:username;type:varchar(64);not null;uniqueIndex;comment:'用户名'"`
	// Password 密码（加密存储）
	Password string `json:"-" gorm:"column:password;type:varchar(128);not null;comment:'密码'"`
	// Name 姓名
	Name string `json:"name" gorm:"column:name;type:varchar(128);comment:'姓名'"`
	// Email 邮箱
	Email string `json:"email" gorm:"column:email;type:varchar(128);comment:'邮箱'"`
	// Phone 手机号
	Phone string `json:"phone" gorm:"column:phone;type:varchar(32);comment:'手机号'"`
	// Role 角色（已废弃，使用 UserRole 关联表）
	Role string `json:"role" gorm:"column:role;type:varchar(64);comment:'角色'"`
	// Status 状态：0-禁用，1-启用
	Status int `json:"status" gorm:"column:status;type:int;default:1;comment:'状态'"`
	// Avatar 头像（base64 编码的 PNG 图片）
	Avatar string `json:"avatar" gorm:"column:avatar;type:text;comment:'头像 (base64 编码)'"`
	// LastLoginTime 最后登录时间
	LastLoginTime int64 `json:"last_login_time" gorm:"column:last_login_time;type:bigint;default:0;comment:'最后登录时间'"`
	// LastLoginIP 最后登录 IP
	LastLoginIP string `json:"last_login_ip" gorm:"column:last_login_ip;type:varchar(64);comment:'最后登录 IP'"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}
