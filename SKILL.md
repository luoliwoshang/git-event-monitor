---
name: git-race-pr-commit-detect
description: Batch audit CSV/Excel repositories for coding competition compliance. Checks repository accessibility, submission deadline compliance, commit counts, and PR statistics (total/open/merged/closed). Use this skill when the user needs to check if participants submitted code on time in a coding race/contest.
---

# Git Race PR Commit Detect

> **GATE CHECK: Tokens are MANDATORY.** If the user does not have both GitHub and Gitee tokens configured, DO NOT proceed past this point. Direct them to [Token Setup](#token-setup--mandatory) immediately. No exceptions.

Batch audit tool for coding competitions. Reads a CSV/Excel file containing repository URLs, then for each repo checks:

- **Accessibility** — whether the repo is reachable via GitHub/Gitee API
- **Start time compliance** — whether the first commit was after a specified start time (optional)
- **Deadline compliance** — whether the latest push was before a specified cutoff (optional)
- **Commit count** — total number of commits (capped at 1000)
- **PR stats** — total PRs, plus breakdown by Open / Merged / Closed

## Prerequisites

This tool is part of the `git-event-monitor` project. If the user does not have it cloned yet:

```bash
git clone https://github.com/luoliwoshang/git-event-monitor.git
cd git-event-monitor
```

**Requirements:**

- Go 1.23+ must be installed (`go version`)
- The project must be the current working directory for `go run ./cmd/...` to work

Always verify the user is inside the project directory before running any commands.

## Required Input Format

The input file (CSV or Excel) **must** have a column whose header **contains** `代码仓库地址` (repository URL). All other columns are optional.

Supported URL formats:

- `https://github.com/owner/repo` or `https://github.com/owner/repo.git`
- `git@github.com:owner/repo.git`
- `https://gitee.com/owner/repo` or `https://gitee.com/owner/repo.git`
- `git@gitee.com:owner/repo.git`

Example valid CSV:

```csv
姓名,代码仓库地址
张三,https://github.com/zhangsan/my-project
李四,https://gitee.com/lisi/demo
```

## Token Setup — MANDATORY

> **CRITICAL: Tokens are MANDATORY. The tool WILL NOT work without them.**
> Without a token, GitHub/Gitee API requests will be rate-limited to the point of failure. Do NOT attempt to run the tool before the user has configured tokens.

Guide the user to obtain BOTH tokens before running:

**GitHub Token:**

1. Visit https://github.com/settings/tokens
2. Click "Generate new token (classic)"
3. Select `public_repo` scope (read-only access to public repositories)
4. Copy the generated token (starts with `ghp_`)
5. Use with `--github-token=ghp_xxxxxxxxxxxx`

**Gitee Token:**

1. Visit https://gitee.com/personal_access_tokens
2. Click "Generate new token"
3. Grant basic read permissions
4. Copy the generated token
5. Use with `--gitee-token=xxxxxxxxxxxx`

**Both tokens are required** even if the input CSV only contains repos from one platform. The tool needs them to function reliably.

**Token Not Ready?** If the user does not have tokens yet, STOP. Do not proceed further. Send them the links above and wait until they confirm both tokens are ready.

## How to Run

```bash
go run ./cmd/csv-processor/main.go \
  --github-token=ghp_xxxxxxxxxxxx \
  --gitee-token=xxxxxxxxxxxx \
  --start-time=2026-05-22T00:00:00+08:00 \
  --deadline=2026-05-26T00:00:00+08:00 \
  "input.csv" <start-row> <end-row>
```

**Arguments:**

| Argument         | Description                                                                                                                             |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `--github-token` | GitHub personal access token (required)                                                                                                 |
| `--gitee-token`  | Gitee personal access token (required)                                                                                                  |
| `--start-time`   | Earliest allowed first-commit time in RFC3339 format. If set, checks that each repo's first commit is at or after this time. (optional) |
| `--deadline`     | Submission cutoff in RFC3339 format. If set, checks that each repo's latest push is at or before this time. (optional)                  |
| `<file>`         | Path to the CSV or Excel file                                                                                                           |
| `<start-row>`    | First data row (1-indexed, row 1 is header, so start >= 2)                                                                              |
| `<end-row>`      | Last data row (1-indexed, inclusive)                                                                                                    |

Tokens are always required. `--start-time` and `--deadline` are optional but at least one of them should be provided for meaningful results.

## Output Columns Appended

The tool generates a `_processed` file next to the original. It appends these 10 columns:

| Column | Values |
|---|---|
| `是否可访问` | `可访问` / `不可访问` |
| `是否准时提交` | `准时提交` / `超时提交` / `空仓库` / `初始提交` / `分析失败` |
| `Commit数量` | Number of commits (`1000+` if capped) |
| `PR总数` | Total PR count |
| `PR-Open` | Open PR count |
| `PR-Merged` | Merged PR count |
| `PR-Closed` | Closed (unmerged) PR count |
| `是否在起始时间后提交` | `在起始时间后提交` / `在起始时间前提交` / `空仓库` / `获取失败` |
| `超时详情` | Time difference (e.g. "5 minutes after deadline"), only when overdue |
| `起始违规详情` | First commit timestamp, only when earlier than start time |

Existing columns with matching names (via substring match on headers) are updated in place rather than duplicated.

## CSV Validation & Preprocessing

**Before running the tool, always validate the input CSV.** Common issues that cause rows to be skipped:

### 1. Messy URL Cells

The `代码仓库地址` cell must contain **only a single GitHub or Gitee repo URL**. Cells with multiple URLs, descriptions, or extra text will be skipped with the message:

> ⏭️ Skipping: Cannot parse repository URL (multiple URLs, unsupported platform, or invalid format)

**Bad — cell contains multiple links and descriptions:**

```
Repository: https://github.com/user/repo-name
【Demo:Some Video】 https://b23.tv/xxxxx
文档链接：https://drive.google.com/file/d/xxxxx/view?usp=sharing
```

**Good — extract only the repo URL:**

```
https://github.com/user/repo-name
```

### 2. Validation Checklist

When inspecting a user's CSV, check for these issues:

| Issue                     | Detection                                                                      | Fix                                               |
| ------------------------- | ------------------------------------------------------------------------------ | ------------------------------------------------- |
| Multiple URLs in one cell | Cell contains more than one `http`                                             | Keep only the GitHub/Gitee repo URL               |
| Extra descriptions        | Cell contains text beyond the URL (e.g. "Repository: ", demo links, doc links) | Strip descriptions, keep only the repo URL        |
| Newlines in cell          | Cell contains `\n`                                                             | Remove newlines and extra content                 |
| Missing repo URL          | Cell is empty                                                                  | Flag the row, cannot process                      |
| Unsupported platforms     | URL is not GitHub or Gitee (e.g. Google Drive, Bilibili, personal GitLab)      | Remove non-repo links, keep only GitHub/Gitee URL |
| `.git` suffix             | URL ends with `.git`                                                           | This is fine, the tool handles it automatically   |

### 3. Preprocessing Steps for the Agent

1. **Inspect first** — run `go run ./cmd/csv-reader/main.go <file>` to see headers, row count, and sample data
2. **Check the URL column** — look for cells with multiple `http` occurrences, newlines, or extra text
3. **Flag bad rows** — tell the user which rows have messy URL cells before running
4. **Help clean** — if there are only a few bad rows, help the user fix them manually. For many bad rows, suggest they clean the spreadsheet by extracting only `https://github.com/...` or `https://gitee.com/...` patterns from each cell
5. **Re-run** — after cleaning, validate again before executing the processor

### 4. Quick Sanity Check Command

```bash
# Count how many rows will be skipped due to bad URLs
grep -c 'http.*http' "input.csv"
# Count rows with newlines in cells (potential multi-line mess)
grep -c '^\|\n' "input.csv"
```

## Tips for the Agent

1. **Tokens first, ALWAYS** — This is non-negotiable. Before doing anything else, ask the user: "Do you have both GitHub and Gitee tokens ready?" If they say no or are unsure, STOP immediately and direct them to the Token Setup section. Do not proceed to CSV inspection, do not run any commands, do not offer to "try anyway." The tool will fail without tokens.
2. **Check the CSV first** — once tokens are confirmed, use `go run ./cmd/csv-reader/main.go <file>` to inspect headers and row count.
3. **Count rows** — row 1 is the header, data starts at row 2. The `<end-row>` should be the total line count.
4. **Start time and deadline timezone** — always use `+08:00` for China Standard Time unless the user specifies otherwise. Format: `YYYY-MM-DDTHH:MM:SS+08:00`. Both `--start-time` and `--deadline` accept the same format.
5. **Large files** — the tool supports both `.csv` and `.xlsx` / `.xls` formats.

go run ./cmd/csv-processor/main.go \
 --github-token=ghp_xxxxxxxxxxxx \
 --gitee-token=xxxxxxxxxxxx \
 --start-time=2026-05-22T00:00:00+08:00 \
 --deadline=2026-05-25T00:00:00+08:00 \
 "2026 暑期实训项目- HR 初筛表-5月22日至5月24日-HR 初筛.csv" 2 246
