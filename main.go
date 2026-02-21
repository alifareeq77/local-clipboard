package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"local-clipboard/internal/client"
	"local-clipboard/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <server|client> [flags]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		fs := flag.NewFlagSet("server", flag.ExitOnError)
		addr := fs.String("addr", ":8080", "listen address for the web server")
		dbPath := fs.String("db", "clipboard.db", "path to sqlite database")
		staticDir := fs.String("static", "web/dist", "directory containing built Vue app (e.g. web/dist); empty = embedded fallback")
		noBuild := fs.Bool("no-build", false, "skip automatic Vue build before starting")
		_ = fs.Parse(os.Args[2:])

		if !*noBuild && *staticDir != "" {
			buildVue(*staticDir)
		}
		server.Run(server.Config{Addr: *addr, DBPath: *dbPath, StaticDir: *staticDir})
	case "client":
		fs := flag.NewFlagSet("client", flag.ExitOnError)
		serverURL := fs.String("server", "http://127.0.0.1:8080", "base URL of clipboard server")
		interval := fs.Duration("interval", 1*time.Second, "poll interval for local clipboard")
		source := fs.String("source", client.HostName(), "source label for this machine")
		_ = fs.Parse(os.Args[2:])
		client.Run(client.Config{ServerURL: *serverURL, Interval: *interval, Source: *source})
	default:
		fmt.Printf("unknown mode %q, expected server or client\n", os.Args[1])
		os.Exit(1)
	}
}

// buildVue runs "npm run build" in the web directory (parent of staticDir, e.g. web).
// On failure, logs a warning and returns so the server can still start with embedded fallback.
func buildVue(staticDir string) {
	webDir := filepath.Dir(staticDir) // e.g. "web/dist" -> "web"
	if webDir == "." || webDir == staticDir {
		return
	}
	packageJSON := filepath.Join(webDir, "package.json")
	if _, err := os.Stat(packageJSON); err != nil {
		return
	}
	absWeb, err := filepath.Abs(webDir)
	if err != nil {
		log.Printf("vue build: resolve path: %v", err)
		return
	}
	cmd := exec.Command("npm", "run", "build")
	cmd.Dir = absWeb
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("building Vue app in %s ...", absWeb)
	if err := cmd.Run(); err != nil {
		log.Printf("vue build failed (server will use embedded fallback if available): %v", err)
		return
	}
	log.Printf("Vue build done.")
}
