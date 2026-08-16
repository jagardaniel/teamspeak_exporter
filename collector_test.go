package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestSingleOnlineServer(t *testing.T) {
	responses := map[string]string{
		"/version":      mockVersionJSON,
		"/serverlist":   mockServerListSingleJSON,
		"/1/serverinfo": mockServer1InfoJSON,
	}

	ts := newMockServer(t, http.StatusOK, responses)
	client := NewClient(ts.URL, "test-api-key")
	collector := NewCollector(client)

	expectedOutput := `
		# HELP teamspeak_up Was the last scrape of the TeamSpeak ServerQuery API successful (1 = yes, 0 = no)
        # TYPE teamspeak_up gauge
        teamspeak_up 1
		# HELP teamspeak_version_info Servers version information including platform and build number
		# TYPE teamspeak_version_info gauge
		teamspeak_version_info{build="1779874471",platform="Linux",version="3.13.8"} 1
		# HELP teamspeak_virtualserver_clients_online Number of clients connected to the virtual server (including query clients)
		# TYPE teamspeak_virtualserver_clients_online gauge
        teamspeak_virtualserver_clients_online{id="1",virtualserver="First server"} 2
		# HELP teamspeak_virtualserver_ping_seconds The average ping of all clients connected to the virtual server
        # TYPE teamspeak_virtualserver_ping_seconds gauge
        teamspeak_virtualserver_ping_seconds{id="1",virtualserver="First server"} 0
		# HELP teamspeak_virtualserver_sent_packets_total Total amount of packets sent
        # TYPE teamspeak_virtualserver_sent_packets_total counter
        teamspeak_virtualserver_sent_packets_total{id="1",virtualserver="First server"} 1.8339208e+07
		# HELP teamspeak_virtualserver_status Status of the virtual server (0 = offline, 1 = online, 2 = virtual online, 3 = booting up, 4 = shutting down)
		# TYPE teamspeak_virtualserver_status gauge
        teamspeak_virtualserver_status{id="1",virtualserver="First server"} 1
		# HELP teamspeak_virtualserver_up Virtual server online status (1 = online, 0 = not online)
        # TYPE teamspeak_virtualserver_up gauge
        teamspeak_virtualserver_up{id="1",virtualserver="First server"} 1
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expectedOutput),
		"teamspeak_up",
		"teamspeak_version_info",
		"teamspeak_virtualserver_status",
		"teamspeak_virtualserver_up",
		"teamspeak_virtualserver_ping_seconds",
		"teamspeak_virtualserver_clients_online",
		"teamspeak_virtualserver_sent_packets_total",
	)
	if err != nil {
		t.Fatalf("unexpected metric output:\n%v", err)
	}
}

func TestMultiOnlineOfflineServers(t *testing.T) {
	responses := map[string]string{
		"/version":      mockVersionJSON,
		"/serverlist":   mockServerListMultiJSON,
		"/1/serverinfo": mockServer1InfoJSON,
		"/2/serverinfo": mockServer2InfoJSON,
		// Server 3 doesn't need mock data. Serverinfo is not called for offline servers.
	}

	ts := newMockServer(t, http.StatusOK, responses)
	client := NewClient(ts.URL, "test-api-key")
	collector := NewCollector(client)

	expectedOutput := `
		# HELP teamspeak_up Was the last scrape of the TeamSpeak ServerQuery API successful (1 = yes, 0 = no)
        # TYPE teamspeak_up gauge
        teamspeak_up 1
		# HELP teamspeak_virtualserver_channels_online Number of channels created on the virtual server
		# TYPE teamspeak_virtualserver_channels_online gauge
		teamspeak_virtualserver_channels_online{id="1",virtualserver="First server"} 53
		teamspeak_virtualserver_channels_online{id="2",virtualserver="Second server"} 1
		# HELP teamspeak_virtualserver_sent_bytes_total Total amount of bytes sent
		# TYPE teamspeak_virtualserver_sent_bytes_total counter
		teamspeak_virtualserver_sent_bytes_total{id="1",virtualserver="First server"} 2.187625219e+09
		teamspeak_virtualserver_sent_bytes_total{id="2",virtualserver="Second server"} 0
		# HELP teamspeak_virtualserver_status Status of the virtual server (0 = offline, 1 = online, 2 = virtual online, 3 = booting up, 4 = shutting down)
        # TYPE teamspeak_virtualserver_status gauge
        teamspeak_virtualserver_status{id="1",virtualserver="First server"} 1
        teamspeak_virtualserver_status{id="2",virtualserver="Second server"} 1
        teamspeak_virtualserver_status{id="6",virtualserver="Third server"} 0
        # HELP teamspeak_virtualserver_up Virtual server online status (1 = online, 0 = not online)
        # TYPE teamspeak_virtualserver_up gauge
        teamspeak_virtualserver_up{id="1",virtualserver="First server"} 1
        teamspeak_virtualserver_up{id="2",virtualserver="Second server"} 1
        teamspeak_virtualserver_up{id="6",virtualserver="Third server"} 0
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expectedOutput),
		"teamspeak_up",
		"teamspeak_virtualserver_status",
		"teamspeak_virtualserver_up",
		"teamspeak_virtualserver_channels_online",
		"teamspeak_virtualserver_sent_bytes_total",
	)
	if err != nil {
		t.Fatalf("unexpected metric output:\n%v", err)
	}
}

