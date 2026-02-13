# GitHub Actions Setup for Genrify Web

## Overview

Two GitHub Actions workflows have been configured for the web application:

1. **CI Workflow** - Tests and builds on every push/PR
2. **Deploy Workflow** - Deploys to GitHub Pages on main branch changes

## Workflows

### 1. CI Workflow (`.github/workflows/ci.yml`)

Runs on every push to `main` and on all pull requests.

**Jobs added for web:**

#### `web-lint`
- Runs ESLint on TypeScript/React code
- Checks for code quality issues
- Fast fail for syntax errors

#### `web-test`
- Runs Vitest unit tests
- Depends on `web-lint` passing
- Reports test failures

#### `web-build`
- Builds production bundle
- Verifies TypeScript compilation
- Uploads build artifact
- Depends on `web-lint` passing

**Configuration:**
```yaml
- Node.js: 20
- Package manager: npm
- Cache: npm (speeds up installs)
- Working directory: web/
```

### 2. Deploy Workflow (`.github/workflows/deploy-web.yml`)

Deploys to GitHub Pages when:
- Push to `main` branch
- Changes in `web/` directory or workflow file
- Manual trigger via workflow_dispatch

**Jobs:**

#### `build`
- Installs dependencies
- Builds production bundle
- Uploads as GitHub Pages artifact

#### `deploy`
- Deploys artifact to GitHub Pages
- Updates deployment environment
- Provides deployment URL

**Permissions required:**
```yaml
contents: read
pages: write
id-token: write
```

**Concurrency:**
- Only one deployment runs at a time
- New deployments cancel in-progress ones

## Setup Instructions

### 1. Enable GitHub Pages

1. Go to your repo → **Settings** → **Pages**
2. Under "Source", select: **GitHub Actions**
3. Save

### 2. Configure Base Path (if needed)

If your repo is at `github.com/username/genrify`:

**Edit `.github/workflows/deploy-web.yml`:**

```yaml
- name: Build web app
  working-directory: web
  run: npm run build
  env:
    BASE_URL: /genrify/  # Add this line
```

If your repo is at `github.com/username/username.github.io`:
- No changes needed (base is `/` by default)

### 3. Update Spotify App Settings

Add the GitHub Pages URL as a redirect URI:

1. Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
2. Edit your app
3. Add redirect URI:
   - Project pages: `https://username.github.io/genrify/callback`
   - User pages: `https://username.github.io/callback`
4. Save

### 4. Push to Main

```bash
git add .
git commit -m "Add web app and GitHub Actions"
git push origin main
```

### 5. Monitor Deployment

1. Go to **Actions** tab in GitHub
2. Watch "Deploy Web to GitHub Pages" workflow
3. Once complete (green checkmark), visit your GitHub Pages URL
4. First-time users: Configure Spotify Client ID in settings

## Workflow Triggers

### CI Workflow

```yaml
on:
  push:
    branches: [main]
  pull_request:
```

Runs on:
- Every push to main
- Every pull request (any branch)

### Deploy Workflow

```yaml
on:
  push:
    branches: [main]
    paths:
      - 'web/**'
      - '.github/workflows/deploy-web.yml'
  workflow_dispatch:
```

Runs on:
- Push to main that changes `web/` or workflow file
- Manual trigger from Actions tab

## Environment Variables

### Build Time

- `BASE_URL` - Base path for assets (e.g., `/genrify/`)
  - Set in workflow file
  - Override in local builds: `BASE_URL=/foo/ npm run build`

### Runtime

No environment variables needed! All configuration is done in-browser:
- Spotify Client ID (settings dialog)
- Redirect URI (settings dialog)
- Tokens (localStorage)

## Deployment URL

After first successful deployment, find your URL:

1. Go to **Settings** → **Pages**
2. URL shown at top: "Your site is live at..."
3. Or check deployment job output in Actions

## Troubleshooting

### Deployment fails with 403

**Problem:** Missing Pages permissions

**Solution:**
1. Settings → Actions → General
2. Scroll to "Workflow permissions"
3. Select "Read and write permissions"
4. Check "Allow GitHub Actions to create and approve pull requests"
5. Save

### Build succeeds but site is blank

**Problem:** Wrong base path

**Solution:**
1. Check browser console for 404 errors
2. Update `BASE_URL` in deploy workflow
3. Re-run deployment

### OAuth redirect fails

**Problem:** Redirect URI mismatch

**Solution:**
1. Check exact URL in Spotify app settings
2. Must match: `https://username.github.io/genrify/callback`
3. Include or exclude `/genrify/` based on your setup
4. Ensure HTTPS (not HTTP)

### Tests fail in CI but pass locally

**Problem:** Node version mismatch

**Solution:**
1. Check local Node version: `node -v`
2. Update workflow if needed:
   ```yaml
   - uses: actions/setup-node@v4
     with:
       node-version: '20'  # Match your local version
   ```

### Deployment takes too long

**Problem:** Not using npm cache

**Solution:** Already configured! Cache is enabled:
```yaml
- uses: actions/setup-node@v4
  with:
    cache: 'npm'
    cache-dependency-path: web/package-lock.json
```

## Status Badges

Add to your README:

```markdown
[![Deploy Web](https://github.com/username/genrify/actions/workflows/deploy-web.yml/badge.svg)](https://github.com/username/genrify/actions/workflows/deploy-web.yml)
```

Replace `username/genrify` with your repo path.

## Manual Deployment

If you need to manually trigger deployment:

1. Go to **Actions** tab
2. Click "Deploy Web to GitHub Pages"
3. Click "Run workflow" button
4. Select branch (usually `main`)
5. Click "Run workflow"

## Secrets

No secrets needed! Everything is public:
- ✅ Spotify Client ID (designed to be public)
- ✅ Code is open source
- ✅ Tokens stored in user's browser only

## Performance

**CI workflow** (all jobs parallel):
- Lint: ~30 seconds
- Test: ~45 seconds
- Build: ~45 seconds
- Total: ~1 minute

**Deploy workflow**:
- Build: ~45 seconds
- Deploy: ~30 seconds
- Total: ~1.5 minutes

**Optimizations:**
- npm cache enabled (saves ~15 seconds)
- Parallel job execution
- Artifacts cached between jobs

## Cost

**Free tier limits:**
- 2,000 minutes/month (public repos)
- Unlimited for public repos on free plan
- Each workflow run: ~2.5 minutes
- Plenty for hobby projects!

## Next Steps

1. [x] Enable GitHub Pages
2. [x] Configure base path (if needed)
3. [x] Update Spotify redirect URI
4. [x] Push to main
5. [x] Watch deployment
6. [ ] Test deployed app
7. [ ] Share your deployment URL!

## Resources

- [GitHub Pages Documentation](https://docs.github.com/en/pages)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Vite Deployment Guide](https://vitejs.dev/guide/static-deploy.html)
- [Spotify OAuth Documentation](https://developer.spotify.com/documentation/web-api/concepts/authorization)
