# GitHub Release Instructions

## Current Status

✅ **Code committed locally** with tag `v1.1.0`
✅ **Binary built** and ready: `s3-proxy` (16MB)
✅ **Release notes prepared** in `RELEASE.md`

## Steps to Push and Create Release

### 1. Authenticate with GitHub

You'll need to authenticate to push. Choose one method:

**Option A: Personal Access Token (HTTPS)**
```bash
git remote set-url origin https://YOUR_TOKEN@github.com/jcomo/s3-proxy.git
```

**Option B: SSH Key**
```bash
# Ensure your SSH key is added to GitHub
ssh -T git@github.com
git remote set-url origin git@github.com:jcomo/s3-proxy.git
```

### 2. Push Code and Tag

```bash
cd /home/linux/projects/s3-proxy

# Push main branch
git push -u origin main

# Push tag
git push origin v1.1.0
```

### 3. Create GitHub Release

**Via GitHub Web UI:**
1. Go to: https://github.com/jcomo/s3-proxy/releases/new
2. **Tag:** Select `v1.1.0` (or create new tag)
3. **Title:** `v1.1.0 - Multipart Upload Support`
4. **Description:** Copy from `RELEASE.md` file
5. **Attach binary:** Upload `s3-proxy` binary
6. Click **"Publish release"**

**Via GitHub CLI (if installed):**
```bash
gh release create v1.1.0 \
  --title "v1.1.0 - Multipart Upload Support" \
  --notes-file RELEASE.md \
  s3-proxy
```

## Release Assets

The following files are ready:
- **Binary:** `s3-proxy` (Linux x86-64, 16MB)
- **Release Notes:** `RELEASE.md`
- **Documentation:** `MULTIPART-UPLOADS.md`

## What's Included in v1.1.0

- ✅ Full S3 multipart upload API support
- ✅ POST request handling for multipart operations
- ✅ S3-compatible XML responses
- ✅ Fixes 405 Method Not Allowed errors
- ✅ All 5 multipart operations implemented:
  - CreateMultipartUpload
  - UploadPart
  - CompleteMultipartUpload
  - AbortMultipartUpload
  - ListMultipartUploads

## Verification

After pushing, verify:
```bash
# Check remote
git remote -v

# Check tags
git tag -l

# Check commits
git log --oneline -5
```

