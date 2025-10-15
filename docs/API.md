# API 文档

本文档描述了 Git Event Monitor 项目中使用的各个 API 接口及其参数。

## 目录

- [GitHub API](#github-api)
  - [Events API](#github-events-api)
  - [Commits API](#github-commits-api)
- [Gitee API](#gitee-api)
  - [Events API](#gitee-events-api)
  - [Commits API](#gitee-commits-api)
- [API 调用流程](#api-调用流程)
- [错误处理](#错误处理)

---

## GitHub API

### GitHub Events API

获取 GitHub 仓库的事件列表，用于检查代码提交活动。

**端点**
```
GET https://api.github.com/repos/{owner}/{repo}/events
```

**请求头**
```http
Accept: application/vnd.github+json
X-GitHub-Api-Version: 2022-11-28
Authorization: Bearer {token}  # 可选
```

**查询参数**
- 无（使用默认分页）

**响应示例**
```json
[
  {
    "id": "12345",
    "type": "PushEvent",
    "created_at": "2025-09-30T12:29:00Z",
    "payload": {
      "commits": [...]
    }
  }
]
```

**响应状态码**
- `200`: 成功
- `404`: 仓库不存在
- `401`: 未授权
- `403`: 速率限制

---

### GitHub Commits API

检查 GitHub 仓库是否有提交记录。

**端点**
```
GET https://api.github.com/repos/{owner}/{repo}/commits
```

**请求头**
```http
Accept: application/vnd.github+json
X-GitHub-Api-Version: 2022-11-28
Authorization: Bearer {token}  # 可选
```

**查询参数**
- `per_page`: 每页返回数量（默认：30，最大：100）
- `page`: 页码（默认：1）
- `sha`: 分支名或 commit SHA（默认：默认分支）

**响应示例**
```json
[
  {
    "sha": "5dd93ba...",
    "commit": {
      "author": {
        "name": "...",
        "date": "2025-09-30T12:29:00Z"
      },
      "message": "Add files via upload"
    }
  }
]
```

**响应状态码**
- `200`: 仓库有提交记录
- `409`: 仓库为空
- `404`: 仓库不存在

---

## Gitee API

### Gitee Events API

获取 Gitee 仓库的事件列表。

**端点**
```
GET https://gitee.com/api/v5/repos/{owner}/{repo}/events
```

**请求头**
```http
Authorization: token {access_token}  # 可选
```

**查询参数**
- `limit`: 限制返回数量（默认：20，最大：100）
- `page`: 页码（默认：1）
- `access_token`: 认证 token（可通过 header 或 query 传递）

**响应示例**
```json
[
  {
    "id": 12345,
    "type": "PushEvent",
    "created_at": "2025-09-30T18:55:22+08:00",
    "payload": {
      "commits": [...]
    }
  }
]
```

**响应状态码**
- `200`: 成功
- `404`: 仓库不存在
- `401`: 未授权

---

### Gitee Commits API

检查 Gitee 仓库是否有提交记录。

**端点**
```
GET https://gitee.com/api/v5/repos/{owner}/{repo}/commits
```

**请求头**
```http
Authorization: token {access_token}  # 可选
```

**查询参数**
- `page`: 页码（默认：1）
- `per_page`: 每页返回数量（默认：20，最大：100）
- `sha`: 分支名或 commit SHA（默认：默认分支）
- `access_token`: 认证 token（可通过 header 或 query 传递）

**响应示例**
```json
[
  {
    "sha": "f9252ea...",
    "commit": {
      "author": {
        "name": "...",
        "date": "2025-09-24T12:37:37Z"
      },
      "message": "增加FreeRTOS"
    }
  }
]
```

**响应状态码**
- `200`: 仓库有提交记录
- `404`: 仓库为空或不存在

---

## 平台差异对比

| 特性 | GitHub | Gitee |
|------|--------|-------|
| **Events API 路径** | `/repos/{owner}/{repo}/events` | `/repos/{owner}/{repo}/events` |
| **Commits API 路径** | `/repos/{owner}/{repo}/commits` | `/repos/{owner}/{repo}/commits` |
| **认证方式** | `Authorization: Bearer {token}` | `Authorization: token {token}` |
| **空仓库状态码** | 409 Conflict | 404 Not Found |
| **时间格式** | RFC3339 (UTC) | RFC3339 with timezone |
| **API 版本 Header** | 需要 `X-GitHub-Api-Version` | 不需要 |
| **速率限制（无认证）** | 60 次/小时 | 60 次/小时 |
| **速率限制（有认证）** | 5000 次/小时 | 5000 次/小时 |

---

## 相关文档

- [GitHub REST API 文档](https://docs.github.com/en/rest)
- [Gitee API 文档](https://gitee.com/api/v5/swagger)
- [项目设计文档](./DESIGN.md)

---

*最后更新: 2025-10-01*
