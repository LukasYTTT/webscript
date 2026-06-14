# WebScript (wbs) 🚀

WebScript is a modern, high-performance Domain-Specific Language (DSL) and package manager for reverse proxies and static web servers. Written in Go, it replaces complex Nginx setups with clean, readable code.

Additionally, WebScript **automatically handles HTTPS certificates** (Let's Encrypt) for all your domains!

## Features
- **Readable Syntax**: Much simpler than Nginx or Apache.
- **High-Performance**: Powered by Go's extremely fast `net/http` engine.
- **Auto-HTTPS**: No more manual Certbot certificates.
- **Package Manager**: Install third-party modules via `wbs install`.

---

## 📦 Installation

### 🚀 Universal Install (Linux & macOS)
The absolute easiest way to install WebScript anywhere is via our automated installation script. Just open your terminal and run:
```bash
curl -sSL https://raw.githubusercontent.com/LukasYTTT/webscript/main/install.sh | bash
```
*(Requires Go and Git to be installed)*

### For Arch Linux Users (via AUR)
You can easily install WebScript using the Arch User Repository (AUR):

```bash
git clone https://github.com/LukasYTTT/webscript.git
cd webscript
makepkg -si
```
*(This will soon be available directly via `yay -S webscript`!)*
### Arch Linux (AUR)
If you are using Arch Linux, you can easily install WebScript via the AUR using `yay`:
```bash
yay -S webscript-git
```

### Debian / Ubuntu
For Debian or Ubuntu, you can build a standard `.deb` package to easily install and manage WebScript using `apt`:
```bash
git clone https://github.com/LukasYTTT/webscript.git
cd webscript
chmod +x build_deb.sh
./build_deb.sh
sudo apt install ./webscript_1.0.0_amd64.deb
```

### Manual Installation (All Linux)
You need [Go](https://golang.org/) installed on your system.

```bash
git clone https://github.com/LukasYTTT/webscript.git
cd webscript
go build -o wbs .
sudo mv wbs /usr/local/bin/wbs
```

---

## 🛠 Usage

### 1. Interactive Setup
You no longer need to write configs by hand! Just use the interactive generator:
```bash
sudo wbs create
```
This will ask you for your domain, path, and proxy settings, and automatically generate a perfect configuration file in `/etc/wbs/confs/`.

### 2. Start the System Service
To start WebScript as a professional background daemon (like Nginx):
```bash
sudo wbs service
sudo systemctl start wbs
```

You can now manage it via systemctl:
- `sudo systemctl restart wbs`
- `sudo systemctl status wbs`

### Syntax Checking
To test if your configs are valid without restarting the server:
```bash
wbs -t /etc/wbs/confs/
```

### Local Testing (Dev Mode)
To test a config locally without HTTPS (starts on port 8080):
```bash
wbs run /etc/wbs/confs/ --dev
```

---

## 📚 Custom Packages & Wiki
You can install libraries from GitHub:
```bash
wbs install github.com/username/awesome-library
```
Want to build your own WebScript packages? **[Check out our official Wiki in the `/docs` folder!](./docs/index.md)**
