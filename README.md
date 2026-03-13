# LLM100X Tester

Automated testing tool for the LLM100X course.

## Option 1: Build from Source

```bash
git clone https://github.com/tensorhero/llm100x-tester
cd llm100x-tester
go build .
./llm100x-tester -s hello -d ~/my-solution/hello
```

**Dependencies:** Go 1.24+, clang, python3, sqlite3

## Option 2: Docker Image

**Quick Start**

```bash
cd ~/my-solution  # your solution root directory
docker pull tensorhero/llm100x-tester
docker run --rm --user $(id -u):$(id -g) -v "$(pwd):/workspace" tensorhero/llm100x-tester -s hello -d /workspace/hello
```

**Simplified script (recommended)**

Create `test.sh` in your solution root:

```bash
#!/bin/bash
docker run --rm --user $(id -u):$(id -g) -v "$(pwd):/workspace" tensorhero/llm100x-tester \
  -s "${1:-hello}" -d "/workspace/${1:-hello}"
```

Usage: `chmod +x test.sh && ./test.sh hello`

**Local build (optional)**

```bash
git clone https://github.com/tensorhero/llm100x-tester
cd llm100x-tester
docker build -t my-tester .
# Usage: docker run --rm --user $(id -u):$(id -g) -v ~/my-solution:/workspace my-tester -s hello -d /workspace/hello
```

## License

MIT
