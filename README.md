# Git Event Monitor

🔍 一个用于监控 Git 仓库代码提交事件的工具，专为代码竞赛公平性检查而设计。

## 🚀 在线体验

**[👉 立即使用 - https://luoliwoshang.github.io/git-event-monitor/](https://luoliwoshang.github.io/git-event-monitor/)**

无需安装，直接在浏览器中使用！

## ✨ 功能特点

- 🌍 **多平台支持**：同时支持 GitHub 和 Gitee 平台
- 🔍 **智能事件检测**：自动识别推送事件和合并的 Pull Request
- ⏰ **截止时间检查**：验证代码提交是否在指定截止时间之前
- 🕐 **多格式时间显示**：显示原始时间、UTC 时间和本地时间
- 📱 **响应式设计**：完美支持桌面和移动设备
- 🎨 **现代化界面**：基于 Ant Design 的专业级 UI

## 🎯 使用场景

- **代码竞赛监督**：检查参赛者是否在截止时间前提交代码
- **项目截止检查**：验证团队成员的提交时间合规性
- **仓库活动分析**：查看最近的代码提交活动

## 🛠️ 技术栈

- **前端**：React + TypeScript + Vite
- **UI 框架**：Ant Design
- **API 集成**：GitHub API v3 + Gitee API v5
- **部署**：GitHub Pages + GitHub Actions
- **时间处理**：Day.js

## 📖 使用方法

1. 访问 [在线工具](https://luoliwoshang.github.io/git-event-monitor/)
2. 输入仓库名称（格式：`owner/repo`）
3. 选择平台（GitHub 或 Gitee）
4. 可选：添加 API Token 以提高请求限制
5. 可选：设置截止时间进行合规性检查
6. 点击"分析仓库"查看结果

## 🔑 API Token 说明

### GitHub Token
- 访问：https://github.com/settings/tokens
- 权限：只需要 `public_repo` 权限即可

### Gitee Token
- 访问：https://gitee.com/personal_access_tokens
- 权限：只需要基础读取权限即可

## 🖥️ 命令行工具

### 批量检查提交合规性（csv-processor）

读取包含仓库地址的 CSV/Excel 表格，批量检查仓库可访问性和是否在截止时间前提交。

```bash
go run ./cmd/csv-processor/main.go \
  --github-token=YOUR_GH_TOKEN \
  --gitee-token=YOUR_GITEE_TOKEN \
  --deadline=2025-09-30T23:59:59Z \
  "议题三 待筛选 名单.xlsx" 2 67
```

参数说明：
- `--github-token`：GitHub API Token（可选，提高速率限制）
- `--gitee-token`：Gitee API Token（可选，提高速率限制）
- `--deadline`：截止时间，RFC3339 格式（如 `2025-09-30T23:59:59Z`）
- 位置参数：`<文件路径> <起始行> <结束行>`（行号从 1 开始，第 1 行为表头）

输出结果会在原文件同目录生成一个 `_processed` 文件，新增「是否可访问」「是否准时提交」两列。

### 其他工具

| 工具 | 路径 | 用途 |
|------|------|------|
| 主 CLI | `cmd/git-event-monitor/` | 检查单个仓库的提交合规性 |
| AI Coding 统计 | `cmd/ai-coding-stats/` | 统计仓库中 AI 生成 PR 的占比 |
| CSV 查看器 | `cmd/csv-reader/` | 快速查看 CSV 文件内容 |
| Excel 查看器 | `cmd/excel-reader/` | 快速查看 Excel 文件内容 |

```bash
# 检查单个仓库
go run ./cmd/git-event-monitor/main.go check owner/repo

# 统计 AI Coding 浓度
go run ./cmd/ai-coding-stats/main.go input.csv > output.csv
```

## 🏗️ 本地开发

```bash
# 克隆仓库
git clone https://github.com/luoliwoshang/git-event-monitor.git

# 进入项目目录
cd git-event-monitor/web-app

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build
```

## 📝 更新日志

- **v2.0** - 添加 Gitee 支持，UI 重构为 Ant Design
- **v1.0** - 基础 GitHub 事件监控功能

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License