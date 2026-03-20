package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/panjjo/gosip/api/model"
	"github.com/panjjo/gosip/db"
	"github.com/panjjo/gosip/m"
	"github.com/panjjo/gosip/utils"
)

// UsersListResponse 用户列表响应
type UsersListResponse struct {
	Total int64        `json:"total"`
	List  []model.User `json:"list"`
}

// @Summary     用户列表接口
// @Description 获取用户列表，支持分页和筛选
// @Tags        users
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       limit   query    integer false "条数 (0-100) 默认 20"
// @Param       skip    query    integer false "间隔 默认 0"
// @Param       sort    query    string  false "排序，例:-key，根据 key 倒序，key，根据 key 正序"
// @Param       filters query    string  false "查询条件，使用规则详情请看帮助"
// @Success     0       {object} UsersListResponse
// @Failure     1000    {object} string
// @Failure     1001    {object} string
// @Failure     1002    {object} string
// @Failure     1003    {object} string
// @Router      /api/v1/users [get]
func UsersList(c *gin.Context) {
	limit := m.GetLimit(c)
	skip := m.GetSkip(c)
	sort := m.GetSort(c)
	users := []model.User{}
	total, err := db.FindWithJson(db.DBClient, new(model.User), &users, c.Query("filters"), sort, skip, limit, true)
	if err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}
	m.JsonResponse(c, m.StatusSucc, UsersListResponse{
		Total: total,
		List:  users,
	})
}

// @Summary     用户新增接口
// @Description 新增一个用户
// @Tags        users
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       username formData string true "用户名"
// @Param       password formData string true "密码"
// @Param       name     formData string false "姓名"
// @Param       email    formData string false "邮箱"
// @Param       phone    formData string false "手机号"
// @Param       role     formData string false "角色：admin-管理员，operator-操作员，viewer-观察者"
// @Success     0        {object} sipapi.User
// @Failure     1000     {object} string
// @Failure     1001     {object} string
// @Failure     1002     {object} string
// @Failure     1003     {object} string
// @Router      /api/v1/users/create [post]
func UsersCreate(c *gin.Context) {
	username := c.PostForm("username")
	if username == "" {
		m.JsonResponse(c, m.StatusParamsERR, "用户名不能为空")
		return
	}
	password := c.PostForm("password")
	if password == "" {
		m.JsonResponse(c, m.StatusParamsERR, "密码不能为空")
		return
	}

	// 检查用户名是否已存在
	if err := db.Get(db.DBClient, &model.User{Username: username}); err == nil {
		m.JsonResponse(c, m.StatusParamsERR, "用户名已存在")
		return
	}

	user := model.User{
		Username: username,
		Password: utils.EncryptPassword(password),
		Name:     c.PostForm("name"),
		Email:    c.PostForm("email"),
		Phone:    c.PostForm("phone"),
		Role:     c.DefaultPostForm("role", "viewer"),
		Status:   1,
		Avatar:   utils.GenerateAvatar(username, 200), // 生成随机马赛克头像
	}

	tx, err := db.NewTx(db.DBClient)
	if err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}
	defer tx.End()

	if err := db.Create(tx.DB(), &user); err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	tx.Commit()
	m.JsonResponse(c, m.StatusSucc, user)
}

// @Summary     用户修改接口
// @Description 修改用户信息
// @Tags        users
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       id    path     string true "用户 ID"
// @Param       name  formData string false "姓名"
// @Param       email formData string false "邮箱"
// @Param       phone formData string false "手机号"
// @Param       role  formData string false "角色：admin-管理员，operator-操作员，viewer-观察者"
// @Success     0     {object} sipapi.User
// @Failure     1000  {object} string
// @Failure     1001  {object} string
// @Failure     1002  {object} string
// @Failure     1003  {object} string
// @Router      /api/v1/users/{id} [post]
func UsersUpdate(c *gin.Context) {
	userid := c.Param("id")
	if userid == "" {
		m.JsonResponse(c, m.StatusParamsERR, "用户 ID 不能为空")
		return
	}

	id, err := strconv.Atoi(userid)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "用户 ID 格式错误")
		return
	}

	user := &model.User{}
	if err := db.DBClient.First(user, uint(id)).Error; err != nil {
		if db.RecordNotFound(err) {
			m.JsonResponse(c, m.StatusParamsERR, "用户不存在")
			return
		}
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	// 更新字段
	if name := c.PostForm("name"); name != "" {
		user.Name = name
	}
	if email := c.PostForm("email"); email != "" {
		user.Email = email
	}
	if phone := c.PostForm("phone"); phone != "" {
		user.Phone = phone
	}
	if role := c.PostForm("role"); role != "" {
		user.Role = role
	}
	// 支持上传自定义头像（base64 编码）
	if avatar := c.PostForm("avatar"); avatar != "" {
		user.Avatar = avatar
	}

	if err := db.Save(db.DBClient, user); err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, user)
}

