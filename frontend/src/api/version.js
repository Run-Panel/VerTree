import request from './index'

// Get versions list
export function getVersions(params) {
  return request({
    url: '/versions',
    method: 'get',
    params
  })
}

// Get single version
export function getVersion(id) {
  return request({
    url: `/versions/${id}`,
    method: 'get'
  })
}

// Create version
export function createVersion(data) {
  return request({
    url: '/versions',
    method: 'post',
    data
  })
}

// Update version
export function updateVersion(id, data) {
  return request({
    url: `/versions/${id}`,
    method: 'put',
    data
  })
}

// Delete version
export function deleteVersion(id) {
  return request({
    url: `/versions/${id}`,
    method: 'delete'
  })
}

// Publish version
export function publishVersion(id) {
  return request({
    url: `/versions/${id}/publish`,
    method: 'post'
  })
}

// Unpublish version
export function unpublishVersion(id) {
  return request({
    url: `/versions/${id}/unpublish`,
    method: 'post'
  })
}
