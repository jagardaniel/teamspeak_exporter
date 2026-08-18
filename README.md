# TeamSpeak exporter

Prometheus exporter for TeamSpeak servers. It uses the TeamSpeak WebQuery API to collect metrics.

### Metrics

The exporter collects some overall status and more detailed statistics for each virtual server that is online.

Example metrics:

- **Overall**: `teamspeak_up`, `teamspeak_version_info`
- **Per virtual server**: `teamspeak_virtualserver_up`, `teamspeak_virtualserver_clients_online`, `teamspeak_virtualserver_uptime`, `teamspeak_virtualserver_packetloss_control_percent`, `teamspeak_virtualserver_sent_bytes_total`

Some notes:

- `...clients_online` also includes query clients. You can subtract `...query_clients_online` to get the count of regular voice clients.
- `...received_bytes_total`, `...sent_bytes_total`, and their packet counterparts (`...packets_total`) do not include file transfer data. File transfers have their own dedicated bytes metrics (`...*_file_transfer_bytes_total`).

### Prerequisites

To use TeamSpeak's WebQuery you need an API key. An API key is created when you start the server for the first time.

```
------------------------------------------------------------------
                      I M P O R T A N T
------------------------------------------------------------------
               Server Query Admin Account created
         loginname= "serveradmin", password= "cIIxkT0T"
         apikey= "BADy2WBwAHAeFyRqTXOoT5kKgIR1889C1WDxlhm"
------------------------------------------------------------------
```

It is also possible to manually create a new API key over `raw` (telnet) or `SSH` query using `apikeyadd`. The scope has to be set to `manage` and `lifetime=0` ensures the API key does not expire.

```
serveradmin> apikeyadd scope=manage lifetime=0
apikey=BADy2WBwAHAeFyRqTXOoT5kKgIR1889C1WDxlhm id=2 sid=0 cldbid=1 scope=manage time_left=unlimited created_at=1786880137 expires_at=1786880137
error id=0 msg=ok
```

Make sure to enable `http` as query protocol when you start the TeamSpeak server. You can use a tool like `curl` to verify that the WebQuery API is reachable and that your API key has the correct scope:

```bash
curl -X GET "http://192.168.0.164:10080/serverlist?api-key=BADy2WBwAHAeFyRqTXOoT5kKgIR1889C1WDxlhm"
```

TeamSpeak's documentation recommends placing the WebQuery port (10080) behind a reverse proxy with HTTPS enabled. Especially if you need the WebQuery exposed to the internet.

### Usage

| Flag      | Environment Variable | Default                | Description                                           |
| --------- | -------------------- | ---------------------- | ----------------------------------------------------- |
| --listen  | TS_EXPORTER_LISTEN   | :9289                  | Address on which to expose metrics and web interface. |
| --api-key | TS_EXPORTER_API_KEY  | _(Required)_           | API key for TeamSpeak WebQuery authentication.        |
| --url     | TS_EXPORTER_URL      | http://127.0.0.1:10080 | URL for TeamSpeak WebQuery endpoint.                  |

```bash
# Using command-line flags
./teamspeak_exporter --url "http://192.168.0.100:10080" --api-key "BADy2WBwAHAeFyRqTXOoT5kKgIR1889C1WDxlhm"

# Using environment variables
TS_EXPORTER_URL="http://192.168.0.100:10080" TS_EXPORTER_API_KEY="BADy2WBwAHAeFyRqTXOoT5kKgIR1889C1WDxlhm" ./teamspeak_exporter
```

### Build and test

```bash
# Build
go build -o teamspeak_exporter .

# Run tests
go test ./...
```

### Docker / Podman

A container image is also available on Docker Hub. Example using the latest tag:

```bash
podman run --rm -p 9289:9289 -e TS_EXPORTER_URL="http://192.168.0.100:10080" -e TS_EXPORTER_API_KEY="BADy2WBwAHAeFyRqTXOoT5kKgIR1889C1WDxlhm" docker.io/jagardaniel/teamspeak_exporter:latest
```

Build a local container:

```bash
podman build -t teamspeak_exporter:local .
```

### TeamSpeak version support

It has been tested and working against **TeamSpeak 3** and **TeamSpeak 6 beta**. I haven't been able to to verify with multiple virtual servers for TeamSpeak 6 but the WebQuery responses seems to be almost the same. Note that TeamSpeak 6 is still under active development so things can change quickly.

### About

This project is mainly for learning purposes and to build something that I use. There are other TeamSpeak exporters out there that do a better job.

I would lie if I said that I didn't get a lot of help from Google Gemini, especially when it comes to split things up and where they should be (which I now realize is a big part of programming). I can usually figure out how to do the things I want with some old StackOverflow and blog posts, but it ends up in the same function and without any structure. If something looks clever, it's not me.
