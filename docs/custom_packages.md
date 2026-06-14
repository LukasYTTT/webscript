# 📦 Building Custom WebScript Packages

WebScript is not just a configuration file format—it's a complete Domain-Specific Language (DSL). This means you can write reusable logic and share it with the world via the official WebScript Mono-Repo.

---

## 1. How the Package System Works
Whenever you define a variable or a function in the root scope of your `.ws` file, it is **automatically exported**. When someone imports your package via `wbs install <package>`, WebScript downloads your code and converts your package name into an object that contains all your exported functions.

## 2. Publishing Your Package

To ensure all packages are safe and centralized, WebScript uses an official Mono-Repo architecture. All packages live inside the official `LukasYTTT/webscript-packages` repository on GitHub.

### Step 1: Fork the Official Repository
Go to [github.com/LukasYTTT/webscript-packages](https://github.com/LukasYTTT/webscript-packages) and fork the repository to your own GitHub account.

### Step 2: Create Your Package Folder
Inside your forked repository, create a new folder with the name of your package (e.g., `logger`). 
The entry point of your package inside this folder **must** be named `index.ws`.

```webscript
# logger/index.ws

# Exported Variable
let defaultPrefix = "[INFO]"

# Exported Function
let logRequest = fn(domain, path) {
    print(defaultPrefix + " Incoming request for " + domain + " on path: " + path)
}
```

### Step 3: Submit a Pull Request
Commit your new folder to your fork and open a **Pull Request** to the original `LukasYTTT/webscript-packages` repository. Once the repository maintainers review and merge your Pull Request, your package becomes officially available to everyone!

---

## 3. How Users Install Your Package

Any WebScript user can open their terminal and type:
```bash
wbs install logger
```

**What happens in the background?**
1. WebScript connects to the official repository and downloads ONLY your `logger` folder.
2. It places your code into the user's `wbs_modules/logger/` directory.
3. It adds your package to their `webscript.json` dependency tracker.

---

## 4. How Users Use Your Package

Users can now import your library in their `config.ws`.

> [!NOTE]
> WebScript uses the package name (the name of your folder) as the variable name. Slashes and dashes are replaced with underscores.

```webscript
# Import the package
import "logger"

http.server("example.com") {
    
    # Call the function from your package!
    logger.logRequest("example.com", "/*")

    http.route("/*", http.static("/var/www/html"))
}
```

## 💡 Best Practices
1. **Always name your main file `index.ws`**: WebScript looks for this file when importing a directory.
2. **Keep it stateless**: WebScript configurations are evaluated once at startup. Avoid using global state variables that change during runtime unless you explicitly understand the evaluator's memory model.
3. **Write a README**: Always include a `README.md` inside your package folder so others know what functions you have exported!
