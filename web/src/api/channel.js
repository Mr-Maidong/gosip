import request from './request'

/**
 * 获取通道列表
 */
export function getChannels(params) {
  return request({
    url: '/channels',
    method: 'get',
    params
  })
}

/**
 * 创建通道
 */
export function createChannel(deviceId, data) {
  return request({
    url: `/devices/${deviceId}/channels`,
    method: 'post',
    data
  })
}

/**
 * 获取录像列表
 */
export function getRecords(channelId, params) {
  return request({
    url: `/channels/${channelId}/records`,
    method: 'get',
    params
  })
}
