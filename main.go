package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"webscript/evaluator"
	"webscript/lexer"
	"webscript/object"
	"webscript/parser"
	"webscript/server"
)

type WebScriptConfig struct {
	Dependencies map[string]string `json:"dependencies"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		handleInit()
	case "service":
		handleService()
	case "install":
		if len(os.Args) < 3 {
			handleInstallAll()
		} else {
			handleInstall(os.Args[2])
		}
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a file: wbs run config.ws")
			os.Exit(1)
		}
		filename := os.Args[2]
		devMode := false
		if len(os.Args) > 3 && os.Args[3] == "--dev" {
			devMode = true
		}
		handleRun(filename, devMode)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("WebScript (wbs) - Package Manager & Runtime")
	fmt.Println("\nUsage:")
	fmt.Println("  wbs init                  Creates a webscript.json")
	fmt.Println("  wbs install <url>         Installs a library (e.g. github.com/user/lib)")
	fmt.Println("  wbs install               Installs all libraries listed in webscript.json")
	fmt.Println("  wbs run <file.ws> [--dev] Executes a WebScript configuration")
	fmt.Println("  wbs service               Installs WebScript as a systemd service")
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
ExecStart=/usr/bin/wbs run /etc/wbs/config.ws
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`
	fmt.Println("Installing WebScript as a systemd service...")
	
	os.MkdirAll("/etc/wbs", 0755)
	
	if _, err := os.Stat("/etc/wbs/config.ws"); os.IsNotExist(err) {
		defaultConf := "import \"std/http\"\n\nhttp.server(\"localhost\") {\n    http.route(\"/*\", http.static(\"/var/www/html\"))\n}\n"
		ioutil.WriteFile("/etc/wbs/config.ws", []byte(defaultConf), 0644)
		fmt.Println("Created default config at /etc/wbs/config.ws")
	}

	err := ioutil.WriteFile("/etc/systemd/system/wbs.service", []byte(serviceContent), 0644)
	if err != nil {
		log.Fatalf("Failed to write service file: %v", err)
	}

	// Make sure wbs is in /usr/bin/wbs if we aren't there already
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

func handleRun(filename string, devMode bool) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading file: %v\n", err)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, msg := range p.Errors() {
			fmt.Printf("\t- %s\n", msg)
		}
		os.Exit(1)
	}

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

	evaluated := evaluator.Eval(program, env)
	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		log.Fatalf("Runtime Error: %s\n", evaluated.Inspect())
	}

	if err := engine.Start(devMode); err != nil {
		log.Fatalf("Server Error: %v\n", err)
	}
}
