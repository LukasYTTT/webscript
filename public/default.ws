import "std/http"

http.server("localhost") {
    http.route("/*", http.static("/var/www/html"))
}
