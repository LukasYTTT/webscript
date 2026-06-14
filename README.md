# WebScript (wbs) 🚀

WebScript ist eine moderne, hochperformante Domain-Specific Language (DSL) und ein Package-Manager für Reverse-Proxys und statische Webserver. Es ist in Go geschrieben und ersetzt komplizierte Nginx-Setups durch sauberen, lesbaren Code.

Außerdem kümmert sich WebScript **automatisch um HTTPS-Zertifikate** (Let's Encrypt) für alle deine Domains!

## Features
- **Eigene lesbare Syntax**: Einfacher als Nginx oder Apache.
- **High-Performance**: Basiert auf der extrem schnellen `net/http` Engine von Go.
- **Auto-HTTPS**: Keine manuellen Certbot-Zertifikate mehr.
- **Package Manager**: Lade fremde Module mit `wbs install` herunter.

---

## 📦 Installation

### Für Arch Linux Nutzer (via AUR)
Du kannst WebScript ganz einfach über das Arch User Repository (AUR) installieren:

```bash
git clone https://github.com/LukasYTTT/webscript.git
cd webscript
makepkg -si
```
*(Dies wird bald auch direkt über `yay -S webscript` verfügbar sein!)*

### Für alle anderen (Linux / macOS / Windows)
Du benötigst [Go](https://golang.org/) auf deinem System.

```bash
git clone https://github.com/LukasYTTT/webscript.git
cd webscript
go build -o wbs .
sudo mv wbs /usr/local/bin/wbs
```

---

## 🛠 Nutzung

### 1. Projekt initialisieren
Erstelle einen neuen Ordner für deine Web-Konfiguration und führe aus:
```bash
wbs init
```
Das erstellt eine `webscript.json` Datei für deine Abhängigkeiten.

### 2. Konfiguration schreiben
Erstelle eine Datei `config.ws`:

```webscript
import "std/http"

# Beispiel 1: Statische Website ausliefern
http.server("meinedomain.de") {
    http.route("/*", http.static("./public"))
}

# Beispiel 2: API Reverse-Proxy
http.server("api.meinedomain.de") {
    http.route("/*", http.proxy("localhost:3000"))
}
```

### 3. Server starten
Um lokal ohne HTTPS zu testen (startet auf Port 8080):
```bash
wbs run config.ws --dev
```

Um in der Produktion mit **automatischem HTTPS** zu starten (benötigt Root-Rechte für Port 80 und 443):
```bash
sudo wbs run config.ws
```

---

## 📚 Eigene Packages installieren
Du kannst Libraries von GitHub installieren:
```bash
wbs install github.com/username/tolle-library
```
Diese werden dann in einem `wbs_modules` Ordner gespeichert und können via `import` in deinem Skript genutzt werden.
