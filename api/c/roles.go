package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/panjjo/gosip/api/model"
	"github.com/panjjo/gosip/api/service"
	"github.com/panjjo/gosip/db"
	"github.com/panjjo/gosip/m"
)

// RolesListResponse 角色列表响应
type RolesListResponse struct {
	Total int64         `json:"total"`
	List  []model.Role `json:"list"`
}

// @Summary     角色列表接口
// @Description 获取角色列表，支持分页和筛选
// @Tags        roles
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       limit   query    integer false "条数 (0-100) 默认 20"
// @Param       skip    query    integer false "间隔 默认 0"
// @Param       sort    query    string  false "排序，例:-sort 根据 sort 倒序"
// @Param       filters query    string  false "查询条件"
// @Success     0       {object} RolesListResponse
// @Failure     1000    {object} string
// @Failure     1001    {object} string
// @Router      /api/v1/roles [get]
func RolesList(c *gin.Context) {
	limit := m.GetLimit(c)
	skip := m.GetSkip(c)
	sort := m.GetSort(c)

	roles := []model.Role{}
	total, err := db.FindWithJson(db.DBClient, new(model.Role), &roles, c.Query("filters"), sort, skip, limit, true)
	if err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, RolesListResponse{
		Total: total,
		List:  roles,
	})
}

// @Summary     创建角色
// @Description 创建一个新的角色
// @Tags        roles
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       name        formData string true "角色名称"
// @Param       code        formData string true "角色编码"
// @Param       description formData string false "描述"
// @Param       sort        formData int    false "排序"
// @Success     0           {object} model.Role
// @Failure     1000        {object} string
// @Failure     1001        {object} string
// @Failure     1002        {object} string
// @Router      /api/v1/roles/create [post]
func RolesCreate(c *gin.Context) {
	name := c.PostForm("name")
	if name == "" {
		m.JsonResponse(c, m.StatusParamsERR, "角色名称不能为空")
		return
	}

	code := c.PostForm("code")
	if code == "" {
		m.JsonResponse(c, m.StatusParamsERR, "角色编码不能为空")
		return
	}

	// 检查角色编码是否已存在
	if err := db.Get(db.DBClient.Where("code = ?", code), &model.Role{}); err == nil {
		m.JsonResponse(c, m.StatusParamsERR, "角色编码已存在")
		return
	}

	sort, _ := strconv.Atoi(c.DefaultPostForm("sort", "0"))

	role := model.Role{
		Name:        name,
		Code:        code,
		Description: c.PostForm("description"),
		Sort:        sort,
		Status:      1,
	}

	if err := db.Create(db.DBClient, &role); err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, role)
}

// @Summary     更新角色
// @Description 更新角色信息
// @Tags        roles
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       id          path     int    true "角色 ID"
// @Param       name        formData string false "角色名称"
// @Param       description formData string false "描述"
// @Param       sort        formData int    false "排序"
// @Param       status      formData int    false "状态：0-禁用，1-启用"
// @Success     0           {object} model.Role
// @Failure     1000        {object} string
// @Failure     1001        {object} string
// @Failure     1002        {object} string
// @Router      /api/v1/roles/:id [post]
func RolesUpdate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "角色 ID 格式错误")
		return
	}

	role := &model.Role{}
	if err := db.DBClient.First(role, uint(id)).Error; err != nil {
		if db.RecordNotFound(err) {
			m.JsonResponse(c, m.StatusParamsERR, "角色不存在")
			return
		}
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	// 更新字段
	if name := c.PostForm("name"); name != "" {
		role.Name = name
	}
	if description := c.PostForm("description"); description != "" {
		role.Description = description
	}
	if sort := c.PostForm("sort"); sort != "" {
		role.Sort, _ = strconv.Atoi(sort)
	}
	if status := c.PostForm("status"); status != "" {
		role.Status, _ = strconv.Atoi(status)
	}

	if err := db.Save(db.DBClient, role); err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, role)
}

// @Summary     删除角色
// @Description 删除角色
// @Tags        roles
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       id   path     int true "角色 ID"
// @Success     0    {object} string
// @Failure     1000 {object} string
// @Failure     1001 {object} string
// @Router      /api/v1/roles/:id [delete]
func RolesDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "角色 ID 格式错误")
		return
	}

	role := &model.Role{}
	if err := db.DBClient.First(role, uint(id)).Error; err != nil {
		if db.RecordNotFound(err) {
			m.JsonResponse(c, m.StatusParamsERR, "角色不存在")
			return
		}
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	tx := db.DBClient.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除角色权限关联
	if err := tx.Where("role_id = ?", role.ID).Delete(&model.RolePermission{}).Error; err != nil {
		tx.Rollback()
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	// 删除用户角色关联
	if err := tx.Where("role_id = ?", role.ID).Delete(&model.UserRole{}).Error; err != nil {
		tx.Rollback()
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	// 删除角色
	if err := db.Del(tx, role); err != nil {
		tx.Rollback()
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	tx.Commit()
	m.JsonResponse(c, m.StatusSucc, "删除成功")
}

// @Summary     启用角色
// @Description 启用角色
// @Tags        roles
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       id   path     int true "角色 ID"
// @Success     0    {object} string
// @Failure     1000 {object} string
// @Failure     1001 {object} string
// @Router      /api/v1/roles/:id/enable [post]
func RolesEnable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "角色 ID 格式错误")
		return
	}

	role := &model.Role{}
	if err := db.DBClient.First(role, uint(id)).Error; err != nil {
		if db.RecordNotFound(err) {
			m.JsonResponse(c, m.StatusParamsERR, "角色不存在")
			return
		}
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	role.Status = 1
	if err := db.Save(db.DBClient, role); err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, "启用成功")
}

