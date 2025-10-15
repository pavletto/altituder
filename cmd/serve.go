package cmd

import (
	"altituder/cmd/ddm"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	readTimeout  = 5
	writeTimeout = 10
	idleTimeout  = 120
)

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
	Short: "Hello World web server",
	Long:  `Hello World web server`,
	Run: func(cmd *cobra.Command, args []string) {
		cacheDir := getenv("DDM_CACHE_DIR", "./cache")
		urlTpl := getenv("DDM_URL_TEMPLATE", "https://{s}.geodata.microavia.com/srtm/{z}/{y}/{x}.ddm")
		subs := getenv("DDM_SUBDOMAINS", "a,b,c")
		heightFactor := getenvFloat("DDM_HEIGHT_FACTOR", 1.0)
		defaultZoom := getenvInt("DDM_DEFAULT_Z", 14) // как в вашем примере maxZoom=17
		maxNativeZoom := getenvInt("DDM_MAX_NATIVE_Z", 14)
		noDataCSV := os.Getenv("DDM_NODATA_CSV") // например: "-32768,3.4028235e+38"

		cfg := ddm.StoreConfig{
			CacheDir:          cacheDir,
			URLTemplate:       urlTpl,
			Subdomains:        strings.Split(subs, ","),
			PermitDownload:    urlTpl != "",
			HTTPClientTimeout: 15 * time.Second,
			DefaultZoom:       defaultZoom,
			MaxNativeZoom:     maxNativeZoom,
			HeightFactor:      float32(heightFactor),
			NoDataValues:      ddm.ParseNoData(noDataCSV),
		}

		store, err := ddm.NewStore(cfg)
		if err != nil {
			log.Fatal(err)
		}

		s := &ddm.Server{Store: store}

		mux := http.NewServeMux()
		mux.HandleFunc("/height", s.HandleHeight)
		mux.HandleFunc("/health", s.HandleHealth)

		addr := getenv("ADDR", ":8080")
		log.Printf("listening on %s cache=%s download=%v tpl=%s", addr, cacheDir, cfg.PermitDownload, urlTpl)
		log.Fatal(http.ListenAndServe(addr, mux))
	},
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func getenvInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return d
}
func getenvFloat(k string, d float64) float64 {
	if v := os.Getenv(k); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return d
}
func init() {
	rootCmd.AddCommand(serveCmd)
}
