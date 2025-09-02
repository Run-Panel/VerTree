# GitHub Integration 使用示例

## 快速开始

### 1. 创建GitHub Personal Access Token

1. 访问 GitHub Settings > Developer settings > Personal access tokens
2. 点击 "Generate new token"
3. 选择以下权限：
   - `repo` - 完整的仓库访问权限
   - `admin:repo_hook` - 管理仓库webhook

### 2. 绑定GitHub仓库

```bash
# 使用curl创建仓库绑定
curl -X POST http://localhost:8080/admin/api/v1/github/repositories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "app_id": "myapp_001",
    "repository_url": "https://github.com/mycompany/myapp",
    "branch_name": "main",
    "access_token": "ghp_xxxxxxxxxxxxxxxxxxxxx",
    "is_active": true,
    "auto_sync": true,
    "auto_publish": false,
    "default_channel": "stable"
  }'
```

### 3. 获取版本信息

```bash
# 获取最新版本信息
curl http://localhost:8080/api/v1/version-info/myapp_001/stable

# 响应示例
{
  "version": {
    "id": 123,
    "app_id": "myapp_001",
    "version": "v1.2.3",
    "channel": "stable",
    "title": "Release v1.2.3",
    "description": "新功能和bug修复",
    "file_url": "/api/v1/download/cached/456",
    "file_size": 52428800,
    "is_published": true,
    "publish_time": "2024-01-20T10:30:00Z"
  },
  "download_url": "/api/v1/download/latest/myapp_001/stable"
}
```

### 4. 下载最新版本

```bash
# 直接下载最新版本
curl -L http://localhost:8080/api/v1/download/latest/myapp_001/stable \
  -o myapp-latest.zip

# 下载特定版本
curl -L http://localhost:8080/api/v1/download/version/myapp_001/v1.2.3 \
  -o myapp-v1.2.3.zip
```

## 实际场景示例

### 场景1：自动化部署应用

假设你有一个Web应用，需要从GitHub自动获取最新版本并部署。

#### 应用配置
```json
{
  "app_id": "webapp_production",
  "repository_url": "https://github.com/company/webapp",
  "auto_sync": true,
  "auto_publish": true,
  "default_channel": "stable"
}
```

#### 部署脚本
```bash
#!/bin/bash

APP_ID="webapp_production"
CHANNEL="stable"
DEPLOY_DIR="/var/www/webapp"

# 检查更新
LATEST_INFO=$(curl -s "http://vertree.internal/api/v1/version-info/${APP_ID}/${CHANNEL}")
LATEST_VERSION=$(echo $LATEST_INFO | jq -r '.version.version')
CURRENT_VERSION=$(cat ${DEPLOY_DIR}/version.txt 2>/dev/null || echo "none")

if [ "$LATEST_VERSION" != "$CURRENT_VERSION" ]; then
    echo "发现新版本: $LATEST_VERSION，开始部署..."
    
    # 下载新版本
    curl -L "http://vertree.internal/api/v1/download/latest/${APP_ID}/${CHANNEL}" \
      -o "/tmp/webapp-${LATEST_VERSION}.zip"
    
    # 停止服务
    systemctl stop webapp
    
    # 备份当前版本
    mv $DEPLOY_DIR "${DEPLOY_DIR}.backup"
    
    # 解压新版本
    mkdir -p $DEPLOY_DIR
    unzip "/tmp/webapp-${LATEST_VERSION}.zip" -d $DEPLOY_DIR
    
    # 记录版本
    echo $LATEST_VERSION > ${DEPLOY_DIR}/version.txt
    
    # 启动服务
    systemctl start webapp
    
    echo "部署完成: $LATEST_VERSION"
else
    echo "已是最新版本: $CURRENT_VERSION"
fi
```

### 场景2：桌面应用自动更新

桌面应用程序检查并下载更新。

