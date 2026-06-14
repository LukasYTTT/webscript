import "std/http"

http.server("blitzmails.de") {
    http.route("/*", http.proxy("http://localhost:3000"))
}

http.server("www.blitzmails.de") {
    http.route("/*", http.proxy("http://localhost:3000"))
}

http.server("myaccount.blitzmails.de") {
    http.route("/api/*", http.proxy("http://localhost:3000/api"))
    http.route("/uploads/*", http.proxy("http://localhost:3000/uploads"))
    http.route("/*", http.proxy("http://localhost:3000/account"))
}
