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

### 1. Initialize a Project
Create a new folder for your web configuration and run:
```bash
wbs init
```
This creates a `webscript.json` file to track your dependencies.

### 2. Write Configuration
Create a `config.ws` file:

```webscript
import "std/http"

# Example 1: Serve a static website
http.server("mydomain.com") {
    http.route("/*", http.static("./public"))
}

# Example 2: API Reverse Proxy
http.server("api.mydomain.com") {
    http.route("/*", http.proxy("localhost:3000"))
}
```

### 3. Start the Server
To test locally without HTTPS (starts on port 8080):
```bash
wbs run config.ws --dev
```

To run in production with **automatic HTTPS** (requires root privileges for ports 80 and 443):
```bash
sudo wbs run config.ws
```

---

## 📚 Installing Custom Packages
You can install libraries from GitHub:
```bash
wbs install github.com/username/awesome-library
```
These will be downloaded into a `wbs_modules` folder and can be used via `import` in your script.
