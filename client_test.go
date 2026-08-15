package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newMockServer(t *testing.T, statusCode int, responseBody string, expectedPath string) *httptest.Server {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify header for api-key is set for the tests
		if r.Header.Get("x-api-key") == "" {
			t.Error("expected x-api-key header to be set, got empty string")
		}

		if r.URL.Path != expectedPath {
			t.Errorf("expected path %q, got %q", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Server", "TeamSpeak Server 3.13.8")
		w.WriteHeader(statusCode)
		fmt.Fprintln(w, responseBody)
	}))

	t.Cleanup(ts.Close)

	return ts
}

func TestPing(t *testing.T) {
	mockJSON := `
	{
		"body": [
			{
			"virtualserver_autostart": "1",
			"virtualserver_clientsonline": "1",
			"virtualserver_id": "1",
			"virtualserver_machine_id": "",
			"virtualserver_maxclients": "100",
			"virtualserver_name": "First server",
			"virtualserver_port": "9987",
			"virtualserver_queryclientsonline": "0",
			"virtualserver_status": "online",
			"virtualserver_uptime": "685689"
			}
		],
		"status": {
			"code": 0,
			"message": "ok"
		}
	}
	`

	ts := newMockServer(t, http.StatusOK, mockJSON, "/serverlist")
	client := NewClient(ts.URL, "test-api-key")

	if err := client.Ping(); err != nil {
		t.Fatalf("expected Ping() to succeed, got error: %v", err)
	}
}

func TestPingInvalidAPIKey(t *testing.T) {
	mockJSON := `
	{
		"status": {
			"code": 5122,
			"extra_message": "invalid api key",
			"message": "invalid apikey"
		}
	}
	`

	ts := newMockServer(t, http.StatusUnauthorized, mockJSON, "/serverlist")
	client := NewClient(ts.URL, "test-api-key")

	err := client.Ping()
	if err == nil {
		t.Fatalf("expected Ping() error, got nil: %v", err)
	}

	expectedErr := "webquery api error 5122: invalid apikey"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got %q", expectedErr, err.Error())
	}
}

func TestPingAPIKeyOutOfScope(t *testing.T) {
	mockJSON := `
	{
		"status": {
			"code": 5120,
			"extra_message": "command not in api key scope",
			"message": "out of scope"
		}
	}
	`
	ts := newMockServer(t, http.StatusUnauthorized, mockJSON, "/serverlist")
	client := NewClient(ts.URL, "test-api-key")

	err := client.Ping()
	if err == nil {
		t.Fatalf("expected Ping() error, got nil: %v", err)
	}

	expectedErr := "webquery api error 5120: out of scope"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got %q", expectedErr, err.Error())
	}
}

func TestVersion(t *testing.T) {
	mockJSON := `
	{
  		"body": [
    		{
				"build": "1779874471",
				"platform": "Linux",
				"version": "3.13.8"
    		}
  		],
  		"status": {
    		"code": 0,
    		"message": "ok"
  		}
	}`

	ts := newMockServer(t, http.StatusOK, mockJSON, "/version")
	client := NewClient(ts.URL, "test-api-key")

	versionInfo, err := client.Version()
	if err != nil {
		t.Fatalf("expected Version() to succeed, got error: %v", err)
	}

	if versionInfo.Build != 1779874471 {
		t.Errorf("expected build to be 1779874471, got %d", versionInfo.Build)
	}

	if versionInfo.Platform != "Linux" {
		t.Errorf("expected platform to be 'Linux', got %q", versionInfo.Platform)
	}

	if versionInfo.Version != "3.13.8" {
		t.Errorf("expected version to be '3.13.8', got %q", versionInfo.Version)
	}
}

func TestVirtualServerList(t *testing.T) {
	mockJSON := `
	{
		"body": [
			{
				"virtualserver_autostart": "1",
				"virtualserver_clientsonline": "1",
				"virtualserver_id": "1",
				"virtualserver_machine_id": "",
				"virtualserver_maxclients": "100",
				"virtualserver_name": "First server",
				"virtualserver_port": "9987",
				"virtualserver_queryclientsonline": "0",
				"virtualserver_status": "online",
				"virtualserver_uptime": "688078"
			},
			{
				"virtualserver_autostart": "1",
				"virtualserver_clientsonline": "0",
				"virtualserver_id": "2",
				"virtualserver_machine_id": "",
				"virtualserver_maxclients": "15",
				"virtualserver_name": "Second server",
				"virtualserver_port": "9988",
				"virtualserver_queryclientsonline": "0",
				"virtualserver_status": "online",
				"virtualserver_uptime": "120921"
			},
			{
				"virtualserver_autostart": "1",
				"virtualserver_id": "6",
				"virtualserver_machine_id": "",
				"virtualserver_name": "Third server",
				"virtualserver_port": "9989",
				"virtualserver_status": "offline"
			}
		],
		"status": {
			"code": 0,
			"message": "ok"
		}
	}
	`

	ts := newMockServer(t, http.StatusOK, mockJSON, "/serverlist")
	client := NewClient(ts.URL, "test-api-key")

	servers, err := client.VirtualServerList()
	if err != nil {
		t.Fatalf("expected VirtualServerList() to succeed, got error: %v", err)
	}

	if len(servers) != 3 {
		t.Fatalf("expected 3 entries in server list, got %d", len(servers))
	}

	if servers[0].ID != 1 {
		t.Errorf("expected server[0].ID to be 1, got %d", servers[0].ID)
	}
	if servers[0].Name != "First server" {
		t.Errorf("expected server[0].Name to be 'First server', got %q", servers[0].Name)
	}
	if servers[0].Status != "online" {
		t.Errorf("expected server[0].Status to be 'online', got %q", servers[0].Status)
	}

	if servers[2].ID != 6 {
		t.Errorf("expected server[2].ID to be 6, got %d", servers[2].ID)
	}
	if servers[2].Status != "offline" {
		t.Errorf("expected server[2].Status to be 'offline', got %q", servers[2].Status)
	}
}

