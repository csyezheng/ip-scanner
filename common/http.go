package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
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

//func testTLS(destination string, destinationPort uint16, config *Config) bool {
//	timeout := config.HTTP.Timeout
//	addr := fmt.Sprintf("%s:%d", destination, destinationPort)
//	dialer := &net.Dialer{
//		Timeout:   timeout,
//		KeepAlive: 2 * time.Second,
//	}
//	conf := &tls.Config{
//		InsecureSkipVerify: true,
//	}
//	conn, err := tls.DialWithDialer(dialer, "tcp", addr, conf)
//	if err != nil {
//		slog.Debug("tls dial:", err)
//		return false
//	}
//	for _, cert := range conn.ConnectionState().PeerCertificates {
//		verifyHost := "google.com"
//		if cert.VerifyHostname(verifyHost) != nil {
//			slog.Debug("host not match!")
//			return false
//		} else {
//			return true
//		}
//	}
//	return true
//}

func reqHEAD(destination string, destinationPort uint16, config *Config) error {
	slog.Debug("Https request using:", "IP", destination)
	timeout := config.HTTP.Timeout
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
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
	url := ""
	if config.General.UsedFor == "Cloudflare" {
		url = config.UsedFor.Cloudflare.HttpsURL
	} else if config.General.UsedFor == "GoogleTranslate" {
		url = config.UsedFor.GoogleTranslate.HttpsURL
	} else {
		slog.Error("UsedFor should be Cloudflare or GoogleTranslate")
		os.Exit(0)
	}
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		slog.Debug("http request:", slog.String("url", url), slog.Any("Error", err))
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		slog.Debug("Http response error:", "Error", err)
		return err
	}
	//if resp.StatusCode >= 400 {
	//	slog.Debug("Http response", "status code", resp.StatusCode)
	//	return false
	//}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)
	return nil
}

func reqTranslate(destination string, destinationPort uint16, config *Config) error {
	slog.Debug("Https request using:", "IP", destination)
	timeout := config.HTTP.Timeout
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		DialContext:     dialContext(destination, destinationPort),
	}
	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}
	translateAPI := "https://translate.googleapis.com/translate_a/single?client=gtx&sl=zh-CN&tl=en&dt=t&q=%E4%BD%A0%E5%A5%BD"
	req, err := http.NewRequest("GET", translateAPI, nil)
	if err != nil {
		return err
	}
	req.Header.Add("host", "translate.googleapis.com")
	res, err := client.Do(req)
	if err != nil {
		slog.Debug("translate:", "error", err)
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	slog.Debug(string(body))
	if strings.Contains(string(body), "你好") {
		return nil
	} else {
		return fmt.Errorf("response content not contain correct translated word")
	}
}
