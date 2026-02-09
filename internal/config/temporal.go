package config

type TemporalConfig struct {
	ServerURL string `mapstructure:"SERVER_URL"`

	// TaskQueueRAG string `mapstructure:"TASK_QUEUE_RAG"`

	// WorkflowGenerateField string `mapstructure:"WORKFLOW_GENERATE_FIELD"`
	// WorkflowParseDocument string `mapstructure:"WORKFLOW_PARSE_DOCUMENT"`
}
