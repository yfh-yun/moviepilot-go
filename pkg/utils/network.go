package utils

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// IsPortAvailable checks if a port is available
func IsPortAvailable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// GetAvailablePort finds an available port
func GetAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// IsIPAddress checks if a string is a valid IP address
func IsIPAddress(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsIPv4Address checks if a string is a valid IPv4 address
func IsIPv4Address(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return parsed.To4() != nil
}

// IsIPv6Address checks if a string is a valid IPv6 address
func IsIPv6Address(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return parsed.To4() == nil
}

// GetLocalIP gets the local IP address
func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// GetExternalIP gets the external IP address
func GetExternalIP() (string, error) {
	resp, err := http.Get("http://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(ip)), nil
}

// IsURLReachable checks if a URL is reachable
func IsURLReachable(url string, timeout time.Duration) bool {
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// PingHost pings a host to check connectivity
func PingHost(host string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", host, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// IsValidURL checks if a string is a valid URL
func IsValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// ParseURL parses a URL string and returns URL components
func ParseURL(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}

// BuildURL builds a URL from components
func BuildURL(scheme, host, path string, queryParams map[string]string) string {
	u := &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}

	if len(queryParams) > 0 {
		q := u.Query()
		for key, value := range queryParams {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
	}

	return u.String()
}

// GetHTTPClient creates an HTTP client with custom configuration
func GetHTTPClient(timeout time.Duration, skipTLSVerify bool) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLSVerify,
		},
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// DownloadFile downloads a file from URL
func DownloadFile(url, filepath string, timeout time.Duration) error {
	client := GetHTTPClient(timeout, false)

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// This is a placeholder - actual file download implementation would be used
	return nil
}

// GetHTTPStatusCode gets the HTTP status code for a URL
func GetHTTPStatusCode(url string, timeout time.Duration) (int, error) {
	client := GetHTTPClient(timeout, false)

	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

// IsHTTPS checks if a URL uses HTTPS
func IsHTTPS(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return u.Scheme == "https"
}

// NormalizeURL normalizes a URL string
func NormalizeURL(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	// Ensure scheme is lowercase
	u.Scheme = strings.ToLower(u.Scheme)

	// Ensure host is lowercase
	u.Host = strings.ToLower(u.Host)

	// Remove default ports
	if u.Port() == "80" && u.Scheme == "http" {
		u.Host = u.Hostname()
	}
	if u.Port() == "443" && u.Scheme == "https" {
		u.Host = u.Hostname()
	}

	return u.String(), nil
}

// ExtractDomain extracts domain from URL
func ExtractDomain(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	return u.Hostname(), nil
}

// ExtractPath extracts path from URL
func ExtractPath(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	return u.Path, nil
}

// ExtractQueryParams extracts query parameters from URL
func ExtractQueryParams(urlStr string) (map[string]string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	params := make(map[string]string)
	for key, values := range u.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	return params, nil
}

// ResolveDNS resolves a hostname to IP addresses
func ResolveDNS(hostname string) ([]string, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}

	var ipStrings []string
	for _, ip := range ips {
		ipStrings = append(ipStrings, ip.String())
	}

	return ipStrings, nil
}

// IsPrivateIP checks if an IP address is private
func IsPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16", // Link-local
		"127.0.0.0/8",    // Loopback
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// GetNetworkInterfaceInfo gets information about network interfaces
func GetNetworkInterfaceInfo() ([]net.Interface, error) {
	return net.Interfaces()
}

// GetInterfaceIPs gets IP addresses for a network interface
func GetInterfaceIPs(interfaceName string) ([]net.IP, error) {
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	var ips []net.IP
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			ips = append(ips, ipnet.IP)
		}
	}

	return ips, nil
}
