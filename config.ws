import "std/http"

http.server("localhost") {
    http.route("/*", http.static("./public"))
}

http.server("api.localhost") {
    http.route("/*", http.proxy("localhost:3000"))
}
