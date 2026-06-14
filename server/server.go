package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"webscript/evaluator"
	"webscript/object"

	"golang.org/x/crypto/acme/autocert"
)

type RouteTarget struct {
	Type      string
	Value     string
	IndexFile string
	Action    *object.Function
}

type ServerConfig struct {
	Domain string
	Routes map[string]RouteTarget
}

type Engine struct {
	servers  map[string]*ServerConfig
	current  *ServerConfig
	SecureIP bool
}

func NewEngine() *Engine {
	return &Engine{
		servers: make(map[string]*ServerConfig),
	}
}

// Builtin: http.secure_ip()
func (e *Engine) BuiltinSecureIp(args ...object.Object) object.Object {
	e.SecureIP = true
	return evaluator.NULL
}

// Builtin: http.server(domain, func() { ... })
func (e *Engine) BuiltinServer(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: fmt.Sprintf("http.server requires 2 arguments, got %d", len(args))}
	}

	domainObj, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "http.server first argument must be a string"}
	}

	blockFunc, ok := args[1].(*object.Function)
	if !ok {
		return &object.Error{Message: "http.server second argument must be a block"}
	}

	srv := &ServerConfig{
		Domain: domainObj.Value,
		Routes: make(map[string]RouteTarget),
	}
	e.servers[domainObj.Value] = srv

	prev := e.current
	e.current = srv

	evaluator.Eval(blockFunc.Body, blockFunc.Env)

	e.current = prev
	return evaluator.NULL
}

// Builtin: http.route(path, target)
func (e *Engine) BuiltinRoute(args ...object.Object) object.Object {
	if e.current == nil {
		return &object.Error{Message: "http.route must be called inside http.server"}
	}
	if len(args) != 2 {
		return &object.Error{Message: "http.route requires 2 arguments"}
	}
	pathObj, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "http.route first argument must be a string"}
	}

	target := RouteTarget{}

	switch obj := args[1].(type) {
	case *object.String:
		val := obj.Value
		if strings.HasPrefix(val, "proxy:") {
			target.Type = "proxy"
			target.Value = strings.TrimPrefix(val, "proxy:")
		} else if strings.HasPrefix(val, "static:") {
			target.Type = "static"
			target.Value = strings.TrimPrefix(val, "static:")
		} else if strings.HasPrefix(val, "php:") {
			parts := strings.SplitN(strings.TrimPrefix(val, "php:"), "|", 2)
			target.Type = "php"
			target.Value = parts[0]
			if len(parts) > 1 {
				target.IndexFile = parts[1]
			} else {
				target.IndexFile = "index.php"
			}
		} else {
			return &object.Error{Message: "invalid target string"}
		}
	case *object.Function:
		target.Type = "function"
		target.Action = obj
	default:
		return &object.Error{Message: "http.route second argument must be proxy, static, php, or func"}
	}

	e.current.Routes[pathObj.Value] = target
	return evaluator.NULL
}

func (e *Engine) BuiltinProxy(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "http.proxy requires 1 argument"}
	}
	urlObj, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "http.proxy argument must be string"}
	}
	return &object.String{Value: "proxy:" + urlObj.Value}
}

func (e *Engine) BuiltinStatic(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "http.static requires 1 argument"}
	}
	pathObj, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "http.static argument must be string"}
	}
	return &object.String{Value: "static:" + pathObj.Value}
}

func (e *Engine) BuiltinPhp(args ...object.Object) object.Object {
	if len(args) < 1 {
		return &object.Error{Message: "http.php requires at least 1 argument"}
	}
	folderObj, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "http.php first argument must be string"}
	}
	indexFile := "index.php"
	if len(args) == 2 {
		if idxObj, ok := args[1].(*object.String); ok {
			indexFile = idxObj.Value
		}
	}
	return &object.String{Value: "php:" + folderObj.Value + "|" + indexFile}
}

func (e *Engine) Start(devMode bool) error {
	var autocertDomains []string
	var localDomains bool
	customPorts := make(map[string]bool)
	needsDefaultPorts := false

	for dom := range e.servers {
		host := dom
		port := ""
		if strings.Contains(dom, ":") {
			parts := strings.Split(dom, ":")
			host = parts[0]
			port = parts[1]
		}

		if port != "" {
			customPorts[port] = true
		} else {
			needsDefaultPorts = true
		}

		if host == "localhost" || host == "127.0.0.1" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") {
			localDomains = true
		} else if port == "" {
			autocertDomains = append(autocertDomains, host)
		}
	}

	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(autocertDomains...),
		Cache:      autocert.DirCache("certs"),
	}

	handler := &routerHandler{engine: e}

	for port := range customPorts {
		log.Printf("Starting custom listener on port :%s...\n", port)
		go func(p string) {
			log.Fatal(http.ListenAndServe(":"+p, handler))
		}(port)
	}

	if !needsDefaultPorts {
		// Block forever if we only have custom ports
		select {}
	}

	if devMode {
		log.Println("Starting in Dev mode (HTTP) on port 8080...")
		return http.ListenAndServe(":8080", handler)
	}

	server := &http.Server{
		Addr:    ":443",
		Handler: handler,
	}

	if localDomains {
		log.Println("Local domains detected. Generating self-signed HTTPS certificates...")
		cert, err := generateDevCert(append(autocertDomains, "localhost"))
		if err != nil {
			return err
		}
		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{*cert},
		}
	} else if len(autocertDomains) > 0 {
		server.TLSConfig = &tls.Config{
			GetCertificate: certManager.GetCertificate,
		}
		go func() {
			log.Fatal(http.ListenAndServe(":80", certManager.HTTPHandler(nil)))
		}()
	} else {
		// Fallback for default HTTPS without specific domains
		log.Println("No public domains found for Auto-HTTPS. Waiting for connections...")
	}

	log.Printf("Starting WebScript Server on port 443 (HTTPS) for %v...\n", autocertDomains)

	return server.ListenAndServeTLS("", "")
}

