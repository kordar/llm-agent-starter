module github.com/kordar/llm-agent-starter

go 1.25.5

require github.com/spf13/cast v1.10.0

replace github.com/kordar/llm-agent => ../llm-agent

require github.com/kordar/llm-agent v0.0.0-00010101000000-000000000000

require github.com/kordar/llm-memory v0.0.0-00010101000000-000000000000

require github.com/kordar/llm-memory-rag v0.0.0-00010101000000-000000000000

replace github.com/kordar/llm-memory => ../llm-memory

replace github.com/kordar/llm-memory-rag => ../llm-memory-rag

require github.com/kordar/llm-rag v0.0.0-00010101000000-000000000000

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/lib/pq v1.12.0 // indirect
	github.com/pgvector/pgvector-go v0.3.0 // indirect
)

replace github.com/kordar/llm-rag => ../llm-rag

replace github.com/kordar/llm-tool => ../llm-tool
