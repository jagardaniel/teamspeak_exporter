package main

import (
	"log"

	"github.com/alecthomas/kong"
)

func main() {
	config := &Config{}
	kong.Parse(config)

	client := NewClient(config.URL, config.APIKey)

	version, err := client.Version()
	if err != nil {
		log.Println(err)
	}

	servers, err := client.VirtualServerList()
	if err != nil {
		log.Println(err)
	}

	for _, server := range servers {
		info, err := client.VirtualServerInfo(server.ID)
		if err != nil {
			log.Println(err)
		}

		log.Println(info)
	}

	log.Println(version)
	log.Println(servers)
}
