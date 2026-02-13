# Deploying Genrify Web

This guide covers deploying Genrify Web to various platforms.

## GitHub Pages

### Automatic Deployment

The repo includes a GitHub Actions workflow that automatically deploys to GitHub Pages on every push to `main` that changes files in the `web/` directory.

### Setup

1. **Enable GitHub Pages in your repository:**
   - Go to Settings → Pages
   - Source: "GitHub Actions"
   - Save

2. **Update Spotify Redirect URI:**
   - Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
   - Edit your app
   - Add redirect URI: `https://yourusername.github.io/genrify/callback`
   - Save

3. **Configure Base Path (if needed):**

   If your GitHub Pages is at `https://yourusername.github.io/genrify/` (project pages):
   - Edit `.github/workflows/deploy-web.yml`
   - Uncomment and set `BASE_URL: /genrify/` in the build step

   If your GitHub Pages is at `https://yourusername.github.io/` (user/org pages):
   - No changes needed (base is `/` by default)

4. **Push to main:**
   ```bash
   git add .
   git commit -m "Deploy web to GitHub Pages"
   git push
   ```

5. **Wait for deployment:**
   - Go to Actions tab in GitHub
   - Watch the "Deploy Web to GitHub Pages" workflow
   - Once complete, visit your GitHub Pages URL

### Manual Deployment

```bash
cd web

# Build for GitHub Pages with base path
BASE_URL=/genrify/ npm run build

# Or use the npm script
npm run build:gh-pages

# Deploy the dist/ folder to gh-pages branch using gh-pages package
npx gh-pages -d dist
```

## Netlify

### Via UI

1. Connect your GitHub repo to Netlify
2. Configure build settings:
   - **Base directory:** `web`
   - **Build command:** `npm run build`
   - **Publish directory:** `web/dist`
3. Add environment variables (if needed):
   - `BASE_URL=/` (default)
4. Deploy

### Via CLI

```bash
cd web
npm install -g netlify-cli
npm run build
netlify deploy --prod --dir=dist
```

### Redirect URI

Add to your Spotify app:
```
https://your-app.netlify.app/callback
```

## Vercel

### Via UI

1. Import your GitHub repo
2. Configure:
   - **Framework Preset:** Vite
   - **Root Directory:** `web`
   - **Build Command:** `npm run build`
   - **Output Directory:** `dist`
3. Deploy

### Via CLI

```bash
cd web
npm install -g vercel
npm run build
vercel --prod
```

### Redirect URI

Add to your Spotify app:
```
https://your-app.vercel.app/callback
```

## Custom Domain

If deploying to a custom domain (e.g., `https://genrify.example.com`):

1. **No base path needed:**
   ```bash
   BASE_URL=/ npm run build
   ```

2. **Update Spotify redirect URI:**
   ```
   https://genrify.example.com/callback
   ```

3. **Configure HTTPS:**
   - Spotify requires HTTPS for redirect URIs
   - Most platforms (Netlify, Vercel, GitHub Pages) provide HTTPS automatically

## Docker

Create `web/Dockerfile`:

```dockerfile
FROM node:20-alpine as builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

Build and run:

```bash
cd web
docker build -t genrify-web .
docker run -p 8080:80 genrify-web
```

## Environment Variables

### Build Time

- `BASE_URL` - Base path for the app (default: `/`)
  - GitHub Pages project: `/genrify/`
  - GitHub Pages user: `/`
  - Custom domain: `/`

### Runtime (in browser)

All configuration is done through the Settings dialog in the app:
- Spotify Client ID
- Redirect URI
- Scopes

No server-side environment variables needed!

## Troubleshooting

### 404 on refresh

Single-page apps need special routing configuration:

**Netlify** - Create `web/public/_redirects`:
```
/*    /index.html   200
```

**Vercel** - Create `web/vercel.json`:
```json
{
  "rewrites": [{ "source": "/(.*)", "destination": "/" }]
}
```

**GitHub Pages** - Create `web/public/404.html` (copy of `index.html`)

**Nginx** - Add to `nginx.conf`:
```nginx
location / {
  try_files $uri $uri/ /index.html;
}
```

### OAuth redirect fails

1. Check redirect URI matches exactly in Spotify app settings
2. Check HTTPS is enabled
3. Check base path is correct
4. Check for CORS issues (shouldn't be any - API calls go direct to Spotify)

### Assets not loading

1. Check `BASE_URL` environment variable during build
2. Check browser console for 404 errors
3. Verify paths in built `index.html` match deployment path

## Security Checklist

- [ ] HTTPS enabled
- [ ] Redirect URI matches exactly
- [ ] Client ID is public (not a secret)
- [ ] No API keys or secrets in code
- [ ] Content Security Policy configured (optional)
- [ ] CORS not needed (direct API calls)

## Post-Deployment

After deploying:

1. Visit your deployed URL
2. Open Settings
3. Configure Spotify Client ID
4. Set correct Redirect URI
5. Save
6. Test login flow
7. Verify all features work

## Cost

All recommended platforms have generous free tiers:
- **GitHub Pages:** Free for public repos
- **Netlify:** 100GB bandwidth/month free
- **Vercel:** 100GB bandwidth/month free

Perfect for personal use or small teams!
