package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	iradix "github.com/hashicorp/go-immutable-radix"
	"github.com/jedisct1/dlog"
	"github.com/miekg/dns"
)

type PluginAllowedIP struct {
	allowedPrefixes *iradix.Tree
	logger          io.Writer
	format          string
}

func (plugin *PluginAllowedIP) Name() string {
	return "allow_ip"
}

func (plugin *PluginAllowedIP) Description() string {
	return "Allows DNS queries only from specific IP ranges"
}

func (plugin *PluginAllowedIP) Init(proxy *Proxy) error {
	dlog.Notice("Initializing allowed IP rules for 192.168.0.0/24")

	plugin.allowedPrefixes = iradix.New()
	_, cidr, _ := net.ParseCIDR("192.168.0.0/24")
	plugin.allowedPrefixes, _, _ = plugin.allowedPrefixes.Insert([]byte(cidr.String()), true)

	if len(proxy.allowedIPLogFile) > 0 {
		plugin.logger = Logger(proxy.logMaxSize, proxy.logMaxAge, proxy.logMaxBackups, proxy.allowedIPLogFile)
		plugin.format = proxy.allowedIPFormat
	}

	return nil
}

func (plugin *PluginAllowedIP) Drop() error {
	return nil
}

func (plugin *PluginAllowedIP) Reload() error {
	return nil
}

func (plugin *PluginAllowedIP) Eval(pluginsState *PluginsState, msg *dns.Msg) error {
	clientIPStr := ""
	switch pluginsState.clientProto {
	case "udp":
		clientIPStr = (*pluginsState.clientAddr).(*net.UDPAddr).IP.String()
	case "tcp", "local_doh":
		clientIPStr = (*pluginsState.clientAddr).(*net.TCPAddr).IP.String()
	default:
		return nil
	}

	clientIP := net.ParseIP(clientIPStr)
	if clientIP == nil {
		return errors.New("invalid client IP")
	}

	_, cidr, _ := net.ParseCIDR("192.168.0.0/24")
	if !cidr.Contains(clientIP) {
		dlog.Warnf("Blocked DNS request from unauthorized IP: %s", clientIPStr)
		return errors.New("unauthorized IP")
	}

	if plugin.logger != nil {
		qName := pluginsState.qName
		now := time.Now()
		logEntry := fmt.Sprintf("[%s] Allowed: %s queried %s\n", now.Format(time.RFC3339), clientIPStr, qName)
		_, _ = plugin.logger.Write([]byte(logEntry))
	}

	return nil
}