#### 客户端代码示例 (Python)
```python
import requests
import json
import os
from pathlib import Path

class AppUpdater:
    def __init__(self, app_id, channel, current_version):
        self.app_id = app_id
        self.channel = channel
        self.current_version = current_version
        self.base_url = "https://updates.mycompany.com"
    
    def check_update(self):
        """检查是否有新版本"""
        try:
            response = requests.get(
                f"{self.base_url}/api/v1/check-update",
                json={
                    "app_id": self.app_id,
                    "current_version": self.current_version,
                    "channel": self.channel,
                    "client_id": "unique_client_id"
                }
            )
            
            if response.status_code == 200:
                data = response.json()
                return data.get("has_update"), data.get("latest_version")
            else:
                return False, None
                
        except Exception as e:
            print(f"检查更新失败: {e}")
            return False, None
    
    def download_update(self, version):
        """下载更新文件"""
        try:
            download_url = f"{self.base_url}/api/v1/download/version/{self.app_id}/{version}"
            
            response = requests.get(download_url, stream=True)
            response.raise_for_status()
            
            # 获取文件名
            filename = f"update_{version}.zip"
            download_path = Path.home() / "Downloads" / filename
            
            # 下载文件
            with open(download_path, 'wb') as f:
                for chunk in response.iter_content(chunk_size=8192):
                    f.write(chunk)
            
            return str(download_path)
            
        except Exception as e:
            print(f"下载更新失败: {e}")
            return None
    
    def auto_update(self):
        """自动更新流程"""
        print("检查更新...")
        has_update, latest_version = self.check_update()
        
        if has_update:
            print(f"发现新版本: {latest_version}")
            
            # 下载更新
            update_file = self.download_update(latest_version)
            if update_file:
                print(f"更新已下载到: {update_file}")
                # 这里可以添加安装逻辑
                return True
            else:
                print("下载失败")
                return False
        else:
            print("已是最新版本")
            return False

# 使用示例
if __name__ == "__main__":
    updater = AppUpdater(
        app_id="mydesktopapp",
        channel="stable", 
        current_version="v1.0.0"
    )
    
    updater.auto_update()
```

### 场景3：微服务版本管理

在微服务架构中管理多个服务的版本。

#### Docker Compose 自动更新
```yaml
# docker-compose.yml
version: '3.8'

services:
  service-updater:
    image: alpine/curl
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./update-script.sh:/update-script.sh
    environment:
      - VERTREE_URL=http://vertree.internal
    command: ["/bin/sh", "/update-script.sh"]
    restart: always
    
  api-service:
    image: mycompany/api-service:v1.0.0
    ports:
      - "8080:8080"
    
  worker-service:
    image: mycompany/worker-service:v1.0.0
```

#### 更新脚本
```bash
#!/bin/bash
# update-script.sh

SERVICES=("api-service:myapi_prod:stable" "worker-service:myworker_prod:stable")

while true; do
    for service_config in "${SERVICES[@]}"; do
        IFS=':' read -r service_name app_id channel <<< "$service_config"
        
        # 获取最新版本信息
        latest_info=$(curl -s "${VERTREE_URL}/api/v1/version-info/${app_id}/${channel}")
        latest_version=$(echo $latest_info | jq -r '.version.version')
        
        # 获取当前运行的版本
        current_image=$(docker ps --format "table {{.Image}}" | grep $service_name | head -1)
        current_version=$(echo $current_image | cut -d':' -f2)
        
        if [ "$latest_version" != "$current_version" ]; then
            echo "更新 $service_name 从 $current_version 到 $latest_version"
            
            # 下载新镜像（这里假设镜像已经在registry中）
            docker pull "mycompany/${service_name}:${latest_version}"
            
            # 更新docker-compose.yml中的版本
            sed -i "s/${service_name}:${current_version}/${service_name}:${latest_version}/g" docker-compose.yml
            
            # 重启服务
            docker-compose up -d $service_name
            
            echo "$service_name 更新完成"
        fi
    done
    
    # 每5分钟检查一次
    sleep 300
done
```

## 高级配置示例

### 多环境版本管理

