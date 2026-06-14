package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"webscript/lexer"
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
	case "install":
		if len(os.Args) < 3 {
			handleInstallAll()
		} else {
			handleInstall(os.Args[2])
		}
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Bitte gib eine Datei an: wbs run config.ws")
			os.Exit(1)
		}
		filename := os.Args[2]
		devMode := false
		if len(os.Args) > 3 && os.Args[3] == "--dev" {
			devMode = true
		}
		handleRun(filename, devMode)
	default:
		fmt.Printf("Unbekannter Befehl: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("WebScript (wbs) - Package Manager & Runtime")
	fmt.Println("\nNutzung:")
	fmt.Println("  wbs init                  Erstellt eine webscript.json")
	fmt.Println("  wbs install <url>         Installiert eine Library (z.B. github.com/user/lib)")
	fmt.Println("  wbs install               Installiert alle in webscript.json gelisteten Libraries")
	fmt.Println("  wbs run <datei.ws> [--dev] Führt ein WebScript aus")
}

func handleInit() {
	config := WebScriptConfig{
		Dependencies: make(map[string]string),
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	err := ioutil.WriteFile("webscript.json", data, 0644)
	if err != nil {
		log.Fatalf("Fehler beim Erstellen von webscript.json: %v", err)
	}
	fmt.Println("webscript.json erfolgreich erstellt!")
}

func handleInstallAll() {
	data, err := ioutil.ReadFile("webscript.json")
	if err != nil {
		log.Fatalf("Keine webscript.json gefunden. Bitte führe 'wbs init' aus.")
	}
	var config WebScriptConfig
	json.Unmarshal(data, &config)

	for pkg, url := range config.Dependencies {
		fmt.Printf("Installiere %s aus %s...\n", pkg, url)
		installGitRepo(pkg, url)
	}
}

func handleInstall(repoURL string) {
	// For simplicity, we assume repoURL is a github URL like github.com/user/lib
	// We extract the library name from the URL
	pkgName := filepath.Base(repoURL)
	
	installGitRepo(pkgName, "https://"+repoURL)

	// Update webscript.json
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
	
	fmt.Printf("Library '%s' installiert!\n", pkgName)
}

func installGitRepo(pkgName, url string) {
	targetDir := filepath.Join("wbs_modules", pkgName)
	
	// Remove if exists
	os.RemoveAll(targetDir)
	
	// Clone
	cmd := exec.Command("git", "clone", url, targetDir)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Warnung: Konnte Repository %s nicht klonen. (Hast du git installiert?)\n", url)
	}
}

func handleRun(filename string, devMode bool) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Fehler beim Lesen der Datei: %v\n", err)
	}

	// NOTE: For v2, we need to rewrite the lexer/parser.
	// For now, this still uses the v1 parser, but we will upgrade it in the next steps!
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Fehler beim Parsen:")
		for _, msg := range p.Errors() {
			fmt.Printf("\t- %s\n", msg)
		}
		os.Exit(1)
	}

	srv := server.New(program)
	if err := srv.Start(devMode); err != nil {
		log.Fatalf("Server Fehler: %v\n", err)
	}
}
