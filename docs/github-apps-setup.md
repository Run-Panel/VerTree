# GitHub Apps 设置指南

当您尝试连接GitHub组织的仓库时，可能会遇到以下错误：

> **GitHub说，由于我这是组织，所以只允许GitHub Apps**

这是因为GitHub组织出于安全考虑，限制了Personal Access Token的使用，要求使用GitHub Apps进行认证。

## 什么是GitHub Apps？

GitHub Apps是一种更安全、更精细的认证方式，相比Personal Access Token有以下优势：

- ✅ **更细粒度的权限控制**：只授予必要的权限
- ✅ **更好的安全性**：基于JWT和短期访问令牌
- ✅ **组织支持**：专为组织设计，受到组织管理员信任
- ✅ **审计跟踪**：更好的操作日志和审计功能

## 第一步：创建GitHub App

### 1. 访问GitHub Apps设置页面

以组织管理员身份访问：
```
https://github.com/organizations/YOUR_ORG_NAME/settings/apps
```

或者个人账户：
```
https://github.com/settings/apps
```

### 2. 点击 "New GitHub App"

### 3. 填写基本信息

- **GitHub App name**: `VerTree-YourOrgName` （必须全局唯一）
- **Description**: `VerTree version management integration`
- **Homepage URL**: `https://your-domain.com` （您的VerTree部署地址）

### 4. 配置权限 (Permissions)

在 **Repository permissions** 部分设置：

| 权限 | 级别 | 说明 |
|------|------|------|
| Contents | Read | 读取仓库内容和releases |
| Metadata | Read | 读取仓库基本信息 |
| Webhooks | Write | 创建和管理webhooks |

### 5. 配置事件订阅 (Subscribe to events)

勾选以下事件：
- ✅ **Release** - 监听release发布事件

### 6. 设置安装范围

- **Where can this GitHub App be installed?**
  - 选择 "Only on this account" （推荐）

### 7. 创建应用

点击 **Create GitHub App** 完成创建。

## 第二步：获取配置信息

创建成功后，您需要收集以下信息：

### 1. 获取 App ID

在应用详情页面，找到 **App ID**，例如：`123456`

### 2. 生成私钥

1. 滚动到 **Private keys** 部分
2. 点击 **Generate a private key**
3. 下载 `.pem` 文件并安全保存
4. 用文本编辑器打开，复制完整内容（包括 `-----BEGIN RSA PRIVATE KEY-----` 和 `-----END RSA PRIVATE KEY-----`）

### 3. 安装应用到仓库

1. 点击 **Install App**
2. 选择要安装的组织或个人账户
3. 选择仓库访问权限：
   - **All repositories** - 所有仓库
   - **Only select repositories** - 指定仓库（推荐）
4. 点击 **Install**

### 4. 获取 Installation ID

安装完成后，浏览器地址栏会显示类似：
```
https://github.com/settings/installations/12345678
```

其中 `12345678` 就是您的 **Installation ID**。

## 第三步：在VerTree中配置

### 1. 访问VerTree管理界面

打开：`http://your-vertree-domain:8080/admin-ui`

### 2. 添加GitHub仓库绑定

1. 进入 **GitHub管理** 页面
2. 点击 **添加仓库**
3. 在认证类型中选择 **GitHub App**

### 3. 填写配置信息

- **仓库URL**: `https://github.com/YOUR_ORG/YOUR_REPO`
- **认证类型**: 选择 `GitHub App`
- **GitHub App ID**: 输入第二步获取的App ID
- **Installation ID**: 输入第二步获取的Installation ID  
- **私钥**: 粘贴完整的私钥内容

### 4. 测试连接

点击 **验证仓库** 按钮，系统会：
- ✅ 验证GitHub App凭据
- ✅ 测试仓库访问权限
- ✅ 获取仓库基本信息和最新release

## 故障排除

### 问题1：提示 "Invalid GitHub App credentials"

**解决方案**：
- 检查App ID是否正确
- 确认私钥格式完整（包含开始和结束标记）
- 验证私钥没有多余的空格或换行

### 问题2：提示 "No installation found for repository"

**解决方案**：
- 确认GitHub App已安装到目标仓库/组织
- 检查Installation ID是否正确
- 验证仓库URL格式是否正确

### 问题3：提示 "insufficient permissions"

**解决方案**：
- 检查GitHub App权限设置
- 确认已授予 Contents(Read) 和 Metadata(Read) 权限
- 如需webhook功能，确认已授予 Webhooks(Write) 权限

### 问题4：如何找到Installation ID？

**方法1 - 从安装URL获取**：
1. 访问 `https://github.com/settings/installations`
2. 点击您的GitHub App
3. 地址栏显示的数字即为Installation ID

**方法2 - 使用VerTree的检查功能**：
1. 在VerTree中选择 "GitHub App" 认证
2. 输入App ID和私钥
3. 点击 "获取安装列表" 查看所有可用的installation

## 最佳实践

### 1. 安全管理私钥

- ❌ 不要将私钥提交到版本控制系统
- ❌ 不要在日志中记录私钥内容
- ✅ 将私钥存储在安全的密钥管理系统中
- ✅ 定期轮换私钥（建议一年一次）

### 2. 最小权限原则

- ✅ 只授予必要的权限
- ✅ 定期审查和清理不需要的权限
- ✅ 使用 "Only select repositories" 而不是 "All repositories"

### 3. 监控和审计

- ✅ 定期检查GitHub App的使用情况
- ✅ 监控API调用频率和错误
- ✅ 启用webhook日志记录

## 相关链接

- [GitHub Apps 官方文档](https://docs.github.com/en/developers/apps/getting-started-with-apps)
- [GitHub Apps 权限参考](https://docs.github.com/en/developers/apps/building-github-apps/setting-permissions-for-github-apps)
- [JWT 令牌生成](https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps)

---

如果您在设置过程中遇到问题，请查看VerTree日志或联系技术支持。