```json
{
  "environments": [
    {
      "name": "development",
      "app_id": "myapp_dev",
      "repository_url": "https://github.com/company/myapp",
      "branch_name": "develop",
      "auto_sync": true,
      "auto_publish": true,
      "default_channel": "alpha"
    },
    {
      "name": "staging", 
      "app_id": "myapp_staging",
      "repository_url": "https://github.com/company/myapp",
      "branch_name": "release",
      "auto_sync": true,
      "auto_publish": false,
      "default_channel": "beta"
    },
    {
      "name": "production",
      "app_id": "myapp_prod", 
      "repository_url": "https://github.com/company/myapp",
      "branch_name": "main",
      "auto_sync": true,
      "auto_publish": false,
      "default_channel": "stable"
    }
  ]
}
```

### 智能版本发布策略

```bash
#!/bin/bash
# 发布策略脚本

APP_ID="myapp_prod"
REPO_ID=123

# 获取最新同步的版本
latest_release=$(curl -s "http://vertree.internal/admin/api/v1/github/repositories/${REPO_ID}/releases" | jq -r '.[0]')
version=$(echo $latest_release | jq -r '.tag_name')
is_prerelease=$(echo $latest_release | jq -r '.is_prerelease')

# 根据版本号判断发布策略
if [[ $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    # 正式版本 (如 v1.2.3)
    if [ "$is_prerelease" == "false" ]; then
        echo "检测到正式版本 $version，自动发布到stable渠道"
        # 发布到stable渠道的逻辑
    fi
elif [[ $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+-rc\.[0-9]+$ ]]; then
    # RC版本 (如 v1.2.3-rc.1)
    echo "检测到RC版本 $version，发布到beta渠道"
    # 发布到beta渠道的逻辑
elif [[ $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+-alpha\.[0-9]+$ ]]; then
    # Alpha版本 (如 v1.2.3-alpha.1)
    echo "检测到Alpha版本 $version，发布到alpha渠道"
    # 发布到alpha渠道的逻辑
fi
```

## 监控和告警

### Prometheus 指标

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'vertree'
    static_configs:
      - targets: ['vertree:8080']
    metrics_path: /metrics
    scrape_interval: 30s
```

### Grafana 仪表板

```json
{
  "dashboard": {
    "title": "VerTree GitHub Integration",
    "panels": [
      {
        "title": "同步成功率",
        "type": "stat",
        "targets": [
          {
            "expr": "rate(github_sync_success_total[5m]) / rate(github_sync_total[5m]) * 100"
          }
        ]
      },
      {
        "title": "下载量",
        "type": "graph", 
        "targets": [
          {
            "expr": "increase(file_download_total[1h])"
          }
        ]
      }
    ]
  }
}
```

### 告警规则

```yaml
# alerts.yml
groups:
  - name: vertree
    rules:
      - alert: GitHubSyncFailure
        expr: increase(github_sync_errors_total[5m]) > 5
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "GitHub同步失败率过高"
          
      - alert: FileDownloadFailure  
        expr: increase(file_download_errors_total[5m]) > 10
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "文件下载失败率过高"
```

## 故障处理示例

### 同步失败恢复

```bash
#!/bin/bash
# 同步恢复脚本

REPO_ID=$1
if [ -z "$REPO_ID" ]; then
    echo "用法: $0 <repository_id>"
    exit 1
fi

echo "开始恢复仓库 $REPO_ID 的同步..."

# 强制重新同步
curl -X POST "http://vertree.internal/admin/api/v1/github/repositories/${REPO_ID}/sync" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"force": true}'

echo "同步恢复请求已发送"
```

### 缓存清理

```bash
#!/bin/bash
# 清理过期缓存

# 清理30天前的缓存文件
find /var/vertree/cache -type f -mtime +30 -delete

# 清理数据库中的过期缓存记录
psql -d vertree -c "DELETE FROM file_cache WHERE last_accessed < NOW() - INTERVAL '30 days';"

echo "缓存清理完成"
```

---

*更多示例和最佳实践，请参考 [GitHub Integration Guide](../docs/GITHUB_INTEGRATION.md)*

