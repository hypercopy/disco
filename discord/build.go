package discord

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/valyala/fasthttp"
)

var (
	JS_FILE_REGEX    = regexp.MustCompile(`<script\s+src="([^"]+\.js)"\s+defer>\s*</script>`)
	BUILD_INFO_REGEX = regexp.MustCompile(`Build Number: \"\).concat\(\"(\d+)\"`)
	BUILD_HEADERS    = map[string]string{
		"Accept":             "*/*",
		"Accept-Language":    "en-GB,en-US;q=0.9,en;q=0.8",
		"Cache-Control":      "no-cache",
		"Pragma":             "no-cache",
		"Referer":            "https://discord.com/login",
		"Sec-Ch-Ua":          `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`,
		"Sec-Ch-Ua-Mobile":   "?0",
		"Sec-Ch-Ua-Platform": `"macOS"`,
		"Sec-Fetch-Dest":     "script",
		"Sec-Fetch-Mode":     "no-cors",
		"Sec-Fetch-Site":     "same-origin",
		"User-Agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	}
)

func getLatestBuild() (string, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.Header.SetMethod(fasthttp.MethodGet)
	for h, v := range BUILD_HEADERS {
		req.Header.Set(h, v)
	}
	req.SetRequestURI("https://discord.com/login")

	if err := requestClient.Do(req, resp); err != nil {
		return "", err
	}

	// Extract Asset JS Files
	matches := JS_FILE_REGEX.FindAllStringSubmatch(string(resp.Body()), -1)
	if len(matches) == 0 {
		return "", errors.New("unable to fetch discord build number (block?)")
	}
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		asset := match[1]
		if asset == "" {
			continue
		}
		// fetch the asset file and check for errors
		req.Header.SetMethod(fasthttp.MethodGet)
		req.SetRequestURI(fmt.Sprintf("https://discord.com%s", asset))
		for h, v := range BUILD_HEADERS {
			req.Header.Set(h, v)
		}
		if err := requestClient.Do(req, resp); err != nil {
			continue
		}

		if !strings.Contains(string(resp.Body()), "buildNumber") {
			continue
		}
		build_number := strings.Split(strings.Split(string(resp.Body()), `build_number:"`)[1], `"`)[0]
		return build_number, nil
	}

	return "", errors.New("unable to find a valid build number")
}

func mustGetLatestBuild() string {
	if build, err := getLatestBuild(); err != nil {
		panic(err)
	} else {
		return build
	}
}
