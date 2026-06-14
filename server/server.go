package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"webscript/evaluator"
	"webscript/object"

	"golang.org/x/crypto/acme/autocert"
)

type RouteTarget struct {
	Type   string
	Value  string
	Action *object.Function
}

type ServerConfig struct {
	Domain string
	Routes map[string]RouteTarget
}

type Engine struct {
	servers map[string]*ServerConfig
	current *ServerConfig
}

func NewEngine() *Engine {
	return &Engine{
		servers: make(map[string]*ServerConfig),
	}
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

	// Set current context so routes know where to register
	prev := e.current
	e.current = srv

	// Execute block
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
		// Either proxy or static, marked by prefix
		val := obj.Value
		if strings.HasPrefix(val, "proxy:") {
			target.Type = "proxy"
			target.Value = strings.TrimPrefix(val, "proxy:")
		} else if strings.HasPrefix(val, "static:") {
			target.Type = "static"
			target.Value = strings.TrimPrefix(val, "static:")
		} else {
			return &object.Error{Message: "invalid target string"}
		}
	case *object.Function:
		target.Type = "function"
		target.Action = obj
	default:
		return &object.Error{Message: "http.route second argument must be proxy, static, or func"}
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

func (e *Engine) Start(devMode bool) error {
	var domains []string
	for dom := range e.servers {
		domains = append(domains, dom)
	}

	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domains...),
		Cache:      autocert.DirCache("certs"),
	}

	handler := &routerHandler{engine: e}

	if devMode {
		log.Println("Starting in Dev mode (HTTP) on port 8080...")
		return http.ListenAndServe(":8080", handler)
	}

	server := &http.Server{
		Addr:    ":443",
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	log.Printf("Starting WebScript Server on port 80 and 443 for %v...\n", domains)

	go func() {
		log.Fatal(http.ListenAndServe(":80", certManager.HTTPHandler(nil)))
	}()

	return server.ListenAndServeTLS("", "")
}

type routerHandler struct {
	engine *Engine
}

func (h *routerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	srv, ok := h.engine.servers[host]
	if !ok {
		http.Error(w, "Domain not configured in WebScript", http.StatusNotFound)
		return
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

	case "function":
		reqObj := &object.String{Value: "Request to " + r.URL.Path}
		
		// Create a local scope for the function execution
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
