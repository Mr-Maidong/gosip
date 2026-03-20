import request from './request'

/**
 * 用户登录
 * @param {Object} data - 登录信息
 * @param {string} data.username - 用户名
 * @param {string} data.password - 密码
 * @returns {Promise}
 */
export function login(data) {
  return request({
    url: '/v1/login',
    method: 'post',
    data,
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    }
  })
}

/**
 * 用户登出
 * @returns {Promise}
 */
export function logout() {
  return request({
    url: '/v1/logout',
    method: 'post'
  })
}

/**
 * 获取当前用户信息
 * @returns {Promise}
 */
export function getCurrentUser() {
  return request({
    url: '/v1/users/current',
    method: 'get'
  })
}

/**
 * 获取用户列表
 * @param {Object} params - 查询参数
 * @returns {Promise}
 */
export function getUserList(params) {
  return request({
    url: '/v1/users',
    method: 'get',
    params
  })
}

/**
 * 创建用户
 * @param {Object} data - 用户信息
 * @returns {Promise}
 */
export function createUser(data) {
  return request({
    url: '/v1/users/create',
    method: 'post',
    data
  })
}

/**
 * 更新用户
 * @param {number} id - 用户 ID
 * @param {Object} data - 用户信息
 * @returns {Promise}
 */
export function updateUser(id, data) {
  return request({
    url: `/v1/users/${id}`,
    method: 'post',
    data
  })
}

/**
 * 删除用户
 * @param {number} id - 用户 ID
 * @returns {Promise}
 */
export function deleteUser(id) {
  return request({
    url: `/v1/users/${id}`,
    method: 'delete'
  })
}

/**
 * 启用用户
 * @param {number} id - 用户 ID
 * @returns {Promise}
 */
export function enableUser(id) {
  return request({
    url: `/v1/users/${id}/enable`,
    method: 'post'
  })
}

/**
 * 禁用用户
 * @param {number} id - 用户 ID
 * @returns {Promise}
 */
export function disableUser(id) {
  return request({
    url: `/v1/users/${id}/disable`,
    method: 'post'
  })
}

/**
 * 修改用户密码
 * @param {number} id - 用户 ID
 * @param {Object} data - 密码信息
 * @returns {Promise}
 */
export function changePassword(id, data) {
  return request({
    url: `/v1/users/${id}/password`,
    method: 'post',
    data
  })
}
