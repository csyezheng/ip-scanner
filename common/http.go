package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
)

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
	siteCfg := RetrieveSiteCfg(config)
	url := siteCfg.HttpsURL
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
	if resp.StatusCode >= 400 {
		slog.Debug("Http response", "status code", resp.StatusCode)
		return fmt.Errorf("http response status code %d", resp.StatusCode)
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}
	return nil
}
