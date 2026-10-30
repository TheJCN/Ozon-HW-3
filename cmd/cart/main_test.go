package main

import (
	"errors"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type closedListener struct{}

func (closedListener) Accept() (net.Conn, error) { return nil, errors.New("closed") }
func (closedListener) Close() error              { return nil }
func (closedListener) Addr() net.Addr            { return &net.TCPAddr{} }

func TestServeWithClosedListener(t *testing.T) {
	t.Parallel()
	require.Error(t, serve(closedListener{}, http.NewServeMux()))
}
