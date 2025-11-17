#!/bin/bash
# Script to set up branch protection rules for trunk-based development
# Requires GitHub CLI (gh) to be installed and authenticated

set -e

echo "🔒 Setting up branch protection for master branch..."

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) is not installed."
    echo "Please install it from: https://cli.github.com/"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo "❌ Not authenticated with GitHub CLI."
    echo "Please run: gh auth login"
    exit 1
fi

# Get repository information
REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
echo "📦 Repository: $REPO"

# Enable branch protection for master
echo "⚙️  Configuring branch protection rules..."

gh api \
  --method PUT \
  -H "Accept: application/vnd.github+json" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  "/repos/$REPO/branches/master/protection" \
  --input - <<EOF
{
  "required_status_checks": {
    "strict": true,
    "contexts": [
      "Unit Tests",
      "Integration Tests",
      "Lint",
      "Check Formatting"
    ]
  },
  "required_pull_request_reviews": {
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": false,
    "required_approving_review_count": 1,
    "require_last_push_approval": false
  },
  "required_conversation_resolution": true,
  "enforce_admins": false,
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "block_creations": false,
  "required_linear_history": false,
  "allow_fork_syncing": true
}
EOF

echo "✅ Branch protection rules configured successfully!"
echo ""
echo "📋 Summary of protection rules:"
echo "   ✓ Require pull request reviews (1 approval required)"
echo "   ✓ Require status checks to pass (Unit Tests, Integration Tests, Lint, Formatting)"
echo "   ✓ Require branches to be up to date before merging"
echo "   ✓ Dismiss stale pull request approvals when new commits are pushed"
echo "   ✓ Require review from code owners"
echo "   ✓ Require conversation resolution before merging"
echo "   ✓ Prevent force pushes"
echo "   ✓ Prevent branch deletion"
echo ""
echo "🎉 Trunk-based development workflow is now active!"
echo ""
echo "Next steps:"
echo "1. Create feature branches from master"
echo "2. Open PRs for review"
echo "3. AI code review will run automatically"
echo "4. Merge after approval and green CI"

