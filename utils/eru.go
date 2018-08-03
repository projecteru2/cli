package utils

import (
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// ConnectEru connect to eru
func ConnectEru(server string, opts []grpc.DialOption) *grpc.ClientConn {
	conn, err := grpc.Dial(server, opts...)
	if err != nil {
		log.Fatalf("[ConnectEru] Can not connect %v", err)
	}
	log.Debugf("[ConnectEru] Init eru connection %s", server)
	return conn
}