// @Summary     禁用角色
// @Description 禁用角色
// @Tags        roles
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       id   path     int true "角色 ID"
// @Success     0    {object} string
// @Failure     1000 {object} string
// @Failure     1001 {object} string
// @Router      /api/v1/roles/:id/disable [post]
func RolesDisable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "角色 ID 格式错误")
		return
	}

	role := &model.Role{}
	if err := db.DBClient.First(role, uint(id)).Error; err != nil {
		if db.RecordNotFound(err) {
			m.JsonResponse(c, m.StatusParamsERR, "角色不存在")
			return
		}
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	role.Status = 0
	if err := db.Save(db.DBClient, role); err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, "禁用成功")
}

// @Summary     给角色分配权限
// @Description 给角色分配权限
// @Tags        roles
// @Accept      json
// @Produce     json
// @Param       id   path     int  true "角色 ID"
// @Param       data body     model.AssignPermissionsRequest true "权限 ID 列表"
// @Success     0    {object} string
// @Failure     1000 {object} string
// @Failure     1001 {object} string
// @Failure     1002 {object} string
// @Router      /api/v1/roles/:id/permissions [post]
func RolesAssignPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "角色 ID 格式错误")
		return
	}

	var req model.AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		m.JsonResponse(c, m.StatusParamsERR, err)
		return
	}

	permissionService := &service.PermissionService{}
	if err := permissionService.AssignPermissionsToRole(uint(id), req.PermissionIDs); err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, "分配成功")
}

// @Summary     获取角色的权限列表
// @Description 获取角色的所有权限
// @Tags        roles
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       id   path     int true "角色 ID"
// @Success     0    {object} []model.Permission
// @Failure     1000 {object} string
// @Failure     1001 {object} string
// @Router      /api/v1/roles/:id/permissions [get]
func RolesGetPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "角色 ID 格式错误")
		return
	}

	permissionService := &service.PermissionService{}
	permissions, err := permissionService.GetRolePermissions(uint(id))
	if err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, permissions)
}

// @Summary     获取用户的角色列表
// @Description 获取用户的所有角色
// @Tags        roles
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       id   path     int true "用户 ID"
// @Success     0    {object} []model.Role
// @Failure     1000 {object} string
// @Failure     1001 {object} string
// @Router      /api/v1/users/:id/roles [get]
func UsersGetRoles(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "用户 ID 格式错误")
		return
	}

	permissionService := &service.PermissionService{}
	roles, err := permissionService.GetUserRoles(uint(id))
	if err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, roles)
}

// @Summary     给用户分配角色
// @Description 给用户分配角色
// @Tags        roles
// @Accept      json
// @Produce     json
// @Param       id   path     int  true "用户 ID"
// @Param       data body     model.AssignRolesRequest true "角色 ID 列表"
// @Success     0    {object} string
// @Failure     1000 {object} string
// @Failure     1001 {object} string
// @Failure     1002 {object} string
// @Router      /api/v1/users/:id/roles [post]
func UsersAssignRoles(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "用户 ID 格式错误")
		return
	}

	var req model.AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		m.JsonResponse(c, m.StatusParamsERR, err)
		return
	}

	permissionService := &service.PermissionService{}
	if err := permissionService.AssignRolesToUser(uint(id), req.RoleIDs); err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, "分配成功")
}

// @Summary     移除用户角色
// @Description 移除用户的单个角色
// @Tags        roles
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       userId path     int true "用户 ID"
// @Param       roleId path     int true "角色 ID"
// @Success     0      {object} string
// @Failure     1000   {object} string
// @Failure     1001   {object} string
// @Router      /api/v1/users/:userId/roles/:roleId [delete]
func UsersRemoveRole(c *gin.Context) {
	userIDStr := c.Param("userId")
	roleIDStr := c.Param("roleId")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "用户 ID 格式错误")
		return
	}

	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "角色 ID 格式错误")
		return
	}

	permissionService := &service.PermissionService{}
	if err := permissionService.RemoveRoleFromUser(uint(userID), uint(roleID)); err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, "移除成功")
}

// @Summary     检查用户权限
// @Description 检查用户是否有某权限
// @Tags        roles
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       id       path     int    true "用户 ID"
// @Param       code     query    string true "权限编码"
// @Success     0        {object} model.CheckPermissionResponse
// @Failure     1000     {object} string
// @Failure     1001     {object} string
// @Router      /api/v1/users/:id/hasPermission [get]
func UsersHasPermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "用户 ID 格式错误")
		return
	}

	code := c.Query("code")
	if code == "" {
		m.JsonResponse(c, m.StatusParamsERR, "权限编码不能为空")
		return
	}

	permissionService := &service.PermissionService{}
	hasPermission, err := permissionService.CheckPermission(uint(id), code)
	if err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, model.CheckPermissionResponse{
		HasPermission: hasPermission,
	})
}

// @Summary     获取用户权限列表
// @Description 获取用户的所有权限
// @Tags        roles
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       id   path     int true "用户 ID"
// @Success     0    {object} model.UserPermissionsResponse
// @Failure     1000 {object} string
// @Failure     1001 {object} string
// @Router      /api/v1/users/:id/permissions [get]
func UsersGetPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "用户 ID 格式错误")
		return
	}

	permissionService := &service.PermissionService{}
	permissions, err := permissionService.GetUserPermissions(uint(id))
	if err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	// 提取权限编码列表
	codes := make([]string, len(permissions))
	for i, perm := range permissions {
		codes[i] = perm.Code
	}

	m.JsonResponse(c, m.StatusSucc, model.UserPermissionsResponse{
		Permissions: permissions,
		Codes:       codes,
	})
}
