package main

import (
	"os"

	grpc "google.golang.org/grpc"

	"github.com/bserdar/watermelon-modules/pkg/yum"
	"github.com/bserdar/watermelon/client"
)

func main() {
	yumServer := yum.Server{}
	client.Run(os.Args[1:], nil, func(server *grpc.Server, rt *client.Runtime) {
		rt.RegisterGRPCServer(&yumServer, "yum")
		yum.RegisterYumServer(server, yumServer)
	})
}