func TestVirtualServerInfo(t *testing.T) {
	mockJSON := `
	{
		"body": [
			{
				"connection_bandwidth_received_last_minute_total": "87",
				"connection_bandwidth_received_last_second_total": "83",
				"connection_bandwidth_sent_last_minute_total": "91",
				"connection_bandwidth_sent_last_second_total": "81",
				"connection_bytes_received_control": "10490005",
				"connection_bytes_received_keepalive": "121589710",
				"connection_bytes_received_speech": "1223862593",
				"connection_bytes_received_total": "1355942308",
				"connection_bytes_sent_control": "21893449",
				"connection_bytes_sent_keepalive": "118665357",
				"connection_bytes_sent_speech": "2047066413",
				"connection_bytes_sent_total": "2187625219",
				"connection_filetransfer_bandwidth_received": "0",
				"connection_filetransfer_bandwidth_sent": "0",
				"connection_filetransfer_bytes_received_total": "4807",
				"connection_filetransfer_bytes_sent_total": "65669",
				"connection_packets_received_control": "110641",
				"connection_packets_received_keepalive": "2894928",
				"connection_packets_received_speech": "9530456",
				"connection_packets_received_total": "12536025",
				"connection_packets_sent_control": "110268",
				"connection_packets_sent_keepalive": "2894277",
				"connection_packets_sent_speech": "15334663",
				"connection_packets_sent_total": "18339208",
				"virtualserver_antiflood_points_needed_command_block": "150",
				"virtualserver_antiflood_points_needed_ip_block": "250",
				"virtualserver_antiflood_points_needed_plugin_block": "0",
				"virtualserver_antiflood_points_tick_reduce": "5",
				"virtualserver_ask_for_privilegekey": "0",
				"virtualserver_autostart": "1",
				"virtualserver_capability_extensions": "",
				"virtualserver_channel_temp_delete_delay_default": "0",
				"virtualserver_channelsonline": "53",
				"virtualserver_client_connections": "170",
				"virtualserver_clientsonline": "2",
				"virtualserver_codec_encryption_mode": "0",
				"virtualserver_complain_autoban_count": "5",
				"virtualserver_complain_autoban_time": "1200",
				"virtualserver_complain_remove_time": "3600",
				"virtualserver_created": "0",
				"virtualserver_default_channel_admin_group": "5",
				"virtualserver_default_channel_group": "8",
				"virtualserver_default_server_group": "8",
				"virtualserver_download_quota": "18446744073709551615",
				"virtualserver_file_storage_class": "",
				"virtualserver_filebase": "files/virtualserver_1",
				"virtualserver_flag_password": "1",
				"virtualserver_hostbanner_gfx_interval": "0",
				"virtualserver_hostbanner_gfx_url": "https://i.imgur.com/UniMfif.png",
				"virtualserver_hostbanner_mode": "2",
				"virtualserver_hostbanner_url": "",
				"virtualserver_hostbutton_gfx_url": "",
				"virtualserver_hostbutton_tooltip": "",
				"virtualserver_hostbutton_url": "",
				"virtualserver_hostmessage": "",
				"virtualserver_hostmessage_mode": "0",
				"virtualserver_icon_id": "0",
				"virtualserver_id": "1",
				"virtualserver_ip": "0.0.0.0, ::",
				"virtualserver_log_channel": "1",
				"virtualserver_log_client": "1",
				"virtualserver_log_filetransfer": "1",
				"virtualserver_log_permissions": "1",
				"virtualserver_log_query": "0",
				"virtualserver_log_server": "1",
				"virtualserver_machine_id": "",
				"virtualserver_max_download_total_bandwidth": "18446744073709551615",
				"virtualserver_max_upload_total_bandwidth": "18446744073709551615",
				"virtualserver_maxclients": "100",
				"virtualserver_min_android_version": "1559834030",
				"virtualserver_min_client_version": "1560850141",
				"virtualserver_min_clients_in_channel_before_forced_silence": "100",
				"virtualserver_min_ios_version": "1559144369",
				"virtualserver_month_bytes_downloaded": "1831515",
				"virtualserver_month_bytes_uploaded": "4807",
				"virtualserver_name": "First server",
				"virtualserver_name_phonetic": "",
				"virtualserver_needed_identity_security_level": "8",
				"virtualserver_nickname": "",
				"virtualserver_password": "",
				"virtualserver_platform": "Linux",
				"virtualserver_port": "9987",
				"virtualserver_priority_speaker_dimm_modificator": "-18.0000",
				"virtualserver_query_client_connections": "10857",
				"virtualserver_queryclientsonline": "1",
				"virtualserver_reserved_slots": "0",
				"virtualserver_status": "online",
				"virtualserver_total_bytes_downloaded": "124232499867",
				"virtualserver_total_bytes_uploaded": "1137607415",
				"virtualserver_total_packetloss_control": "0.0000",
				"virtualserver_total_packetloss_keepalive": "0.0000",
				"virtualserver_total_packetloss_speech": "0.0000",
				"virtualserver_total_packetloss_total": "0.0000",
				"virtualserver_total_ping": "0.0000",
				"virtualserver_unique_identifier": "6bKzOpQtexkScXKtJ9MDTEArwnk=",
				"virtualserver_upload_quota": "18446744073709551615",
				"virtualserver_uptime": "688902",
				"virtualserver_version": "3.13.8 [Build: 1779874471]",
				"virtualserver_weblist_enabled": "0",
				"virtualserver_welcomemessage": "Welcome to this server."
			}
		],
		"status": {
			"code": 0,
			"message": "ok"
		}
	}
	`

	ts := newMockServer(t, http.StatusOK, mockJSON, "/1/serverinfo")
	client := NewClient(ts.URL, "test-api-key")

	info, err := client.VirtualServerInfo(1)
	if err != nil {
		t.Fatalf("expected VirtualServerInfo() to succeed, got error: %v", err)
	}

	if info.Uptime != 688902 {
		t.Errorf("expected Uptime to be 688902, got %d", info.Uptime)
	}

	if info.TotalPing != 0.000 {
		t.Errorf("expected TotalPing to be 0.000, got %.4f", info.TotalPing)
	}

	if info.BytesSentTotal != 2187625219 {
		t.Errorf("expected BytesSentTotal to be 2187625219, got %d", info.BytesSentTotal)
	}
}

