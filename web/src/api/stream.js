import request from './request'

/**
 * 获取流列表
 */
export function getStreams(params) {
  return request({
    url: '/v1/streams',
    method: 'get',
    params
  })
}

/**
 * 开始播放（直播/回放）
 */
export function startStream(channelId, data) {
  return request({
    url: `/v1/channels/${channelId}/streams`,
    method: 'post',
    data
  })
}

/**
 * 停止播放
 */
export function stopStream(id) {
  return request({
    url: `/v1/streams/${id}`,
    method: 'delete'
  })
}
