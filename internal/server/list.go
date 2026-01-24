package server

import (
	"strconv"
	"strings"
)

type Client struct {
	Name   string
	Type   string
	Source string
	Target string
}

func (server *ServerData) GetClientList() []Client {
	length := 0

	if server.VTSApi != nil {
		length += len(server.VTSApi.clients)
	}

	if server.VTSPlugin != nil && server.VTSPlugin.authenticated {
		length++
	}

	list := make([]Client, 0, length)

	if server.VTSApi != nil {
		for clientId, client := range server.VTSApi.clients {
			list_item := Client{
				Name:   clientId,
				Type:   "VTS 3rd Party API",
				Source: client.source,
			}

			ports := make([]string, 0, len(client.message.Ports))

			for _, port := range client.message.Ports {
				ports = append(ports, strconv.FormatFloat(port, 'f', 0, 64))
			}

			list_item.Target = strings.Join(ports, ", ")

			list = append(list, list_item)
		}
	}

	if server.VTSPlugin != nil && server.VTSPlugin.authenticated {
		list_item := Client{
			Name:   "VTube Studio",
			Type:   "VTS Plugin",
			Target: strconv.Itoa(Config.VTSPlugin.Port), // convert int to string
		}

		list = append(list, list_item)
	}

	return list
}
