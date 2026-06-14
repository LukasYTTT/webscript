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

	srv := server.New(program)
	if err := srv.Start(devMode); err != nil {
		log.Fatalf("Server Error: %v\n", err)
	}
}
