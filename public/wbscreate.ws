import "std/http"

http.server("wbscreate") {
    http.route("/*", http.proxy("wbs create"))
}
