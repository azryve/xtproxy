package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/netip"
	"net/url"
	"os"
	"strings"

	"github.com/azryve/xtproxy/pkg/aferomount"
	"github.com/azryve/xtproxy/pkg/xtproxy"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var debugFlag bool
var writableFlag bool
var ifacesListen []string
var ftpPort = 21
var tftpPort = 69
var httpPort = 80
var defaultAddr = netip.MustParseAddr("::")
var errUsage = errors.New("error usage")

type mountFs struct {
	URL  *url.URL
	Path string
	Fs   afero.Fs
}

var rootCmd = &cobra.Command{
	Use:   "xtproxy",
	Short: "xtproxy serves files with ftp/tftp",
	RunE: func(cmd *cobra.Command, args []string) error {
		return mainServe(args)
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "enable debuging")
	rootCmd.Flags().StringArrayVarP(&ifacesListen, "ifaces-listen", "i", []string{}, "listen all addreses on specific ifaces")
	rootCmd.Flags().IntVar(&ftpPort, "port-ftp", ftpPort, "ftp tcp port")
	rootCmd.Flags().IntVar(&tftpPort, "port-tftp", tftpPort, "tftp udp port")
	rootCmd.Flags().IntVar(&httpPort, "port-http", httpPort, "http tcp port")
	// disabled until testing
	// rootCmd.Flags().BoolVar(&writableFlag, "writable", false, "allow uploading")
}

func setupMountFs(args []string) ([]mountFs, error) {
	mounts := make([]mountFs, 0, len(args))
	for _, arg := range args {
		urlAndPath := strings.SplitN(arg, " ", 2)
		if len(urlAndPath) != 2 {
			return nil, fmt.Errorf("invalid arg expected <url> <path> got '%s': %w", arg, errUsage)
		}
		URL, err := url.Parse(urlAndPath[0])
		if err != nil {
			return nil, fmt.Errorf("invalid url '%s': %w: %w", urlAndPath[0], err, errUsage)
		}
		if URL.Scheme == "s3" {
			s3creds, ok := os.LookupEnv("XTPROXY_S3_CREDENTIALS")
			if !ok {
				return nil, errors.New("missing XTPROXY_S3_CREDENTIALS=<access_key>:<secret>")
			}
			userPass := strings.SplitN(s3creds, ":", 2)
			if len(userPass) != 2 {
				return nil, errors.New("invalid XTPROXY_S3_CREDENTIALS=<access_key>:<secret>")
			}
			URL.User = url.UserPassword(userPass[0], userPass[1])
		}
		fs, err := xtproxy.FsByURL(URL.String())
		if err != nil {
			return nil, fmt.Errorf("invalid fs url '%s': %w: %w", urlAndPath[0], err, errUsage)
		}
		if debugFlag {
			fs = &xtproxy.DebugFs{Fs: fs}
		}
		mounts = append(mounts, mountFs{
			URL:  URL,
			Fs:   fs,
			Path: urlAndPath[1],
		})
	}
	return mounts, nil
}

func setupListenAddrs() ([]netip.AddrPort, error) {
	listenaddrs := make([]netip.AddrPort, 0)
	if len(ifacesListen) == 0 {
		listenaddrs = append(listenaddrs, netip.AddrPortFrom(defaultAddr, uint16(ftpPort)))
		listenaddrs = append(listenaddrs, netip.AddrPortFrom(defaultAddr, uint16(tftpPort)))
		listenaddrs = append(listenaddrs, netip.AddrPortFrom(defaultAddr, uint16(httpPort)))
	}
	for _, ifaceName := range ifacesListen {
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			return nil, err
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip := netIPAddr(addr).WithZone(ifaceName)
			listenaddrs = append(listenaddrs, netip.AddrPortFrom(ip, uint16(ftpPort)))
			listenaddrs = append(listenaddrs, netip.AddrPortFrom(ip, uint16(tftpPort)))
			listenaddrs = append(listenaddrs, netip.AddrPortFrom(ip, uint16(httpPort)))
		}
	}
	return listenaddrs, nil
}

func netIPAddr(addr net.Addr) netip.Addr {
	switch v := addr.(type) {
	case *net.IPNet:
		return netip.MustParseAddr(v.IP.String())
	case *net.IPAddr:
		return netip.MustParseAddr(v.IP.String())
	case *net.TCPAddr:
		return netip.MustParseAddr(v.IP.String())
	case *net.UDPAddr:
		return netip.MustParseAddr(v.IP.String())
	default:
		return netip.Addr{}
	}
}

func masked(URL *url.URL) *url.URL {
	URL, _ = url.Parse(URL.String())
	URL.User = nil
	return URL
}

func mainServe(args []string) error {
	if len(args) == 0 {
		mountVal, ok := os.LookupEnv("XTPROXY_S3_MOUNTS")
		if !ok {
			return fmt.Errorf("missing mounts via args or XTPROXY_S3_MOUNTS=<url> <path>: %w", errUsage)
		}
		args = []string{mountVal}
	}
	mounts, err := setupMountFs(args)
	if err != nil {
		return err
	}
	rootfs := aferomount.NewMountFS(afero.NewMemMapFs())
	if !writableFlag {
		for i, m := range mounts {
			mounts[i].Fs = afero.NewReadOnlyFs(m.Fs)
		}
	}
	for _, m := range mounts {
		log.Printf("mounts %s -> %s\n", masked(m.URL).String(), m.Path)
		rootfs.Mount(m.Fs, m.Path)
	}
	listenaddrs, err := setupListenAddrs()
	if err != nil {
		return err
	}
	opts := make([]xtproxy.XTProxyOpt, 0)
	for _, addrport := range listenaddrs {
		switch int(addrport.Port()) {
		case ftpPort:
			tcpaddr := net.TCPAddrFromAddrPort(addrport)
			opts = append(opts, xtproxy.WithFTPAddr(tcpaddr))
		case tftpPort:
			udpaddr := net.UDPAddrFromAddrPort(addrport)
			opts = append(opts, xtproxy.WithTFTPAddr(udpaddr))
		case httpPort:
			tcpaddr := net.TCPAddrFromAddrPort(addrport)
			opts = append(opts, xtproxy.WithHTTPAddr(tcpaddr))
		default:
			return fmt.Errorf("unknown port %d: %s", int(addrport.Port()), errUsage)
		}
	}
	fproxy, err := xtproxy.NewXTProxy(rootfs, opts...)
	if err != nil {
		return err
	}
	log.Printf("listens on %s\n", listenaddrs)
	return fproxy.Wait()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		if errors.Is(err, errUsage) {
			rootCmd.Usage()
		}
		log.Fatal(err)
	}
}
