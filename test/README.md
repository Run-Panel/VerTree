# VerTree 测试文档

## 概述

这个目录包含了 VerTree 应用的各种测试文件和工具，用于验证版本管理功能的正确性。

## 文件结构

```
test/
├── README.md                 # 本文档
├── version_demo.go          # 版本比较功能演示程序
├── integration_test.go      # 集成测试（WIP）
├── run_tests.sh            # 测试运行脚本
└── vertree                 # 编译后的测试应用（动态生成）
```

## 版本比较功能

### 核心功能

VerTree 现在支持完整的语义化版本(SemVer)比较功能：

1. **版本比较** - 判断两个版本的大小关系
2. **更新检查** - 判断是否需要更新到新版本
3. **最小版本要求** - 检查当前版本是否满足最小要求
4. **版本排序** - 对版本列表进行正确排序
5. **版本信息提取** - 解析版本号的各个组成部分

### 支持的版本格式

- 标准语义化版本：`1.2.3`
- 带 v 前缀：`v1.2.3`
- 预发布版本：`1.2.3-alpha.1`
- 构建元数据：`1.2.3+build.123`
- 复合格式：`v2.0.0-beta.1+build.456`
- 非标准版本：自动回退到字符串比较

### 演示程序

运行版本比较演示：

```bash
cd test
go run version_demo.go
```

演示程序会展示：
- 各种版本比较场景
- 版本排序功能
- 最小版本要求检查
- 版本信息提取

## 测试结果示例

### 版本比较测试

```
📊 Version Comparison Tests:
 1. Patch update: 1.0.0 < 1.0.1 ✅ Update needed
 2. Minor update: 1.0.0 < 1.1.0 ✅ Update needed  
 3. Major update: 1.0.0 < 2.0.0 ✅ Update needed
 4. Same version: 2.0.0 = 2.0.0 ❌ No update needed
 5. Prerelease: 1.0.0-alpha < 1.0.0 ✅ Update needed
```

### 版本排序测试

```
📋 Version Sorting Test:
Original: [2.1.0 1.0.0 v1.2.0 1.0.1 2.0.0-beta.1 2.0.0 1.0.0-alpha v3.0.0]
Sorted:   [1.0.0-alpha 1.0.0 1.0.1 v1.2.0 2.0.0-beta.1 2.0.0 2.1.0 v3.0.0]
```

## 单元测试

运行完整的单元测试套件：

```bash
go test ./internal/utils/ -v
```

测试覆盖：
- ✅ 版本比较逻辑
- ✅ 更新需求检查
- ✅ 最小版本要求
- ✅ 版本有效性验证
- ✅ 版本排序算法
- ✅ 边界情况处理

## 集成测试（开发中）

集成测试将验证：
- API 端点的版本管理功能
- 数据库存储和检索
- 用户认证和授权
- 完整的版本生命周期

运行集成测试：

```bash
./run_tests.sh
```

## 技术实现

### 依赖库

- `github.com/Masterminds/semver/v3` - 语义化版本处理
- 标准库 `sort` - 版本排序
- 标准库 `strings` - 字符串处理

### 核心算法

1. **版本规范化** - 统一版本格式（添加 v 前缀）
2. **SemVer 解析** - 使用 Masterminds/semver 库
3. **回退机制** - 非 SemVer 版本使用字符串比较
4. **排序算法** - 基于 SemVer 规则的自定义排序

### 性能特点

- 高效的版本比较算法
- 内存友好的排序实现
- 支持大量版本的批量处理
- 错误容错机制

## 使用示例

```go
// 创建版本比较器
vc := utils.NewVersionComparer()

// 检查是否需要更新
needsUpdate := vc.IsUpdateNeeded("1.0.0", "1.0.1") // true

// 比较两个版本
comparison := vc.CompareVersions("1.0.0", "2.0.0") // -1

// 检查最小版本要求
meets := vc.MeetsMinimumVersion("1.2.0", "1.0.0") // true

// 排序版本列表
versions := []string{"2.0.0", "1.0.0", "1.1.0"}
sorted := vc.SortVersions(versions) // ["1.0.0", "1.1.0", "2.0.0"]

// 获取版本信息
info := vc.GetVersionInfo("v1.2.3-beta.1")
// info["major"] = 1, info["minor"] = 2, info["patch"] = 3
```

## 测试覆盖率

当前测试覆盖了以下场景：
- ✅ 标准 SemVer 版本
- ✅ 带前缀版本
- ✅ 预发布版本
- ✅ 构建元数据
- ✅ 混合格式
- ✅ 无效版本
- ✅ 边界情况
- ✅ 空值处理

## 下一步计划

- [ ] 完善集成测试
- [ ] 添加性能基准测试
- [ ] 支持更多版本格式
- [ ] 添加版本建议功能
- [ ] 实现版本冲突检测

---

*最后更新: 2025年8月*
