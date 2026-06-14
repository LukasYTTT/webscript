# 🏗️ How WebScript Works on Your Server

WebScript is not just a language; it is a full-featured web server and automation engine designed to replace traditional tools like Nginx or Apache while being significantly easier to configure. 

This guide explains what happens behind the scenes when you run WebScript on a production server.

---

## 1. The SystemD Daemon (`wbs service`)

When you run the command `wbs service` or use the official `install.sh` script, WebScript installs itself as a **background daemon** using Linux SystemD.

- **Autostart:** WebScript will automatically start whenever your server reboots.
- **Process Manager:** If WebScript crashes for any reason, SystemD will immediately restart it.
- **Root Privileges:** The service runs as `root` so it can bind to the protected web ports (Port 80 and 443).

You can control the background service using standard Linux commands:
- `sudo systemctl status wbs` (View logs and status)
- `sudo systemctl restart wbs` (Apply new configurations)
- `sudo systemctl stop wbs` (Stop the server)

---

## 2. Configuration Loading (`/etc/wbs/confs/`)

When the WebScript daemon starts, it does not look for a single `config.ws` file. Instead, it scans the entire `/etc/wbs/confs/` directory.

- Every `.ws` file inside this directory is evaluated by the engine.
- You can split your domains into different files (e.g., `example.com.ws` and `api.example.com.ws`), or group multiple domains in a single file!
- The engine collects all `http.server("domain")` blocks and builds a high-performance routing map in memory.

---

## 3. Let's Encrypt Auto-HTTPS

WebScript handles SSL/TLS certificates completely automatically. You don't need to install `certbot` or manage cron jobs.

When a user connects to your domain via HTTPS (Port 443):
1. WebScript checks if it already has a valid Let's Encrypt certificate in its cache (`/etc/wbs/certs/`).
2. If the certificate is missing or expired, WebScript temporarily pauses the user's connection.
3. It asks Let's Encrypt for a new certificate. Let's Encrypt verifies your domain by sending a challenge to Port 80 (HTTP).
4. WebScript intercepts this challenge automatically, proves you own the domain, downloads the green certificate, and resumes the user's connection!

> [!WARNING]
> For Auto-HTTPS to work, your DNS (A-Record) must point directly to your server's IP, and **Port 80 and Port 443 must be open in your firewall**. If you are using Cloudflare Proxy (Orange Cloud), ensure your SSL/TLS encryption mode is set to **"Full"** or **"Full (strict)"** to prevent redirect loops.

---

## 4. The Routing Engine

WebScript's Go-based HTTP engine intercepts incoming requests and matches them against your routes.

### Static Hosting (`http.static`)
When you use `http.static("/var/www/html")`, WebScript serves files directly from the hard drive into the browser. It automatically handles MIME types (CSS, JS, Images) and serves `index.html` by default.

### Reverse Proxy (`http.proxy`)
When you use `http.proxy("http://localhost:3000")`, WebScript acts as a middleman.
- It receives the request on Port 443 (HTTPS).
- It forwards the request internally to your Docker container or Node.js app on Port 3000.
- It automatically passes all necessary headers (like `X-Forwarded-For`, `X-Forwarded-Proto`, and preserves the original `Host`) so your backend application knows the true origin of the request.

### PHP FastCGI (`http.php`)
When using `http.php`, WebScript acts as a FastCGI client. It executes the `php-cgi` binary on your server, passes the HTTP request data to PHP via environment variables, and streams the generated HTML back to the user.
