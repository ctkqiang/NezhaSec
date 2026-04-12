package security

import (
	"net"
	"net/url"
	"strings"
)

func IsTargetAllowed(target string, allowLocal bool) (bool, string) {

	parsedURL, err := url.Parse(target)
	if err != nil {

		ip := net.ParseIP(target)
		if ip != nil {
			return isIPAllowed(ip, allowLocal)
		}

		if isLocalDomain(target) && !allowLocal {
			return false, "目标为本地域名，禁止扫描"
		}

		return true, ""
	}

	host := parsedURL.Hostname()
	if host == "" {

		host = target
		if host == "" {
			return false, "无效的主机名"
		}
	}

	if isSensitiveDomain(host) {
		return false, "目标包含敏感域名，禁止扫描"
	}

	ip := net.ParseIP(host)
	if ip != nil {
		return isIPAllowed(ip, allowLocal)
	}

	if isLocalDomain(host) && !allowLocal {
		return false, "目标为本地域名，禁止扫描"
	}

	return true, ""
}

func isSensitiveDomain(domain string) bool {
	sensitiveTLDs := []string{
		".gov",
		".mil",
		".gov.cn",
		".mil.cn",
	}

	for _, tld := range sensitiveTLDs {
		if strings.HasSuffix(domain, tld) {
			return true
		}
	}

	return false
}

func isIPAllowed(ip net.IP, allowLocal bool) (bool, string) {

	if ip.IsLoopback() && !allowLocal {
		return false, "目标为本地回环地址，禁止扫描"
	}

	if ip.IsPrivate() && !allowLocal {
		return false, "目标为内网IP地址，禁止扫描"
	}

	if ip.IsLinkLocalUnicast() && !allowLocal {
		return false, "目标为链路本地地址，禁止扫描"
	}

	return true, ""
}

func isLocalDomain(domain string) bool {
	localDomains := []string{
		"localhost",
		"local",
		"test",
		"example",
		"example.com",
	}

	for _, localDomain := range localDomains {
		if domain == localDomain {
			return true
		}
	}

	return false
}
