package utils

import (
	log "github.com/Sirupsen/logrus"
	"google.golang.org/grpc"
)

func ConnectEru(server string, timeout int) *grpc.ClientConn {
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("[ConnectEru] Can not connect %v", err)
	}
	log.Debugf("[ConnectEru] Init eru connection %s", server)
	return conn
}
