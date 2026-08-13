package main

type Config struct {
	Listen string `env:"TS_EXPORTER_LISTEN" help:"Address on which to expose metrics and web interface." default:":9800"`
	APIKey string `env:"TS_EXPORTER_API_KEY" help:"API key for WebQuery authentication." required:""`
	URL    string `env:"TS_EXPORTER_URL" help:"URL for TeamSpeak WebQuery endpoint." default:"http://127.0.0.1:10080"`
}
