# Contributing to Genrify Web

Thank you for your interest in contributing to Genrify Web!

## Development Setup

```bash
# Clone the repo
git clone https://github.com/yourusername/genrify.git
cd genrify/web

# Install dependencies
npm install

# Start dev server
npm run dev
```

## Code Style

- **TypeScript**: Strict mode enabled
- **Formatting**: Prettier (run `npm run format`)
- **Linting**: ESLint (run `npm run lint`)
- **Naming**: camelCase for variables/functions, PascalCase for components

## Testing

```bash
# Run tests
npm test

# Run tests with UI
npm test:ui

# Run tests in watch mode
npm test -- --watch
```

All new features should include tests in `src/__tests__/`.

## Submitting Changes

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Run tests: `npm test`
5. Run linter: `npm run lint`
6. Format code: `npm run format`
7. Commit: `git commit -m "Add my feature"`
8. Push: `git push origin feature/my-feature`
9. Open a Pull Request

## Commit Messages

Follow conventional commits:

- `feat: Add playlist export`
- `fix: Handle empty playlist names`
- `docs: Update README with examples`
- `test: Add tests for merge service`
- `refactor: Simplify auth context`
- `style: Fix formatting in DataTable`

## Architecture

```
src/
├── types/       # TypeScript types
├── lib/         # Core logic (ported from Go)
├── contexts/    # React contexts
├── hooks/       # TanStack Query hooks
├── components/  # Shared components
└── pages/       # Route pages
```

## Porting from Go

When porting features from the Go codebase:

1. Find the Go source file
2. Port types first (e.g., `types.go` → `types/spotify.ts`)
3. Port core logic (e.g., `service.go` → `lib/playlist-service.ts`)
4. Add React hooks if needed
5. Create UI components
6. Add tests mirroring Go tests

## Guidelines

- ✅ Keep library layer (`lib/`) React-free
- ✅ Use TypeScript's strict mode
- ✅ Add JSDoc comments to public functions
- ✅ Use CSS Modules for styling
- ✅ Follow existing patterns
- ✅ Write tests for new features
- ❌ Don't add new dependencies without discussion
- ❌ Don't break existing APIs
- ❌ Don't commit `node_modules/` or `dist/`

## Questions?

Open an issue or discussion on GitHub!
