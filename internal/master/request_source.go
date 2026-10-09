package master

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
)

type sourceOriginKey struct{}
type sourceTransportKey struct{}

// ParseTrustedProxies accepts explicit IPs/CIDRs; an empty setting trusts nobody.
func ParseTrustedProxies(value string) ([]netip.Prefix, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	var prefixes []netip.Prefix
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		prefix, err := netip.ParsePrefix(item)
		if err != nil {
			ip, e := netip.ParseAddr(item)
			if e != nil || ip.Zone() != "" {
				return nil, errors.New("trusted proxies must be IP addresses or CIDRs")
			}
			ip = ip.Unmap()
			prefix = netip.PrefixFrom(ip, ip.BitLen())
		}
		if prefix.Addr().Zone() != "" {
			return nil, errors.New("proxy zones are unsupported")
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return prefixes, nil
}

func (s *Server) trustedProxy(ip netip.Addr) bool {
	if s.trustAllProxies {
		return true
	}
	for _, prefix := range s.trustedProxies {
		if prefix.Contains(ip.Unmap()) {
			return true
		}
	}
	return false
}

func oneSourceHeader(r *http.Request, name string) (string, error) {
	values := r.Header.Values(name)
	if len(values) == 0 {
		return "", nil
	}
	if len(values) != 1 || strings.TrimSpace(values[0]) == "" {
		return "", errors.New("ambiguous proxy header")
	}
	return strings.TrimSpace(values[0]), nil
}

func sourceHost(value, scheme string) (string, bool, error) {
	if value == "" || strings.ContainsAny(value, ",/\\?#@ \t\r\n") {
		return "", false, errors.New("invalid request host")
	}
	u, err := url.Parse(scheme + "://" + value)
	if err != nil || u.Host != value || u.User != nil || u.Hostname() == "" {
		return "", false, errors.New("invalid request host")
	}
	host := strings.ToLower(u.Hostname())
	ip, ipErr := netip.ParseAddr(host)
	if ipErr != nil {
		if strings.Contains(host, ":") {
			return "", false, errors.New("invalid host address")
		}
		for _, label := range strings.Split(host, ".") {
			if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
				return "", false, errors.New("invalid DNS host")
			}
			for _, c := range label {
				if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
					return "", false, errors.New("invalid DNS host")
				}
			}
		}
	}
	if strings.HasSuffix(value, ":") {
		return "", false, errors.New("empty port")
	}
	if port := u.Port(); port != "" {
		n, e := strconv.Atoi(port)
		if e != nil || n < 1 || n > 65535 {
			return "", false, errors.New("invalid port")
		}
		// Browser URL.origin omits HTTPS's default port, even if Host includes it.
		if scheme == "https" && n == 443 || scheme == "http" && n == 80 {
			value = strings.TrimSuffix(value, ":"+port)
		}
	}
	if ipErr == nil && ip.Zone() != "" {
		return "", false, errors.New("host zones are unsupported")
	}
	local := host == "localhost" || ipErr == nil && ip.Unmap().IsLoopback()
	return strings.ToLower(value), local, nil
}

// Resolve before authorization and rate limiting. Cleartext management is
// permitted only for a real loopback peer, never because a header claims TLS.
func (s *Server) resolveSource(r *http.Request) (*http.Request, error) {
	peer, _, _ := net.SplitHostPort(r.RemoteAddr)
	ip, _ := netip.ParseAddr(peer)
	transportOK := r.TLS != nil || s.allowLoopbackHTTP && ip.IsValid() && ip.Unmap().IsLoopback()
	if !transportOK {
		return nil, errors.New("management transport requires TLS or configured loopback HTTP")
	}
	trusted := ip.IsValid() && s.trustedProxy(ip)
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	host := r.Host
	forwardedHost := false
	client := ip
	if trusted {
		xhost, err := oneSourceHeader(r, "X-Forwarded-Host")
		if err != nil {
			return nil, err
		}
		proto, err := oneSourceHeader(r, "X-Forwarded-Proto")
		if err != nil {
			return nil, err
		}
		if xhost != "" || proto != "" {
			if xhost == "" || proto != "https" && proto != "http" {
				return nil, errors.New("proxy source must be a single HTTP/HTTPS origin")
			}
			host, forwardedHost = xhost, true
			scheme = proto
		}
		xff, err := oneSourceHeader(r, "X-Forwarded-For")
		if err != nil {
			return nil, err
		}
		if xff != "" {
			parts := strings.Split(xff, ",")
			if len(parts) > 32 || len(xff) > 4096 {
				return nil, errors.New("proxy chain too long")
			}
			addresses := make([]netip.Addr, len(parts))
			for i, part := range parts {
				addr, e := netip.ParseAddr(strings.TrimSpace(part))
				if e != nil || addr.Zone() != "" {
					return nil, errors.New("invalid proxy client address")
				}
				addresses[i] = addr.Unmap()
			}
			for i := len(addresses) - 1; i >= 0 && s.trustedProxy(client); i-- {
				client = addresses[i]
			}
		} else {
			realIP, e := oneSourceHeader(r, "X-Real-IP")
			if e != nil {
				return nil, e
			}
			if realIP != "" {
				client, e = netip.ParseAddr(realIP)
				if e != nil || client.Zone() != "" {
					return nil, errors.New("invalid proxy client address")
				}
				client = client.Unmap()
			}
		}
	}
	origin := s.origin
	if origin == "" || forwardedHost {
		normalized, local, err := sourceHost(host, scheme)
		if err != nil {
			return nil, err
		}
		host = normalized
		if scheme == "http" && !local {
			return nil, errors.New("external browser origins require HTTPS")
		}
		if origin == "" {
			if !forwardedHost && !local {
				return nil, errors.New("direct access requires a loopback host or explicit origin")
			}
			origin = scheme + "://" + host
		}
	}
	ctx := context.WithValue(r.Context(), sourceOriginKey{}, origin)
	next := r.Clone(context.WithValue(ctx, sourceTransportKey{}, transportOK))
	next.Host = host
	if client.IsValid() {
		next.RemoteAddr = net.JoinHostPort(client.Unmap().String(), "0")
	}
	return next, nil
}

func managementTransport(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	allowed, _ := r.Context().Value(sourceTransportKey{}).(bool)
	return allowed
}

func (s *Server) requestOrigin(r *http.Request) string {
	if origin, ok := r.Context().Value(sourceOriginKey{}).(string); ok {
		return origin
	}
	return s.origin
}
