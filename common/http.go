package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
)

func isIPv4(ip string) bool {
	return strings.Contains(ip, ".")
}

func dialContext(destination string, destinationPort uint16) func(ctx context.Context, network, address string) (net.Conn, error) {
	var addr string
	if isIPv4(destination) {
		addr = fmt.Sprintf("%s:%d", destination, destinationPort)
	} else {
		addr = fmt.Sprintf("[%s]:%d", destination, destinationPort)
	}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, addr)
	}
}

func HttpPing(destination string, destinationPort uint16, config *Config) bool {
	timeout := config.HTTP.Timeout
	//addr := fmt.Sprintf("%s:%d", destination, destinationPort)
	//dialer := &net.Dialer{
	//	Timeout:   timeout,
	//	KeepAlive: 2 * time.Second,
	//}
	//conf := &tls.Config{
	//	InsecureSkipVerify: true,
	//}
	//conn, err := tls.DialWithDialer(dialer, "tcp", addr, conf)
	//if err != nil {
	//	slog.Debug("tls dial:", err)
	//	return false
	//}
	////hostname := []string{"google.com"}
	//for i, cert := range conn.ConnectionState().PeerCertificates {
	//	subject := cert.Subject
	//	issuer := cert.Issuer
	//	slog.Info(" %d s:/C=%v/ST=%v/L=%v/O=%v/OU=%v/CN=%s", i, subject.Country, subject.Province, subject.Locality, subject.Organization, subject.OrganizationalUnit, subject.CommonName)
	//	slog.Info("   i:/C=%v/ST=%v/L=%v/O=%v/OU=%v/CN=%s", issuer.Country, issuer.Province, issuer.Locality, issuer.Organization, issuer.OrganizationalUnit, issuer.CommonName)
	//	for _, verifyHost := range hostname {
	//		if cert.VerifyHostname(verifyHost) != nil {
	//			return false
	//		} else {
	//			success = true
	//		}
	//	}
	//	if success {
	//		break
	//	}
	//}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext:     dialContext(destination, destinationPort),
	}
	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// ErrUseLastResponse can be returned by Client.CheckRedirect hooks to control how redirects are processed.
			//If returned, the next request is not sent and the most recent response is returned with its body unclosed.
			return http.ErrUseLastResponse
		},
	}
	url := config.Domains.Cloudflare.HttpsURL
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		slog.Debug("http request:", slog.String("url", url), slog.Any("Error", err))
		return false
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		slog.Debug("Http response error:", "Error", err)
		return false
	}
	if resp.StatusCode >= 400 {
		slog.Debug("Http response", "status code", resp.StatusCode)
		return false
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)
	return true
}
