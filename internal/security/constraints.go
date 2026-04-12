package security

import (
	"net"
	"net/url"
	"strings"
)

// IsTargetAllowed 检查目标是否允许扫描
// 参数列表：
//   - target 目标URL或IP地址
//   - allowLocal 是否允许扫描本地地址
//
// 返回值列表：
//   - bool 是否允许扫描
//   - string 不允许的原因
func IsTargetAllowed(target string, allowLocal bool) (bool, string) {
	// 解析目标
	parsedURL, err := url.Parse(target)
	if err != nil {
		// 尝试作为IP地址解析
		ip := net.ParseIP(target)
		if ip != nil {
			return isIPAllowed(ip, allowLocal)
		}

		// 尝试作为域名解析
		if isLocalDomain(target) && !allowLocal {
			return false, "目标为本地域名，禁止扫描"
		}

		return true, ""
	}

	// 检查域名
	host := parsedURL.Hostname()
	if host == "" {
		// 如果主机名为空，可能是因为没有协议部分，尝试使用整个输入作为域名
		host = target
		if host == "" {
			return false, "无效的主机名"
		}
	}

	// 检查是否为敏感域名
	if isSensitiveDomain(host) {
		return false, "目标包含敏感域名，禁止扫描"
	}

	// 尝试解析为IP地址
	ip := net.ParseIP(host)
	if ip != nil {
		return isIPAllowed(ip, allowLocal)
	}

	// 检查是否为本地域名
	if isLocalDomain(host) && !allowLocal {
		return false, "目标为本地域名，禁止扫描"
	}

	return true, ""
}

// isSensitiveDomain 检查是否为敏感域名
// 参数列表：
//   - domain 域名
//
// 返回值列表：
//   - bool 是否为敏感域名
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

// isIPAllowed 检查IP地址是否允许扫描
// 参数列表：
//   - ip IP地址
//   - allowLocal 是否允许扫描本地地址
//
// 返回值列表：
//   - bool 是否允许扫描
//   - string 不允许的原因
func isIPAllowed(ip net.IP, allowLocal bool) (bool, string) {
	// 检查是否为内网IP
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

// isLocalDomain 检查是否为本地域名
// 参数列表：
//   - domain 域名
//
// 返回值列表：
//   - bool 是否为本地域名
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
