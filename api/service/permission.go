package service

import (
	"github.com/panjjo/gosip/api/model"
	"github.com/panjjo/gosip/db"
	"github.com/panjjo/gosip/utils"
)

// PermissionService 权限服务
type PermissionService struct{}

// GetRoleByCode 根据编码获取角色（包级别函数）
func GetRoleByCode(code string) (*model.Role, error) {
	role := &model.Role{}
	err := db.Get(db.DBClient.Where("code = ?", code), role)
	return role, err
}

// GetRoleByCode 根据编码获取角色（方法）
func (s *PermissionService) GetRoleByCode(code string) (*model.Role, error) {
	role := &model.Role{}
	err := db.Get(db.DBClient.Where("code = ?", code), role)
	return role, err
}

// GetRoleByID 根据 ID 获取角色
func (s *PermissionService) GetRoleByID(id uint) (*model.Role, error) {
	role := &model.Role{}
	err := db.Get(db.DBClient, role)
	return role, err
}

// GetRolePermissions 获取角色的所有权限
func (s *PermissionService) GetRolePermissions(roleID uint) ([]model.Permission, error) {
	// 查询角色权限关联
	var rolePermissions []model.RolePermission
	if err := db.DBClient.Where("role_id = ?", roleID).Find(&rolePermissions).Error; err != nil {
		return nil, err
	}

	if len(rolePermissions) == 0 {
		return []model.Permission{}, nil
	}

	// 提取权限 ID 列表
	permissionIDs := make([]uint, len(rolePermissions))
	for i, rp := range rolePermissions {
		permissionIDs[i] = rp.PermissionID
	}

	// 查询权限详情
	var permissions []model.Permission
	if err := db.DBClient.Where("id IN (?)", permissionIDs).Order("id ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}

// GetUserRoles 获取用户的所有角色
func (s *PermissionService) GetUserRoles(userID uint) ([]model.Role, error) {
	// 查询用户角色关联
	var userRoles []model.UserRole
	if err := db.DBClient.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return nil, err
	}

	if len(userRoles) == 0 {
		return []model.Role{}, nil
	}

	// 提取角色 ID 列表
	roleIDs := make([]uint, len(userRoles))
	for i, ur := range userRoles {
		roleIDs[i] = ur.RoleID
	}

	// 查询角色详情
	var roles []model.Role
	if err := db.DBClient.Where("id IN (?) AND status = 1", roleIDs).Find(&roles).Error; err != nil {
		return nil, err
	}

	return roles, nil
}

// GetUserPermissions 获取用户的所有权限
func (s *PermissionService) GetUserPermissions(userID uint) ([]model.Permission, error) {
	// 获取用户的所有角色
	roles, err := s.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		return []model.Permission{}, nil
	}

	// 提取角色 ID 列表
	roleIDs := make([]uint, len(roles))
	for i, role := range roles {
		roleIDs[i] = role.ID
	}

	// 查询角色权限关联
	var rolePermissions []model.RolePermission
	if err := db.DBClient.Where("role_id IN (?)", roleIDs).Find(&rolePermissions).Error; err != nil {
		return nil, err
	}

	if len(rolePermissions) == 0 {
		return []model.Permission{}, nil
	}

	// 提取权限 ID 列表（去重）
	permissionMap := make(map[uint]bool)
	for _, rp := range rolePermissions {
		permissionMap[rp.PermissionID] = true
	}

	permissionIDs := make([]uint, 0, len(permissionMap))
	for id := range permissionMap {
		permissionIDs = append(permissionIDs, id)
	}

	// 查询权限详情
	var permissions []model.Permission
	if err := db.DBClient.Where("id IN (?)", permissionIDs).Order("id ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}

// CheckPermission 检查用户是否有某权限（通过权限编码）
func (s *PermissionService) CheckPermission(userID uint, permissionCode string) (bool, error) {
	permissions, err := s.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm.Code == permissionCode {
			return true, nil
		}
	}

	return false, nil
}

// CheckPermissionByPath 检查用户是否有某权限（通过路径和方法）
func (s *PermissionService) CheckPermissionByPath(userID uint, path, method string) (bool, error) {
	permissions, err := s.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		// 精确匹配
		if perm.Path == path && (perm.Method == "" || perm.Method == method) {
			return true, nil
		}
		// 通配符匹配（如 /api/v1/devices/:id 匹配 /api/v1/devices/123）
		if s.matchPathWithParams(perm.Path, path) && (perm.Method == "" || perm.Method == method) {
			return true, nil
		}
	}

	return false, nil
}

