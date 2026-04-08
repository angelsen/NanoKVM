# Makefile for NanoKVM Project

# Configuration
IMAGE_NAME := nanokvm-builder
UID := $(shell id -u)
GID := $(shell id -g)
PWD := $(shell pwd)

# Container runtime (docker, podman, nerdctl, etc). Override via env:
#   DOCKER_CMD=podman make app
DOCKER_CMD ?= docker

# Rootless podman maps in-container UIDs via subuid/subgid by default, so the
# bind mount isn't writable as the host user. --userns=keep-id maps the host
# UID/GID 1:1 inside the container. Docker uses a different userns model and
# does not accept this flag, so it's only set when DOCKER_CMD is podman.
USERNS_FLAG := $(shell [ "$(DOCKER_CMD)" = "podman" ] && echo "--userns=keep-id")

# Container run common parameters
DOCKER_RUN_BASE := $(DOCKER_CMD) run $(USERNS_FLAG) -e UID=$(UID) -e GID=$(GID) -v $(PWD):/home/build/NanoKVM --rm

# Build commands
# CGO_LDFLAGS adds an rpath so libkvm.so resolves at runtime relative to the
# binary's own location. Without it the dynamic loader fails to find
# libkvm.so on the device since dl_lib/ is not on any system search path.
# Use \$$ORIGIN: Make expands $$ → $, then bash double-quote treats \$ as
# literal $, so the linker receives the literal string $ORIGIN/dl_lib.
GO_BUILD_CMD := cd /home/build/NanoKVM/server && go mod tidy && CGO_ENABLED=1 GOOS=linux GOARCH=riscv64 CC=riscv64-unknown-linux-musl-gcc CGO_CFLAGS="-mcpu=c906fdv -march=rv64imafdcv0p7xthead -mcmodel=medany -mabi=lp64d" CGO_LDFLAGS="-Wl,-rpath,\$$ORIGIN/dl_lib" go build
SUPPORT_BUILD_CMD := . ./home/build/MaixCDK/bin/activate && cd /home/build/NanoKVM/support/sg2002 && ./build kvm_system && ./build kvm_system add_to_kvmapp

.PHONY: help check-root builder-image rebuild-image check-image shell app support all clean

# Default target
all: app support

# Help target
help:
	@echo "NanoKVM Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  help          - Show this help message"
	@echo "  check-image   - Check builder image and show versions"
	@echo "  builder-image - Build builder image if not exists"
	@echo "  rebuild-image - Force rebuild builder image"
	@echo "  shell         - Enter interactive builder environment"
	@echo "  app           - Build Go application server"
	@echo "  support       - Build hardware support libraries"
	@echo "  all           - Build both app and support (default)"
	@echo "  clean         - Clean build artifacts"
	@echo ""
	@echo "Prerequisites:"
	@echo "  - A container runtime (docker or podman) must be installed and running"
	@echo "  - Override with DOCKER_CMD=podman if docker is not present"
	@echo "  - Must not run as root user"

# Security check - prevent running as root
check-root:
	@if [ "$$(id -u)" -eq 0 ]; then \
		echo "Can't run as root"; \
		exit 1; \
	fi

# Check if builder image exists and show versions
check-image: check-root
	@echo "Checking builder image..."
	@echo "Golang version: " && \
		$(DOCKER_CMD) run --rm -i $(IMAGE_NAME) go version && \
		echo "" && \
		echo "Host-tools version:" && \
		$(DOCKER_CMD) run --rm -i $(IMAGE_NAME) riscv64-unknown-linux-musl-gcc -v && \
		echo ""

# Build builder image if it doesn't exist
builder-image: check-root
	@if ! $(DOCKER_CMD) image inspect $(IMAGE_NAME) >/dev/null 2>&1; then \
		echo "Building builder image..."; \
		$(DOCKER_CMD) build -t $(IMAGE_NAME) -f docker/Dockerfile ./; \
	else \
		echo "Builder image $(IMAGE_NAME) already exists."; \
	fi

# Force rebuild builder image
rebuild-image: check-root
	@echo "Force rebuilding builder image..."
	@$(DOCKER_CMD) build --no-cache -t $(IMAGE_NAME) -f docker/Dockerfile ./

# Enter interactive shell (equivalent to build.sh with no arguments)
shell: check-root builder-image
	@echo "Switching into builder..."
	@$(DOCKER_RUN_BASE) -it $(IMAGE_NAME) /bin/bash -c ". ./home/build/MaixCDK/bin/activate && cd /home/build/NanoKVM ; exec bash"

# Build Go application
app: check-root builder-image
	@echo "Building app..."
	@$(DOCKER_RUN_BASE) -it $(IMAGE_NAME) /bin/bash -c '$(GO_BUILD_CMD)'

# Build hardware support libraries
support: check-root builder-image
	@echo "Building support..."
	@$(DOCKER_RUN_BASE) -it $(IMAGE_NAME) /bin/bash -c '$(SUPPORT_BUILD_CMD)'

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@if [ -f server/NanoKVM-Server ]; then \
		rm -f server/NanoKVM-Server; \
		echo "Removed server/NanoKVM-Server"; \
	fi
	@if [ -d support/sg2002/build ]; then \
		rm -rf support/sg2002/build; \
		echo "Removed support/sg2002/build"; \
	fi
	@echo "Clean completed."