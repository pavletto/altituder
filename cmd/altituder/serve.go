package main

import (
	"fmt"
	"github.com/pavletto/altituder/internal/elevation"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

const (
	readTimeout  = 5
	writeTimeout = 10
	idleTimeout  = 120
)

func getenv(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	returnStatus := http.StatusOK
	w.WriteHeader(returnStatus)
	message := fmt.Sprintf("Hello World! %s", r.UserAgent())
	w.Write([]byte(message))
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP API server",
	Long: `Start an HTTP server that provides REST API endpoints for:
  - /height - Get terrain elevation at a location
  - /intersection - Find terrain intersection via raycast
  - /health - Health check endpoint

Configuration can be provided via environment variables or command-line flags.
Flags take precedence over environment variables.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg := LoadConfig(cmd)

		// Create store
		store, err := cfg.CreateStore()
		if err != nil {
			log.Fatal(err)
		}

		// Create server
		s := &elevation.Server{Store: store}

		// Setup routes
		mux := http.NewServeMux()
		mux.HandleFunc("/intersection", s.HandleIntersection)
		mux.HandleFunc("/height", s.HandleHeight)
		mux.HandleFunc("/health", s.HandleHealth)

		// Get listen address
		addr := getenv("ADDR", ":8080")
		if addrFlag, _ := cmd.Flags().GetString("addr"); cmd.Flags().Changed("addr") {
			addr = addrFlag
		}

		log.Printf("Starting server on %s", addr)
		log.Printf("  Cache dir: %s", cfg.CacheDir)
		log.Printf("  Download: %v", cfg.URLTemplate != "")
		log.Fatal(http.ListenAndServe(addr, mux))
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("addr", "a", ":8080", "Address to listen on")
}
