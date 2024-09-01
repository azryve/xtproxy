package xtproxy

import (
	"io"
	"net"
	"os"

	"github.com/pin/tftp"
	"github.com/spf13/afero"
)

type XTProxyTFTP struct {
	Fs         afero.Fs
	ListenAddr *net.UDPAddr
	server     *tftp.Server
}

func (m *XTProxyTFTP) Wait() error {
	if err := m.init(); err != nil {
		return err
	}
	return m.server.ListenAndServe(m.ListenAddr.String())
}

func (m *XTProxyTFTP) init() error {
	if m.server != nil {
		return nil
	}
	m.server = tftp.NewServer(
		m.readHandler,
		m.writeHandler,
	)
	return nil
}

// readHandler is called when client starts file download from server
func (m *XTProxyTFTP) readHandler(filename string, rf io.ReaderFrom) error {
	fs := m.Fs
	file, err := fs.Open(filename)
	if err != nil {
		return err
	}
	_, err = rf.ReadFrom(file)
	if err != nil {
		return err
	}
	return nil
}

// writeHandler is called when client starts file upload to server
func (m *XTProxyTFTP) writeHandler(filename string, wt io.WriterTo) error {
	fs := m.Fs
	file, err := fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	_, err = wt.WriteTo(file)
	if err != nil {
		return err
	}
	return nil
}
