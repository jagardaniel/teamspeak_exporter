package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/alecthomas/kong"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var version = "dev"

type Config struct {
	Listen  string           `env:"TS_EXPORTER_LISTEN" help:"Address on which to expose metrics and web interface." default:":9800"`
	APIKey  string           `env:"TS_EXPORTER_API_KEY" help:"API key for TeamSpeak WebQuery authentication." required:""`
	URL     string           `env:"TS_EXPORTER_URL" help:"URL for TeamSpeak WebQuery endpoint." default:"http://127.0.0.1:10080"`
	Version kong.VersionFlag `help:"Show version and exit."`
}

func main() {
	config := &Config{}
	kong.Parse(config,
		kong.Name("teamspeak_exporter"),
		kong.Description("Prometheus exporter for TeamSpeak servers."),
		kong.Vars{
			"version": version,
		},
	)

	slog.Info("Starting teamspeak_exporter",
		"version", version,
		"listen", config.Listen,
	)

	client := NewClient(config.URL, config.APIKey)

	// Verify that we can reach the WebQuery API and that the api-key is valid and has the correct scope
	if err := client.Ping(); err != nil {
		slog.Error("Failed to verify WebQuery connection", "error", err)
		os.Exit(1)
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

	slog.Info("Listening for metrics requests", "address", config.Listen)

	if err := http.ListenAndServe(config.Listen, nil); err != nil {
		slog.Error("HTTP server error", "error", err)
		os.Exit(1)
	}
}
