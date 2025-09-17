# SSH Secret Keeper Site (Development)

Landing page for SSH Secret Keeper - currently in active development.

## Quick Start

```bash
# Start development server (default make target)
make

# Or explicitly
make dev
```

The development server will start at http://localhost:3000 and automatically open in your browser.

## Available Commands

### Development
- `make` or `make dev` - Start development server (default)
- `make install` - Install dependencies
- `make build` - Test build process
- `make preview` - Preview production build

### Code Quality
- `make lint` - Run ESLint
- `make lint-fix` - Run ESLint with auto-fix
- `make type-check` - Run TypeScript type check

### Maintenance
- `make clean` - Clean all build artifacts and dependencies
- `make clean-build` - Clean build directory only
- `make help` - Show all available commands

## Development Server

- **URL**: http://localhost:3000
- **Auto-reload**: Enabled for instant feedback
- **Host binding**: Enabled for network access (useful for mobile testing)
- **Auto-open**: Browser opens automatically

## Technology Stack

- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite 5
- **Styling**: Tailwind CSS
- **Icons**: Lucide React
- **Theme**: Terminal/hacker aesthetic

## Project Structure

```
site/
├── src/
│   ├── App.tsx          # Main application component
│   ├── main.tsx         # React entry point
│   ├── index.css        # Tailwind CSS imports
│   └── config.ts        # Development configuration
├── public/              # Static assets
├── dist/               # Build output (generated)
├── package.json        # Dependencies and scripts
├── vite.config.ts      # Vite configuration
├── tailwind.config.js  # Tailwind configuration
├── Makefile           # Development commands
└── README.md          # This file
```

## Development Notes

- **No versioning**: This is a development build without version constraints
- **Hot reload**: Changes are reflected immediately
- **Source maps**: Enabled for debugging
- **Linting**: ESLint configured for React + TypeScript
- **Type checking**: Full TypeScript support

## Common Development Tasks

```bash
# Start fresh development session
make clean
make dev

# Check code quality
make lint
make type-check

# Test build process
make build
make preview

# Fix common issues
make lint-fix
```

## Troubleshooting

### Port Already in Use
If port 3000 is busy, Vite will automatically try the next available port.

### Dependencies Issues
```bash
make clean
make install
```

### Build Issues
```bash
make clean-build
make build
```

## CI/CD Pipeline

The site has automated building and deployment through GitHub Actions:

### Pipeline Features
- **Quality Assurance**: ESLint and TypeScript type checking
- **Multi-Node Testing**: Tests against Node.js 18 and 20
- **Smart Caching**: Optimized dependency caching for faster builds
- **Path-Based Triggers**: Only runs when site files change
- **Artifact Storage**: Build artifacts retained for 30 days
- **GitHub Pages Deployment**: Automatic deployment on main branch

### Pipeline Triggers
- Push to `main` or `developer` branches (when site files change)
- Pull requests to `main` (when site files change)
- Manual workflow dispatch with optional deployment

### Build Status
[![Site Build](https://github.com/rafaelvzago/ssh-secret-keeper/actions/workflows/site.yml/badge.svg)](https://github.com/rafaelvzago/ssh-secret-keeper/actions/workflows/site.yml)

### Deployment
- **Production**: Automatically deployed to GitHub Pages on main branch pushes
- **Preview**: Build artifacts available for manual testing
- **URL**: Available at GitHub Pages URL once deployed

## Status

🚧 **In Active Development** - Site pipeline implemented with automated building and deployment capabilities.
