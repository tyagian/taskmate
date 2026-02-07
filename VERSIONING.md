# Versioning Guide

TaskMate follows [Semantic Versioning](https://semver.org/) with beta releases during development.

## Version Format

- **Beta releases**: `v0.x.y-beta.z` (e.g., `v0.1.0-beta.1`, `v0.2.0-beta.3`)
- **Stable releases**: `v1.x.y` (e.g., `v1.0.0`, `v1.2.3`)

## Beta Stage (Current)

While in beta (v0.x.x), we use the following convention:

- **v0.MINOR.PATCH-beta.BUILD**
  - `MINOR`: Incremented for new features or significant changes
  - `PATCH`: Incremented for bug fixes and minor improvements
  - `BUILD`: Beta build number (incremented for each beta release)

### Examples:

- `v0.1.0-beta.1` - First beta release with initial features
- `v0.1.0-beta.2` - Second beta with bug fixes
- `v0.2.0-beta.1` - New features added, new minor version
- `v0.2.1-beta.1` - Bug fix release

## Creating a Release

### 1. Beta Release

```bash
# For a new feature
git tag v0.2.0-beta.1
git push origin v0.2.0-beta.1

# For a bug fix
git tag v0.1.1-beta.1
git push origin v0.1.1-beta.1

# For subsequent beta builds
git tag v0.2.0-beta.2
git push origin v0.2.0-beta.2
```

### 2. Stable Release (Future)

When ready for production:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Automated Release Process

When you push a tag, GitHub Actions will automatically:

1. Run all tests and linting
2. Build binaries for multiple platforms:
   - Linux (AMD64, ARM64)
   - macOS (AMD64, ARM64)
   - Windows (AMD64)
3. Create a GitHub release with:
   - Changelog from commits
   - Downloadable binaries
   - Release notes
4. Build and push Docker images with appropriate tags

## Version Bumping Guidelines

### Increment MINOR (0.X.0) when:
- Adding new API endpoints
- Adding new features
- Making significant changes to existing functionality
- Breaking changes (during beta)

### Increment PATCH (0.1.X) when:
- Fixing bugs
- Making minor improvements
- Updating documentation
- Refactoring without behavior changes

### Increment BUILD (0.1.0-beta.X) when:
- Testing releases
- Iterating on the same feature set
- Making experimental changes

## Transition to Stable

When transitioning from beta to stable (v1.0.0):

1. Complete all planned features
2. Fix all critical bugs
3. Complete documentation
4. Run extensive testing
5. Tag as `v1.0.0`

After v1.0.0, follow strict semantic versioning:
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

## Current Roadmap

- **v0.1.0-beta.x**: Initial release with core features
  - Task CRUD operations
  - Token-based authentication
  - Web UI
  - REST API

- **v0.2.0-beta.x**: Enhanced features
  - Task filtering and search
  - User management
  - API rate limiting

- **v1.0.0**: First stable release
  - Production-ready
  - Complete documentation
  - Comprehensive test coverage
