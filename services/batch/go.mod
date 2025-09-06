module github.com/tokane888/test-mcp/services/batch

go 1.24

require (
	github.com/joho/godotenv v1.5.1
	github.com/tokane888/test-mcp/pkg/logger v0.0.0
	go.uber.org/zap v1.27.0
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/tokane888/test-mcp/pkg/logger => ../../pkg/logger
