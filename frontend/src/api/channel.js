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

// Get enabled channels by app
export function getChannelsByApp(appId) {
  return request({
    url: `/applications/${appId}/channels`,
    method: 'get'
  })
}

// Get all channels (enabled and disabled) for app
export function getAllChannelsForApp(appId) {
  return request({
    url: `/applications/${appId}/channels/all`,
    method: 'get'
  })
}

// Enable/configure channel for app
export function enableChannelForApp(appId, channelName, data) {
  return request({
    url: `/applications/${appId}/channels/${channelName}`,
    method: 'put',
    data
  })
}

// Disable channel for app
export function disableChannelForApp(appId, channelName) {
  return request({
    url: `/applications/${appId}/channels/${channelName}`,
    method: 'delete'
  })
}
