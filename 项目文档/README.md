# Bubbles-CN 项目文档

## 概述

Bubbles-CN 是一个为 Bubble Tea 应用程序提供 TUI（终端用户界面）组件的中文本地化项目。本文档体系提供了全面的项目说明、技术架构、开发指南和运维支持。

## 文档目录

### 架构设计

| 文档 | 描述 |
|------|------|
| [01-项目架构概述](architecture/01-项目架构概述.md) | 详细描述系统整体架构设计、核心组件关系及技术架构图 |
| [02-技术栈说明](architecture/02-技术栈说明.md) | 列出前端、后端、数据库、中间件等所有技术组件及其版本信息 |
| [03-模块划分](architecture/03-模块划分.md) | 明确系统功能模块划分、模块职责及模块间交互关系 |

### API 文档

| 文档 | 描述 |
|------|------|
| [01-API接口文档](api/01-API接口文档.md) | 提供完整的API接口文档，包含接口路径、请求方法、参数说明、返回格式及错误码定义 |
| [02-数据流程文档](api/02-数据流程文档.md) | 绘制关键业务流程的数据流转图，说明数据在各模块间的传递过程 |

### 开发指南

| 文档 | 描述 |
|------|------|
| [01-开发规范](development/01-开发规范.md) | 制定编码规范、命名规范、代码审查标准及文档编写规范 |

### 部署指南

| 文档 | 描述 |
|------|------|
| [01-构建部署流程](deployment/01-构建部署流程.md) | 提供详细的环境配置说明、构建步骤、部署流程及环境变量配置 |

### 测试指南

| 文档 | 描述 |
|------|------|
| [01-测试策略](testing/01-测试策略.md) | 明确单元测试、集成测试、系统测试的实施方法及测试工具使用规范 |

### 版本控制

| 文档 | 描述 |
|------|------|
| [01-版本控制策略](version-control/01-版本控制策略.md) | 制定分支管理策略、代码合并流程及版本号命名规则 |

### 故障排除

| 文档 | 描述 |
|------|------|
| [01-常见问题解决方案](troubleshooting/01-常见问题解决方案.md) | 整理开发、测试、部署过程中常见问题的诊断方法和解决策略 |

## 快速开始

### 1. 环境准备

```bash
# 安装 Go 1.24.2 或更高版本
wget https://go.dev/dl/go1.24.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 克隆项目
git clone https://github.com/purpose168/bubbles-cn.git
cd bubbles-cn
```

### 2. 构建项目

```bash
# 下载依赖
go mod download

# 运行测试
go test ./...

# 构建项目
make build
```

### 3. 使用组件

