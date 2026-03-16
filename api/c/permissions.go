package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/panjjo/gosip/api/model"
	"github.com/panjjo/gosip/api/service"
	"github.com/panjjo/gosip/db"
	"github.com/panjjo/gosip/m"
)

// PermissionsListResponse 权限列表响应
type PermissionsListResponse struct {
	Total int64             `json:"total"`
	List  []model.Permission `json:"list"`
}

// @Summary     权限列表接口
// @Description 获取权限列表，支持树形结构
// @Tags        permissions
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       type     query    string  false "权限类型：menu/button/api"
// @Param       parentId query    int     false "父权限 ID，0 表示根节点"
// @Param       limit    query    integer false "条数 (0-100) 默认 100"
// @Param       skip     query    integer false "间隔 默认 0"
// @Param       sort     query    string  false "排序"
// @Success     0        {object} PermissionsListResponse
// @Failure     1000     {object} string
// @Failure     1001     {object} string
// @Router      /api/v1/permissions [get]
func PermissionsList(c *gin.Context) {
	limit := m.GetLimit(c)
	skip := m.GetSkip(c)
	sort := m.GetSort(c)
	permType := c.Query("type")
	parentIDStr := c.Query("parentId")

	query := db.DBClient

	// 类型筛选
	if permType != "" {
		query = query.Where("type = ?", permType)
	}

	// 父 ID 筛选
	if parentIDStr != "" {
		parentID, err := strconv.Atoi(parentIDStr)
		if err == nil {
			query = query.Where("parent_id = ?", parentID)
		}
	}

	permissions := []model.Permission{}
	total, err := db.FindWithJson(query, new(model.Permission), &permissions, c.Query("filters"), sort, skip, limit, true)
	if err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, PermissionsListResponse{
		Total: total,
		List:  permissions,
	})
}

// @Summary     权限树接口
// @Description 获取权限树形结构
// @Tags        permissions
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Success     0 {object} []model.PermissionTree
// @Failure     1000 {object} string
// @Failure     1001 {object} string
// @Router      /api/v1/permissions/tree [get]
func PermissionsTree(c *gin.Context) {
	permissionService := &service.PermissionService{}
	tree, err := permissionService.GetPermissionTree()
	if err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, tree)
}

// @Summary     创建权限
// @Description 创建一个新的权限
// @Tags        permissions
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       name        formData string true "权限名称"
// @Param       code        formData string true "权限编码"
// @Param       type        formData string true "权限类型：menu/button/api"
// @Param       path        formData string false "路径"
// @Param       method      formData string false "HTTP 方法"
// @Param       parentId    formData int    false "父权限 ID"
// @Param       description formData string false "描述"
// @Success     0           {object} model.Permission
// @Failure     1000        {object} string
// @Failure     1001        {object} string
// @Failure     1002        {object} string
// @Router      /api/v1/permissions/create [post]
func PermissionsCreate(c *gin.Context) {
	name := c.PostForm("name")
	if name == "" {
		m.JsonResponse(c, m.StatusParamsERR, "权限名称不能为空")
		return
	}

	code := c.PostForm("code")
	if code == "" {
		m.JsonResponse(c, m.StatusParamsERR, "权限编码不能为空")
		return
	}

	// 检查权限编码是否已存在
	if err := db.Get(db.DBClient.Where("code = ?", code), &model.Permission{}); err == nil {
		m.JsonResponse(c, m.StatusParamsERR, "权限编码已存在")
		return
	}

	permType := c.DefaultPostForm("type", "api")
	if permType != "menu" && permType != "button" && permType != "api" {
		m.JsonResponse(c, m.StatusParamsERR, "权限类型必须是 menu、button 或 api")
		return
	}

	parentID, _ := strconv.Atoi(c.DefaultPostForm("parentId", "0"))

	permission := model.Permission{
		Name:        name,
		Code:        code,
		Type:        permType,
		Path:        c.PostForm("path"),
		Method:      c.PostForm("method"),
		ParentID:    uint(parentID),
		Description: c.PostForm("description"),
	}

	if err := db.Create(db.DBClient, &permission); err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, permission)
}

// @Summary     更新权限
// @Description 更新权限信息
// @Tags        permissions
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       id          path     int    true "权限 ID"
// @Param       name        formData string false "权限名称"
// @Param       type        formData string false "权限类型"
// @Param       path        formData string false "路径"
// @Param       method      formData string false "HTTP 方法"
// @Param       parentId    formData int    false "父权限 ID"
// @Param       description formData string false "描述"
// @Success     0           {object} model.Permission
// @Failure     1000        {object} string
// @Failure     1001        {object} string
// @Failure     1002        {object} string
// @Router      /api/v1/permissions/:id [post]
func PermissionsUpdate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "权限 ID 格式错误")
		return
	}

	permission := &model.Permission{}
	if err := db.DBClient.First(permission, uint(id)).Error; err != nil {
		if db.RecordNotFound(err) {
			m.JsonResponse(c, m.StatusParamsERR, "权限不存在")
			return
		}
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	// 更新字段
	if name := c.PostForm("name"); name != "" {
		permission.Name = name
	}
	if permType := c.PostForm("type"); permType != "" {
		if permType != "menu" && permType != "button" && permType != "api" {
			m.JsonResponse(c, m.StatusParamsERR, "权限类型必须是 menu、button 或 api")
			return
		}
		permission.Type = permType
	}
	if path := c.PostForm("path"); path != "" {
		permission.Path = path
	}
	if method := c.PostForm("method"); method != "" {
		permission.Method = method
	}
	if parentID := c.PostForm("parentId"); parentID != "" {
		parentIDUint, _ := strconv.ParseUint(parentID, 10, 32)
		permission.ParentID = uint(parentIDUint)
	}
	if description := c.PostForm("description"); description != "" {
		permission.Description = description
	}

	if err := db.Save(db.DBClient, permission); err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, permission)
}

// @Summary     删除权限
// @Description 删除权限
// @Tags        permissions
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       id   path     int true "权限 ID"
// @Success     0    {object} string
// @Failure     1000 {object} string
// @Failure     1001 {object} string
// @Router      /api/v1/permissions/:id [delete]
func PermissionsDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "权限 ID 格式错误")
		return
	}

	permission := &model.Permission{}
	if err := db.DBClient.First(permission, uint(id)).Error; err != nil {
		if db.RecordNotFound(err) {
			m.JsonResponse(c, m.StatusParamsERR, "权限不存在")
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
	if err := tx.Where("permission_id = ?", permission.ID).Delete(&model.RolePermission{}).Error; err != nil {
		tx.Rollback()
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	// 删除权限
	if err := db.Del(tx, permission); err != nil {
		tx.Rollback()
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	tx.Commit()
	m.JsonResponse(c, m.StatusSucc, "删除成功")
}

// @Summary     移除角色权限
// @Description 移除角色的单个权限
// @Tags        permissions
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       roleId       path     int true "角色 ID"
// @Param       permissionId path     int true "权限 ID"
// @Success     0            {object} string
// @Failure     1000         {object} string
// @Failure     1001         {object} string
// @Router      /api/v1/roles/:roleId/permissions/:permissionId [delete]
func RolesRemovePermission(c *gin.Context) {
	roleIDStr := c.Param("roleId")
	permissionIDStr := c.Param("permissionId")

	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "角色 ID 格式错误")
		return
	}

	permissionID, err := strconv.Atoi(permissionIDStr)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "权限 ID 格式错误")
		return
	}

	permissionService := &service.PermissionService{}
	if err := permissionService.RemovePermissionFromRole(uint(roleID), uint(permissionID)); err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, "移除成功")
}
