package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/netip"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/azryve/xtproxy/pkg/xtproxy"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var debugFlag bool
var writableFlag bool
var ifacesListen []string
var ftpPort = 21
var tftpPort = 69
var httpPort = 80
var webdavHandle = "/.webdav"
var defaultAddr = netip.MustParseAddr("::")
var errUsage = errors.New("error usage")
var envPrefix = "XTPROXY"

type mountFs struct {
	URL  *url.URL
	Path string
	Fs   afero.Fs
}

var rootCmd = &cobra.Command{
	Use:   "xtproxy",
	Short: "xtproxy serves files with ftp/tftp",
	RunE: func(cmd *cobra.Command, args []string) error {
		parseArgs()
		return mainServe(args)
	},
}

func fatal(err error) {
	if err != nil {
		panic(err)
	}
}

func parseArgs() {
	debugFlag = viper.GetBool("debug")
	ifacesListen = viper.GetStringSlice("ifaces-listen")
	ftpPort = viper.GetInt("port-ftp")
	tftpPort = viper.GetInt("port-tftp")
	httpPort = viper.GetInt("port-http")
	webdavHandle = viper.GetString("webdav-handle")
}

func bindArgs() {
	var key string

	key = "debug"
	rootCmd.PersistentFlags().BoolVar(&debugFlag, key, false, "enable debuging")
	fatal(viper.BindPFlag(key, rootCmd.PersistentFlags().Lookup(key)))
	viper.SetDefault(key, false)

	key = "ifaces-listen"
	rootCmd.Flags().StringArrayVarP(&ifacesListen, key, "i", []string{}, "listen all addreses on specific ifaces")
	fatal(viper.BindPFlag(key, rootCmd.Flags().Lookup(key)))
	viper.SetDefault(key, []string{})

	key = "port-ftp"
	rootCmd.Flags().IntVar(&ftpPort, key, ftpPort, "ftp tcp port")
	fatal(viper.BindPFlag(key, rootCmd.Flags().Lookup(key)))
	viper.SetDefault(key, ftpPort)

	key = "port-tftp"
	rootCmd.Flags().IntVar(&tftpPort, key, tftpPort, "tftp udp port")
	fatal(viper.BindPFlag(key, rootCmd.Flags().Lookup(key)))
	viper.SetDefault(key, tftpPort)

	key = "port-http"
	rootCmd.Flags().IntVar(&httpPort, key, httpPort, "http tcp port")
	fatal(viper.BindPFlag(key, rootCmd.Flags().Lookup(key)))
	viper.SetDefault(key, httpPort)

	key = "webdav-handle"
	rootCmd.Flags().StringVar(&webdavHandle, key, webdavHandle, "webdav handle for http server")
	fatal(viper.BindPFlag(key, rootCmd.Flags().Lookup(key)))
	viper.SetDefault(key, webdavHandle)

	// disabled until testing
	// rootCmd.Flags().BoolVar(&writableFlag, "writable", false, "allow uploading")

	envReplacer := strings.NewReplacer("-", "_")
	envDecorator := func(f *pflag.Flag) {
		name := strings.ToUpper(envPrefix + "_" + f.Name)
		name = envReplacer.Replace(name)
		tag := fmt.Sprintf(" (env %s)", name)
		if !strings.Contains(f.Usage, tag) {
			f.Usage += tag
		}
	}
	rootCmd.PersistentFlags().VisitAll(envDecorator)
	rootCmd.Flags().VisitAll(envDecorator)
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(envReplacer)
	viper.AutomaticEnv()
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
	listenaddrs = slices.DeleteFunc(listenaddrs, func(addr netip.AddrPort) bool {
		return addr.Port() == 0
	})
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
	opts := make([]xtproxy.XTProxyOpt, 0)
	mounts, err := setupMountFs(args)
	if err != nil {
		return err
	}
	if !writableFlag {
		for i, m := range mounts {
			mounts[i].Fs = afero.NewReadOnlyFs(m.Fs)
		}
	}
	for _, m := range mounts {
		log.Printf("mounts %s -> %s\n", masked(m.URL).String(), m.Path)
		opts = append(opts, xtproxy.WithMount(m.Fs, m.Path))
	}
	listenaddrs, err := setupListenAddrs()
	if err != nil {
		return err
	}
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
			opts = append(opts, xtproxy.WithHTTPWebdavAddr(tcpaddr, webdavHandle))
		default:
			return fmt.Errorf("unknown port %d: %s", int(addrport.Port()), errUsage)
		}
	}
	fproxy, err := xtproxy.NewXTProxy(opts...)
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

func init() {
	bindArgs()
}
