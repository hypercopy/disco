package discord

import (
	"fmt"
	"regexp"

	"github.com/valyala/fasthttp"
)

var (
	JS_FILE_REGEX    = regexp.MustCompile(`<script src=\"(/assets/\d{4,5}\.[^\"]+\.js)\" defer></script>`)
	BUILD_INFO_REGEX = regexp.MustCompile(`Build Number: \"\).concat\(\"(\d+)\"`)
)

func getLatestBuild() (string, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI("https://discord.com/app")

	if err := requestClient.Do(req, resp); err != nil {
		return "", err
	}

	matches := JS_FILE_REGEX.FindAllStringSubmatch(string(resp.Body()), -1)
	if len(matches) == 0 {
		return "9999", nil
	}
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		asset := match[1]
		if asset == "" {
			continue
		}
		req.Header.SetMethod(fasthttp.MethodGet)
		req.SetRequestURI(fmt.Sprintf("https://discord.com%s", asset))
		if err := requestClient.Do(req, resp); err != nil {
			continue
		}
		match := BUILD_INFO_REGEX.FindStringSubmatch(string(resp.Body()))
		if len(match) < 2 {
			continue
		}
		return match[1], nil
	}

	return "9999", nil
}

func mustGetLatestBuild() string {
	if build, err := getLatestBuild(); err != nil {
		panic(err)
	} else {
		return build
	}
}
