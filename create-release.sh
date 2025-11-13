#!/bin/bash
# Script to help create GitHub release for s3-proxy v1.1.0

set -e

echo "=== S3-Proxy Release Helper ==="
echo ""
echo "This script will help you push code and create a GitHub release."
echo ""

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "Error: Must run from s3-proxy directory"
    exit 1
fi

# Check git status
echo "📋 Current git status:"
git status --short
echo ""

# Check if tag exists
if git tag -l | grep -q "v1.1.0"; then
    echo "✅ Tag v1.1.0 exists"
else
    echo "❌ Tag v1.1.0 not found"
    exit 1
fi

# Check remote
echo "📡 Remote configuration:"
git remote -v
echo ""

# Check if binary exists
if [ -f "s3-proxy" ]; then
    echo "✅ Binary s3-proxy exists ($(du -h s3-proxy | cut -f1))"
else
    echo "❌ Binary s3-proxy not found. Building..."
    go build -o s3-proxy .
    echo "✅ Binary built"
fi
echo ""

echo "🚀 Ready to push! Run these commands:"
echo ""
echo "1. Push main branch:"
echo "   git push -u origin main"
echo ""
echo "2. Push tag:"
echo "   git push origin v1.1.0"
echo ""
echo "3. Create GitHub release:"
echo "   - Go to: https://github.com/jcomo/s3-proxy/releases/new"
echo "   - Select tag: v1.1.0"
echo "   - Title: v1.1.0 - Multipart Upload Support"
echo "   - Description: Copy from RELEASE.md"
echo "   - Attach binary: s3-proxy"
echo "   - Click 'Publish release'"
echo ""
echo "Or use GitHub CLI:"
echo "   gh release create v1.1.0 --title 'v1.1.0 - Multipart Upload Support' --notes-file RELEASE.md s3-proxy"
echo ""