func TestVirtualServerInfoBadID(t *testing.T) {
	mockJSON := `
	{
		"status": {
			"code": 7,
			"message": "canceled"
		}
	}
	`

	ts := newMockServer(t, http.StatusBadRequest, mockJSON, "/999/serverinfo")
	client := NewClient(ts.URL, "test-api-key")

	_, err := client.VirtualServerInfo(999)
	if err == nil {
		t.Fatalf("expected VirtualServerInfo() error, got nil: %v", err)
	}

	expectedErr := "webquery api error 7: canceled"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got %q", expectedErr, err.Error())
	}
}

func TestTrimTrailingSlash(t *testing.T) {
	client := NewClient("http://127.0.0.1:10080/", "test-api-key")

	if client.baseURL != "http://127.0.0.1:10080" {
		t.Errorf("expected baseURL to be 'http://127.0.0.1:10080', got %q", client.baseURL)
	}
}

func TestBadJSONResponse(t *testing.T) {
	mockJSON := `
	{
  		"body": [
    		{
				"build": "1779874471",
				"platform":
				"version": "3.13.8"
	}`

	ts := newMockServer(t, http.StatusOK, mockJSON, "/version")
	client := NewClient(ts.URL, "test-api-key")

	_, err := client.Version()
	if err == nil {
		t.Fatalf("expected Version() error, got nil: %v", err)
	}

	expectedErr := "failed to decode JSON response"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got %q", expectedErr, err.Error())
	}
}

func TestServerUnreachable(t *testing.T) {
	client := NewClient("http://127.0.0.1:10090", "test-api-key")

	_, err := client.Version()
	if err == nil {
		t.Fatal("expected error for unreachable server, got nil")
	}

	expectedErr := "unable to send HTTP request"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got %q", expectedErr, err.Error())
	}
}

func TestEmptyBody(t *testing.T) {
	mockJSON := `
	{
  		"body": [],
  		"status": {
    		"code": 0,
    		"message": "ok"
  		}
	}`

	ts := newMockServer(t, http.StatusOK, mockJSON, "/version")
	client := NewClient(ts.URL, "test-api-key")

	_, err := client.Version()
	if err == nil {
		t.Fatalf("expected Version() error, got nil: %v", err)
	}

	expectedErr := "expected exactly 1 version entry, got 0"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got %q", expectedErr, err.Error())
	}
}
