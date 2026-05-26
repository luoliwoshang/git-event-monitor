---
name: git-race-pr-commit-detect
description: Batch audit CSV/Excel repositories for coding competition compliance. Checks repository accessibility, submission deadline compliance, commit counts, and PR statistics (total/open/merged/closed). Use this skill when the user needs to check if participants submitted code on time in a coding race/contest.
---

# Git Race PR Commit Detect

Batch audit tool for coding competitions. Reads a CSV/Excel file containing repository URLs, then for each repo checks:

- **Accessibility** — whether the repo is reachable via GitHub/Gitee API
- **Deadline compliance** — whether the latest push was before a specified cutoff
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

## Token Setup

Guide the user to obtain API tokens before running:

**GitHub:**
1. Visit https://github.com/settings/tokens
2. Create a classic token with `public_repo` scope (read-only is sufficient)
3. Use with `--github-token=ghp_xxxxxxxxxxxx`

**Gitee:**
1. Visit https://gitee.com/personal_access_tokens
2. Create a token with basic read permissions
3. Use with `--gitee-token=xxxxxxxxxxxx`

At least one token is required. If the input file contains repos from both platforms, both tokens are needed.

## How to Run

```bash
go run ./cmd/csv-processor/main.go \
  --github-token=ghp_xxxxxxxxxxxx \
  --gitee-token=xxxxxxxxxxxx \
  --deadline=2026-05-26T00:00:00+08:00 \
  "input.csv" <start-row> <end-row>
```

**Arguments:**

| Argument | Description |
|---|---|
| `--github-token` | GitHub personal access token |
| `--gitee-token` | Gitee personal access token |
| `--deadline` | Submission cutoff in RFC3339 format (e.g. `2026-05-26T00:00:00+08:00`) |
| `<file>` | Path to the CSV or Excel file |
| `<start-row>` | First data row (1-indexed, row 1 is header, so start >= 2) |
| `<end-row>` | Last data row (1-indexed, inclusive) |

All arguments are required.

## Output Columns Appended

The tool generates a `_processed` file next to the original. It appends these 7 columns:

| Column | Values |
|---|---|
| `是否可访问` | `可访问` / `不可访问` |
| `是否准时提交` | `准时提交` / `超时提交` / `空仓库` / `初始提交` / `分析失败` |
| `Commit数量` | Number of commits (`1000+` if capped) |
| `PR总数` | Total PR count |
| `PR-Open` | Open PR count |
| `PR-Merged` | Merged PR count |
| `PR-Closed` | Closed (unmerged) PR count |

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

| Issue | Detection | Fix |
|---|---|---|
| Multiple URLs in one cell | Cell contains more than one `http` | Keep only the GitHub/Gitee repo URL |
| Extra descriptions | Cell contains text beyond the URL (e.g. "Repository: ", demo links, doc links) | Strip descriptions, keep only the repo URL |
| Newlines in cell | Cell contains `\n` | Remove newlines and extra content |
| Missing repo URL | Cell is empty | Flag the row, cannot process |
| Unsupported platforms | URL is not GitHub or Gitee (e.g. Google Drive, Bilibili, personal GitLab) | Remove non-repo links, keep only GitHub/Gitee URL |
| `.git` suffix | URL ends with `.git` | This is fine, the tool handles it automatically |

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

1. **Check the CSV first** — use `go run ./cmd/csv-reader/main.go <file>` to inspect headers and row count before running.
2. **Count rows** — row 1 is the header, data starts at row 2. The `<end-row>` should be the total line count.
3. **Deadline timezone** — always use `+08:00` for China Standard Time unless the user specifies otherwise. Format: `YYYY-MM-DDTHH:MM:SS+08:00`.
4. **Missing tokens** — if the user has no tokens, guide them through the setup URLs above before running.
5. **Large files** — the tool supports both `.csv` and `.xlsx` / `.xls` formats.
