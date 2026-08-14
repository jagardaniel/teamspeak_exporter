package main

import (
	"log"
	"net/http"

	"github.com/alecthomas/kong"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	Listen string `env:"TS_EXPORTER_LISTEN" help:"Address on which to expose metrics and web interface." default:":9800"`
	APIKey string `env:"TS_EXPORTER_API_KEY" help:"API key for WebQuery authentication." required:""`
	URL    string `env:"TS_EXPORTER_URL" help:"URL for TeamSpeak WebQuery endpoint." default:"http://127.0.0.1:10080"`
}

func main() {
	config := &Config{}
	kong.Parse(config)

	client := NewClient(config.URL, config.APIKey)

	// Verify that we can reach the WebQuery API and that the api-key is valid and has the correct scope
	if err := client.Ping(); err != nil {
		log.Fatalf("unable to reach webquery api: %v", err)
	}

	reg := prometheus.NewRegistry()
	collector := NewCollector(client)
	reg.MustRegister(collector)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
             <head><title>TeamSpeak	Exporter</title></head>
             <body>
             <h1>TeamSpeak Exporter</h1>
             <p><a href="/metrics">Metrics</a></p>
             </body>
             </html>`))
	})

	log.Printf("listening on %s", config.Listen)

	if err := http.ListenAndServe(config.Listen, nil); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
