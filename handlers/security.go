package handlers

import (
	"net"
	"net/url"
	"strings"
)

var shellMetaChars = []string{"|", ";", "$", "`", "(", ")", "{", "}", "<", ">", "&", "\\", "\n", "\r"}

func validateCommand(cmd string) bool {
	if cmd == "" {
		return true
	}
	for _, c := range shellMetaChars {
		if strings.Contains(cmd, c) {
			return false
		}
	}
	return true
}

func isPrivateURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return true
	}
	host := u.Hostname()
	return isPrivateIP(host)
}

func isPrivateIP(host string) bool {
	if host == "" {
		return true
	}
	if strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.") || strings.HasPrefix(host, "0.") {
		return true
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return true
		}
	}
	return false
}
