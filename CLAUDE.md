# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a kubectl plugins repository designed to host custom kubectl plugins. The repository follows Go conventions based on the .gitignore file.

## Project Structure

Since this is a kubectl plugins repository, each plugin should typically:
- Be placed in its own directory (e.g., `/kubectl-example`)
- Follow the kubectl plugin naming convention (prefix with `kubectl-`)
- Be executable and installable in the user's PATH

## Development Commands

### For Go-based plugins:
```bash
# Initialize a new Go module for a plugin
go mod init github.com/[username]/kubectl-plugins/[plugin-name]

# Build a plugin
go build -o kubectl-[plugin-name] ./[plugin-name]

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Format code
go fmt ./...

# Vet code
go vet ./...
```

### For shell-based plugins:
```bash
# Make plugin executable
chmod +x kubectl-[plugin-name]

# Test plugin locally
./kubectl-[plugin-name] [args]
```

## kubectl Plugin Development Guidelines

1. **Plugin Naming**: All kubectl plugins must be named with the prefix `kubectl-` (e.g., `kubectl-debug`, `kubectl-status`)

2. **Plugin Discovery**: kubectl automatically discovers plugins by searching for executables in PATH that start with `kubectl-`

3. **Plugin Structure**: 
   - Plugins can be written in any language (Go, Bash, Python, etc.)
   - Must be executable
   - Should handle `--help` flag
   - Should follow kubectl conventions for output formatting

4. **Testing Plugins**:
   - Place the plugin in your PATH or test locally with `./kubectl-[plugin-name]`
   - Once in PATH, can be invoked as `kubectl [plugin-name]`

## Important Notes

- The .gitignore is configured for Go development, suggesting Go as the primary language for plugins
- Test artifacts and coverage files are automatically ignored
- The repository currently has no existing plugins - new plugins should be created as needed