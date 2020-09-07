package connections

import (
	"crypto/tls"
	"fmt"
	"github.com/derickr/dbgp-tools/lib/logger"
	"hash/crc32"
	"net"
)

func ConnectTo(address string, ssl bool) (net.Conn, error) {
	var conn net.Conn
	var err error
	// var cert tls.Certificate

	if ssl {
		/*
		cert, err = tls.LoadX509KeyPair("client-certs/client.pem", "client-certs/client.key")
		if err != nil {
			return nil, err
		}
		config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
		*/
		conn, err = tls.Dial("tcp", address, nil)

		if err != nil {
			return nil, err
		}
	} else {
		conn, err = net.Dial("tcp", address)

		if err != nil {
			return nil, err
		}
	}

	return conn, nil
}

func CloudHostFromUserId(domain string, port string, uid string) string {
	crc32v := crc32.ChecksumIEEE([]byte(uid))

	host := fmt.Sprintf("%c.%s:%s", (crc32v&0x0f)+'a'-1, domain, port)

	return host
}

func ConnectToCloud(domain string, port string, uid string, logger logger.Logger) (net.Conn, error) {
	host := CloudHostFromUserId(domain, port, uid)

	logger.LogInfo("utils", "Connecting to cloud host '%s'", host)

	return ConnectTo(host, true)
}
