package agentstarter

import (
	"fmt"
	"io"
	"strings"
	"sync"

	llmmemory "github.com/kordar/llm-memory"
	memoryrag "github.com/kordar/llm-memory-rag"
	"github.com/kordar/llm-rag/embedding"
	"github.com/kordar/llm-rag/rerank"
	"github.com/spf13/cast"
)

type MemoryStoreBuilder func(cfg map[string]string) (llmmemory.Store, io.Closer, error)

var (
	memoryStoreBuilderMu sync.RWMutex
	memoryStoreBuilders  = make(map[string]MemoryStoreBuilder)
)

func init() {
	_ = RegisterMemoryStoreBuilder("memory-rag", buildMemoryRAGStore)
}

func RegisterMemoryStoreBuilder(storeType string, builder MemoryStoreBuilder) error {
	storeType = strings.ToLower(strings.TrimSpace(storeType))
	if storeType == "" {
		return fmt.Errorf("memory store type is empty")
	}
	if builder == nil {
		return fmt.Errorf("memory store builder is nil for type=%s", storeType)
	}
	memoryStoreBuilderMu.Lock()
	memoryStoreBuilders[storeType] = builder
	memoryStoreBuilderMu.Unlock()
	return nil
}

func BuildMemoryStoreFromConfig(cfg map[string]string) (llmmemory.Store, io.Closer, error) {
	storeType := strings.ToLower(strings.TrimSpace(cfg["type"]))
	if storeType == "" {
		storeType = "memory-rag"
	}

	memoryStoreBuilderMu.RLock()
	builder, ok := memoryStoreBuilders[storeType]
	memoryStoreBuilderMu.RUnlock()
	if !ok {
		return nil, nil, fmt.Errorf("unsupported memory store type: %s", storeType)
	}
	return builder(cfg)
}

func buildMemoryStoreByConfig(cfg map[string]string) (llmmemory.Store, io.Closer, error) {
	return BuildMemoryStoreFromConfig(cfg)
}

func buildMemoryRAGStore(cfg map[string]string) (llmmemory.Store, io.Closer, error) {
	embeddingEndpoint := strings.TrimSpace(cfg["embedding_endpoint"])
	if embeddingEndpoint == "" {
		embeddingEndpoint = "http://localhost:8000"
	}
	embeddingModel := strings.TrimSpace(cfg["embedding_model"])
	if embeddingModel == "" {
		embeddingModel = "bge-large-zh"
	}

	rerankEndpoint := strings.TrimSpace(cfg["rerank_endpoint"])
	if rerankEndpoint == "" {
		rerankEndpoint = "http://localhost:8001"
	}
	rerankModel := strings.TrimSpace(cfg["rerank_model"])
	if rerankModel == "" {
		rerankModel = "bge-reranker"
	}

	commonHeaders, err := parseHeaderExpr(cfg["headers"])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid headers: %w", err)
	}
	embeddingHeaders := cloneHeaderMap(commonHeaders)
	if extra, err := parseHeaderExpr(cfg["embedding_headers"]); err != nil {
		return nil, nil, fmt.Errorf("invalid embedding_headers: %w", err)
	} else {
		mergeHeaderMap(embeddingHeaders, extra)
	}
	rerankHeaders := cloneHeaderMap(commonHeaders)
	if extra, err := parseHeaderExpr(cfg["rerank_headers"]); err != nil {
		return nil, nil, fmt.Errorf("invalid rerank_headers: %w", err)
	} else {
		mergeHeaderMap(rerankHeaders, extra)
	}

	dsn := strings.TrimSpace(cfg["dsn"])
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/rag?sslmode=disable"
	}
	table := strings.TrimSpace(cfg["table"])
	if table == "" {
		table = "memory_records"
	}
	dimension := cast.ToInt(cfg["dimension"])
	if dimension <= 0 {
		dimension = 1024
	}

	commonToken := strings.TrimSpace(cfg["token"])
	if commonToken == "" {
		commonToken = strings.TrimSpace(cfg["bearer_token"])
	}

	embeddingToken := strings.TrimSpace(cfg["embedding_token"])
	if embeddingToken == "" {
		embeddingToken = strings.TrimSpace(cfg["embedding_bearer_token"])
	}
	if embeddingToken == "" {
		embeddingToken = commonToken
	}
	embeddingAuthorization := strings.TrimSpace(cfg["embedding_authorization"])
	if embeddingAuthorization == "" && embeddingToken != "" {
		embeddingAuthorization = "Bearer " + embeddingToken
	}
	if embeddingAuthorization != "" {
		embeddingHeaders["Authorization"] = embeddingAuthorization
	}

	rerankToken := strings.TrimSpace(cfg["rerank_token"])
	if rerankToken == "" {
		rerankToken = strings.TrimSpace(cfg["rerank_bearer_token"])
	}
	if rerankToken == "" {
		rerankToken = commonToken
	}
	rerankAuthorization := strings.TrimSpace(cfg["rerank_authorization"])
	if rerankAuthorization == "" && rerankToken != "" {
		rerankAuthorization = "Bearer " + rerankToken
	}
	if rerankAuthorization != "" {
		rerankHeaders["Authorization"] = rerankAuthorization
	}

	embeddingOpts := make([]embedding.VLLMOption, 0, len(embeddingHeaders))
	for k, v := range embeddingHeaders {
		embeddingOpts = append(embeddingOpts, embedding.WithHeader(k, v))
	}
	embedder := embedding.New(
		embedding.NewVLLMProvider(embeddingEndpoint, embeddingOpts...),
		embedding.WithModel(embeddingModel),
	)
	rerankOpts := make([]rerank.Option, 0, len(rerankHeaders))
	for k, v := range rerankHeaders {
		rerankOpts = append(rerankOpts, rerank.WithHeader(k, v))
	}
	reranker := rerank.New(rerankEndpoint, rerankModel, rerankOpts...)

	store, err := memoryrag.NewWithRAGClients(
		dsn,
		embedder,
		reranker,
		memoryrag.Config{
			Table:     table,
			Dimension: dimension,
		},
	)
	if err != nil {
		return nil, nil, err
	}
	return store, store, nil
}

func parseHeaderExpr(raw string) (map[string]string, error) {
	out := map[string]string{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return out, nil
	}
	items := strings.Split(raw, "|")
	for _, item := range items {
		part := strings.TrimSpace(item)
		if part == "" {
			continue
		}
		key, value, ok := strings.Cut(part, "::")
		if !ok {
			return nil, fmt.Errorf("invalid pair %q, expected key::value", part)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return nil, fmt.Errorf("empty header key in %q", part)
		}
		out[key] = value
	}
	return out, nil
}

func cloneHeaderMap(src map[string]string) map[string]string {
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func mergeHeaderMap(dst map[string]string, src map[string]string) {
	for k, v := range src {
		dst[k] = v
	}
}
