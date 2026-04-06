package handlers

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"

	"lightpanel/config"
)

func LoadJSON(path string, v any) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, v)
}

func WriteJSON(path string, v any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmp := path + ".tmp"
	f, e := os.Create(tmp)
	if e != nil {
		return e
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err := enc.Encode(v)
	if err == nil {
		err = f.Close()
	} else {
		f.Close()
	}

	if err != nil {
		_ = os.Remove(tmp)
		return err
	}

	_ = os.Chmod(tmp, 0600)
	return os.Rename(tmp, path)
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func getLogoURL() string {
	type LogoConfig struct {
		URL string `json:"url"`
	}
	var logo LogoConfig
	_ = LoadJSON(config.ConfigDir+"/logo.json", &logo)
	return logo.URL
}

func getBgURL() string {
	type BgConfig struct {
		URL string `json:"url"`
	}
	var bg BgConfig
	_ = LoadJSON(config.ConfigBg, &bg)
	return bg.URL
}

func getLocalIP() string {
	addrs, _ := net.InterfaceAddrs()
	var privateIPs []net.IP
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP
				if isPrivateIPAddr(ip) {
					privateIPs = append(privateIPs, ip)
				}
			}
		}
	}
	if len(privateIPs) > 0 {
		return privateIPs[0].String()
	}
	return "127.0.0.1"
}

func isPrivateIPAddr(ip net.IP) bool {
	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}
	if ip4[0] == 10 {
		return true
	}
	if ip4[0] == 192 && ip4[1] == 168 {
		return true
	}
	if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
		return true
	}
	return false
}
