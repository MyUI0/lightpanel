package handlers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func validateRemoteURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Hostname() == "" || isPrivateIP(u.Hostname()) {
		return fmt.Errorf("仅允许公共网络 HTTP/HTTPS 地址")
	}
	return nil
}

func publicIP(ip net.IP) bool {
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	// Shared address space is also not an Internet destination.
	_, shared, _ := net.ParseCIDR("100.64.0.0/10")
	return !shared.Contains(ip)
}

// Resolve once, check every result and dial that exact IP (no second DNS lookup).
func publicDial(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("域名没有可用地址")
	}
	for _, ip := range ips {
		if !publicIP(ip.IP) || ip.Zone != "" {
			return nil, fmt.Errorf("目标解析到非公共网络地址")
		}
	}
	dialer := net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	for _, ip := range ips {
		conn, e := dialer.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
		if e == nil {
			return conn, nil
		}
		err = e
	}
	return nil, err
}

var publicTransport http.RoundTripper = &http.Transport{
	DialContext: publicDial, TLSHandshakeTimeout: 10 * time.Second, ResponseHeaderTimeout: 15 * time.Second,
	IdleConnTimeout: 60 * time.Second, MaxIdleConns: 20, MaxIdleConnsPerHost: 2,
	// Deliberately do not inherit an environment proxy that could resolve an unchecked target.
}

type checkedTransport struct{}

func (checkedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if err := validateRemoteURL(r.URL.String()); err != nil {
		return nil, err
	}
	return publicTransport.RoundTrip(r)
}
func remoteClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout, Transport: checkedTransport{}, CheckRedirect: func(r *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("重定向次数过多")
		}
		if len(via) > 0 && via[0].URL.Scheme == "https" && r.URL.Scheme != "https" {
			return fmt.Errorf("拒绝 HTTPS 降级重定向")
		}
		return validateRemoteURL(r.URL.String())
	}}
}
func downloadFilename(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "download"
	}
	parts := strings.Split(u.Path, "/")
	name := parts[len(parts)-1]
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "\\/:?*\x00\"<>|") {
		return "download"
	}
	return name
}
