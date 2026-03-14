import request from './request'

/**
 * 获取设备列表
 */
export function getDevices(params) {
  return request({
    url: '/devices',
    method: 'get',
    params
  })
}

/**
 * 创建设备
 */
export function createDevice(data) {
  return request({
    url: '/devices/create',
    method: 'post',
    data
  })
}

/**
 * 删除设备
 */
export function deleteDevice(id) {
  return request({
    url: `/devices/${id}`,
    method: 'delete'
  })
}

/**
 * PTZ 控制
 */
export function ptzControl(data) {
  return request({
    url: '/devices/ptz',
    method: 'post',
    data
  })
}
