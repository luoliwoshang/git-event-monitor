---
name: demo-video-check
description: Check GitHub/Gitee repositories for demo videos, presentation videos, or walkthrough videos in any form (video files, external links to Bilibili/YouTube, cloud storage links like Baidu Pan/Google Drive). Use this skill when the user needs to verify whether a repo has a demo video.
---

# Demo Video Check

Check whether a GitHub or Gitee repository has a demo video, presentation video, or walkthrough in **any form** — video files committed to the repo, external links to video platforms, or cloud storage shares.

## Strategy: Remote First, Clone as Fallback

Prefer API-based remote inspection. If the API is unavailable (rate limited, no `gh` CLI, network issues), fall back to `git clone --depth 1`. Cloning is slower but works when the API path is blocked.

## Prerequisites & Fallbacks

### gh CLI Not Installed?

Check with `which gh`. If not installed, guide the user:

**macOS:** `brew install gh`
**Linux:** Follow https://github.com/cli/cli/blob/trunk/docs/install_linux.md
**Windows:** `winget install --id GitHub.cli`

If the user cannot install `gh`, offer to use `curl` + GitHub API directly, or fall back to `git clone --depth 1`.

### Rate Limited by GitHub API?

Without a token, GitHub's unauthenticated API allows only 60 requests/hour. **As soon as you hit a rate limit, STOP and ask the user:**

> "I'm being rate-limited by GitHub. Do you have a GitHub personal access token? Create one at https://github.com/settings/tokens (only needs `public_repo` scope), then set it with: `export GH_TOKEN=ghp_xxx`"

Similarly for Gitee — ask for a Gitee token and use `?access_token=xxx` in API URLs.

### Git Clone Fallback

If API approaches fail, use shallow clone to inspect the repo locally:

```bash
git clone --depth 1 https://github.com/{owner}/{repo}.git /tmp/demo-check/{repo}
find /tmp/demo-check/{repo} -type f -not -path '*/.git/*'
```

Then scan the local files for video files and README content. Clean up with `rm -rf /tmp/demo-check/{repo}` when done.

## Step 1: Scan the File Tree for Committed Video Files

Use `gh api` to list all files in the repo and look for video file extensions:

```bash
gh api repos/{owner}/{repo}/git/trees/HEAD?recursive=1 --jq '.tree[].path' | grep -iE '\.(mp4|mov|webm|avi|mkv|flv|wmv|m4v|ogv)'
```

Also check for files or directories with demo-related names:

```bash
gh api repos/{owner}/{repo}/git/trees/HEAD?recursive=1 --jq '.tree[].path' | grep -iE 'demo|视频|演示|video|walkthrough|tutorial|screen.?record'
```

If a video file is found in the repo, that's a hit. Note the file path.

## Step 2: Check README for External Links

Fetch the README (without cloning):

```bash
gh api repos/{owner}/{repo}/readme --jq '.content' | base64 -d
```

Scan the README for video or demo links. **The platforms listed below are common examples, NOT an exhaustive list.** Look for ANY URL that could host a demo video:

### Video platform links (examples)

```
bilibili.com  b23.tv  youtube.com  youtu.be  vimeo.com  youku.com
tudou.com  iqiyi.com  acfun.cn  xigua.com  douyin.com
```

### Cloud storage / file share links (examples)

```
pan.baidu.com   drive.google.com   aliyundrive.com   lanzou (蓝奏云)
weiyun.com   115.com   mega.nz   dropbox.com   onedrive.com
quark.cn   xunlei.com   caiyun.com
```

### Archive files that may contain videos (examples)

```
.zip  .rar  .7z  .tar.gz  .tgz
```

### Generic demo keywords in link context (examples)

```
演示  视频  demo  预览  preview  walkthrough  教程  tutorial
showcase  overview  introduction  quickstart  getting.started
```

Search patterns to use:

```bash
# Video platforms
grep -iE 'bilibili|b23\.tv|youtube|youtu\.be|vimeo|youku|tudou|iqiyi|acfun|xigua|douyin'

# Cloud storage / file sharing
grep -iE 'pan\.baidu|drive\.google|aliyundrive|lanzou|weiyun|mega\.nz|dropbox|115\.com|onedrive|quark|xunlei|caiyun'

# Any URL near demo keywords — cast a wide net
grep -iE 'https?://[^ ]*' | grep -iE 'demo|视频|演示|preview|walkthrough|tutorial|教程|showcase|overview|introduction'

# .zip / .rar / .7z archive links (might contain video)
grep -iE 'https?://[^ ]*\.(zip|rar|7z|tar\.gz|tgz)'
```

**Pro tip**: If a README is very long, don't dump the whole thing. Pipe through grep first. If nothing matches, skim the full content for any unexpected URLs.

## Step 3: Check Repo Metadata

The repo's description and homepage URL may contain demo links:

```bash
gh api repos/{owner}/{repo} --jq '{desc: .description, homepage: .homepage, topics: .topics}'
```

- **homepage** — often points to a demo site or video link
- **description** — may mention a demo link
- **topics** — may indicate the project type

## Step 4: Check Other Markdown Files

Beyond README.md, repos may have dedicated demo docs like `DEMO.md`, `演示.md`, etc. List them:

```bash
gh api repos/{owner}/{repo}/git/trees/HEAD?recursive=1 --jq '.tree[].path' | grep -iE 'demo|演示|video|walkthrough|tutorial' | grep -iE '\.(md|txt|html)$'
```

Then fetch each one and scan for links using the same patterns as Step 2.

## Step 5: Check Recent Commits / Release Assets

Check if any release has video attachments:

```bash
gh api repos/{owner}/{repo}/releases --jq '.[].assets[].browser_download_url' | grep -iE '\.(mp4|mov|webm|avi|mkv)'
```

## Summary Report Format

After checking all 5 steps, report for each repo:

```
| Repo | Video File | External Link | Source |
|------|-----------|---------------|--------|
| owner/repo | demo.mp4 | - | In repo |
| owner/repo | - | bilibili.com/video/xxx | README |
| owner/repo | - | pan.baidu.com/s/xxx | README (Baidu Pan) |
| owner/repo | - | - | NONE |
```

## Tips for the Agent

1. **Remote first, clone as backup** — Start with `gh api` / `curl`. If rate limited or blocked, ask for tokens. If that also fails, fall back to `git clone --depth 1`.
2. **Ask for tokens proactively** — at the start of checking multiple repos, ask: "Do you have GitHub and Gitee tokens? I may need them to avoid rate limits." If the user doesn't have them, guide them through creation.
3. **Gitee repos** — the approach is the same, but use Gitee API endpoints. `gh api` only works for GitHub. For Gitee, use `curl` with `https://gitee.com/api/v5/repos/{owner}/{repo}/...` and add `?access_token=xxx` for auth.
4. **Link lists are examples** — The domains listed in Step 2 are common ones, not a complete catalog. Any URL could be a demo link. If you see an unfamiliar domain next to demo keywords, report it.
5. **Multi-line README** — some READMEs are long. Always pipe through `grep` to find relevant lines rather than dumping the entire content.
6. **Be thorough but efficient** — report `NONE` only after checking all 5 steps. Don't give false negatives.
7. **Clean up clones** — if you used `git clone --depth 1` as fallback, remove the temp directory when done.
