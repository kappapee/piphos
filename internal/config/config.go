// Package config defines application-wide constants used throughout piphos.
package config

import "time"

const (
	// HTTPClientTimeout is the maximum duration for HTTP requests.
	HTTPClientTimeout = 10 * time.Second

	// PiphosUserAgent is the User-Agent header value for HTTP requests.
	PiphosUserAgent = "piphos/1.0"

	// PiphosStamp is the identifier used for gist descriptions and filenames.
	PiphosStamp = "_piphos_"
)
