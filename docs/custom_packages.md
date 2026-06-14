# How to Build Custom WebScript Packages

WebScript is designed to be highly modular. Just like in Node.js or Python, anyone can build a custom library (package) and publish it for the community to use!

## 1. Create a GitHub Repository

A WebScript package is simply a public Git repository. 
Create a new folder and a Git repo, for example on GitHub: `github.com/YourName/wbs-logger`.

## 2. Write your WebScript Code

Create an `index.ws` file in your repository. This file will be the entry point of your package.

To make functions available to other people, you define them just like regular variables and functions. WebScript automatically exports all variables that you define at the top level of your file!

**Example (`index.ws`):**
```webscript
# This is our custom logging library!

let logRequest = fn(domain) {
    # You could write custom logic here, 
    # e.g., saving to a database or formatting the output.
    print("Incoming request to: " + domain)
}
```

## 3. Publish your Package

Simply commit your code and push it to GitHub:
```bash
git add index.ws
git commit -m "Initial release of my awesome logger"
git push origin main
```

## 4. How Others Can Use Your Package

Any WebScript user can now install your package using the built-in package manager:

```bash
wbs install github.com/YourName/wbs-logger
```

This will download your code into their local `wbs_modules` folder and add it to their `webscript.json`.

They can then `import` it in their `config.ws` files:

```webscript
# Import your package
import "github.com/YourName/wbs-logger"

http.server("example.com") {
    # Use the function from your package!
    wbs_logger.logRequest("example.com")

    http.route("/*", http.static("/var/www/html"))
}
```
*(Note: When importing, the package name automatically becomes the variable name with slashes/dashes replaced by underscores).*

## Best Practices
1. **Always name your entry file `index.ws`** (or create a folder structure if needed).
2. Document your exported functions in a `README.md` so others know how to use them.
