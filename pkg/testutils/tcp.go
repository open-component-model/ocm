// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package testutils

import (
	"context"
	"github.com/mandelsoft/goutils/errors"
	"net"
	"time"
)

func PingTCPServer(address string, dur time.Duration) error {
	var conn net.Conn
	var d net.Dialer

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	end := time.Now().Add(dur)
	err := errors.New("timed out waiting for server to start")
	for time.Now().Before(end) {
		conn, err = d.DialContext(ctx, "tcp", address)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		conn.Close()
		break
	}
	return err
}
