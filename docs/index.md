# 📚 WebScript Official Wiki

Welcome to the **WebScript (wbs)** official documentation! 

WebScript is designed to make configuring web servers as simple as writing a basic script, while retaining the raw power of a full programming language.

## 🌟 Core Features
- **Auto-HTTPS:** Zero configuration Let's Encrypt certificates.
- **Built-in PHP Support:** FastCGI integration out of the box.
- **Reverse Proxy:** Proxy API traffic with a single line of code.
- **Package Manager:** Create, share, and import custom libraries from GitHub.

---

## 📖 Table of Contents

### 1. Getting Started
- **[Installation Guide](../README.md)** (Learn how to install via curl)
- **[CLI Reference](#cli-reference)** (Overview of all `wbs` commands)

### 2. Developer Guides
- **[Server Architecture](./architecture.md)** (How WebScript works under the hood on a production server)
- **[Building Custom Packages](./custom_packages.md)** (How to build and publish your own WebScript libraries)

---

## 💻 CLI Reference
| Command | Description |
|---|---|
| `wbs create` | Interactive wizard to generate perfect configuration files |
| `wbs service` | Installs WebScript as an automatic background SystemD daemon |
| `wbs -t <path>` | Syntax-checks your configuration files without starting the server |
| `wbs run <path>` | Manually start the server (use `--dev` for local testing) |
| `wbs install <url>` | Installs a third-party package from GitHub |
