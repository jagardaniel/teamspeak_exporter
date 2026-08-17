package main

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// https://pyhilmi.online/blog/building-prometheus-exporter-in-go/

var serverStatusMap = map[string]float64{
	"offline":        0,
	"online":         1,
	"virtual online": 2,
	"booting up":     3,
	"shutting down":  4,
}

func parseServerStatus(status string) float64 {
	if val, ok := serverStatusMap[strings.ToLower(status)]; ok {
		return val
	}
	return -1
}

type collector struct {
	client *Client

	// Available for all virtual servers regardless of their status
	apiUp       *prometheus.Desc
	versionInfo *prometheus.Desc
	status      *prometheus.Desc
	serverUp    *prometheus.Desc

	// Only available if the virtual server has status "online"
	maxClients                     *prometheus.Desc
	clientsOnline                  *prometheus.Desc
	channelsOnline                 *prometheus.Desc
	uptime                         *prometheus.Desc
	queryClientsOnline             *prometheus.Desc
	clientConnections              *prometheus.Desc
	queryClientConnections         *prometheus.Desc
	totalPacketLossSpeech          *prometheus.Desc
	totalPacketLossKeepalive       *prometheus.Desc
	totalPacketLossControl         *prometheus.Desc
	totalPacketLossTotal           *prometheus.Desc
	totalPing                      *prometheus.Desc
	bytesReceivedControl           *prometheus.Desc
	bytesReceivedKeepalive         *prometheus.Desc
	bytesReceivedSpeech            *prometheus.Desc
	bytesReceivedTotal             *prometheus.Desc
	bytesSentControl               *prometheus.Desc
	bytesSentKeepalive             *prometheus.Desc
	bytesSentSpeech                *prometheus.Desc
	bytesSentTotal                 *prometheus.Desc
	fileTransferBytesSentTotal     *prometheus.Desc
	fileTransferBytesReceivedTotal *prometheus.Desc
	packetsReceivedControl         *prometheus.Desc
	packetsReceivedKeepalive       *prometheus.Desc
	packetsReceivedSpeech          *prometheus.Desc
	packetsReceivedTotal           *prometheus.Desc
	packetsSentControl             *prometheus.Desc
	packetsSentKeepalive           *prometheus.Desc
	packetsSentSpeech              *prometheus.Desc
	packetsSentTotal               *prometheus.Desc
}

