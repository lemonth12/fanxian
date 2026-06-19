package jd

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var pcPattern = regexp.MustCompile(`item\.jd\.com/(\d+)\.html`)
var mobilePattern = regexp.MustCompile(`item\.m\.jd\.com/product/(\d+)\.html`)

func ParseURL(rawURL string) (string, error) {
	u := strings.TrimSpace(rawURL)
	if u == "" {
		return "", fmt.Errorf("empty URL")
	}

	if strings.Contains(u, "3.cn") || (strings.Contains(u, "m.jd.com") && !strings.Contains(u, "item.m.jd.com")) {
		resp, err := http.Head(u)
		if err != nil {
			return "", fmt.Errorf("无法识别该链接，请粘贴京东商品链接")
		}
		resp.Body.Close()
		if resp.Header.Get("Location") != "" {
			u = resp.Header.Get("Location")
		}
	}

	m := pcPattern.FindStringSubmatch(u)
	if len(m) == 2 {
		return m[1], nil
	}

	m = mobilePattern.FindStringSubmatch(u)
	if len(m) == 2 {
		return m[1], nil
	}

	return "", fmt.Errorf("无法识别该链接，请粘贴京东商品链接")
}
