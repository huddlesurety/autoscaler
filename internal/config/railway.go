package config

type RailwayConfig struct {
	Token         string `mapstructure:"API_TOKEN"`
	EnvironmentID string `mapstructure:"ENVIRONMENT_ID"`

	ServiceRAG string `mapstructure:"SERVICE_RAG"`
	ServiceAPI string `mapstructure:"SERVICE_API"`
	ServiceApp string `mapstructure:"SERVICE_APP"`
}
