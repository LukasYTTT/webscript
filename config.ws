import "std/http"

# Security: Block raw IP scans (scanners that don't know the domain get 403 Forbidden)
http.secure_ip()

http.server("localhost") {
    # Serve PHP! 
    # Everything inside /var/www/test is served. If a user goes to "/", it loads "index.php".
    # All .php files are automatically executed via php-cgi!
    http.route("/*", http.php("/var/www/test", "index.php"))
}

http.server("api.localhost") {
    # You can still use the proxy for your NodeJS apps on other subdomains
    http.route("/*", http.proxy("localhost:3000"))
}
