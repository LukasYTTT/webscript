import "std/http"

http.server("sudowbscreate") {
    http.route("/*", http.proxy("sudo wbs create"))
}
