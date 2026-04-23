import request from './request'

/**
 * 获取设备列表
 */
export function getDevices(params) {
  return request({
    url: '/v1/devices',
    method: 'get',
    params
  })
}

/**
 * 创建设备
 */
export function createDevice(data) {
  return request({
    url: '/v1/devices/create',
    method: 'post',
    data
  })
}

/**
 * 删除设备
 */
export function deleteDevice(id) {
  return request({
    url: `/v1/devices/${id}`,
    method: 'delete'
  })
}

/**
 * PTZ 控制
 */
export function ptzControl(data) {
  return request({
    url: '/v1/devices/ptz',
    method: 'post',
    data
  })
}

/**
 * 同步设备通道
 */
export function syncChannels(id) {
  return request({
    url: `/v1/devices/${id}/channels_sync`,
    method: 'post'
  })
}

/**
 * 获取设备上下线事件
 */
export function getDeviceEvents(id, params) {
  return request({
    url: `/v1/devices/${id}/events`,
    method: 'get',
    params
  })
}

/**
 * 获取设备GPS轨迹
 */
export function getDevicePositions(id, params) {
  return request({
    url: `/v1/devices/${id}/positions`,
    method: 'get',
    params
  })
}

/**
 * 获取设备告警记录
 */
export function getDeviceAlarms(id, params) {
  return request({
    url: `/v1/devices/${id}/alarms`,
    method: 'get',
    params
  })
}
