package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type responseWrapper[T any] struct {
	Body   T              `json:"body"`
	Status statusResponse `json:"status"`
}

type statusResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type VersionResponse struct {
	Build    int    `json:"build,string"`
	Platform string `json:"platform"`
	Version  string `json:"version"`
}

type VirtualServerListItemResponse struct {
	ID     int    `json:"virtualserver_id,string"`
	Name   string `json:"virtualserver_name"`
	Status string `json:"virtualserver_status"`
}

type VirtualServerInfoResponse struct {
	MaxClients                     int     `json:"virtualserver_maxclients,string"`
	ClientsOnline                  int     `json:"virtualserver_clientsonline,string"`
	ChannelsOnline                 int     `json:"virtualserver_channelsonline,string"`
	Uptime                         int     `json:"virtualserver_uptime,string"`
	QueryClientsOnline             int     `json:"virtualserver_queryclientsonline,string"`
	ClientConnections              int     `json:"virtualserver_client_connections,string"`
	QueryClientConnections         int     `json:"virtualserver_query_client_connections,string"`
	TotalPacketLossSpeech          float64 `json:"virtualserver_total_packetloss_speech,string"`
	TotalPacketLossKeepalive       float64 `json:"virtualserver_total_packetloss_keepalive,string"`
	TotalPacketLossControl         float64 `json:"virtualserver_total_packetloss_control,string"`
	TotalPacketLossTotal           float64 `json:"virtualserver_total_packetloss_total,string"`
	TotalPing                      float64 `json:"virtualserver_total_ping,string"`
	BytesReceivedControl           uint64  `json:"connection_bytes_received_control,string"`
	BytesReceivedKeepalive         uint64  `json:"connection_bytes_received_keepalive,string"`
	BytesReceivedSpeech            uint64  `json:"connection_bytes_received_speech,string"`
	BytesReceivedTotal             uint64  `json:"connection_bytes_received_total,string"`
	BytesSentControl               uint64  `json:"connection_bytes_sent_control,string"`
	BytesSentKeepalive             uint64  `json:"connection_bytes_sent_keepalive,string"`
	BytesSentSpeech                uint64  `json:"connection_bytes_sent_speech,string"`
	BytesSentTotal                 uint64  `json:"connection_bytes_sent_total,string"`
	FileTransferBytesSentTotal     uint64  `json:"connection_filetransfer_bytes_sent_total,string"`
	FileTransferBytesReceivedTotal uint64  `json:"connection_filetransfer_bytes_received_total,string"`
	PacketsReceivedControl         uint64  `json:"connection_packets_received_control,string"`
	PacketsReceivedKeepalive       uint64  `json:"connection_packets_received_keepalive,string"`
	PacketsReceivedSpeech          uint64  `json:"connection_packets_received_speech,string"`
	PacketsReceivedTotal           uint64  `json:"connection_packets_received_total,string"`
	PacketsSentControl             uint64  `json:"connection_packets_sent_control,string"`
	PacketsSentKeepalive           uint64  `json:"connection_packets_sent_keepalive,string"`
	PacketsSentSpeech              uint64  `json:"connection_packets_sent_speech,string"`
	PacketsSentTotal               uint64  `json:"connection_packets_sent_total,string"`
}

func NewClient(baseURL, apiKey string) *Client {
	cleanedURL := strings.TrimRight(baseURL, "/")

	return &Client{
		baseURL: cleanedURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func get[T any](c *Client, endpoint string) (T, error) {
	var zero T

	reqURL := fmt.Sprintf("%s/%s", c.baseURL, endpoint)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return zero, fmt.Errorf("unable to create HTTP request for %s: %w", reqURL, err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return zero, fmt.Errorf("unable to send HTTP request to %s: %w", reqURL, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return zero, fmt.Errorf("invalid HTTP status code %d from %s", res.StatusCode, endpoint)
	}

	var env responseWrapper[T]
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		return zero, fmt.Errorf("failed to decode JSON response from %s: %w", endpoint, err)
	}

	if env.Status.Code != 0 {
		return zero, fmt.Errorf("webquery api error %d: %s", env.Status.Code, env.Status.Message)
	}

	return env.Body, nil
}

func (c *Client) Version() (VersionResponse, error) {
	versionInfo, err := get[[]VersionResponse](c, "version")
	if err != nil {
		return VersionResponse{}, fmt.Errorf("unable to get version info: %w", err)
	}

	if len(versionInfo) != 1 {
		return VersionResponse{}, fmt.Errorf("expected exactly 1 version entry, got %d", len(versionInfo))
	}

	return versionInfo[0], nil
}

func (c *Client) VirtualServerList() ([]VirtualServerListItemResponse, error) {
	servers, err := get[[]VirtualServerListItemResponse](c, "serverlist")
	if err != nil {
		return nil, fmt.Errorf("unable to get serverlist: %w", err)
	}

	return servers, nil
}

func (c *Client) VirtualServerInfo(id int) (VirtualServerInfoResponse, error) {
	path := fmt.Sprintf("%d/serverinfo", id)
	info, err := get[[]VirtualServerInfoResponse](c, path)
	if err != nil {
		return VirtualServerInfoResponse{}, fmt.Errorf("unable to get serverinfo for id %d: %w", id, err)
	}

	if len(info) != 1 {
		return VirtualServerInfoResponse{}, fmt.Errorf("expected exactly 1 serverinfo entry, got %d", len(info))
	}

	return info[0], nil
}
