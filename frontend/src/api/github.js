import request from './index'

// GitHub Repository management APIs

// Get GitHub repositories for an application
export function getGitHubRepositories(appId) {
  return request({
    url: `/github/repositories`,
    method: 'get',
    params: { app_id: appId }
  })
}

// Get all GitHub repositories (for admin)
export function getAllGitHubRepositories(params = {}) {
  return request({
    url: `/github/repositories`,
    method: 'get',
    params
  })
}

// Get single GitHub repository
export function getGitHubRepository(id) {
  return request({
    url: `/github/repositories/${id}`,
    method: 'get'
  })
}

// Create GitHub repository binding
export function createGitHubRepository(data) {
  return request({
    url: `/github/repositories`,
    method: 'post',
    data
  })
}

// Update GitHub repository binding
export function updateGitHubRepository(id, data) {
  return request({
    url: `/github/repositories/${id}`,
    method: 'put',
    data
  })
}

// Delete GitHub repository binding
export function deleteGitHubRepository(id) {
  return request({
    url: `/github/repositories/${id}`,
    method: 'delete'
  })
}

// Manually sync GitHub repository
export function syncGitHubRepository(id, data = {}) {
  return request({
    url: `/github/repositories/${id}/sync`,
    method: 'post',
    data
  })
}

// Get GitHub repository releases
export function getGitHubReleases(repositoryId, params = {}) {
  return request({
    url: `/github/repositories/${repositoryId}/releases`,
    method: 'get',
    params
  })
}

// Get GitHub repository sync status
export function getGitHubSyncStatus(repositoryId) {
  return request({
    url: `/github/repositories/${repositoryId}/sync-status`,
    method: 'get'
  })
}

// Validate GitHub repository URL
export function validateGitHubRepository(data) {
  return request({
    url: `/github/repositories/validate`,
    method: 'post',
    data
  })
}

// Test GitHub access token
export function testGitHubToken(data) {
  return request({
    url: `/github/test-token`,
    method: 'post',
    data
  })
}

// Get GitHub repository statistics
export function getGitHubStats(repositoryId) {
  return request({
    url: `/github/repositories/${repositoryId}/stats`,
    method: 'get'
  })
}

// GitHub Apps APIs

// Get GitHub App installations
export function getGitHubAppInstallations(data) {
  return request({
    url: `/github/app/installations`,
    method: 'post',
    data
  })
}

// Test GitHub App connection
export function testGitHubApp(data) {
  return request({
    url: `/github/app/test`,
    method: 'post',
    data
  })
}