func TestBadJSONOnServerList(t *testing.T) {
	responses := map[string]string{
		"/version":    mockVersionJSON,
		"/serverlist": mockMalformedJSON,
	}

	ts := newMockServer(t, http.StatusOK, responses)
	client := NewClient(ts.URL, "test-api-key")
	collector := NewCollector(client)

	expectedOutput := `
		# HELP teamspeak_up Was the last scrape of the TeamSpeak ServerQuery API successful (1 = yes, 0 = no)
        # TYPE teamspeak_up gauge
        teamspeak_up 0
        # HELP teamspeak_version_info Servers version information including platform and build number
        # TYPE teamspeak_version_info gauge
        teamspeak_version_info{build="1779874471",platform="Linux",version="3.13.8"} 1
    `

	err := testutil.CollectAndCompare(collector, strings.NewReader(expectedOutput))
	if err != nil {
		t.Fatalf("unexpected metric output:\n%v", err)
	}
}

func TestBadJSONOnServerInfo(t *testing.T) {
	responses := map[string]string{
		"/version":      mockVersionJSON,
		"/serverlist":   mockServerListMultiJSON,
		"/1/serverinfo": mockMalformedJSON,
		"/2/serverinfo": mockServer2InfoJSON,
	}

	ts := newMockServer(t, http.StatusOK, responses)
	client := NewClient(ts.URL, "test-api-key")
	collector := NewCollector(client)

	expectedOutput := `
        # HELP teamspeak_virtualserver_status Status of the virtual server (0 = offline, 1 = online, 2 = virtual online, 3 = booting up, 4 = shutting down)
        # TYPE teamspeak_virtualserver_status gauge
        teamspeak_virtualserver_status{id="1",virtualserver="First server"} 1
        teamspeak_virtualserver_status{id="2",virtualserver="Second server"} 1
		teamspeak_virtualserver_status{id="6",virtualserver="Third server"} 0
        # HELP teamspeak_virtualserver_channels_online Number of channels created on the virtual server
        # TYPE teamspeak_virtualserver_channels_online gauge
        teamspeak_virtualserver_channels_online{id="2",virtualserver="Second server"} 1
    `

	err := testutil.CollectAndCompare(collector, strings.NewReader(expectedOutput),
		"teamspeak_virtualserver_status",
		"teamspeak_virtualserver_channels_online",
	)
	if err != nil {
		t.Fatalf("unexpected metric output:\n%v", err)
	}
}

func TestAPIErrorResponse(t *testing.T) {
	responses := map[string]string{
		"/version": "Internal Server Error",
	}

	ts := newMockServer(t, http.StatusInternalServerError, responses)
	client := NewClient(ts.URL, "test-api-key")
	collector := NewCollector(client)

	expectedOutput := `
	# HELP teamspeak_up Was the last scrape of the TeamSpeak ServerQuery API successful (1 = yes, 0 = no)
    # TYPE teamspeak_up gauge
    teamspeak_up 0
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expectedOutput))
	if err != nil {
		t.Fatalf("unexpected metric output:\n%v", err)
	}
}

func TestUnreachableServer(t *testing.T) {
	client := NewClient("http://127.0.0.1:0", "test-api-key")
	collector := NewCollector(client)

	expectedOutput := `
	# HELP teamspeak_up Was the last scrape of the TeamSpeak ServerQuery API successful (1 = yes, 0 = no)
    # TYPE teamspeak_up gauge
    teamspeak_up 0
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expectedOutput))
	if err != nil {
		t.Fatalf("unexpected metric output:\n%v", err)
	}
}
