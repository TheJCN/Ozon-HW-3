package main

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"

	"OzonHW2/internal/product"
	"github.com/stretchr/testify/require"
)

type closedListener struct{}

func (closedListener) Accept() (net.Conn, error) { return nil, errors.New("closed") }
func (closedListener) Close() error              { return nil }
func (closedListener) Addr() net.Addr            { return &net.TCPAddr{} }

func TestServeWithClosedListener(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "products.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"1":{"name":"one","price":10}}`), 0600))
	catalog, err := product.Load(path)
	require.NoError(t, err)
	require.Error(t, serve(closedListener{}, catalog))
}
