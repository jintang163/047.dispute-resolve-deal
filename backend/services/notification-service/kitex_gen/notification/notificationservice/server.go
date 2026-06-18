package notificationservice

import (
	"github.com/cloudwego/kitex/pkg/remote/trans/netpoll"

	"github.com/cloudwego/kitex/server"
)

func NewServer(handler NotificationService, opts ...server.Option) server.Server {
	svr := server.NewServer(opts...)
	return svr
}

func init() {
	netpoll.NewTransporter()
}
