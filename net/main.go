package main

import (
	"os"

	grpc "google.golang.org/grpc"

	"github.com/bserdar/watermelon-modules/net/firewalld"
	"github.com/bserdar/watermelon/client"
)

func main() {
	firewallServer := firewalld.Server{}
	client.Run(os.Args[1:], nil, func(server *grpc.Server, rt *client.Runtime) {
		rt.RegisterGRPCServer(&firewallServer, "firewalld")
		firewalld.RegisterFirewalldServer(server, firewallServer)
	})
}
