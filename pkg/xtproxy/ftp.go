package xtproxy

import (
	"crypto/tls"
	"net"
	"os"

	ftpserverlib "github.com/fclairamb/ftpserverlib"
	"github.com/spf13/afero"
)

type XTProxyFTP struct {
	Fs         afero.Fs
	ListenAddr *net.TCPAddr
	server     *ftpserverlib.FtpServer
}

func (m *XTProxyFTP) Wait() error {
	if m.server == nil {
		m.server = ftpserverlib.NewFtpServer(m)
	}
	return m.server.ListenAndServe()
}

type cdriver struct {
	afero.Fs
}

var _ ftpserverlib.MainDriver = &XTProxyFTP{}
var _ ftpserverlib.ClientDriver = &cdriver{}
var _ ftpserverlib.ClientDriverExtentionFileTransfer = &cdriver{}

// GetSettings returns some general settings around the server setup
func (m *XTProxyFTP) GetSettings() (*ftpserverlib.Settings, error) {
	return &ftpserverlib.Settings{
		ListenAddr: m.ListenAddr.String(),
	}, nil
}

// ClientConnected is called to send the very first welcome message
func (m *XTProxyFTP) ClientConnected(cc ftpserverlib.ClientContext) (string, error) {
	return "xtproxy ftp server", nil
}

// ClientDisconnected is called when the user disconnects, even if he never authenticated
func (m *XTProxyFTP) ClientDisconnected(cc ftpserverlib.ClientContext) {
}

// AuthUser is called when the user disconnects, even if he never authenticated
func (m *XTProxyFTP) AuthUser(cc ftpserverlib.ClientContext, user, pass string) (ftpserverlib.ClientDriver, error) {
	return &cdriver{m.Fs}, nil
}

// GetTLSConfig returns a TLS Certificate to use
func (m *XTProxyFTP) GetTLSConfig() (*tls.Config, error) {
	return nil, ErrNotImplemented
}

// ClientDriverExtentionFileTransfer is a convenience extension to allow to transfer files
// GetHandle return an handle to upload or download a file based on flags:
// os.O_RDONLY indicates a download
// os.O_WRONLY indicates an upload and can be combined with os.O_APPEND (resume) or
// os.O_CREATE (upload to new file/truncate)
// offset is the argument of a previous REST command, if any, or 0
func (m *cdriver) GetHandle(name string, flags int, offset int64) (ftpserverlib.FileTransfer, error) {
	return m.Fs.OpenFile(name, flags, os.ModePerm)
}
