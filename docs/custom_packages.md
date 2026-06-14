# 📦 Building Custom WebScript Packages

WebScript is not just a configuration file format—it's a complete Domain-Specific Language (DSL). This means you can write reusable logic and share it with the world via GitHub.

---

## 1. How the Package System Works
Whenever you define a variable or a function in the root scope of your `.ws` file, it is **automatically exported**. When someone imports your package, WebScript converts your GitHub repository name into an object that contains all your exported functions.

## 2. Creating Your Package

### Step 1: Initialize a Git Repository
Create a new folder and initialize Git. Let's assume you will publish it to `github.com/YourName/wbs-logger`.

### Step 2: Write your Code (`index.ws`)
The entry point of your package must be named `index.ws`.

```webscript
# index.ws

# Exported Variable
let defaultPrefix = "[INFO]"

# Exported Function
let logRequest = fn(domain, path) {
    print(defaultPrefix + " Incoming request for " + domain + " on path: " + path)
}
```

### Step 3: Publish to GitHub
Commit your code and push it to a public GitHub repository.
```bash
git add index.ws
git commit -m "First release of my logger"
git push origin main
```

---

## 3. How Users Install Your Package

Any WebScript user can open their terminal and type:
```bash
wbs install github.com/YourName/wbs-logger
```

**What happens in the background?**
1. WebScript clones your repository into their `wbs_modules/` folder.
2. It adds your repository to their `webscript.json` dependency tracker.

---

## 4. How Users Use Your Package

Users can now import your library in their `config.ws`.

> [!NOTE]
> WebScript automatically replaces slashes `/` and dashes `-` in your repository URL with underscores `_` to create a valid variable name.
> `github.com/YourName/wbs-logger` becomes `github_com_YourName_wbs_logger`.

```webscript
# Import the package
import "github.com/YourName/wbs-logger"

http.server("example.com") {
    
    # Call the function from your package!
    github_com_YourName_wbs_logger.logRequest("example.com", "/*")

    http.route("/*", http.static("/var/www/html"))
}
```

## 💡 Best Practices
1. **Always name your main file `index.ws`**: WebScript looks for this file when importing a directory.
2. **Keep it stateless**: WebScript configurations are evaluated once at startup. Avoid using global state variables that change during runtime unless you explicitly understand the evaluator's memory model.
3. **Write a README**: Always include a `README.md` in your GitHub repository so others know what functions you have exported!
