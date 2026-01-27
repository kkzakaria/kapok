# Kapok Installation Guide

Detailed installation instructions for all platforms.

## Table of Contents

- [macOS](#macos)
- [Linux](#linux)
- [Windows (WSL)](#windows-wsl)
- [Building from Source](#building-from-source)
- [Docker](#docker-coming-soon)
- [Verification](#verification)

---

## macOS

### Option 1: Install with Go (Recommended)

**Prerequisites**: Go 1.21+ installed

```bash
go install github.com/kapok/kapok/cmd/kapok@latest
```

The binary will be installed to `$GOPATH/bin` (usually `~/go/bin`).

**Add to PATH** (if not already):

```bash
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```

### Option 2: Homebrew (Coming Soon)

```bash
brew install kapok
```

### Option 3: Download Binary

Download the latest release for macOS:

```bash
# For Apple Silicon (M1/M2/M3)
curl -LO https://github.com/kapok/kapok/releases/latest/download/kapok-darwin-arm64
chmod +x kapok-darwin-arm64
sudo mv kapok-darwin-arm64 /usr/local/bin/kapok

# For Intel Macs
curl -LO https://github.com/kapok/kapok/releases/latest/download/kapok-darwin-amd64
chmod +x kapok-darwin-amd64
sudo mv kapok-darwin-amd64 /usr/local/bin/kapok
```

---

## Linux

### Option 1: Install with Go (Recommended)

**Prerequisites**: Go 1.21+ installed

```bash
go install github.com/kapok/kapok/cmd/kapok@latest
```

**Add to PATH** (if not already):

```bash
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

### Option 2: Download Binary

#### Ubuntu/Debian

```bash
# Download
curl -LO https://github.com/kapok/kapok/releases/latest/download/kapok-linux-amd64
chmod +x kapok-linux-amd64
sudo mv kapok-linux-amd64 /usr/local/bin/kapok

# Verify
kapok version
```

#### Fedora/RHEL/CentOS

```bash
# Download
curl -LO https://github.com/kapok/kapok/releases/latest/download/kapok-linux-amd64
chmod +x kapok-linux-amd64
sudo mv kapok-linux-amd64 /usr/local/bin/kapok

# Verify
kapok version
```

#### Arch Linux (AUR - Coming Soon)

```bash
yay -S kapok
```

---

## Windows (WSL)

Kapok runs best on Windows using Windows Subsystem for Linux (WSL).

### Step 1: Install WSL

If you haven't already:

```powershell
wsl --install
```

Restart your computer when prompted.

### Step 2: Install Go in WSL

Open WSL terminal and install Go:

```bash
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Step 3: Install Kapok

```bash
go install github.com/kapok/kapok/cmd/kapok@latest
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

### Step 4: Install PostgreSQL in WSL

```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo service postgresql start
```

---

## Building from Source

### Prerequisites

- Go 1.21 or higher
- Git

### Steps

```bash
# Clone the repository
git clone https://github.com/kapok/kapok.git
cd kapok

# Build
go build -o kapok ./cmd/kapok

# Install to $GOPATH/bin
go install ./cmd/kapok

# Or move to /usr/local/bin
sudo mv kapok /usr/local/bin/
```

### Development Build

For development with additional tooling:

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build with race detector
go build -race -o kapok ./cmd/kapok

# Run
./kapok version
```

---

## Docker (Coming Soon)

Docker support is planned for a future release.

Expected usage:

```bash
docker run -p 8080:8080 kapok/kapok
```

---

## Verification

After installation, verify Kapok is working:

```bash
# Check version
kapok version

# Check available commands
kapok --help

# Test init (creates test project)
mkdir test-kapok && cd test-kapok
kapok init my-test
ls -la
```

You should see:

```
Kapok version X.X.X
```

### System Requirements

**Minimum**:

- CPU: 1 core
- RAM: 512 MB
- Disk: 100 MB

**Recommended**:

- CPU: 2+ cores
- RAM: 2 GB+
- Disk: 1 GB+

### Runtime Dependencies

Kapok also requires:

- **PostgreSQL 12+** (running locally or remotely)
- **Node.js 18+** (only for SDK generation)

Install PostgreSQL:

- **macOS**: `brew install postgresql`
- **Ubuntu**: `sudo apt install postgresql`
- **Windows WSL**: `sudo apt install postgresql`

Install Node.js:

- **All platforms**: [nodejs.org](https://nodejs.org/)
- **macOS**: `brew install node`
- **Ubuntu**: `sudo apt install nodejs npm`

---

## Updating Kapok

### Using Go Install

```bash
go install github.com/kapok/kapok/cmd/kapok@latest
```

### Check for Updates

```bash
kapok version
# Compare with latest: https://github.com/kapok/kapok/releases
```

---

## Uninstallation

### If installed via Go

```bash
rm $(which kapok)
```

Or:

```bash
rm $GOPATH/bin/kapok
# Or
rm /usr/local/bin/kapok
```

### If installed via Homebrew

```bash
brew uninstall kapok
```

---

## Troubleshooting Installation

### "kapok: command not found"

**Solution**: Add Go bin directory to PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

Make permanent by adding to `~/.bashrc` or `~/.zshrc`.

### Permission Denied

**Solution**: Make binary executable:

```bash
chmod +x /path/to/kapok
```

### SSL/TLS Errors

**Solution**: Update certificates:

```bash
# macOS
brew install ca-certificates

# Ubuntu/Debian
sudo apt-get install ca-certificates

# Update
sudo update-ca-certificates
```

---

## Next Steps

Once installed, see the [Quick Start Guide](./quickstart.md) to build your first
Kapok application!

---

**Need help?** [Open an issue](https://github.com/kapok/kapok/issues) or
[start a discussion](https://github.com/kapok/kapok/discussions).
