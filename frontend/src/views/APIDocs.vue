<template>
  <div class="api-docs">
    <div class="page-header">
      <h1>{{ $t('docs.title', 'VerTree API文档') }}</h1>
      <p class="page-description">
        VerTree 提供了完整的应用版本管理 API，支持版本发布、更新检查、统计分析等功能。
      </p>
    </div>

    <el-card class="docs-card">
      <template #header>
        <div class="card-header">
          <h2>快速开始</h2>
        </div>
      </template>
      
      <div class="docs-content">
        <h3>API地址</h3>
        <ul>
          <li><strong>客户端API:</strong> <code>{{ clientApiBaseUrl }}</code></li>
          <li><strong>管理端API:</strong> <code>{{ adminApiBaseUrl }}</code></li>
        </ul>

        <h3>认证方式</h3>
        <div class="auth-section">
          <h4>🔹 客户端API认证</h4>
          <p>使用应用ID和API密钥进行认证：</p>
          <el-card class="code-card">
            <pre><code>Authorization: Bearer &lt;app_id&gt;:&lt;api_key&gt;</code></pre>
          </el-card>
          
          <h4>🔹 管理端API认证</h4>
          <p>使用JWT令牌进行认证：</p>
          <el-card class="code-card">
            <pre><code>Authorization: Bearer &lt;jwt_token&gt;</code></pre>
          </el-card>
        </div>

        <div class="security-note">
          <h4>🔒 安全隔离</h4>
          <p>每个应用只能访问自己的数据。API会根据认证信息自动过滤应用范围，确保数据安全。</p>
        </div>

        <h3>API Key获取方式</h3>
        <ol>
          <li>登录管理后台，进入 <router-link to="/applications">应用管理</router-link></li>
          <li>创建或选择一个应用（获得App ID）</li>
          <li>在应用详情页点击"密钥管理"或"API Keys"</li>
          <li>点击"创建新密钥"，设置密钥名称</li>
          <li><strong>重要：</strong>设置密钥权限：<code>check_update</code>, <code>download</code>, <code>install</code></li>
          <li>保存后，系统会生成密钥并显示完整的API密钥（仅显示一次）</li>
          <li>记录下App ID（如：app_1234567890）和API密钥（如：sk_test_abcdef123456）</li>
          <li>在客户端中使用格式：<code>Authorization: Bearer app_id:api_key</code></li>
        </ol>
      </div>
    </el-card>

    <el-card class="docs-card">
      <template #header>
        <div class="card-header">
          <h2>客户端API - 更新检查</h2>
        </div>
      </template>

      <el-collapse v-model="activeClientSections">
        <!-- 检查更新 API -->
        <el-collapse-item title="检查更新" name="check-update">
          <div class="api-section">
            <div class="api-header">
              <span class="method post">POST</span>
              <code class="endpoint">/api/v1/check-update</code>
              <span class="description">检查是否有新版本可用</span>
            </div>
            
            <h4>权限要求</h4>
            <p><code>check_update</code> - 在创建API密钥时需要勾选此权限</p>
            
            <div class="warning-note">
              <h4>⚠️ 重要提醒</h4>
              <p><strong>app_id参数是必须的！</strong>此参数用于标识应用身份，确保只能访问对应应用的版本信息。请确保请求中包含正确的app_id。</p>
            </div>
            
            <h4>请求参数</h4>
            <el-table :data="checkUpdateParams" class="params-table">
              <el-table-column prop="name" label="参数名" width="180" />
              <el-table-column prop="type" label="类型" width="100" />
              <el-table-column prop="required" label="必须" width="80">
                <template #default="{ row }">
                  <el-tag :type="row.required ? 'danger' : 'info'" size="small">
                    {{ row.required ? '是' : '否' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="description" label="说明" />
            </el-table>

            <h4>请求示例</h4>
            <el-card class="code-card">
              <pre><code>{{ checkUpdateExample }}</code></pre>
            </el-card>

            <h4>响应示例</h4>
            <el-card class="code-card">
              <pre><code>{{ checkUpdateResponse }}</code></pre>
            </el-card>
          </div>
        </el-collapse-item>

        <!-- 下载开始 API -->
        <el-collapse-item title="下载开始通知" name="download-started">
          <div class="api-section">
            <div class="api-header">
              <span class="method post">POST</span>
              <code class="endpoint">/api/v1/download-started</code>
              <span class="description">通知服务器开始下载更新</span>
            </div>
            
            <h4>权限要求</h4>
            <p><code>download</code> - 用于统计下载次数</p>
            
            <h4>请求参数</h4>
            <el-table :data="downloadStartedParams" class="params-table">
              <el-table-column prop="name" label="参数名" width="180" />
              <el-table-column prop="type" label="类型" width="100" />
              <el-table-column prop="required" label="必须" width="80">
                <template #default="{ row }">
                  <el-tag :type="row.required ? 'danger' : 'info'" size="small">
                    {{ row.required ? '是' : '否' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="description" label="说明" />
            </el-table>

            <h4>请求示例</h4>
            <el-card class="code-card">
              <pre><code>{{ downloadStartedExample }}</code></pre>
            </el-card>
          </div>
        </el-collapse-item>

        <!-- 安装结果 API -->
        <el-collapse-item title="安装结果报告" name="install-result">
          <div class="api-section">
            <div class="api-header">
              <span class="method post">POST</span>
              <code class="endpoint">/api/v1/install-result</code>
              <span class="description">报告更新安装结果</span>
            </div>
            
            <h4>权限要求</h4>
            <p><code>install</code> - 用于统计安装成功率</p>
            
            <h4>请求参数</h4>
            <el-table :data="installResultParams" class="params-table">
              <el-table-column prop="name" label="参数名" width="180" />
              <el-table-column prop="type" label="类型" width="100" />
              <el-table-column prop="required" label="必须" width="80">
                <template #default="{ row }">
                  <el-tag :type="row.required ? 'danger' : 'info'" size="small">
                    {{ row.required ? '是' : '否' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="description" label="说明" />
            </el-table>

            <h4>请求示例</h4>
            <el-card class="code-card">
              <pre><code>{{ installResultExample }}</code></pre>
            </el-card>
          </div>
        </el-collapse-item>

        <!-- 版本列表 API -->
        <el-collapse-item title="获取版本列表" name="get-versions">
          <div class="api-section">
            <div class="api-header">
              <span class="method get">GET</span>
              <code class="endpoint">/api/v1/versions</code>
              <span class="description">获取应用的版本历史列表</span>
            </div>
            
            <h4>权限要求</h4>
            <p><code>check_update</code> - 复用更新检查权限</p>
            
            <div class="warning-note">
              <h4>🎯 应用场景</h4>
              <p>此接口适用于：版本回退、历史展示、灵活更新策略、开发调试等场景。让客户端有更多版本选择权。</p>
            </div>
            
            <h4>查询参数</h4>
            <el-table :data="getVersionsParams" class="params-table">
              <el-table-column prop="name" label="参数名" width="180" />
              <el-table-column prop="type" label="类型" width="100" />
              <el-table-column prop="required" label="必须" width="80">
                <template #default="{ row }">
                  <el-tag :type="row.required ? 'danger' : 'info'" size="small">
                    {{ row.required ? '是' : '否' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="description" label="说明" />
            </el-table>

            <h4>请求示例</h4>
            <el-card class="code-card">
              <pre><code>{{ getVersionsExample }}</code></pre>
            </el-card>

            <h4>响应示例</h4>
            <el-card class="code-card">
              <pre><code>{{ getVersionsResponse }}</code></pre>
            </el-card>
          </div>
        </el-collapse-item>
      </el-collapse>
    </el-card>

    <el-card class="docs-card">
      <template #header>
        <div class="card-header">
          <h2>管理端API</h2>
        </div>
      </template>

      <el-collapse v-model="activeAdminSections">
        <!-- 版本管理 -->
        <el-collapse-item title="版本管理" name="versions">
          <div class="api-section">
            <h4>📋 获取版本列表</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method get">GET</span>
                <code class="endpoint">/admin/api/v1/versions</code>
              </div>
              <p><strong>查询参数:</strong> <code>channel</code>, <code>page</code>, <code>limit</code></p>
            </div>

            <h4>➕ 创建新版本</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method post">POST</span>
                <code class="endpoint">/admin/api/v1/versions</code>
              </div>
              <p><strong>请求体:</strong> JSON格式的版本信息</p>
            </div>

            <h4>📤 上传版本文件</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method post">POST</span>
                <code class="endpoint">/admin/api/v1/applications/:id/versions/upload</code>
                <span class="description">为指定应用上传版本文件</span>
              </div>
              <p><strong>路径参数:</strong> <code>id</code> - 应用ID</p>
              <p><strong>内容类型:</strong> multipart/form-data</p>
              <p><strong>文件字段:</strong> <code>file</code> (最大500MB)</p>
              <p><strong>表单字段:</strong> version, channel, title, description, release_notes, etc.</p>
              <p><strong>支持格式:</strong> .zip, .exe, .dmg, .pkg, .deb, .rpm, .tar.gz, .msi</p>
            </div>

            <h4>📝 更新版本</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method put">PUT</span>
                <code class="endpoint">/admin/api/v1/versions/:id</code>
              </div>
            </div>

            <h4>📝 更新版本（含文件）</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method put">PUT</span>
                <code class="endpoint">/admin/api/v1/applications/:id/versions/:version_id/upload</code>
                <span class="description">为指定应用更新版本和文件</span>
              </div>
              <p><strong>路径参数:</strong> <code>id</code> - 应用ID, <code>version_id</code> - 版本ID</p>
            </div>

            <h4>🚀 发布版本</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method post">POST</span>
                <code class="endpoint">/admin/api/v1/versions/:id/publish</code>
              </div>
            </div>

            <h4>⏸️ 取消发布</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method post">POST</span>
                <code class="endpoint">/admin/api/v1/versions/:id/unpublish</code>
              </div>
            </div>

            <h4>🗑️ 删除版本</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method delete">DELETE</span>
                <code class="endpoint">/admin/api/v1/versions/:id</code>
              </div>
            </div>
          </div>
        </el-collapse-item>

        <!-- 应用管理 -->
        <el-collapse-item title="应用管理" name="applications">
          <div class="api-section">
            <h4>📋 获取应用列表</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method get">GET</span>
                <code class="endpoint">/admin/api/v1/applications</code>
              </div>
            </div>

            <h4>➕ 创建应用</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method post">POST</span>
                <code class="endpoint">/admin/api/v1/applications</code>
              </div>
            </div>

            <h4>🔑 获取API密钥</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method get">GET</span>
                <code class="endpoint">/admin/api/v1/applications/:id/keys</code>
              </div>
            </div>

            <h4>🔑 创建API密钥</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method post">POST</span>
                <code class="endpoint">/admin/api/v1/applications/:id/keys</code>
              </div>
            </div>
          </div>
        </el-collapse-item>

        <!-- 通道管理 -->
        <el-collapse-item title="通道管理" name="channels">
          <div class="api-section">
            <h4>📋 获取通道列表</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method get">GET</span>
                <code class="endpoint">/admin/api/v1/channels</code>
              </div>
            </div>

            <h4>➕ 创建通道</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method post">POST</span>
                <code class="endpoint">/admin/api/v1/channels</code>
              </div>
            </div>

            <h4>📝 更新通道</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method put">PUT</span>
                <code class="endpoint">/admin/api/v1/channels/:id</code>
              </div>
            </div>
          </div>
        </el-collapse-item>

        <!-- 统计分析 -->
        <el-collapse-item title="统计分析" name="stats">
          <div class="api-section">
            <h4>📊 获取统计数据</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method get">GET</span>
                <code class="endpoint">/admin/api/v1/stats</code>
              </div>
            </div>

            <h4>📈 版本分布统计</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method get">GET</span>
                <code class="endpoint">/admin/api/v1/stats/distribution</code>
              </div>
            </div>

            <h4>🌍 地区分布统计</h4>
            <div class="admin-api-item">
              <div class="api-header">
                <span class="method get">GET</span>
                <code class="endpoint">/admin/api/v1/stats/regions</code>
              </div>
            </div>
          </div>
        </el-collapse-item>
      </el-collapse>
    </el-card>

    <el-card class="docs-card">
      <template #header>
        <div class="card-header">
          <h2>SDK 集成示例</h2>
        </div>
      </template>

      <el-tabs v-model="activeSDKTab" type="card">
        <el-tab-pane label="JavaScript" name="javascript">
          <el-card class="code-card">
            <pre><code>{{ javascriptSDK }}</code></pre>
          </el-card>
        </el-tab-pane>
        
        <el-tab-pane label="Python" name="python">
          <el-card class="code-card">
            <pre><code>{{ pythonSDK }}</code></pre>
          </el-card>
        </el-tab-pane>
        
        <el-tab-pane label="Go" name="go">
          <el-card class="code-card">
            <pre><code>{{ goSDK }}</code></pre>
          </el-card>
        </el-tab-pane>

        <el-tab-pane label="cURL" name="curl">
          <el-card class="code-card">
            <pre><code>{{ curlExample }}</code></pre>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <el-card class="docs-card">
      <template #header>
        <div class="card-header">
          <h2>错误码参考</h2>
        </div>
      </template>

      <el-table :data="errorCodes" class="error-table">
        <el-table-column prop="code" label="状态码" width="100" />
        <el-table-column prop="message" label="错误信息" width="200" />
        <el-table-column prop="description" label="说明" />
        <el-table-column prop="solution" label="解决方案" />
      </el-table>
    </el-card>

    <el-card class="docs-card">
      <template #header>
        <div class="card-header">
          <h2>最佳实践</h2>
        </div>
      </template>

      <el-collapse v-model="activeBestPractices">
        <el-collapse-item title="🕐 更新检查频率" name="check-frequency">
          <ul>
            <li><strong>推荐频率：</strong>应用启动时检查一次，避免频繁请求</li>
            <li><strong>后台检查：</strong>可以设置定时任务（如每日或每周检查一次）</li>
            <li><strong>用户触发：</strong>提供手动检查更新的功能</li>
            <li><strong>错误重试：</strong>网络错误时实施指数退避重试策略</li>
          </ul>
        </el-collapse-item>

        <el-collapse-item title="🏷️ 版本号管理" name="version-naming">
          <ul>
            <li><strong>语义化版本：</strong>推荐使用 v1.2.3 格式</li>
            <li><strong>主版本号：</strong>不兼容的API修改</li>
            <li><strong>次版本号：</strong>向下兼容的功能性新增</li>
            <li><strong>修订号：</strong>向下兼容的问题修正</li>
            <li><strong>预发布：</strong>使用 v1.2.3-beta.1 格式</li>
          </ul>
        </el-collapse-item>

        <el-collapse-item title="📁 文件管理" name="file-management">
          <ul>
            <li><strong>文件大小：</strong>单个文件最大500MB，建议压缩后上传</li>
            <li><strong>文件命名：</strong>使用清晰的命名规则，包含版本号</li>
            <li><strong>文件校验：</strong>系统自动计算SHA256校验和</li>
            <li><strong>增量更新：</strong>考虑实施增量更新减少下载量</li>
          </ul>
        </el-collapse-item>

        <el-collapse-item title="⚡ 性能优化" name="performance">
          <ul>
            <li><strong>CDN加速：</strong>使用CDN分发更新文件</li>
            <li><strong>并发控制：</strong>限制同时下载的文件数量</li>
            <li><strong>断点续传：</strong>支持大文件断点续传</li>
            <li><strong>缓存策略：</strong>合理设置HTTP缓存头</li>
          </ul>
        </el-collapse-item>

        <el-collapse-item title="🔒 安全考虑" name="security">
          <ul>
            <li><strong>API密钥保护：</strong>不要在客户端硬编码API密钥</li>
            <li><strong>权限最小化：</strong>API密钥只分配必要的权限</li>
            <li><strong>文件签名：</strong>验证下载文件的数字签名</li>
            <li><strong>HTTPS传输：</strong>始终使用HTTPS传输敏感数据</li>
          </ul>
        </el-collapse-item>
      </el-collapse>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const activeClientSections = ref(['check-update'])
const activeAdminSections = ref(['versions'])
const activeSDKTab = ref('javascript')
const activeBestPractices = ref(['check-frequency'])

// API 基础地址
const clientApiBaseUrl = computed(() => {
  return `${window.location.protocol}//${window.location.host}/api/v1`
})

const adminApiBaseUrl = computed(() => {
  return `${window.location.protocol}//${window.location.host}/admin/api/v1`
})

// 检查更新API参数 (修正后的参数)
const checkUpdateParams = [
  { name: 'app_id', type: 'string', required: true, description: '应用ID，从管理后台获取，如 "app_1234567890"' },
  { name: 'current_version', type: 'string', required: true, description: '当前版本号，如 "v1.2.3"' },
  { name: 'channel', type: 'string', required: true, description: '更新通道：stable, beta, alpha' },
  { name: 'client_id', type: 'string', required: true, description: '客户端唯一标识，用于统计' },
  { name: 'region', type: 'string', required: false, description: '地区代码，如 "CN", "US"' },
  { name: 'arch', type: 'string', required: false, description: '系统架构，如 "x64", "arm64"' },
  { name: 'os', type: 'string', required: false, description: '操作系统，如 "windows", "macos", "linux"' }
]

// 下载开始API参数 (修正后的参数)
const downloadStartedParams = [
  { name: 'version', type: 'string', required: true, description: '下载的版本号' },
  { name: 'client_id', type: 'string', required: true, description: '客户端唯一标识' }
]

// 安装结果API参数 (修正后的参数)
const installResultParams = [
  { name: 'version', type: 'string', required: true, description: '安装的版本号' },
  { name: 'client_id', type: 'string', required: true, description: '客户端唯一标识' },
  { name: 'success', type: 'boolean', required: true, description: '安装是否成功' },
  { name: 'error_message', type: 'string', required: false, description: '错误信息（当success为false时）' }
]

// 获取版本列表API参数
const getVersionsParams = [
  { name: 'channel', type: 'string', required: false, description: '通道过滤：stable, beta, alpha。不指定则返回所有通道' },
  { name: 'limit', type: 'number', required: false, description: '返回数量限制，默认10，最大50' },
  { name: 'published_only', type: 'boolean', required: false, description: '只返回已发布版本，默认true' }
]

// API示例 (修正后的示例)
const checkUpdateExample = `curl -X POST "${window.location.protocol}//${window.location.host}/api/v1/check-update" \\
  -H "Authorization: Bearer app_1234567890:sk_test_abcdef123456" \\
  -H "Content-Type: application/json" \\
  -d '{
    "app_id": "app_1234567890",
    "current_version": "v1.2.3",
    "channel": "stable",
    "client_id": "client_unique_id_12345",
    "region": "CN",
    "arch": "x64",
    "os": "windows"
  }'`

const checkUpdateResponse = `{
  "code": 200,
  "message": "success",
  "data": {
    "has_update": true,
    "latest_version": "v1.3.0",
    "download_url": "${window.location.protocol}//${window.location.host}/uploads/versions/app_v1.3.0.zip",
    "file_size": 52428800,
    "file_checksum": "sha256:abc123def456...",
    "is_forced": false,
    "title": "版本 v1.3.0",
    "description": "修复已知问题，提升性能",
    "release_notes": "新增功能:\\n- 支持自动更新\\n- 修复已知问题",
    "min_upgrade_version": "v1.0.0"
  }
}`

const downloadStartedExample = `curl -X POST "${window.location.protocol}//${window.location.host}/api/v1/download-started" \\
  -H "Authorization: Bearer app_1234567890:sk_test_abcdef123456" \\
  -H "Content-Type: application/json" \\
  -d '{
    "version": "v1.3.0",
    "client_id": "client_unique_id_12345"
  }'`

const installResultExample = `curl -X POST "${window.location.protocol}//${window.location.host}/api/v1/install-result" \\
  -H "Authorization: Bearer app_1234567890:sk_test_abcdef123456" \\
  -H "Content-Type: application/json" \\
  -d '{
    "version": "v1.3.0",
    "client_id": "client_unique_id_12345",
    "success": true
  }'`

const getVersionsExample = `# 获取所有已发布版本（默认）
curl -X GET "${window.location.protocol}//${window.location.host}/api/v1/versions" \\
  -H "Authorization: Bearer app_1234567890:sk_test_abcdef123456"

# 获取stable通道的最近5个版本
curl -X GET "${window.location.protocol}//${window.location.host}/api/v1/versions?channel=stable&limit=5" \\
  -H "Authorization: Bearer app_1234567890:sk_test_abcdef123456"

# 获取所有版本（包括未发布）
curl -X GET "${window.location.protocol}//${window.location.host}/api/v1/versions?published_only=false&limit=20" \\
  -H "Authorization: Bearer app_1234567890:sk_test_abcdef123456"`

const getVersionsResponse = `{
  "code": 200,
  "message": "success",
  "data": [
    {
      "version": "v1.3.0",
      "channel": "stable",
      "title": "稳定版本 v1.3.0",
      "description": "修复重要问题，提升性能",
      "release_notes": "新增功能:\\n- 支持自动更新\\n- 修复已知问题\\n- 性能优化",
      "download_url": "${window.location.protocol}//${window.location.host}/uploads/versions/app_v1.3.0.zip",
      "file_size": 52428800,
      "file_checksum": "sha256:abc123def456...",
      "is_forced": false,
      "min_upgrade_version": "v1.0.0",
      "published_at": "2024-01-15T10:30:00Z"
    },
    {
      "version": "v1.2.5",
      "channel": "stable",
      "title": "稳定版本 v1.2.5",
      "description": "安全更新和bug修复",
      "release_notes": "修复:\\n- 安全漏洞修复\\n- 稳定性提升",
      "download_url": "${window.location.protocol}//${window.location.host}/uploads/versions/app_v1.2.5.zip",
      "file_size": 51200000,
      "file_checksum": "sha256:def789abc123...",
      "is_forced": false,
      "min_upgrade_version": "v1.0.0",
      "published_at": "2024-01-10T14:20:00Z"
    }
  ]
}`

// SDK 示例
const javascriptSDK = `// VerTree JavaScript SDK 示例
class VerTreeClient {
  constructor(appId, apiKey, baseUrl = '${window.location.protocol}//${window.location.host}/api/v1') {
    this.appId = appId;
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
    this.authHeader = \`Bearer \${appId}:\${apiKey}\`;
  }

  async checkUpdate(currentVersion, channel = 'stable', options = {}) {
    const response = await fetch(\`\${this.baseUrl}/check-update\`, {
      method: 'POST',
      headers: {
        'Authorization': this.authHeader,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        app_id: this.appId,
        current_version: currentVersion,
        channel: channel,
        client_id: options.clientId || this.generateClientId(),
        region: options.region,
        arch: options.arch,
        os: options.os
      })
    });
    
    const result = await response.json();
    return result.data;
  }

  async reportDownloadStarted(version, clientId) {
    await fetch(\`\${this.baseUrl}/download-started\`, {
      method: 'POST',
      headers: {
        'Authorization': this.authHeader,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        version: version,
        client_id: clientId
      })
    });
  }

  async reportInstallResult(version, success, errorMessage = null, clientId) {
    await fetch(\`\${this.baseUrl}/install-result\`, {
      method: 'POST',
      headers: {
        'Authorization': this.authHeader,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        version: version,
        client_id: clientId,
        success: success,
        error_message: errorMessage
      })
    });
  }

  async getVersions(options = {}) {
    const params = new URLSearchParams();
    if (options.channel) params.append('channel', options.channel);
    if (options.limit) params.append('limit', options.limit.toString());
    if (options.publishedOnly !== undefined) params.append('published_only', options.publishedOnly.toString());
    
    const url = \`\${this.baseUrl}/versions\` + (params.toString() ? \`?\${params.toString()}\` : '');
    
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Authorization': this.authHeader
      }
    });
    
    const result = await response.json();
    return result.data;
  }

  generateClientId() {
    return 'client_' + Math.random().toString(36).substr(2, 9);
  }
}

// 使用示例
const client = new VerTreeClient('app_1234567890', 'sk_test_abcdef123456');

async function checkForUpdates() {
  try {
    const clientId = client.generateClientId();
    const updateInfo = await client.checkUpdate('v1.2.3', 'stable', {
      clientId: clientId,
      region: 'CN',
      os: 'windows'
    });
    
    if (updateInfo.has_update) {
      console.log('有新版本可用:', updateInfo.latest_version);
      
      // 开始下载
      await client.reportDownloadStarted(updateInfo.latest_version, clientId);
      
      // 模拟下载和安装过程...
      
      // 报告安装结果
      await client.reportInstallResult(updateInfo.latest_version, true, null, clientId);
    }
  } catch (error) {
    console.error('检查更新失败:', error);
  }
}

async function showVersionHistory() {
  try {
    // 获取stable通道的最近10个版本
    const versions = await client.getVersions({
      channel: 'stable',
      limit: 10,
      publishedOnly: true
    });
    
    console.log('版本历史:', versions);
    
    // 展示版本列表供用户选择
    versions.forEach(version => {
      console.log(\`\${version.version} - \${version.title}\`);
      console.log(\`  发布时间: \${version.published_at}\`);
      console.log(\`  文件大小: \${(version.file_size / 1024 / 1024).toFixed(2)} MB\`);
    });
  } catch (error) {
    console.error('获取版本历史失败:', error);
  }
}`

const pythonSDK = `# VerTree Python SDK 示例
import requests
import json
import uuid

class VerTreeClient:
    def __init__(self, app_id, api_key, base_url="${window.location.protocol}//${window.location.host}/api/v1"):
        self.app_id = app_id
        self.api_key = api_key
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {app_id}:{api_key}',
            'Content-Type': 'application/json'
        }
    
    def check_update(self, current_version, channel='stable', client_id=None, **kwargs):
        """检查更新"""
        if client_id is None:
            client_id = str(uuid.uuid4())
            
        data = {
            'app_id': self.app_id,
            'current_version': current_version,
            'channel': channel,
            'client_id': client_id,
            **kwargs  # region, arch, os 等可选参数
        }
        
        response = requests.post(
            f'{self.base_url}/check-update',
            headers=self.headers,
            json=data
        )
        
        result = response.json()
        return result.get('data')
    
    def report_download_started(self, version, client_id):
        """报告下载开始"""
        data = {
            'version': version,
            'client_id': client_id
        }
        
        requests.post(
            f'{self.base_url}/download-started',
            headers=self.headers,
            json=data
        )
    
    def report_install_result(self, version, success, error_message=None, client_id=None):
        """报告安装结果"""
        data = {
            'version': version,
            'client_id': client_id,
            'success': success,
            'error_message': error_message
        }
        
        requests.post(
            f'{self.base_url}/install-result',
            headers=self.headers,
            json=data
        )

# 使用示例
client = VerTreeClient('app_1234567890', 'sk_test_abcdef123456')

def check_for_updates():
    try:
        client_id = str(uuid.uuid4())
        update_info = client.check_update(
            'v1.2.3', 
            'stable', 
            client_id=client_id,
            region='CN',
            os='linux'
        )
        
        if update_info and update_info.get('has_update'):
            print(f"有新版本可用: {update_info['latest_version']}")
            
            # 开始下载
            client.report_download_started(update_info['latest_version'], client_id)
            
            # 模拟下载和安装过程...
            
            # 报告安装结果
            client.report_install_result(update_info['latest_version'], True, None, client_id)
            
    except Exception as e:
        print(f"检查更新失败: {e}")

if __name__ == "__main__":
    check_for_updates()`

const goSDK = `// VerTree Go SDK 示例
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "github.com/google/uuid"
)

type VerTreeClient struct {
    AppID   string
    APIKey  string
    BaseURL string
    Client  *http.Client
}

type UpdateRequest struct {
    AppID          string \`json:"app_id"\`
    CurrentVersion string \`json:"current_version"\`
    Channel        string \`json:"channel"\`
    ClientID       string \`json:"client_id"\`
    Region         string \`json:"region,omitempty"\`
    Arch           string \`json:"arch,omitempty"\`
    OS             string \`json:"os,omitempty"\`
}

type UpdateResponse struct {
    HasUpdate         bool   \`json:"has_update"\`
    LatestVersion     string \`json:"latest_version"\`
    DownloadURL       string \`json:"download_url"\`
    FileSize          int64  \`json:"file_size"\`
    FileChecksum      string \`json:"file_checksum"\`
    IsForced          bool   \`json:"is_forced"\`
    Title             string \`json:"title"\`
    Description       string \`json:"description"\`
    ReleaseNotes      string \`json:"release_notes"\`
    MinUpgradeVersion string \`json:"min_upgrade_version"\`
}

type APIResponse struct {
    Code    int         \`json:"code"\`
    Message string      \`json:"message"\`
    Data    interface{} \`json:"data"\`
}

func NewVerTreeClient(appID, apiKey string) *VerTreeClient {
    return &VerTreeClient{
        AppID:   appID,
        APIKey:  apiKey,
        BaseURL: "${window.location.protocol}//${window.location.host}/api/v1",
        Client:  &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *VerTreeClient) CheckUpdate(req UpdateRequest) (*UpdateResponse, error) {
    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    httpReq, err := http.NewRequest("POST", c.BaseURL+"/check-update", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s:%s", c.AppID, c.APIKey))
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := c.Client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var apiResp APIResponse
    err = json.NewDecoder(resp.Body).Decode(&apiResp)
    if err != nil {
        return nil, err
    }

    updateData, _ := json.Marshal(apiResp.Data)
    var updateResp UpdateResponse
    json.Unmarshal(updateData, &updateResp)

    return &updateResp, nil
}

func (c *VerTreeClient) ReportDownloadStarted(version, clientID string) error {
    data := map[string]string{
        "version":   version,
        "client_id": clientID,
    }

    jsonData, _ := json.Marshal(data)

    httpReq, err := http.NewRequest("POST", c.BaseURL+"/download-started", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }

    httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s:%s", c.AppID, c.APIKey))
    httpReq.Header.Set("Content-Type", "application/json")

    _, err = c.Client.Do(httpReq)
    return err
}

func (c *VerTreeClient) ReportInstallResult(version, clientID string, success bool, errorMessage string) error {
    data := map[string]interface{}{
        "version":       version,
        "client_id":     clientID,
        "success":       success,
        "error_message": errorMessage,
    }

    jsonData, _ := json.Marshal(data)

    httpReq, err := http.NewRequest("POST", c.BaseURL+"/install-result", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }

    httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s:%s", c.AppID, c.APIKey))
    httpReq.Header.Set("Content-Type", "application/json")

    _, err = c.Client.Do(httpReq)
    return err
}

func main() {
    client := NewVerTreeClient("app_1234567890", "sk_test_abcdef123456")
    clientID := uuid.New().String()

    updateInfo, err := client.CheckUpdate(UpdateRequest{
        AppID:          "app_1234567890",
        CurrentVersion: "v1.2.3",
        Channel:        "stable",
        ClientID:       clientID,
        Region:         "CN",
        OS:             "linux",
    })
    
    if err != nil {
        fmt.Printf("检查更新失败: %v\\n", err)
        return
    }

    if updateInfo.HasUpdate {
        fmt.Printf("有新版本可用: %s\\n", updateInfo.LatestVersion)

        // 报告下载开始
        client.ReportDownloadStarted(updateInfo.LatestVersion, clientID)

        // 模拟下载和安装过程...

        // 报告安装结果
        client.ReportInstallResult(updateInfo.LatestVersion, clientID, true, "")
    }
}`

const curlExample = `# 检查更新
curl -X POST "${window.location.protocol}//${window.location.host}/api/v1/check-update" \\
  -H "Authorization: Bearer app_1234567890:sk_test_abcdef123456" \\
  -H "Content-Type: application/json" \\
  -d '{
    "app_id": "app_1234567890",
    "current_version": "v1.2.3",
    "channel": "stable",
    "client_id": "client_unique_id_12345",
    "region": "CN",
    "os": "linux"
  }'

# 获取版本列表
curl -X GET "${window.location.protocol}//${window.location.host}/api/v1/versions?channel=stable&limit=5" \\
  -H "Authorization: Bearer app_1234567890:sk_test_abcdef123456"

# 报告下载开始
curl -X POST "${window.location.protocol}//${window.location.host}/api/v1/download-started" \\
  -H "Authorization: Bearer app_1234567890:sk_test_abcdef123456" \\
  -H "Content-Type: application/json" \\
  -d '{
    "version": "v1.3.0",
    "client_id": "client_unique_id_12345"
  }'

# 报告安装结果
curl -X POST "${window.location.protocol}//${window.location.host}/api/v1/install-result" \\
  -H "Authorization: Bearer app_1234567890:sk_test_abcdef123456" \\
  -H "Content-Type: application/json" \\
  -d '{
    "version": "v1.3.0",
    "client_id": "client_unique_id_12345",
    "success": true
  }'

# 创建版本（管理API）
curl -X POST "${window.location.protocol}//${window.location.host}/admin/api/v1/versions" \\
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{
    "app_id": "app_1234567890",
    "version": "v1.4.0",
    "channel": "stable",
    "title": "新版本发布",
    "description": "修复重要问题",
    "release_notes": "详细的更新说明...",
    "file_url": "https://releases.example.com/v1.4.0/app.zip",
    "file_size": 52428800,
    "file_checksum": "sha256:abc123...",
    "is_forced": false
  }'

# 上传版本文件（管理API）
curl -X POST "${window.location.protocol}//${window.location.host}/admin/api/v1/applications/app_1234567890/versions/upload" \\
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \\
  -F "file=@app_v1.4.0.zip" \\
  -F "version=v1.4.0" \\
  -F "channel=stable" \\
  -F "title=新版本发布" \\
  -F "description=修复重要问题"`

// 错误码表
const errorCodes = [
  { code: '200', message: 'OK', description: '请求成功', solution: '正常响应' },
  { code: '400', message: 'Bad Request', description: '请求参数错误', solution: '检查请求参数格式和必须字段' },
  { code: '401', message: 'Unauthorized', description: 'API Key 无效或缺失', solution: '检查 Authorization 头部和 API Key 格式' },
  { code: '403', message: 'Forbidden', description: 'API Key 权限不足', solution: '确认 API Key 拥有相应权限（check_update, download, install）' },
  { code: '404', message: 'Not Found', description: '请求的资源不存在', solution: '检查 API 路径和应用ID' },
  { code: '429', message: 'Too Many Requests', description: '请求频率过高', solution: '降低请求频率，实施重试策略' },
  { code: '500', message: 'Internal Server Error', description: '服务器内部错误', solution: '稍后重试，如持续出现请联系技术支持' }
]
</script>

<style scoped>
.api-docs {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.page-header {
  margin-bottom: 30px;
}

.page-header h1 {
  margin: 0 0 10px 0;
  color: #2c3e50;
}

.page-description {
  color: #7f8c8d;
  font-size: 16px;
  margin: 0;
}

.docs-card {
  margin-bottom: 24px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.card-header h2 {
  margin: 0;
  color: #2c3e50;
}

.docs-content h3 {
  color: #34495e;
  margin-top: 24px;
  margin-bottom: 12px;
}

.docs-content ul,
.docs-content ol {
  padding-left: 20px;
}

.docs-content li {
  margin-bottom: 8px;
}

.auth-section {
  margin: 20px 0;
}

.auth-section h4 {
  color: #2c3e50;
  margin: 16px 0 8px 0;
}

.security-note {
  background: #f0f9ff;
  border: 1px solid #0ea5e9;
  border-radius: 8px;
  padding: 16px;
  margin: 20px 0;
}

.security-note h4 {
  color: #0369a1;
  margin: 0 0 8px 0;
}

.warning-note {
  background: #fef3c7;
  border: 1px solid #f59e0b;
  border-radius: 8px;
  padding: 16px;
  margin: 16px 0;
}

.warning-note h4 {
  color: #d97706;
  margin: 0 0 8px 0;
}

.warning-note p {
  margin: 0;
  color: #92400e;
}

.code-card {
  margin: 16px 0;
  background-color: #f8f9fa;
}

.code-card pre {
  margin: 0;
  padding: 16px;
  background: #2d3748;
  color: #e2e8f0;
  border-radius: 6px;
  overflow-x: auto;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 14px;
  line-height: 1.5;
}

.api-section {
  padding: 16px 0;
}

.api-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
}

.method {
  padding: 4px 8px;
  border-radius: 4px;
  font-weight: bold;
  font-size: 12px;
  color: white;
  min-width: 60px;
  text-align: center;
}

.method.get {
  background-color: #52c41a;
}

.method.post {
  background-color: #1890ff;
}

.method.put {
  background-color: #faad14;
}

.method.delete {
  background-color: #ff4d4f;
}

.endpoint {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  background: #e6f7ff;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 14px;
}

.description {
  color: #666;
  font-size: 14px;
}

.params-table {
  margin: 16px 0;
}

.admin-api-item {
  margin-bottom: 16px;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.admin-api-item:last-child {
  border-bottom: none;
}

.admin-api-item p {
  margin: 8px 0 0 0;
  color: #666;
  font-size: 14px;
}

.error-table {
  margin-top: 16px;
}

.docs-content h4 {
  color: #2c3e50;
  margin: 20px 0 12px 0;
  font-size: 16px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .api-docs {
    padding: 10px;
  }
  
  .page-header h1 {
    font-size: 24px;
  }
  
  .page-description {
    font-size: 14px;
  }
  
  .api-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .method {
    min-width: auto;
  }
  
  .endpoint {
    font-size: 12px;
    word-break: break-all;
  }
  
  .code-card pre {
    font-size: 12px;
    padding: 12px;
  }
}

@media (max-width: 480px) {
  .mobile-hidden {
    display: none;
  }
}
</style>