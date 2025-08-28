import request from './index'

// Application management APIs

// Get all applications
export function getApplications(params = {}) {
  return request({
    url: '/applications',
    method: 'get',
    params
  })
}

// Get single application
export function getApplication(id) {
  return request({
    url: `/applications/${id}`,
    method: 'get'
  })
}

// Create application
export function createApplication(data) {
  return request({
    url: '/applications',
    method: 'post',
    data
  })
}

// Update application
export function updateApplication(id, data) {
  return request({
    url: `/applications/${id}`,
    method: 'put',
    data
  })
}

// Delete application
export function deleteApplication(id) {
  return request({
    url: `/applications/${id}`,
    method: 'delete'
  })
}

// Application Keys management

// Get application keys
export function getApplicationKeys(appId) {
  return request({
    url: `/applications/${appId}/keys`,
    method: 'get'
  })
}

// Create application key
export function createApplicationKey(appId, data) {
  return request({
    url: `/applications/${appId}/keys`,
    method: 'post',
    data
  })
}

// Update application key
export function updateApplicationKey(appId, keyId, data) {
  return request({
    url: `/applications/${appId}/keys/${keyId}`,
    method: 'put',
    data
  })
}

// Delete application key
export function deleteApplicationKey(appId, keyId) {
  return request({
    url: `/applications/${appId}/keys/${keyId}`,
    method: 'delete'
  })
}

// Get API documentation
export function getAPIDocs() {
  return request({
    url: '/docs',
    method: 'get'
  })
}
