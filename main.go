package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"webscript/evaluator"
	"webscript/lexer"
	"webscript/object"
	"webscript/parser"
	"webscript/server"
)

const Version = "6.0.2"

type WebScriptConfig struct {
	Dependencies map[string]string `json:"dependencies"`
}

func main() {
	if len(os.Args) < 2 {
		checkForUpdates()
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		handleInit()
		checkForUpdates()
	case "create":
		handleCreate()
		checkForUpdates()
	case "service":
		handleService()
		checkForUpdates()
	case "install":
		if len(os.Args) < 3 {
			handleInstallAll()
		} else {
			handleInstall(os.Args[2])
		}
		checkForUpdates()
	case "-t":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a file or folder: wbs -t /path/to/confs")
			os.Exit(1)
		}
		handleRun(os.Args[2], false, true)
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a file or folder: wbs run /path/to/confs")
			os.Exit(1)
		}
		filename := os.Args[2]
		devMode := false
		if len(os.Args) > 3 && os.Args[3] == "--dev" {
			devMode = true
		}
		handleRun(filename, devMode, false)
	case "dns":
		if len(os.Args) < 3 || os.Args[2] == "list" {
			handleDNS("/etc/wbs/confs")
		} else {
			handleDNS(os.Args[2])
		}
	case "version":
		fmt.Printf("WebScript Version v%s\n", Version)
		checkForUpdates()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func checkForUpdates() {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/LukasYTTT/webscript/releases/latest")
	if err != nil {
		return // Silently fail if offline or API is unreachable
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var result struct {
			TagName string `json:"tag_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			if result.TagName != "" && result.TagName != "v"+Version {
				fmt.Printf("\n\033[33m--- UPDATE AVAILABLE ---\033[0m\n")
				fmt.Printf("A new version of WebScript (%s) is available! (Current: v%s)\n", result.TagName, Version)
				fmt.Printf("Update with: curl -sSL https://raw.githubusercontent.com/LukasYTTT/webscript/main/install.sh | bash\n\n")
			}
		}
	}
}

func printUsage() {
	fmt.Printf("WebScript (wbs) - Package Manager & Runtime (v%s)\n", Version)
	fmt.Println("\nUsage:")
	fmt.Println("  wbs init                  Creates a webscript.json")
	fmt.Println("  wbs create                Interactively generate a new .ws config file")
	fmt.Println("  wbs install <url>         Installs a library (e.g. github.com/user/lib)")
	fmt.Println("  wbs install               Installs all libraries listed in webscript.json")
	fmt.Println("  wbs run <path> [--dev]    Executes a WebScript configuration file or folder")
	fmt.Println("  wbs -t <path>             Tests the configuration syntax (like nginx -t)")
	fmt.Println("  wbs dns [path|list]       Checks the DNS A-Records for all configured domains")
	fmt.Println("  wbs service               Installs WebScript as a systemd service")
	fmt.Println("  wbs version               Prints the current version")
}

func handleInit() {
	config := WebScriptConfig{
		Dependencies: make(map[string]string),
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	err := ioutil.WriteFile("webscript.json", data, 0644)
	if err != nil {
		log.Fatalf("Error creating webscript.json: %v", err)
	}
	fmt.Println("webscript.json successfully created!")
}

func handleService() {
	if os.Geteuid() != 0 {
		fmt.Println("Please run this command as root (sudo wbs service)")
		os.Exit(1)
	}

	serviceContent := `[Unit]
Description=WebScript Server
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/bin/wbs run /etc/wbs/confs
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`
	fmt.Println("Installing WebScript as a systemd service...")
	
	os.MkdirAll("/etc/wbs/confs", 0755)
	
	if _, err := os.Stat("/etc/wbs/confs/default.ws"); os.IsNotExist(err) {
		defaultConf := "import \"std/http\"\n\nhttp.server(\"default\") {\n    http.route(\"/*\", http.static(\"/var/www/html\"))\n}\n"
		ioutil.WriteFile("/etc/wbs/confs/default.ws", []byte(defaultConf), 0644)
		fmt.Println("Created default config at /etc/wbs/confs/default.ws")
	}

	os.MkdirAll("/var/www/html", 0755)
	if _, err := os.Stat("/var/www/html/index.html"); os.IsNotExist(err) {
		defaultHtml := "<!DOCTYPE html>\n<html>\n<head>\n<title>WebScript Test</title>\n<style>\n  body { font-family: sans-serif; text-align: center; margin-top: 50px; background: #222; color: #fff; }\n  h1 { color: #00d2ff; }\n</style>\n</head>\n<body>\n  <h1>Welcome to WebScript!</h1>\n  <p>This is a static website, served extremely fast by WebScript, completely without Nginx!</p>\n</body>\n</html>\n"
		ioutil.WriteFile("/var/www/html/index.html", []byte(defaultHtml), 0644)
		fmt.Println("Created default HTML at /var/www/html/index.html")
	}

	err := ioutil.WriteFile("/etc/systemd/system/wbs.service", []byte(serviceContent), 0644)
	if err != nil {
		log.Fatalf("Failed to write service file: %v", err)
	}

	exePath, _ := os.Executable()
	if exePath != "/usr/bin/wbs" && exePath != "/usr/local/bin/wbs" {
		exec.Command("cp", exePath, "/usr/bin/wbs").Run()
	}

	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", "wbs").Run()
	
	fmt.Println("Service installed! You can now manage WebScript like Nginx:")
	fmt.Println("  sudo systemctl start wbs")
	fmt.Println("  sudo systemctl restart wbs")
	fmt.Println("  sudo systemctl status wbs")
}

func handleCreate() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🚀 WebScript Configuration Generator")
	fmt.Println("------------------------------------")
	
	fmt.Print("Domain (e.g. example.com or localhost): ")
	domain, _ := reader.ReadString('\n')
	domain = strings.TrimSpace(domain)
	// Remove hidden terminal escape characters but KEEP the colon for ports
	domain = regexp.MustCompile(`[^a-zA-Z0-9.\-_:]`).ReplaceAllString(domain, "")
	if domain == "" {
		domain = "localhost"
	}

	fmt.Print("Port (leave empty for 80/443, e.g. 8080 for local testing): ")
	portInput, _ := reader.ReadString('\n')
	portInput = strings.TrimSpace(portInput)
	portInput = regexp.MustCompile(`[^0-9]`).ReplaceAllString(portInput, "")
	if portInput != "" && !strings.Contains(domain, ":") {
		domain = fmt.Sprintf("%s:%s", domain, portInput)
	}

	fmt.Print("Static Folder Path (e.g. /var/www/html or ./public): ")
	staticPath, _ := reader.ReadString('\n')
	staticPath = strings.TrimSpace(staticPath)
	if staticPath == "" {
		staticPath = "/var/www/html"
	}

	fmt.Print("Enable PHP Support? (y/N): ")
	phpInput, _ := reader.ReadString('\n')
	phpInput = strings.TrimSpace(strings.ToLower(phpInput))
	enablePhp := phpInput == "y" || phpInput == "yes"

	fmt.Print("Reverse Proxy Target (leave empty if none, e.g. http://localhost:3000): ")
	proxyTarget, _ := reader.ReadString('\n')
	proxyTarget = strings.TrimSpace(proxyTarget)

	fmt.Print("Index File (leave empty for default index.html / index.php): ")
	indexFileInput, _ := reader.ReadString('\n')
	indexFileInput = strings.TrimSpace(indexFileInput)

	var sb strings.Builder
	sb.WriteString("import \"std/http\"\n\n")
	sb.WriteString(fmt.Sprintf("http.server(\"%s\") {\n", domain))
	
	if proxyTarget != "" {
		sb.WriteString(fmt.Sprintf("    http.route(\"/*\", http.proxy(\"%s\"))\n", proxyTarget))
	} else if enablePhp {
		if indexFileInput == "" {
			indexFileInput = "index.php"
		}
		sb.WriteString(fmt.Sprintf("    http.route(\"/*\", http.php(\"%s\", \"%s\"))\n", staticPath, indexFileInput))
	} else {
		if indexFileInput == "" {
			indexFileInput = "index.html"
		}
		sb.WriteString(fmt.Sprintf("    http.route(\"/*\", http.static(\"%s\", \"%s\"))\n", staticPath, indexFileInput))
	}
	sb.WriteString("}\n")

	configStr := sb.String()
	
	// Replace colons with underscores for a clean filename (e.g. localhost:8080 -> localhost_8080.ws)
	safeFilename := strings.ReplaceAll(domain, ":", "_")
	targetPath := fmt.Sprintf("%s.ws", safeFilename)
	
	if os.Geteuid() == 0 {
		if _, err := os.Stat("/etc/wbs/confs"); !os.IsNotExist(err) {
			targetPath = fmt.Sprintf("/etc/wbs/confs/%s.ws", safeFilename)
		}
	}

	err := ioutil.WriteFile(targetPath, []byte(configStr), 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Configuration successfully generated at: %s\n", targetPath)
	fmt.Printf("Run 'wbs -t /etc/wbs/confs' to verify it, or 'sudo systemctl restart wbs' to apply.\n")

	baseDomain := domain
	if strings.Contains(baseDomain, ":") {
		baseDomain = strings.Split(baseDomain, ":")[0]
	}

	isLocal := baseDomain == "localhost" || baseDomain == "127.0.0.1" || strings.HasSuffix(baseDomain, ".localhost") || strings.HasSuffix(baseDomain, ".local")

	if !isLocal {
		fmt.Println("\n🌐 Fetching Server IP...")
		client := http.Client{Timeout: 3 * time.Second}
		// Verwende api.ipify.org, da ifconfig.me bei manchen User-Agents HTML zurückgibt
		resp, err := client.Get("https://api.ipify.org")
		var ip string
		if err == nil {
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			ip = strings.TrimSpace(string(body))
		} else {
			ip = "<YOUR-SERVER-IP>"
		}
		
		fmt.Println("\n============================================================")
		fmt.Println("⚠️  WICHTIG: DNS-EINTRAG SETZEN")
		fmt.Println("============================================================")
		fmt.Printf("Damit das automatische HTTPS funktioniert, musst du bei deinem\n")
		fmt.Printf("Domain-Anbieter einen A-Record erstellen:\n\n")
		fmt.Printf("   Name: %s\n", baseDomain)
		fmt.Printf("   Ziel (IPv4): %s\n\n", ip)
		fmt.Printf("Sobald der Eintrag aktiv ist, holt WebScript vollautomatisch\n")
		fmt.Printf("dein Let's Encrypt HTTPS-Zertifikat!\n")
		fmt.Println("============================================================")
	}
}

func handleInstallAll() {
	data, err := ioutil.ReadFile("webscript.json")
	if err != nil {
		log.Fatalf("No webscript.json found. Please run 'wbs init' first.")
	}
	var config WebScriptConfig
	json.Unmarshal(data, &config)

	for pkg, url := range config.Dependencies {
		fmt.Printf("Installing %s from %s...\n", pkg, url)
		installGitRepo(pkg, url)
	}
}

func handleInstall(repoURL string) {
	pkgName := filepath.Base(repoURL)
	installGitRepo(pkgName, "https://"+repoURL)

	data, err := ioutil.ReadFile("webscript.json")
	var config WebScriptConfig
	if err == nil {
		json.Unmarshal(data, &config)
	} else {
		config.Dependencies = make(map[string]string)
	}
	
	config.Dependencies[pkgName] = "https://" + repoURL
	data, _ = json.MarshalIndent(config, "", "  ")
	ioutil.WriteFile("webscript.json", data, 0644)
	
	fmt.Printf("Library '%s' installed!\n", pkgName)
}

func installGitRepo(pkgName, url string) {
	targetDir := filepath.Join("wbs_modules", pkgName)
	os.RemoveAll(targetDir)
	cmd := exec.Command("git", "clone", url, targetDir)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Warning: Could not clone repository %s. (Do you have git installed?)\n", url)
	}
}

func handleRun(targetPath string, devMode bool, testMode bool) {
	engine := server.NewEngine()
	env := object.NewEnvironment()

	// Inject the standard library builtins
	env.Set("import", &object.Builtin{Fn: func(args ...object.Object) object.Object { return evaluator.NULL }})
	env.Set("http.server", &object.Builtin{Fn: engine.BuiltinServer})
	env.Set("http.route", &object.Builtin{Fn: engine.BuiltinRoute})
	env.Set("http.proxy", &object.Builtin{Fn: engine.BuiltinProxy})
	env.Set("http.static", &object.Builtin{Fn: engine.BuiltinStatic})
	env.Set("http.php", &object.Builtin{Fn: engine.BuiltinPhp})
	env.Set("http.secure_ip", &object.Builtin{Fn: engine.BuiltinSecureIp})

	var files []string
	info, err := os.Stat(targetPath)
	if err != nil {
		log.Fatalf("Error reading path: %v", err)
	}

	if info.IsDir() {
		filepath.Walk(targetPath, func(path string, f os.FileInfo, err error) error {
			if err == nil && !f.IsDir() && strings.HasSuffix(f.Name(), ".ws") {
				files = append(files, path)
			}
			return nil
		})
	} else {
		files = append(files, targetPath)
	}

	if len(files) == 0 {
		log.Fatalf("No .ws files found in %s", targetPath)
	}

	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("Error reading file %s: %v\n", file, err)
		}

		l := lexer.New(string(content))
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			fmt.Printf("Syntax errors in %s:\n", file)
			for _, msg := range p.Errors() {
				fmt.Printf("\t- %s\n", msg)
			}
			os.Exit(1)
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
			log.Fatalf("Runtime Error in %s: %s\n", file, evaluated.Inspect())
		}
	}

	if testMode {
		fmt.Println("Syntax OK! All configuration files are valid.")
		os.Exit(0)
	}

	if err := engine.Start(devMode); err != nil {
		log.Fatalf("Server Error: %v\n", err)
	}
}

func handleDNS(targetPath string) {
	fmt.Println("🔍 WebScript DNS Checker")
	fmt.Println("================================================================================")
	
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://api.ipify.org")
	var expectedIP string
	if err == nil {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		expectedIP = strings.TrimSpace(string(body))
	} else {
		fmt.Println("❌ Could not fetch server IP. Are you offline?")
		os.Exit(1)
	}

	engine := server.NewEngine()
	env := object.NewEnvironment()

	env.Set("import", &object.Builtin{Fn: func(args ...object.Object) object.Object { return evaluator.NULL }})
	env.Set("http.server", &object.Builtin{Fn: engine.BuiltinServer})
	env.Set("http.route", &object.Builtin{Fn: engine.BuiltinRoute})
	env.Set("http.proxy", &object.Builtin{Fn: engine.BuiltinProxy})
	env.Set("http.static", &object.Builtin{Fn: engine.BuiltinStatic})
	env.Set("http.php", &object.Builtin{Fn: engine.BuiltinPhp})
	env.Set("http.secure_ip", &object.Builtin{Fn: engine.BuiltinSecureIp})

	var files []string
	info, err := os.Stat(targetPath)
	if err != nil {
		log.Fatalf("Error reading path: %v", err)
	}

	if info.IsDir() {
		filepath.Walk(targetPath, func(path string, f os.FileInfo, err error) error {
			if err == nil && !f.IsDir() && strings.HasSuffix(f.Name(), ".ws") {
				files = append(files, path)
			}
			return nil
		})
	} else {
		files = append(files, targetPath)
	}

	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err == nil {
			l := lexer.New(string(content))
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) == 0 {
				evaluator.Eval(program, env)
			}
		}
	}

	servers := engine.GetServers()
	if len(servers) == 0 {
		fmt.Println("No domains configured in the specified files.")
		return
	}

	fmt.Printf("%-25s %-18s %-18s %s\n", "Domain", "Expected IP", "Actual DNS IP", "Status")
	fmt.Println("--------------------------------------------------------------------------------")

	for dom := range servers {
		baseDomain := dom
		if strings.Contains(baseDomain, ":") {
			baseDomain = strings.Split(baseDomain, ":")[0]
		}

		if baseDomain == "localhost" || baseDomain == "127.0.0.1" || strings.HasSuffix(baseDomain, ".localhost") || strings.HasSuffix(baseDomain, ".local") {
			continue
		}

		ips, err := net.LookupIP(baseDomain)
		actualIP := "<none>"
		status := "❌ MISSING"

		if err == nil && len(ips) > 0 {
			for _, ip := range ips {
				if ipv4 := ip.To4(); ipv4 != nil {
					actualIP = ipv4.String()
					break
				}
			}
			if actualIP == "<none>" {
				actualIP = ips[0].String() 
			}
		}

		if actualIP == expectedIP {
			status = "✅ OK"
		} else if actualIP != "<none>" {
			status = "❌ WRONG IP"
		}

		fmt.Printf("%-25s %-18s %-18s %s\n", baseDomain, expectedIP, actualIP, status)
	}
	fmt.Println("================================================================================")
}
