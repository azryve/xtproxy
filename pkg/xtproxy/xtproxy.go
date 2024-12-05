package xtproxy

import (
	"errors"
	"net"

	"github.com/azryve/xtproxy/pkg/aferomount"
	"github.com/spf13/afero"
	"golang.org/x/sync/errgroup"
)

type waiter interface {
	Wait() error
}

type XTProxy struct {
	mountfs *aferomount.MountFs
	waiters []waiter
	http    *XTProxyHTTPWebdav
}
type XTProxyOpt func(m *XTProxy) error

func NewXTProxy(opts ...XTProxyOpt) (*XTProxy, error) {
	fproxy := &XTProxy{
		mountfs: aferomount.NewMountFS(afero.NewMemMapFs()),
		waiters: make([]waiter, 0),
	}
	for _, opt := range opts {
		if err := opt(fproxy); err != nil {
			return nil, err
		}
	}
	return fproxy, nil
}

func (m *XTProxy) Wait() error {
	if len(m.waiters) == 0 {
		return errors.New("nothing to wait")
	}
	g := errgroup.Group{}
	for _, w := range m.waiters {
		g.Go(w.Wait)
	}
	return g.Wait()
}

func WithFTPAddr(addr *net.TCPAddr) XTProxyOpt {
	return func(m *XTProxy) error {
		ftp := &XTProxyFTP{Fs: m.mountfs, ListenAddr: addr}
		m.waiters = append(m.waiters, ftp)
		return nil
	}
}

func WithTFTPAddr(addr *net.UDPAddr) XTProxyOpt {
	return func(m *XTProxy) error {
		tftp := &XTProxyTFTP{Fs: m.mountfs, ListenAddr: addr}
		m.waiters = append(m.waiters, tftp)
		return nil
	}
}

func WithHTTPAddr(addr *net.TCPAddr) XTProxyOpt {
	return func(m *XTProxy) error {
		listener, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return err
		}
		if m.http == nil {
			m.http = &XTProxyHTTPWebdav{Fs: m.mountfs}
		}
		m.http.Listener = listener
		m.waiters = append(m.waiters, m.http)
		return nil
	}
}

func WithWebdavHandle(webdavHandle string) XTProxyOpt {
	return func(m *XTProxy) error {
		if m.http == nil {
			m.http = &XTProxyHTTPWebdav{Fs: m.mountfs}
		}
		m.http.WebdavHandle = webdavHandle
		return nil
	}
}

func WithMount(fs afero.Fs, path string) XTProxyOpt {
	return func(m *XTProxy) error {
		return m.mountfs.Mount(fs, path)
	}
}
