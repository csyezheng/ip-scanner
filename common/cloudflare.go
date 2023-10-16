package common

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Response https://mholt.github.io/json-to-go/
type Response struct {
	Result struct {
		Ipv4Cidrs []string `json:"ipv4_cidrs"`
		Ipv6Cidrs []string `json:"ipv6_cidrs"`
		Etag      string   `json:"etag"`
	} `json:"result"`
	Success  bool  `json:"success"`
	Errors   []any `json:"errors"`
	Messages []any `json:"messages"`
}

func GetIps(config Config) []string {
	req, err := http.NewRequest("GET", config.CloudflareAPI, nil)
	log.Println(config.CloudflareAPI)
	log.Println(config.CloudflareToken)
	var bearer = "Bearer " + config.CloudflareToken
	req.Header.Add("Authorization", bearer)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("error on response.\n[ERROR] -", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	//log.Println(string([]byte(body)))
	if err != nil {
		log.Println("error while reading the response bytes:", err)
	}
	var response Response
	if err = json.Unmarshal(body, &response); err != nil {
		fmt.Println("can not unmarshal JSON")
	}
	ipv4 := response.Result.Ipv4Cidrs
	ipv6 := response.Result.Ipv6Cidrs
	return append(ipv4, ipv6...)
}
