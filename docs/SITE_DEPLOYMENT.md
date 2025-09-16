# Site Deployment Guide

This document covers the setup and deployment of the SSH Secret Keeper landing page using GitHub Actions and GitHub Pages.

## Overview

The site is built using a modern React/TypeScript stack and deployed automatically via GitHub Actions:

- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite 5
- **Styling**: Tailwind CSS
- **CI/CD**: GitHub Actions
- **Hosting**: GitHub Pages

## GitHub Actions Pipeline

### Workflow File
The deployment pipeline is defined in `.github/workflows/site.yml` with three main jobs:

1. **Quality Assurance** (`site-quality`)
   - ESLint code quality checks
   - TypeScript type checking
   - Node.js dependency caching

2. **Build** (`site-build`)
   - Production build generation
   - Build artifact upload (30-day retention)
   - Build summary generation

3. **Deploy** (`site-deploy`)
   - GitHub Pages deployment (main branch only)
   - Automatic URL generation
   - Deployment status reporting

### Trigger Conditions

The pipeline runs when:
- **Push Events**: To `main` or `developer` branches (only when site files change)
- **Pull Requests**: To `main` branch (only when site files change)
- **Manual Trigger**: Via workflow dispatch with optional deployment flag

### Path-Based Filtering

The workflow uses intelligent path filtering to only run when relevant files change:
```yaml
paths:
  - 'site/**'
  - '.github/workflows/site.yml'
```

## GitHub Pages Setup

### Repository Configuration

1. **Enable GitHub Pages**:
   - Go to repository Settings → Pages
   - Set Source to "GitHub Actions"
   - No branch selection needed (handled by workflow)

2. **Required Permissions**:
   The workflow requires these permissions:
   ```yaml
   permissions:
     contents: read
     pages: write
     id-token: write
   ```

3. **Environment Setup**:
   - Environment name: `github-pages`
   - Automatically configured by Actions

### Deployment Process

1. **Automatic Deployment**:
   - Triggers on push to `main` branch
   - Only when site files are modified
   - Build artifacts uploaded to Pages

2. **Manual Deployment**:
   ```bash
   # Trigger manual deployment via GitHub UI
   # Actions tab → Site Build Pipeline → Run workflow
   # Enable "Deploy to GitHub Pages" option
   ```

3. **Deployment URL**:
   - Available at: `https://{username}.github.io/{repository-name}/`
   - URL provided in workflow summary

## Local Development

### Prerequisites
- Node.js 18+ (20 recommended)
- npm (comes with Node.js)

### Development Workflow
```bash
# Navigate to site directory
cd site/

# Install dependencies
make install

# Start development server (http://localhost:3000)
make dev

# Build for production testing
make build

# Preview production build
make preview

# Run quality checks
make lint
make type-check

# Auto-fix linting issues
make lint-fix
```

### Development Features
- **Hot Reload**: Instant updates on file changes
- **Auto-open**: Browser opens automatically
- **Network Access**: Accessible from other devices (useful for mobile testing)
- **Source Maps**: Full debugging support

## Build Process

### Build Configuration
The build process is configured in `vite.config.ts`:
```typescript
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist',
    sourcemap: true,
  }
});
```

### Build Outputs
- **Location**: `site/dist/`
- **Assets**: Optimized HTML, CSS, JavaScript
- **Source Maps**: Included for debugging
- **Compression**: Automatic asset optimization

### Build Verification
```bash
# Local build test
cd site/
make build

# Check build contents
ls -la dist/

# Test build locally
make preview
```

## Monitoring and Troubleshooting

### Build Status
Monitor build status via:
- **GitHub Actions Tab**: Real-time build progress
- **README Badge**: Build status indicator
- **Workflow Summaries**: Detailed build reports

### Common Issues

1. **Build Failures**:
   ```bash
   # Check linting issues
   make lint

   # Fix auto-fixable issues
   make lint-fix

   # Check TypeScript errors
   make type-check
   ```

2. **Deployment Failures**:
   - Verify GitHub Pages is enabled
   - Check repository permissions
   - Review workflow logs in Actions tab

3. **Cache Issues**:
   - GitHub Actions automatically handles Node.js caching
   - Local cache issues: `make clean && make install`

### Performance Monitoring
- **Build Time**: Typically <2 minutes
- **Artifact Size**: Monitored in workflow summaries
- **Cache Efficiency**: Node.js dependencies cached between runs

## Security Considerations

### Workflow Security
- **Permissions**: Minimal required permissions only
- **Secrets**: No secrets required for public repositories
- **Environment**: Isolated GitHub Pages environment

### Content Security
- **Static Site**: No server-side code execution
- **HTTPS**: Automatic HTTPS via GitHub Pages
- **Dependencies**: Regular security updates via Dependabot

## Integration with Main Project

### Documentation Updates
- Main README includes site section
- Site README includes CI/CD information
- Build status badges in both locations

### Release Integration
The site pipeline is independent but can be integrated with releases:
```bash
# Manual deployment for releases
# Use workflow dispatch with deployment flag
```

### Branch Strategy
- **Development**: `developer` branch builds but doesn't deploy
- **Production**: `main` branch builds and deploys automatically
- **Pull Requests**: Build verification without deployment

## Future Enhancements

### Potential Improvements
1. **Preview Deployments**: Deploy PR previews to separate URLs
2. **Custom Domain**: Configure custom domain for production
3. **CDN Integration**: Add CloudFlare or similar CDN
4. **Analytics**: Add privacy-friendly analytics
5. **Performance**: Implement advanced build optimizations

### Monitoring Enhancements
1. **Lighthouse CI**: Automated performance testing
2. **Bundle Analysis**: Size and dependency tracking
3. **Uptime Monitoring**: Site availability checks
4. **Error Tracking**: Client-side error monitoring

## Support

### Getting Help
- **Build Issues**: Check Actions tab for detailed logs
- **Local Development**: See `site/README.md`
- **GitHub Pages**: Consult GitHub Pages documentation
- **React/Vite**: Reference official documentation

### Contributing
1. Test changes locally with `make dev`
2. Run quality checks: `make lint && make type-check`
3. Create PR to `main` branch
4. Verify build passes in Actions
5. Deployment happens automatically on merge
