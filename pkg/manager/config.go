package manager

import "time"

type Config struct {
	RailwayEnvironmentID string
	RailwayToken         string
	MetricInterval       time.Duration
	ScaleInterval        time.Duration
}
