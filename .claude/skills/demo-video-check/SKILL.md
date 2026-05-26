---
name: demo-video-check
description: Check GitHub/Gitee repositories for demo videos, presentation videos, or walkthrough videos in any form (video files, external links to Bilibili/YouTube, cloud storage links like Baidu Pan/Google Drive). Use this skill when the user needs to verify whether a repo has a demo video.
---

# Demo Video Check

Check whether a GitHub or Gitee repository has a demo video, presentation video, or walkthrough in **any form** — video files committed to the repo, external links to video platforms, or cloud storage shares.

## Strategy: Remote Inspection First (No Clone Required)

**Never clone the repo.** Cloning is slow and unnecessary. Use the GitHub/Gitee API or `gh` CLI to inspect remotely. Only for repos with committed video files should you consider a shallow clone.

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

Scan the README for:

### Video platform links
```
bilibili.com  b23.tv  youtube.com  youtu.be  vimeo.com
```

### Cloud storage / file share links
```
pan.baidu.com   drive.google.com   aliyundrive.com
lanzou (蓝奏云)   weiyun.com   115.com   mega.nz
dropbox.com
```

### Generic demo keywords in link context
```
演示  视频  demo  预览  preview  walkthrough  教程  tutorial
```

Search patterns to use:
```bash
# Video platforms
grep -iE 'bilibili|b23\.tv|youtube|youtu\.be|vimeo'

# Cloud storage / file sharing
grep -iE 'pan\.baidu|drive\.google|aliyundrive|lanzou|weiyun|mega\.nz|dropbox|115\.com'

# Any URL near demo keywords
grep -iE 'https?://[^ ]*' | grep -iE 'demo|视频|演示|preview|walkthrough|tutorial|教程'

# .zip / .rar / .7z archive links (might be video archives)
grep -iE 'https?://[^ ]*\.(zip|rar|7z)'
```

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

1. **Always check remotely first** — `gh api` is fast, `git clone` is slow and may fail due to network. Only clone as a last resort for repos with committed video files you need to verify.
2. **Token is optional** — public repos can be checked without a token, but you'll hit rate limits quickly. Suggest the user set `GH_TOKEN` if checking many repos.
3. **Gitee repos** — the approach is the same, but use the Gitee API endpoints. `gh api` only works for GitHub. For Gitee, use `curl` with `https://gitee.com/api/v5/repos/{owner}/{repo}/...` and add `?access_token=xxx` for auth.
4. **Multi-line README** — some READMEs are long. Always pipe through `grep` to find relevant lines rather than dumping the entire content.
5. **Be thorough but efficient** — report `NONE` only after checking all 5 steps. Don't give false negatives.
