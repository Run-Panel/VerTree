import request from './index'

// Get statistics
export function getStats(params) {
  return request({
    url: '/stats',
    method: 'get',
    params
  })
}