func NewCollector(client *Client) *collector {
	serverLabels := []string{"id", "virtualserver"}

	return &collector{
		client: client,

		apiUp: prometheus.NewDesc(
			"teamspeak_up",
			"Was the last scrape of the TeamSpeak ServerQuery API successful (1 = yes, 0 = no)",
			nil, nil,
		),
		versionInfo: prometheus.NewDesc(
			"teamspeak_version_info",
			"Servers version information including platform and build number",
			[]string{"version", "build", "platform"}, nil,
		),
		status: prometheus.NewDesc(
			"teamspeak_virtualserver_status",
			"Status of the virtual server (0 = offline, 1 = online, 2 = virtual online, 3 = booting up, 4 = shutting down)",
			serverLabels, nil,
		),
		serverUp: prometheus.NewDesc(
			"teamspeak_virtualserver_up",
			"Virtual server online status (1 = online, 0 = not online)",
			serverLabels, nil,
		),

		maxClients: prometheus.NewDesc(
			"teamspeak_virtualserver_max_clients",
			"Number of slots available on the virtual server",
			serverLabels, nil,
		),
		clientsOnline: prometheus.NewDesc(
			"teamspeak_virtualserver_clients_online",
			"Number of clients connected to the virtual server (including query clients)",
			serverLabels, nil,
		),
		channelsOnline: prometheus.NewDesc(
			"teamspeak_virtualserver_channels_online",
			"Number of channels created on the virtual server",
			serverLabels, nil,
		),
		uptime: prometheus.NewDesc(
			"teamspeak_virtualserver_uptime_seconds",
			"Uptime in seconds",
			serverLabels, nil,
		),
		queryClientsOnline: prometheus.NewDesc(
			"teamspeak_virtualserver_query_clients_online",
			"Number of ServerQuery clients connected to the virtual server",
			serverLabels, nil,
		),
		clientConnections: prometheus.NewDesc(
			"teamspeak_virtualserver_client_connections_total",
			"Total number of clients connected to the virtual server since it was last started",
			serverLabels, nil,
		),
		queryClientConnections: prometheus.NewDesc(
			"teamspeak_virtualserver_query_client_connections_total",
			"Total number of ServerQuery clients connected to the virtual server since it was last started",
			serverLabels, nil,
		),
		totalPacketLossSpeech: prometheus.NewDesc(
			"teamspeak_virtualserver_packetloss_speech_percent",
			"The average packet loss for speech data on the virtual server",
			serverLabels, nil,
		),
		totalPacketLossKeepalive: prometheus.NewDesc(
			"teamspeak_virtualserver_packetloss_keepalive_percent",
			"The average packet loss for keepalive data on the virtual server",
			serverLabels, nil,
		),
		totalPacketLossControl: prometheus.NewDesc(
			"teamspeak_virtualserver_packetloss_control_percent",
			"The average packet loss for control data on the virtual server",
			serverLabels, nil,
		),
		totalPacketLossTotal: prometheus.NewDesc(
			"teamspeak_virtualserver_packetloss_total_percent",
			"The average packet loss for all data on the virtual server",
			serverLabels, nil,
		),
		totalPing: prometheus.NewDesc(
			"teamspeak_virtualserver_ping_seconds",
			"The average ping of all clients connected to the virtual server",
			serverLabels, nil,
		),
		bytesReceivedControl: prometheus.NewDesc(
			"teamspeak_virtualserver_received_control_bytes_total",
			"Total amount of control data bytes received",
			serverLabels, nil,
		),
		bytesReceivedKeepalive: prometheus.NewDesc(
			"teamspeak_virtualserver_received_keepalive_bytes_total",
			"Total amount of keepalive data bytes received",
			serverLabels, nil,
		),
		bytesReceivedSpeech: prometheus.NewDesc(
			"teamspeak_virtualserver_received_speech_bytes_total",
			"Total amount of speech data bytes received",
			serverLabels, nil,
		),
		bytesReceivedTotal: prometheus.NewDesc(
			"teamspeak_virtualserver_received_bytes_total",
			"Total amount of bytes received",
			serverLabels, nil,
		),
		bytesSentControl: prometheus.NewDesc(
			"teamspeak_virtualserver_sent_control_bytes_total",
			"Total amount of control data bytes sent",
			serverLabels, nil,
		),
		bytesSentKeepalive: prometheus.NewDesc(
			"teamspeak_virtualserver_sent_keepalive_bytes_total",
			"Total amount of keepalive data bytes sent",
			serverLabels, nil,
		),
		bytesSentSpeech: prometheus.NewDesc(
			"teamspeak_virtualserver_sent_speech_bytes_total",
			"Total amount of speech data bytes sent",
			serverLabels, nil,
		),
		bytesSentTotal: prometheus.NewDesc(
			"teamspeak_virtualserver_sent_bytes_total",
			"Total amount of bytes sent",
			serverLabels, nil,
		),
		fileTransferBytesSentTotal: prometheus.NewDesc(
			"teamspeak_virtualserver_file_transfer_sent_bytes_total",
			"Total amount of filetransfer data bytes sent",
			serverLabels, nil,
		),
		fileTransferBytesReceivedTotal: prometheus.NewDesc(
			"teamspeak_virtualserver_file_transfer_received_bytes_total",
			"Total amount of filetransfer data bytes received",
			serverLabels, nil,
		),
		packetsReceivedControl: prometheus.NewDesc(
			"teamspeak_virtualserver_received_control_packets_total",
			"Total amount of control data packets received",
			serverLabels, nil,
		),
		packetsReceivedKeepalive: prometheus.NewDesc(
			"teamspeak_virtualserver_received_keepalive_packets_total",
			"Total amount of keepalive data packets received",
			serverLabels, nil,
		),
		packetsReceivedSpeech: prometheus.NewDesc(
			"teamspeak_virtualserver_received_speech_packets_total",
			"Total amount of speech data packets received",
			serverLabels, nil,
		),
		packetsReceivedTotal: prometheus.NewDesc(
			"teamspeak_virtualserver_received_packets_total",
			"Total amount of packets received",
			serverLabels, nil,
		),
		packetsSentControl: prometheus.NewDesc(
			"teamspeak_virtualserver_sent_control_packets_total",
			"Total amount of control data packets sent",
			serverLabels, nil,
		),
		packetsSentKeepalive: prometheus.NewDesc(
			"teamspeak_virtualserver_sent_keepalive_packets_total",
			"Total amount of keepalive data packets sent",
			serverLabels, nil,
		),
		packetsSentSpeech: prometheus.NewDesc(
			"teamspeak_virtualserver_sent_speech_packets_total",
			"Total amount of speech data packets sent",
			serverLabels, nil,
		),
		packetsSentTotal: prometheus.NewDesc(
			"teamspeak_virtualserver_sent_packets_total",
			"Total amount of packets sent",
			serverLabels, nil,
		),
	}
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	apiUp := 0.0

	// Always show 'teamspeak_up' metric, even on error.
	defer func() {
		ch <- prometheus.MustNewConstMetric(c.apiUp, prometheus.GaugeValue, apiUp)
	}()

	versionInfo, err := c.client.Version()
	if err != nil {
		slog.Error("Failed to get version info", "error", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.versionInfo, prometheus.GaugeValue, 1.0, versionInfo.Version, strconv.Itoa(versionInfo.Build), versionInfo.Platform)

	servers, err := c.client.VirtualServerList()
	if err != nil {
		slog.Error("Failed to list virtual servers", "error", err)
		return
	}

	for _, server := range servers {
		idString := strconv.Itoa(server.ID)

		isOnline := 0.0
		if strings.EqualFold(server.Status, "online") {
			isOnline = 1.0
		}

		ch <- prometheus.MustNewConstMetric(c.status, prometheus.GaugeValue, parseServerStatus(server.Status), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.serverUp, prometheus.GaugeValue, isOnline, idString, server.Name)

		// If a virtual server is offline, sending a HTTP request to its endpoint (/<id>/serverinfo) will start up the server in some kind of virtual online mode.
		// To prevent this and get the same behavior as using the raw ServerQuery (where serverinfo is not available), skip more detailed stats.
		if isOnline == 0.0 {
			continue
		}

		// This could be run concurrently. I don't understand how yet, so something for the future. Could be an issue if you have many virtual servers.
		info, err := c.client.VirtualServerInfo(server.ID)
		if err != nil {
			slog.Error("Failed to get info for virtual server",
				"server_id", server.ID,
				"error", err,
			)
			continue
		}

		// All HTTP requests have been successful so mark the scrape status as up
		apiUp = 1.0

		ch <- prometheus.MustNewConstMetric(c.maxClients, prometheus.GaugeValue, float64(info.MaxClients), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.clientsOnline, prometheus.GaugeValue, float64(info.ClientsOnline), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.channelsOnline, prometheus.GaugeValue, float64(info.ChannelsOnline), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.uptime, prometheus.GaugeValue, float64(info.Uptime), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.queryClientsOnline, prometheus.GaugeValue, float64(info.QueryClientsOnline), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.clientConnections, prometheus.CounterValue, float64(info.ClientConnections), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.queryClientConnections, prometheus.CounterValue, float64(info.QueryClientConnections), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.totalPacketLossSpeech, prometheus.GaugeValue, info.TotalPacketLossSpeech, idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.totalPacketLossKeepalive, prometheus.GaugeValue, info.TotalPacketLossKeepalive, idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.totalPacketLossControl, prometheus.GaugeValue, info.TotalPacketLossControl, idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.totalPacketLossTotal, prometheus.GaugeValue, info.TotalPacketLossTotal, idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.totalPing, prometheus.GaugeValue, float64(info.TotalPing)/1000.0, idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.bytesReceivedControl, prometheus.CounterValue, float64(info.BytesReceivedControl), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.bytesReceivedKeepalive, prometheus.CounterValue, float64(info.BytesReceivedKeepalive), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.bytesReceivedSpeech, prometheus.CounterValue, float64(info.BytesReceivedSpeech), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.bytesReceivedTotal, prometheus.CounterValue, float64(info.BytesReceivedTotal), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.bytesSentControl, prometheus.CounterValue, float64(info.BytesSentControl), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.bytesSentKeepalive, prometheus.CounterValue, float64(info.BytesSentKeepalive), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.bytesSentSpeech, prometheus.CounterValue, float64(info.BytesSentSpeech), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.bytesSentTotal, prometheus.CounterValue, float64(info.BytesSentTotal), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.fileTransferBytesSentTotal, prometheus.CounterValue, float64(info.FileTransferBytesSentTotal), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.fileTransferBytesReceivedTotal, prometheus.CounterValue, float64(info.FileTransferBytesReceivedTotal), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.packetsReceivedControl, prometheus.CounterValue, float64(info.PacketsReceivedControl), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.packetsReceivedKeepalive, prometheus.CounterValue, float64(info.PacketsReceivedKeepalive), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.packetsReceivedSpeech, prometheus.CounterValue, float64(info.PacketsReceivedSpeech), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.packetsReceivedTotal, prometheus.CounterValue, float64(info.PacketsReceivedTotal), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.packetsSentControl, prometheus.CounterValue, float64(info.PacketsSentControl), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.packetsSentKeepalive, prometheus.CounterValue, float64(info.PacketsSentKeepalive), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.packetsSentSpeech, prometheus.CounterValue, float64(info.PacketsSentSpeech), idString, server.Name)
		ch <- prometheus.MustNewConstMetric(c.packetsSentTotal, prometheus.CounterValue, float64(info.PacketsSentTotal), idString, server.Name)
	}
}