```go
import (
    tea "github.com/purpose168/bubbletea-cn"
    "github.com/purpose168/bubbles-cn/textinput"
)

func main() {
    ti := textinput.New()
    ti.Placeholder = "请输入内容"
    ti.Focus()
    
    p := tea.NewProgram(ti)
    if _, err := p.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## 项目结构

```
bubbles-cn/
├── 项目文档/                 # 项目文档目录
│   ├── architecture/         # 架构设计文档
│   ├── api/                  # API 文档
│   ├── development/          # 开发指南
│   ├── deployment/           # 部署指南
│   ├── testing/              # 测试指南
│   ├── version-control/      # 版本控制
│   └── troubleshooting/      # 故障排除
├── cursor/                   # 光标组件
├── key/                      # 键绑定组件
├── help/                     # 帮助组件
├── textinput/                # 文本输入组件
├── textarea/                 # 文本区域组件
├── list/                     # 列表组件
├── table/                    # 表格组件
├── viewport/                 # 视口组件
├── spinner/                  # 加载动画组件
├── progress/                 # 进度条组件
├── paginator/                # 分页器组件
├── filepicker/               # 文件选择器组件
├── timer/                    # 定时器组件
├── stopwatch/                # 秒表组件
└── runeutil/                 # 字符工具
```

## 核心组件

### TextInput

文本输入组件，支持单行文本输入、光标导航、粘贴操作等功能。

### TextArea

文本区域组件，支持多行文本输入、滚动、编辑等功能。

### List

列表组件，支持项目选择、过滤、分页等功能。

### Table

表格组件，支持多列数据展示、排序、分页等功能。

### Viewport

视口组件，支持内容滚动、视口管理等功能。

### Spinner

加载动画组件，提供多种加载动画样式。

### Progress

进度条组件，支持进度显示、百分比显示等功能。

### Paginator

分页器组件，支持分页导航、页码显示等功能。

### FilePicker

文件选择器组件，支持文件浏览、选择、过滤等功能。

### Timer

定时器组件，支持定时器启动、停止、重置等功能。

### Stopwatch

秒表组件，支持秒表启动、停止、重置等功能。

## 技术架构

### MVU 架构

Bubbles-CN 采用 Model-View-Update (MVU) 架构模式：

- **Model**: 表示应用的状态
- **View**: 根据状态生成 UI
- **Update**: 处理消息并更新状态

### 组件设计

所有组件都遵循统一的设计模式：

- 实现了 `tea.Model` 接口
- 提供链式调用方法
- 支持自定义样式
- 提供完整的 API

## 开发规范

### 编码规范

- 遵循 Go 语言官方编码规范
- 使用 `gofmt` 格式化代码
- 添加充分的注释
- 编写单元测试

### 命名规范

- 包名使用小写字母
- 导出函数首字母大写
- 使用驼峰命名法
- 变量名应该有意义

### 提交规范

使用语义化提交信息：

```
<类型>(<范围>): <描述>

[可选的正文]

[可选的脚注]
```

## 测试策略

### 测试层次

- **单元测试**: 验证单个函数/方法的行为
- **集成测试**: 验证多个组件的交互
- **系统测试**: 验证整个系统的功能
- **端到端测试**: 验证完整的用户场景

### 测试工具

- `go test`: 运行测试
- `go tool cover`: 覆盖率分析
- `go tool pprof`: 性能分析
- `golangci-lint`: 代码检查

## 版本控制

### 分支策略

- `main`: 主分支，稳定版本
- `develop`: 开发分支，最新开发版本
- `feature/*`: 功能分支
- `bugfix/*`: Bug 修复分支
- `hotfix/*`: 紧急修复分支
- `release/*`: 发布分支

### 版本号

使用语义化版本号：`MAJOR.MINOR.PATCH`

- **MAJOR**: 主版本号，不兼容的 API 修改
- **MINOR**: 次版本号，向下兼容的功能性新增
- **PATCH**: 修订号，向下兼容的问题修正

## 常见问题

### 开发问题

- 依赖下载失败
- 编译错误
- 导入循环
- 类型不匹配

### 测试问题

- 测试失败
- 竞态条件
- 测试超时

### 部署问题

- 构建失败
- 交叉编译失败
- Docker 构建失败

### 运行时问题

- 终端显示异常
- 性能问题
- 内存泄漏

详细的解决方案请参考 [常见问题解决方案](troubleshooting/01-常见问题解决方案.md)。

## 贡献指南

欢迎贡献代码、报告问题或提出建议。

### 贡献流程

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'feat: Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### 代码审查

所有代码都需要经过审查才能合并到主分支。

## 许可证

本项目采用 MIT 许可证。

## 联系方式

- 作者: purpose168
- 邮箱: purpose168@outlook.com
- GitHub: https://github.com/purpose168/bubbles-cn

## 更新日志

详细的更新日志请参考项目的 CHANGELOG.md 文件。

## 相关资源

- [Bubble Tea](https://github.com/purpose168/bubbletea-cn): TUI 框架
- [Lip Gloss](https://github.com/purpose168/lipgloss-cn): 样式库
- [Charm](https://charm.sh): 终端工具集合

## 致谢

感谢所有贡献者和用户的支持！

---

**最后更新**: 2026-02-13