func generateDevCert(domains []string) (*tls.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"WebScript Dev"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              domains,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

type routerHandler struct {
	engine *Engine
}

func (h *routerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hostWithPort := r.Host
	hostWithoutPort := hostWithPort
	if strings.Contains(hostWithPort, ":") {
		hostWithoutPort = strings.Split(hostWithPort, ":")[0]
	}

	if h.engine.SecureIP && net.ParseIP(hostWithoutPort) != nil {
		http.Error(w, "Forbidden - Direct IP Access Blocked", http.StatusForbidden)
		return
	}

	srv, ok := h.engine.servers[hostWithPort]
	if !ok {
		srv, ok = h.engine.servers[hostWithoutPort]
		if !ok {
			http.Error(w, "Domain not configured in WebScript", http.StatusNotFound)
			return
		}
	}

	var matchedTarget *RouteTarget
	var matchedPath string

	for path, target := range srv.Routes {
		if path == r.URL.Path {
			matchedTarget = &target
			matchedPath = path
			break
		}
	}

	if matchedTarget == nil {
		for path, target := range srv.Routes {
			if strings.HasSuffix(path, "/*") {
				prefix := strings.TrimSuffix(path, "/*")
				if strings.HasPrefix(r.URL.Path, prefix) {
					t := target
					matchedTarget = &t
					matchedPath = path
					break
				}
			}
		}
	}

	if matchedTarget == nil {
		http.Error(w, "Path not found", http.StatusNotFound)
		return
	}

	switch matchedTarget.Type {
	case "proxy":
		targetURLStr := matchedTarget.Value
		if !strings.HasPrefix(targetURLStr, "http://") && !strings.HasPrefix(targetURLStr, "https://") {
			targetURLStr = "http://" + targetURLStr
		}
		targetUrl, err := url.Parse(targetURLStr)
		if err != nil {
			http.Error(w, "Invalid proxy target", http.StatusInternalServerError)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(targetUrl)
		r.URL.Host = targetUrl.Host
		r.URL.Scheme = targetUrl.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = targetUrl.Host
		proxy.ServeHTTP(w, r)

	case "static":
		folderPath := matchedTarget.Value
		prefix := strings.TrimSuffix(matchedPath, "/*")
		fs := http.StripPrefix(prefix, http.FileServer(http.Dir(folderPath)))
		fs.ServeHTTP(w, r)
		
	case "php":
		folderPath := matchedTarget.Value
		prefix := strings.TrimSuffix(matchedPath, "/*")
		reqPath := strings.TrimPrefix(r.URL.Path, prefix)
		if reqPath == "" || reqPath == "/" {
			reqPath = "/" + matchedTarget.IndexFile
		}
		fullFilePath := filepath.Join(folderPath, reqPath)

		if strings.HasSuffix(fullFilePath, ".php") {
			cmd := exec.Command("php-cgi", fullFilePath)
			cmd.Env = append(os.Environ(),
				"REQUEST_METHOD="+r.Method,
				"SCRIPT_FILENAME="+fullFilePath,
				"QUERY_STRING="+r.URL.RawQuery,
				"CONTENT_TYPE="+r.Header.Get("Content-Type"),
				"CONTENT_LENGTH="+r.Header.Get("Content-Length"),
				"REMOTE_ADDR="+r.RemoteAddr,
				"SERVER_SOFTWARE=WebScript",
				"REDIRECT_STATUS=200", // Required by some php-cgi versions
			)
			if r.Body != nil {
				cmd.Stdin = r.Body
			}
			out, err := cmd.Output()
			if err != nil {
				http.Error(w, "PHP Error: "+err.Error()+"\n(Make sure php-cgi is installed)", http.StatusInternalServerError)
				return
			}
			parts := strings.SplitN(string(out), "\r\n\r\n", 2)
			if len(parts) == 2 {
				headers := strings.Split(parts[0], "\r\n")
				for _, h := range headers {
					hParts := strings.SplitN(h, ": ", 2)
					if len(hParts) == 2 {
						w.Header().Set(hParts[0], hParts[1])
					}
				}
				w.Write([]byte(parts[1]))
			} else {
				w.Write(out)
			}
		} else {
			http.ServeFile(w, r, fullFilePath)
		}

	case "function":
		reqObj := &object.String{Value: "Request to " + r.URL.Path}
		env := object.NewEnclosedEnvironment(matchedTarget.Action.Env)
		if len(matchedTarget.Action.Parameters) > 0 {
			env.Set(matchedTarget.Action.Parameters[0].Value, reqObj)
		}
		result := evaluator.Eval(matchedTarget.Action.Body, env)
		if result != nil {
			if result.Type() == object.STRING_OBJ {
				w.Write([]byte(result.(*object.String).Value))
			} else {
				w.Write([]byte(result.Inspect()))
			}
		}
	}
}
