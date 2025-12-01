package main

import (
	"bufio" // BSD-3
	"context"
	"net"
	"regexp"

	// BSD-3
	"os"
	"strconv"
)

var re = regexp.MustCompile(`.*\s(@xdebug-ctrl\.(\d+)(yx+)?).*`)

func findFiles() (map[int]string, error) {
	file, _ := os.Open("/proc/net/unix")

	s := bufio.NewScanner(file)
	v := make(map[int]string)

	for s.Scan() {
		matches := re.FindStringSubmatch(s.Text())
		if len(matches) > 0 {
			pid, _ := strconv.Atoi(matches[2])
			v[pid] = matches[1]
		}
	}

	return v, nil
}

func dialCtrlSocket(ctx context.Context, ctrl_socket string) (net.Conn, error) {

	var d net.Dialer

	d.LocalAddr = nil
	raddr := net.UnixAddr{Name: ctrl_socket, Net: "unix"}
	conn, err := d.DialContext(ctx, "unix", raddr.String())

	return conn, err
}
