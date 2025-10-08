## Overview
shpack is a Go-based build tool that bundles shell scripts into a single, portable executable. It enables structured hierarchical scripting while providing easy distribution and management through a single binary.

## Goals
- Bundle multiple shell scripts into one executable
- Maintain script hierarchy and dependencies
- Zero external dependencies for end users
- Cross-platform support (Linux, macOS, BSD)

## Project Structure
```
myproject/
├── scripts/
│   ├── main.sh          # Entry point
│   ├── dependency.sh    # Supporting scripts
│   ├── utils.sh
│   └── ...
├── shpack.yaml          # Build configuration (optional)
└── build/               # Output directory
    └── mytool           # Final executable
```

## Build Process

### Configuration File: `shpack.yaml`
```yaml
name: mytool              # Output executable name
entry: scripts/main.sh    # Entry point script
scripts: scripts          # Scripts folder name
version: 1.0.0
```

### Environment
- `SHPACK_SCRIPT_DIR`: Path to temporary script directory
- `SHPACK_VERSION`: Version of bundled scripts
- All original environment variables passed through
- Scripts can source dependencies using relative paths

## Commands

### `shpack version`
Show shpack version.

### `shpack init`
Initialize new project with template structure.

### `shpack build`
Build the executable from project.

### `shpack make`
Quick build from folder of scripts, output the executable only.

## Dependencies

### Build Time
- Go 1.16+ (for `embed` package)
- YAML parser: `gopkg.in/yaml.v3`

### Runtime
- Bash shell (must be available on target system)
- No other dependencies

