# 🧩 shpack

**shpack** is a Go-based build tool that bundles multiple shell scripts into a single, portable executable.  
It lets you organize scripts hierarchically, distribute them as one binary, and run them anywhere — no dependencies required.

---

## 🚀 Installation

```bash
brew tap luongnguyen1805/shpack
brew install shpack
````

*(Works on macOS and Linux via Homebrew.)*

---

## 💡 Commands

| Command                 | Description                                                      |
| ----------------------- | ---------------------------------------------------------------- |
| `shpack version`        | Show shpack version                                              |
| `shpack init {FOLDER}`  | Initialize a new project with a sample structure                 |
| `shpack build {FOLDER}` | Build an executable from the specified project folder            |
| `shpack make {FOLDER}`  | Quick build from a folder of scripts, output the executable only |

---

## 🧱 Overview

`shpack` enables developers to:

* Bundle multiple shell scripts into a single binary
* Preserve folder hierarchy and script dependencies
* Deliver zero-dependency executables to end users
* Support Linux, macOS, and BSD out of the box

---

## 📁 Project Structure

```
myproject/
├── scripts/
│   ├── main.sh          # Entry point
│   ├── dependency.sh    # Supporting scripts
│   ├── utils.sh
│   └── ...
├── shpack.yaml          # Optional build configuration
└── build/
    └── mytool           # Generated executable
```

---

## ⚙️ Build Configuration (`shpack.yaml`)

```yaml
name: mytool              # Output executable name
entry: scripts/main.sh    # Entry point script
scripts: scripts          # Directory containing scripts
version: 1.0.0
```

---

## 🌍 Runtime Environment

When the executable runs, shpack provides the following environment variables:

| Variable                                               | Description                                  |
| ------------------------------------------------------ | -------------------------------------------- |
| `SHPACK_SCRIPT_DIR`                                    | Path to extracted temporary script directory |
| `SHPACK_VERSION`                                       | Version of the bundled scripts               |
| *(all other user environment variables are preserved)* |                                              |

> Inside your scripts, you can safely `source ./env.sh` or use relative paths — they resolve within `SHPACK_SCRIPT_DIR`.

---

## 🧰 Dependencies

### Build-time

* Go **1.16+** (for `embed` support)
* YAML parser: [`gopkg.in/yaml.v3`](https://pkg.go.dev/gopkg.in/yaml.v3)

### Runtime

* Bash shell (`/bin/bash` or equivalent)
* No additional dependencies

---

## 📦 Example Workflow

```bash
# Initialize project
shpack init sample

# Edit scripts inside sample/scripts/

# Build executable
shpack build sample

# Run built tool
./sample/build/mytool
```

---

## 🧾 License

None