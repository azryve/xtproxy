package xtproxy

import (
	"errors"
	"net"

	"github.com/spf13/afero"
	"golang.org/x/sync/errgroup"
)

type waiter interface {
	Wait() error
}

type XTProxy struct {
	Fs      afero.Fs
	waiters []waiter
}
type XTProxyOpt func(m *XTProxy) error

func NewXTProxy(fs afero.Fs, opts ...XTProxyOpt) (*XTProxy, error) {
	fproxy := &XTProxy{
		Fs:      fs,
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
		ftp := &XTProxyFTP{Fs: m.Fs, ListenAddr: addr}
		m.waiters = append(m.waiters, ftp)
		return nil
	}
}

func WithTFTPAddr(addr *net.UDPAddr) XTProxyOpt {
	return func(m *XTProxy) error {
		tftp := &XTProxyTFTP{Fs: m.Fs, ListenAddr: addr}
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
		http := &XTProxyHTTP{Fs: m.Fs, Listener: listener}
		m.waiters = append(m.waiters, http)
		return nil
	}
}
