package server

import (
	"strings"
)

const (
	StatusOK uint8 = iota
	StatusWarning
	StatusError
)

type Client struct {
	Name      string
	Type      string
	Source    string
	Target    string
	Status    uint8
	StatusMsg string // not sure how to use this in GTK 3. might scratch it
}

func (server *ServerInstance) GetClientList() []Client {
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
				ports = append(ports, int_to_string(int(port)))
			}

			list_item.Target = strings.Join(ports, ", ")

			list = append(list, list_item)
		}
	}

	if server.VTSPlugin != nil && server.VTSPlugin.authenticated {
		list_item := Client{
			Name:   "VTube Studio",
			Type:   "VTS Plugin",
			Target: int_to_string(Config.VTSPlugin.Port),
		}

		list = append(list, list_item)
	}

	if server.VMCApi != nil && server.VMCApi.client != nil {
		target := int_to_string(server.VMCApi.client.Port())

		list_item := Client{
			Name:   "",
			Type:   "VMC Protocol",
			Target: target,
		}

		list = append(list, list_item)
	}

	return list
}
