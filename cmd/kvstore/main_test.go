package main

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	go startServer()
	m.Run()
}

func TestClient(t *testing.T) {
	for _, tc := range []struct {
		desc    string
		request func(*net.UDPConn)
		expResp string
	}{
		{
			desc: "key-value one write one read",
			request: func(uc *net.UDPConn) {
				uc.Write([]byte("foo=bar"))
				uc.Write([]byte("foo"))
			},
			expResp: "foo=bar",
		},
		{
			desc: "write with double delimiter",
			request: func(uc *net.UDPConn) {
				uc.Write([]byte("foo=bar=baz"))
				uc.Write([]byte("foo"))
			},
			expResp: "foo=bar=baz",
		},
		{
			desc: "write with tripple delimiter",
			request: func(uc *net.UDPConn) {
				uc.Write([]byte("foo==="))
				uc.Write([]byte("foo"))
			},
			expResp: "foo===",
		},
		{
			desc: "write with empty value",
			request: func(uc *net.UDPConn) {
				uc.Write([]byte("foo="))
				uc.Write([]byte("foo"))
			},
			expResp: "foo=",
		},
		{
			desc: "write with key",
			request: func(uc *net.UDPConn) {
				uc.Write([]byte("=foo"))
				uc.Write([]byte(""))
			},
			expResp: "=foo",
		},
		{
			desc: "missing key",
			request: func(uc *net.UDPConn) {
				uc.Write([]byte("foo"))
			},
			expResp: "foo=",
		},
		{
			desc: "key-value two writes one read",
			request: func(uc *net.UDPConn) {
				uc.Write([]byte("key2=value2"))
				uc.Write([]byte("key3=value3"))
				uc.Write([]byte("key3"))
			},
			expResp: "key3=value3",
		},
		{
			desc: "update existing key",
			request: func(uc *net.UDPConn) {
				uc.Write([]byte("key3=value3"))
				uc.Write([]byte("key3=value2"))
				uc.Write([]byte("key3"))
			},
			expResp: "key3=value2",
		},
		{
			desc: "version",
			request: func(uc *net.UDPConn) {
				uc.Write([]byte("version"))
			},
			expResp: fmt.Sprintf("version=%s", ProductVersion),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("localhost:%d", serverPort))
			require.NoError(t, err)
			c, err := net.DialUDP("udp4", nil, s)
			require.NoError(t, err)

			tc.request(c)

			buffer := make([]byte, 1000)
			n, _, err := c.ReadFromUDP(buffer)
			require.NoError(t, err)
			require.Equal(t, tc.expResp, string(buffer[0:n]))
		})
	}
}
