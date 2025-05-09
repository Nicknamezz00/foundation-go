package httputil

import (
	"net/http"
	"net/url"
)

func OverwriteHeader(from http.Header, to http.Header) http.Header {
	for key, val := range from {
		to[key] = val
	}
	return to
}

func MergeHeader(from http.Header, to http.Header) http.Header {
	for key, val := range from {
		for _, v := range val {
			to.Add(key, v)
		}
	}
	return to
}

func StringMapStringToURLValues(m map[string]string) url.Values {
	args := url.Values{}
	for key, value := range m {
		args.Add(key, value)
	}
	return args
}