// matchPathWithParams 匹配带参数的路径
func (s *PermissionService) matchPathWithParams(template, actual string) bool {
	templateParts := splitPath(template)
	actualParts := splitPath(actual)

	if len(templateParts) != len(actualParts) {
		return false
	}

	for i, part := range templateParts {
		// 如果模板部分是 :param 格式，则跳过匹配
		if len(part) > 0 && part[0] == ':' {
			continue
		}
		if part != actualParts[i] {
			return false
		}
	}

	return true
}

// splitPath 分割路径
func splitPath(path string) []string {
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				result = append(result, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		result = append(result, path[start:])
	}
	return result
}

// AssignPermissionsToRole 给角色分配权限
func (s *PermissionService) AssignPermissionsToRole(roleID uint, permissionIDs []uint) error {
	tx := db.DBClient.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 删除旧的权限关联
	if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. 添加新的权限关联
	for _, pid := range permissionIDs {
		// 验证权限是否存在
		var perm model.Permission
		if err := tx.First(&perm, pid).Error; err != nil {
			tx.Rollback()
			return err
		}

		rp := model.RolePermission{
			RoleID:       roleID,
			PermissionID: pid,
		}
		if err := tx.Create(&rp).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// AssignRolesToUser 给用户分配角色
func (s *PermissionService) AssignRolesToUser(userID uint, roleIDs []uint) error {
	tx := db.DBClient.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 删除旧的角色关联
	if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. 添加新的角色关联
	for _, rid := range roleIDs {
		// 验证角色是否存在且启用
		var role model.Role
		if err := tx.First(&role, rid).Error; err != nil {
			tx.Rollback()
			return err
		}
		if role.Status != 1 {
			tx.Rollback()
			return db.NewError(nil, "角色已禁用")
		}

		ur := model.UserRole{
			UserID: userID,
			RoleID: rid,
		}
		if err := tx.Create(&ur).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// AddRoleToUser 给用户添加单个角色
func (s *PermissionService) AddRoleToUser(userID, roleID uint) error {
	// 检查是否已存在
	var count int64
	if err := db.DBClient.Model(&model.UserRole{}).Where("user_id = ? AND role_id = ?", userID, roleID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil // 已存在，不重复添加
	}

	// 验证角色是否存在且启用
	var role model.Role
	if err := db.DBClient.First(&role, roleID).Error; err != nil {
		return err
	}
	if role.Status != 1 {
		return db.NewError(nil, "角色已禁用")
	}

	ur := model.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return db.Create(db.DBClient, &ur)
}

// RemoveRoleFromUser 移除用户的角色
func (s *PermissionService) RemoveRoleFromUser(userID, roleID uint) error {
	return db.DBClient.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&model.UserRole{}).Error
}

// RemovePermissionFromRole 移除角色的权限
func (s *PermissionService) RemovePermissionFromRole(roleID, permissionID uint) error {
	return db.DBClient.Where("role_id = ? AND permission_id = ?", roleID, permissionID).Delete(&model.RolePermission{}).Error
}

// GetPermissionTree 获取权限树
func (s *PermissionService) GetPermissionTree() ([]model.PermissionTree, error) {
	// 获取所有权限
	var permissions []model.Permission
	if err := db.DBClient.Order("sort ASC, id ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}

	// 构建树形结构
	return s.buildPermissionTree(permissions, 0), nil
}

// buildPermissionTree 递归构建权限树
func (s *PermissionService) buildPermissionTree(all []model.Permission, parentID uint) []model.PermissionTree {
	var result []model.PermissionTree

	for _, perm := range all {
		if perm.ParentID == parentID {
			node := model.PermissionTree{
				Permission: perm,
				Children:   s.buildPermissionTree(all, perm.ID),
			}
			result = append(result, node)
		}
	}

	return result
}

// InitDefaultRoles 初始化默认角色
func (s *PermissionService) InitDefaultRoles() error {
	defaultRoles := []model.Role{
		{Name: "超级管理员", Code: "admin", Description: "系统超级管理员，拥有所有权限", Status: 1, Sort: 1},
		{Name: "操作员", Code: "operator", Description: "系统操作员，拥有操作权限", Status: 1, Sort: 2},
		{Name: "观察者", Code: "viewer", Description: "只读权限用户", Status: 1, Sort: 3},
	}

	for _, role := range defaultRoles {
		var existing model.Role
		result := db.DBClient.Where("code = ?", role.Code).First(&existing)
		if result.Error != nil {
			if db.RecordNotFound(result.Error) {
				// 角色不存在，创建
				if err := db.Create(db.DBClient, &role); err != nil {
					return err
				}
			} else {
				return result.Error
			}
		}
	}

	return nil
}

// InitDefaultPermissions 初始化默认权限
func (s *PermissionService) InitDefaultPermissions() error {
	defaultPermissions := []model.Permission{
		// 健康检查
		{Name: "健康检查", Code: "health:check", Type: "api", Path: "/api/v1/health", Method: "GET", Description: "健康检查接口"},

		// 设备管理
		{Name: "设备管理", Code: "device", Type: "menu", Path: "/devices", ParentID: 0, Description: "设备管理模块"},
		{Name: "设备列表", Code: "device:list", Type: "api", Path: "/api/v1/devices", Method: "GET", ParentID: 0, Description: "查看设备列表"},
		{Name: "创建设备", Code: "device:create", Type: "api", Path: "/api/v1/devices/create", Method: "POST", ParentID: 0, Description: "创建新设备"},
		{Name: "更新设备", Code: "device:update", Type: "api", Path: "/api/v1/devices/:id", Method: "POST", ParentID: 0, Description: "更新设备信息"},
		{Name: "删除设备", Code: "device:delete", Type: "api", Path: "/api/v1/devices/:id", Method: "DELETE", ParentID: 0, Description: "删除设备"},
		{Name: "PTZ 控制", Code: "device:ptz", Type: "api", Path: "/api/v1/devices/ptz", Method: "POST", ParentID: 0, Description: "云台控制"},

		// 通道管理
		{Name: "通道管理", Code: "channel", Type: "menu", Path: "/channels", ParentID: 0, Description: "通道管理模块"},
		{Name: "通道列表", Code: "channel:list", Type: "api", Path: "/api/v1/channels", Method: "GET", ParentID: 0, Description: "查看通道列表"},
		{Name: "创建通道", Code: "channel:create", Type: "api", Path: "/api/v1/devices/:id/channels", Method: "POST", ParentID: 0, Description: "创建通道"},
		{Name: "更新通道", Code: "channel:update", Type: "api", Path: "/api/v1/channels/:id", Method: "POST", ParentID: 0, Description: "更新通道"},
		{Name: "删除通道", Code: "channel:delete", Type: "api", Path: "/api/v1/channels/:id", Method: "DELETE", ParentID: 0, Description: "删除通道"},
		{Name: "通道同步", Code: "channel:sync", Type: "api", Path: "/api/v1/devices/:id/channels_sync", Method: "POST", ParentID: 0, Description: "同步通道"},

		// 流管理
		{Name: "流管理", Code: "stream", Type: "menu", Path: "/streams", ParentID: 0, Description: "流管理模块"},
		{Name: "流列表", Code: "stream:list", Type: "api", Path: "/api/v1/streams", Method: "GET", ParentID: 0, Description: "查看流列表"},
		{Name: "开始播放", Code: "stream:play", Type: "api", Path: "/api/v1/channels/:id/streams", Method: "POST", ParentID: 0, Description: "开始播放"},
		{Name: "停止播放", Code: "stream:stop", Type: "api", Path: "/api/v1/streams/:id", Method: "DELETE", ParentID: 0, Description: "停止播放"},

		// 录像管理
		{Name: "录像管理", Code: "record", Type: "menu", Path: "/records", ParentID: 0, Description: "录像管理模块"},
		{Name: "录像列表", Code: "record:list", Type: "api", Path: "/api/v1/channels/:id/records", Method: "GET", ParentID: 0, Description: "查看录像列表"},

		// 用户管理
		{Name: "用户管理", Code: "user", Type: "menu", Path: "/users", ParentID: 0, Description: "用户管理模块"},
		{Name: "用户列表", Code: "user:list", Type: "api", Path: "/api/v1/users", Method: "GET", ParentID: 0, Description: "查看用户列表"},
		{Name: "创建用户", Code: "user:create", Type: "api", Path: "/api/v1/users/create", Method: "POST", ParentID: 0, Description: "创建用户"},
		{Name: "更新用户", Code: "user:update", Type: "api", Path: "/api/v1/users/:id", Method: "POST", ParentID: 0, Description: "更新用户"},
		{Name: "删除用户", Code: "user:delete", Type: "api", Path: "/api/v1/users/:id", Method: "DELETE", ParentID: 0, Description: "删除用户"},
		{Name: "启用用户", Code: "user:enable", Type: "api", Path: "/api/v1/users/:id/enable", Method: "POST", ParentID: 0, Description: "启用用户"},
		{Name: "禁用用户", Code: "user:disable", Type: "api", Path: "/api/v1/users/:id/disable", Method: "POST", ParentID: 0, Description: "禁用用户"},
		{Name: "修改密码", Code: "user:password", Type: "api", Path: "/api/v1/users/:id/password", Method: "POST", ParentID: 0, Description: "修改密码"},

		// 角色管理
		{Name: "角色管理", Code: "role", Type: "menu", Path: "/roles", ParentID: 0, Description: "角色管理模块"},
		{Name: "角色列表", Code: "role:list", Type: "api", Path: "/api/v1/roles", Method: "GET", ParentID: 0, Description: "查看角色列表"},
		{Name: "创建角色", Code: "role:create", Type: "api", Path: "/api/v1/roles/create", Method: "POST", ParentID: 0, Description: "创建角色"},
		{Name: "更新角色", Code: "role:update", Type: "api", Path: "/api/v1/roles/:id", Method: "POST", ParentID: 0, Description: "更新角色"},
		{Name: "删除角色", Code: "role:delete", Type: "api", Path: "/api/v1/roles/:id", Method: "DELETE", ParentID: 0, Description: "删除角色"},
		{Name: "启用角色", Code: "role:enable", Type: "api", Path: "/api/v1/roles/:id/enable", Method: "POST", ParentID: 0, Description: "启用角色"},
		{Name: "禁用角色", Code: "role:disable", Type: "api", Path: "/api/v1/roles/:id/disable", Method: "POST", ParentID: 0, Description: "禁用角色"},
		{Name: "分配权限", Code: "role:assign", Type: "api", Path: "/api/v1/roles/:id/permissions", Method: "POST", ParentID: 0, Description: "分配权限"},

		// 权限管理
		{Name: "权限管理", Code: "permission", Type: "menu", Path: "/permissions", ParentID: 0, Description: "权限管理模块"},
		{Name: "权限列表", Code: "permission:list", Type: "api", Path: "/api/v1/permissions", Method: "GET", ParentID: 0, Description: "查看权限列表"},
		{Name: "创建权限", Code: "permission:create", Type: "api", Path: "/api/v1/permissions/create", Method: "POST", ParentID: 0, Description: "创建权限"},
		{Name: "更新权限", Code: "permission:update", Type: "api", Path: "/api/v1/permissions/:id", Method: "POST", ParentID: 0, Description: "更新权限"},
		{Name: "删除权限", Code: "permission:delete", Type: "api", Path: "/api/v1/permissions/:id", Method: "DELETE", ParentID: 0, Description: "删除权限"},

		// ZLM 管理
		{Name: "ZLM 管理", Code: "zlm", Type: "menu", Path: "/zlm", ParentID: 0, Description: "ZLM 管理模块"},
		{Name: "媒体状态", Code: "zlm:status", Type: "api", Path: "/api/v1/media/status", Method: "GET", ParentID: 0, Description: "媒体服务器状态"},
		{Name: "流心跳", Code: "zlm:keepalive", Type: "api", Path: "/api/v1/stream/keepalive", Method: "POST", ParentID: 0, Description: "流心跳接口"},
		{Name: "Webhook", Code: "zlm:webhook", Type: "api", Path: "/zlm/webhook/:method", Method: "POST", ParentID: 0, Description: "ZLM Webhook"},
	}

	for _, perm := range defaultPermissions {
		var existing model.Permission
		result := db.DBClient.Where("code = ?", perm.Code).First(&existing)
		if result.Error != nil {
			if db.RecordNotFound(result.Error) {
				// 权限不存在，创建
				if err := db.Create(db.DBClient, &perm); err != nil {
					return err
				}
			} else {
				return result.Error
			}
		}
	}

	return nil
}

// InitAdminUser 初始化超级管理员用户
func (s *PermissionService) InitAdminUser() error {
	// 检查是否已有 admin 用户
	var adminUser model.User
	result := db.DBClient.Where("username = ?", "admin").First(&adminUser)
	if result.Error == nil {
		// 已存在
		return nil
	}
	if !db.RecordNotFound(result.Error) {
		return result.Error
	}

	// 创建 admin 用户
	adminUser = model.User{
		Username: "admin",
		Password: utils.EncryptPassword("admin123"), // 默认密码
		Name:     "超级管理员",
		Email:    "admin@example.com",
		Role:     "admin",
		Status:   1,
	}

	if err := db.Create(db.DBClient, &adminUser); err != nil {
		return err
	}

	// 获取 admin 角色
	var adminRole model.Role
	if err := db.DBClient.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		return err
	}

	// 将 admin 用户关联到 admin 角色
	ur := model.UserRole{
		UserID: adminUser.ID,
		RoleID: adminRole.ID,
	}
	return db.Create(db.DBClient, &ur)
}

// InitAdminPermissions 给 admin 角色分配所有权限
func (s *PermissionService) InitAdminPermissions() error {
	// 获取 admin 角色
	var adminRole model.Role
	if err := db.DBClient.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		return err
	}

	// 获取所有权限
	var allPermissions []model.Permission
	if err := db.DBClient.Find(&allPermissions).Error; err != nil {
		return err
	}

	// 获取 admin 角色已有的权限
	existingPermissions, err := s.GetRolePermissions(adminRole.ID)
	if err != nil {
		return err
	}

	// 如果已有权限数量等于总权限数，说明已分配
	if len(existingPermissions) == len(allPermissions) {
		return nil
	}

	// 分配所有权限
	permissionIDs := make([]uint, len(allPermissions))
	for i, perm := range allPermissions {
		permissionIDs[i] = perm.ID
	}

	return s.AssignPermissionsToRole(adminRole.ID, permissionIDs)
}

// Init 初始化权限系统
func (s *PermissionService) Init() error {
	// 1. 初始化默认角色
	if err := s.InitDefaultRoles(); err != nil {
		return err
	}

	// 2. 初始化默认权限
	if err := s.InitDefaultPermissions(); err != nil {
		return err
	}

	// 3. 初始化 admin 用户
	if err := s.InitAdminUser(); err != nil {
		return err
	}

	// 4. 给 admin 分配所有权限
	if err := s.InitAdminPermissions(); err != nil {
		return err
	}

	return nil
}
