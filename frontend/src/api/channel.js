import request from './index'

// Get channels list
export function getChannels() {
  return request({
    url: '/channels',
    method: 'get'
  })
}

// Get single channel
export function getChannel(id) {
  return request({
    url: `/channels/${id}`,
    method: 'get'
  })
}

// Create channel
export function createChannel(data) {
  return request({
    url: '/channels',
    method: 'post',
    data
  })
}

// Update channel
export function updateChannel(id, data) {
  return request({
    url: `/channels/${id}`,
    method: 'put',
    data
  })
}

// Delete channel
export function deleteChannel(id) {
  return request({
    url: `/channels/${id}`,
    method: 'delete'
  })
}
