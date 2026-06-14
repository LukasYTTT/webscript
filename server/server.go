package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"webscript/ast"

	"golang.org/x/crypto/acme/autocert"
)

type WebScriptServer struct {
	program *ast.Program
}

func New(program *ast.Program) *WebScriptServer {
	return &WebScriptServer{
		program: program,
	}
}

func (s *WebScriptServer) Start(devMode bool) error {
	// Collect all domains for Let's Encrypt
	var domains []string
	for _, srv := range s.program.Servers {
		domains = append(domains, srv.Domain)
	}

	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domains...),
		Cache:      autocert.DirCache("certs"), // Speichert Zertifikate im Ordner 'certs'
	}

	// Create custom handler
	handler := &routerHandler{program: s.program}

	if devMode {
		log.Println("Starting in Dev mode (HTTP) on port 8080...")
		return http.ListenAndServe(":8080", handler)
	}

	// Production Mode: HTTPS with automatic Let's Encrypt certificates
	server := &http.Server{
		Addr:    ":443",
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	log.Printf("Starting WebScript Server on port 80 and 443 for %v...\n", domains)

	// Start HTTP server that redirects to HTTPS and handles ACME challenges
	go func() {
		log.Fatal(http.ListenAndServe(":80", certManager.HTTPHandler(nil)))
	}()

	// Start HTTPS server
	return server.ListenAndServeTLS("", "")
}

type routerHandler struct {
	program *ast.Program
}

func (h *routerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Find matching server by domain
	var matchedServer *ast.Server
	for _, srv := range h.program.Servers {
		// Clean the host in case it contains port (e.g., localhost:8080)
		host := r.Host
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}

		if srv.Domain == host {
			matchedServer = srv
			break
		}
	}

	if matchedServer == nil {
		http.Error(w, "Domain not configured in WebScript", http.StatusNotFound)
		return
	}

	// Find matching route
	// Note: We need a smarter matching if "/*" vs specific routes exist.
	// For simplicity, we check specific routes first, then wildcards.
	var matchedRoute *ast.Route

	// Exact match first
	for _, route := range matchedServer.Routes {
		if route.Path == r.URL.Path {
			matchedRoute = route
			break
		}
	}

	// Wildcard match (simple implementation)
	if matchedRoute == nil {
		for _, route := range matchedServer.Routes {
			if strings.HasSuffix(route.Path, "/*") {
				prefix := strings.TrimSuffix(route.Path, "/*")
				if strings.HasPrefix(r.URL.Path, prefix) {
					matchedRoute = route
					break
				}
			}
		}
	}

	if matchedRoute == nil {
		http.Error(w, "Path not found", http.StatusNotFound)
		return
	}

	// Handle the target
	switch matchedRoute.Target.Type {
	case ast.TargetProxy:
		targetURLStr := matchedRoute.Target.Value
		if !strings.HasPrefix(targetURLStr, "http://") && !strings.HasPrefix(targetURLStr, "https://") {
			targetURLStr = "http://" + targetURLStr
		}
		targetUrl, err := url.Parse(targetURLStr)
		if err != nil {
			http.Error(w, "Invalid proxy target", http.StatusInternalServerError)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(targetUrl)
		
		// Update headers for proxying
		r.URL.Host = targetUrl.Host
		r.URL.Scheme = targetUrl.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = targetUrl.Host

		proxy.ServeHTTP(w, r)

	case ast.TargetStatic:
		folderPath := matchedRoute.Target.Value
		
		// If path is "/*", we serve directly from folder
		// If path is "/images/*", we need to strip "/images" before looking up in folder
		prefix := strings.TrimSuffix(matchedRoute.Path, "/*")
		
		fs := http.StripPrefix(prefix, http.FileServer(http.Dir(folderPath)))
		fs.ServeHTTP(w, r)
	default:
		http.Error(w, "Unknown target type", http.StatusInternalServerError)
	}
}
