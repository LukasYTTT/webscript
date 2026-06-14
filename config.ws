import "std/http"

port = "3000"

http.server("localhost") {
    # Programmable Logic!
    http.route("/api/custom", func(req, res) {
        "Hello World from your own WebScript logic! Proxying would go to " + port
    })

    http.route("/*", http.static("./public"))
}

http.server("api.localhost") {
    http.route("/*", http.proxy("localhost:" + port))
}