// @Summary     用户删除接口
// @Description 删除用户
// @Tags        users
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       id   path     string true "用户 ID"
// @Success     0    {object} string
// @Failure     1000 {object} string
// @Failure     1001 {object} string
// @Failure     1002 {object} string
// @Failure     1003 {object} string
// @Router      /api/v1/users/{id} [delete]
func UsersDelete(c *gin.Context) {
	userid := c.Param("id")
	if userid == "" {
		m.JsonResponse(c, m.StatusParamsERR, "用户 ID 不能为空")
		return
	}

	id, err := strconv.Atoi(userid)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "用户 ID 格式错误")
		return
	}

	user := &model.User{}
	if err := db.DBClient.First(user, uint(id)).Error; err != nil {
		if db.RecordNotFound(err) {
			m.JsonResponse(c, m.StatusParamsERR, "用户不存在")
			return
		}
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	if err := db.Del(db.DBClient, user); err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, "删除成功")
}

// @Summary     用户启用接口
// @Description 启用用户
// @Tags        users
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       id   path     string true "用户 ID"
// @Success     0    {object} string
// @Failure     1000 {object} string
// @Failure     1001 {object} string
// @Failure     1002 {object} string
// @Failure     1003 {object} string
// @Router      /api/v1/users/{id}/enable [post]
func UsersEnable(c *gin.Context) {
	userid := c.Param("id")
	if userid == "" {
		m.JsonResponse(c, m.StatusParamsERR, "用户 ID 不能为空")
		return
	}

	id, err := strconv.Atoi(userid)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "用户 ID 格式错误")
		return
	}

	user := &model.User{}
	if err := db.DBClient.First(user, uint(id)).Error; err != nil {
		if db.RecordNotFound(err) {
			m.JsonResponse(c, m.StatusParamsERR, "用户不存在")
			return
		}
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	user.Status = 1
	if err := db.Save(db.DBClient, user); err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, "启用成功")
}

// @Summary     用户禁用接口
// @Description 禁用用户
// @Tags        users
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       id   path     string true "用户 ID"
// @Success     0    {object} string
// @Failure     1000 {object} string
// @Failure     1001 {object} string
// @Failure     1002 {object} string
// @Failure     1003 {object} string
// @Router      /api/v1/users/{id}/disable [post]
func UsersDisable(c *gin.Context) {
	userid := c.Param("id")
	if userid == "" {
		m.JsonResponse(c, m.StatusParamsERR, "用户 ID 不能为空")
		return
	}

	id, err := strconv.Atoi(userid)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "用户 ID 格式错误")
		return
	}

	user := &model.User{}
	if err := db.DBClient.First(user, uint(id)).Error; err != nil {
		if db.RecordNotFound(err) {
			m.JsonResponse(c, m.StatusParamsERR, "用户不存在")
			return
		}
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	user.Status = 0
	if err := db.Save(db.DBClient, user); err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, "禁用成功")
}

// @Summary     修改密码接口
// @Description 修改用户密码
// @Tags        users
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       id          path     string true "用户 ID"
// @Param       oldPassword formData string true "旧密码"
// @Param       newPassword formData string true "新密码"
// @Success     0           {object} string
// @Failure     1000        {object} string
// @Failure     1001        {object} string
// @Failure     1002        {object} string
// @Failure     1003        {object} string
// @Router      /api/v1/users/{id}/password [post]
func UsersChangePassword(c *gin.Context) {
	userid := c.Param("id")
	if userid == "" {
		m.JsonResponse(c, m.StatusParamsERR, "用户 ID 不能为空")
		return
	}

	id, err := strconv.Atoi(userid)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "用户 ID 格式错误")
		return
	}

	user := &model.User{}
	if err := db.DBClient.First(user, uint(id)).Error; err != nil {
		if db.RecordNotFound(err) {
			m.JsonResponse(c, m.StatusParamsERR, "用户不存在")
			return
		}
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	oldPassword := c.PostForm("oldPassword")
	if oldPassword == "" {
		m.JsonResponse(c, m.StatusParamsERR, "旧密码不能为空")
		return
	}

	newPassword := c.PostForm("newPassword")
	if newPassword == "" {
		m.JsonResponse(c, m.StatusParamsERR, "新密码不能为空")
		return
	}

	// 验证旧密码并生成新密码
	newEncrypted, success := utils.ChangePassword(oldPassword, user.Password, newPassword)
	if !success {
		m.JsonResponse(c, m.StatusParamsERR, "旧密码错误")
		return
	}

	user.Password = newEncrypted
	if err := db.Save(db.DBClient, user); err != nil {
		m.JsonResponse(c, m.StatusDBERR, err)
		return
	}

	m.JsonResponse(c, m.StatusSucc, "密码修改成功")
}
