# 📚 WebScript Official Wiki

Welcome to the **WebScript (wbs)** official documentation! 

WebScript is designed to make configuring web servers as simple as writing a basic script, while retaining the raw power of a full programming language.

---

## 📖 1. WebScript Guide (How it works)
Everything you need to know about operating WebScript on a production server.
- **[Server Architecture](./architecture.md)** (How WebScript handles HTTPS, background processes, and routing)
- **[Installation Guide](../README.md)** (Learn how to install via curl)
- **[CLI Reference](#cli-reference)** (Overview of all `wbs` commands)

## 📦 2. Custom Libraries
WebScript is extensible! Learn how to use third-party packages or publish your own to the official registry.
- **[Publishing Custom Packages](./custom_packages.md)** (How to publish your own WebScript libraries to the official Mono-Repo)

---

## 💻 CLI Reference
| Command | Description |
|---|---|
| `wbs create` | Interactive wizard to generate perfect configuration files |
| `wbs service` | Installs WebScript as an automatic background SystemD daemon |
| `wbs -t <path>` | Syntax-checks your configuration files without starting the server |
| `wbs run <path>` | Manually start the server (use `--dev` for local testing) |
| `wbs install <package>` | Installs an officially approved third-party package from the registry |
