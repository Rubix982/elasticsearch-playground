package models

import "time"

// IndexRequest represents a request to create an index
type IndexRequest struct {
	IndexName        string                 `json:"index_name" binding:"required"`
	Settings         *IndexSettings         `json:"settings,omitempty"`
	Mappings         map[string]interface{} `json:"mappings,omitempty"`
	Aliases          map[string]interface{} `json:"aliases,omitempty"`
	WriteOptimized   bool                   `json:"write_optimized,omitempty"`
	TextHeavy        bool                   `json:"text_heavy,omitempty"`
	ExpectedVolume   string                 `json:"expected_volume,omitempty"` // low, medium, high
	ExpectedDocSize  string                 `json:"expected_doc_size,omitempty"` // small, medium, large
	IngestionRate    string                 `json:"ingestion_rate,omitempty"` // low, medium, high
}

// IndexSettings represents index settings configuration
type IndexSettings struct {
	NumberOfShards   int    `json:"number_of_shards,omitempty"`
	NumberOfReplicas int    `json:"number_of_replicas,omitempty"`
	RefreshInterval  string `json:"refresh_interval,omitempty"`
	
	// Write optimization settings
	IndexBufferSize           string `json:"index.buffer_size,omitempty"`
	TranslogFlushThresholdSize string `json:"index.translog.flush_threshold_size,omitempty"`
	TranslogSyncInterval      string `json:"index.translog.sync_interval,omitempty"`
	TranslogDurability        string `json:"index.translog.durability,omitempty"`
	
	// Merge policy settings
	MergePolicyMaxMergeSize          string `json:"index.merge.policy.max_merge_size,omitempty"`
	MergePolicySegmentsPerTier       int    `json:"index.merge.policy.segments_per_tier,omitempty"`
	MergePolicyMaxMergedSegmentMB    int    `json:"index.merge.policy.max_merged_segment_mb,omitempty"`
	MergeSchedulerMaxThreadCount     int    `json:"index.merge.scheduler.max_thread_count,omitempty"`
	
	// Codec and compression
	Codec string `json:"index.codec,omitempty"`
	
	// Additional custom settings
	Additional map[string]interface{} `json:"additional,omitempty"`
}

// IndexResponse represents the response after index creation
type IndexResponse struct {
	IndexName    string    `json:"index_name"`
	Acknowledged bool      `json:"acknowledged"`
	Created      bool      `json:"created"`  
	Settings     *IndexSettings `json:"settings,omitempty"`
	Optimizations []string `json:"optimizations,omitempty"`
	RequestID    string    `json:"request_id"`
	Timestamp    time.Time `json:"timestamp"`
}

// IndexInfo represents comprehensive information about an index
type IndexInfo struct {
	IndexName    string                 `json:"index_name"`
	UUID         string                 `json:"uuid"`
	Health       string                 `json:"health"`
	Status       string                 `json:"status"`
	Primary      int                    `json:"pri"`
	Replica      int                    `json:"rep"`
	DocsCount    int64                  `json:"docs.count"`
	DocsDeleted  int64                  `json:"docs.deleted"`
	StoreSize    string                 `json:"store.size"`
	PrimaryStoreSize string            `json:"pri.store.size"`
	Settings     *DetailedIndexSettings `json:"settings"`
	Mappings     interface{}            `json:"mappings"`
	Aliases      map[string]interface{} `json:"aliases"`
	Stats        *IndexStats            `json:"stats,omitempty"`
	WriteMetrics *WriteMetrics          `json:"write_metrics,omitempty"`
	RequestID    string                 `json:"request_id"`
	Timestamp    time.Time              `json:"timestamp"`
}

// DetailedIndexSettings represents detailed index settings
type DetailedIndexSettings struct {
	Index IndexConfig `json:"index"`
}

// IndexConfig represents the index configuration
type IndexConfig struct {
	CreationDate           string `json:"creation_date"`
	NumberOfShards         string `json:"number_of_shards"`
	NumberOfReplicas       string `json:"number_of_replicas"`
	UUID                   string `json:"uuid"`
	Version                map[string]interface{} `json:"version"`
	ProvidedName           string `json:"provided_name"`
	RefreshInterval        string `json:"refresh_interval,omitempty"`
	MaxResultWindow        string `json:"max_result_window,omitempty"`
	
	// Write-related settings
	BufferSize                 string `json:"buffer_size,omitempty"`
	TranslogFlushThresholdSize string `json:"translog.flush_threshold_size,omitempty"`
	TranslogSyncInterval       string `json:"translog.sync_interval,omitempty"`
	TranslogDurability         string `json:"translog.durability,omitempty"`
	
	// Merge settings
	MergePolicyMaxMergeSize       string `json:"merge.policy.max_merge_size,omitempty"`
	MergePolicySegmentsPerTier    string `json:"merge.policy.segments_per_tier,omitempty"`
	MergePolicyMaxMergedSegmentMB string `json:"merge.policy.max_merged_segment_mb,omitempty"`
	
	// Other settings
	Codec string `json:"codec,omitempty"`
	BlocksReadOnlyAllowDelete string `json:"blocks.read_only_allow_delete,omitempty"`
}

// IndexStats represents index statistics
type IndexStats struct {
	Primaries *IndexStatsDetails `json:"primaries"`
	Total     *IndexStatsDetails `json:"total"`
}

// IndexStatsDetails represents detailed index statistics
type IndexStatsDetails struct {
	Docs       DocsStats       `json:"docs"`
	Store      StoreStats      `json:"store"`
	Indexing   IndexingStats   `json:"indexing"`
	Get        GetStats        `json:"get"`
	Search     SearchStats     `json:"search"`
	Merges     MergeStats      `json:"merges"`
	Refresh    RefreshStats    `json:"refresh"`
	Flush      FlushStats      `json:"flush"`
	Warmer     WarmerStats     `json:"warmer"`
	QueryCache QueryCacheStats `json:"query_cache"`
	Fielddata  FielddataStats  `json:"fielddata"`
	Completion CompletionStats `json:"completion"`
	Segments   SegmentsStats   `json:"segments"`
	Translog   TranslogStats   `json:"translog"`
}

// DocsStats represents document statistics
type DocsStats struct {
	Count   int64 `json:"count"`
	Deleted int64 `json:"deleted"`
}

// StoreStats represents storage statistics
type StoreStats struct {
	SizeInBytes          int64 `json:"size_in_bytes"`
	ReservedInBytes      int64 `json:"reserved_in_bytes,omitempty"`
	TotalDataSetSizeInBytes int64 `json:"total_data_set_size_in_bytes,omitempty"`
}

// IndexingStats represents indexing statistics
type IndexingStats struct {
	IndexTotal           int64         `json:"index_total"`
	IndexTimeInMillis    int64         `json:"index_time_in_millis"`
	IndexCurrent         int64         `json:"index_current"`
	IndexFailed          int64         `json:"index_failed"`
	DeleteTotal          int64         `json:"delete_total"`
	DeleteTimeInMillis   int64         `json:"delete_time_in_millis"`
	DeleteCurrent        int64         `json:"delete_current"`
	NoopUpdateTotal      int64         `json:"noop_update_total"`
	IsThrottled          bool          `json:"is_throttled"`
	ThrottleTimeInMillis int64         `json:"throttle_time_in_millis"`
	WriteLoad            float64       `json:"write_load,omitempty"`
}

// GetStats represents get statistics
type GetStats struct {
	Total               int64 `json:"total"`
	TimeInMillis        int64 `json:"time_in_millis"`
	ExistsTotal         int64 `json:"exists_total"`
	ExistsTimeInMillis  int64 `json:"exists_time_in_millis"`
	MissingTotal        int64 `json:"missing_total"`
	MissingTimeInMillis int64 `json:"missing_time_in_millis"`
	Current             int64 `json:"current"`
}

// SearchStats represents search statistics
type SearchStats struct {
	OpenContexts        int64 `json:"open_contexts"`
	QueryTotal          int64 `json:"query_total"`
	QueryTimeInMillis   int64 `json:"query_time_in_millis"`
	QueryCurrent        int64 `json:"query_current"`
	FetchTotal          int64 `json:"fetch_total"`
	FetchTimeInMillis   int64 `json:"fetch_time_in_millis"`
	FetchCurrent        int64 `json:"fetch_current"`
	ScrollTotal         int64 `json:"scroll_total"`
	ScrollTimeInMillis  int64 `json:"scroll_time_in_millis"`
	ScrollCurrent       int64 `json:"scroll_current"`
	SuggestTotal        int64 `json:"suggest_total"`
	SuggestTimeInMillis int64 `json:"suggest_time_in_millis"`
	SuggestCurrent      int64 `json:"suggest_current"`
}

// MergeStats represents merge statistics
type MergeStats struct {
	Current                    int64 `json:"current"`
	CurrentDocs                int64 `json:"current_docs"`
	CurrentSizeInBytes         int64 `json:"current_size_in_bytes"`
	Total                      int64 `json:"total"`
	TotalTimeInMillis          int64 `json:"total_time_in_millis"`
	TotalDocs                  int64 `json:"total_docs"`
	TotalSizeInBytes           int64 `json:"total_size_in_bytes"`
	TotalStoppedTimeInMillis   int64 `json:"total_stopped_time_in_millis"`
	TotalThrottledTimeInMillis int64 `json:"total_throttled_time_in_millis"`
	TotalAutoThrottleInBytes   int64 `json:"total_auto_throttle_in_bytes"`
}

// RefreshStats represents refresh statistics
type RefreshStats struct {
	Total             int64 `json:"total"`
	TotalTimeInMillis int64 `json:"total_time_in_millis"`
	ExternalTotal     int64 `json:"external_total"`
	ExternalTimeInMillis int64 `json:"external_total_time_in_millis"`
	Listeners         int64 `json:"listeners"`
}

// FlushStats represents flush statistics
type FlushStats struct {
	Total             int64 `json:"total"`
	Periodic          int64 `json:"periodic"`
	TotalTimeInMillis int64 `json:"total_time_in_millis"`
}

// WarmerStats represents warmer statistics
type WarmerStats struct {
	Current           int64 `json:"current"`
	Total             int64 `json:"total"`
	TotalTimeInMillis int64 `json:"total_time_in_millis"`
}

// QueryCacheStats represents query cache statistics
type QueryCacheStats struct {
	MemorySizeInBytes int64 `json:"memory_size_in_bytes"`
	TotalCount        int64 `json:"total_count"`
	HitCount          int64 `json:"hit_count"`
	MissCount         int64 `json:"miss_count"`
	CacheSize         int64 `json:"cache_size"`
	CacheCount        int64 `json:"cache_count"`
	Evictions         int64 `json:"evictions"`
}

// FielddataStats represents fielddata statistics
type FielddataStats struct {
	MemorySizeInBytes int64 `json:"memory_size_in_bytes"`
	Evictions         int64 `json:"evictions"`
}

// CompletionStats represents completion statistics
type CompletionStats struct {
	SizeInBytes int64 `json:"size_in_bytes"`
}

// SegmentsStats represents segments statistics
type SegmentsStats struct {
	Count                     int64 `json:"count"`
	MemoryInBytes            int64 `json:"memory_in_bytes"`
	TermsMemoryInBytes       int64 `json:"terms_memory_in_bytes"`
	StoredFieldsMemoryInBytes int64 `json:"stored_fields_memory_in_bytes"`
	TermVectorsMemoryInBytes  int64 `json:"term_vectors_memory_in_bytes"`
	NormsMemoryInBytes        int64 `json:"norms_memory_in_bytes"`
	PointsMemoryInBytes       int64 `json:"points_memory_in_bytes"`
	DocValuesMemoryInBytes    int64 `json:"doc_values_memory_in_bytes"`
	IndexWriterMemoryInBytes  int64 `json:"index_writer_memory_in_bytes"`
	VersionMapMemoryInBytes   int64 `json:"version_map_memory_in_bytes"`
	FixedBitSetMemoryInBytes  int64 `json:"fixed_bit_set_memory_in_bytes"`
	MaxUnsafeAutoIdTimestamp  int64 `json:"max_unsafe_auto_id_timestamp"`
	FileSizes                 map[string]FileSizeInfo `json:"file_sizes"`
}

// FileSizeInfo represents file size information
type FileSizeInfo struct {
	Size        int64  `json:"size_in_bytes"`
	MinSize     int64  `json:"min_size_in_bytes"`
	MaxSize     int64  `json:"max_size_in_bytes"`
	AverageSize int64  `json:"average_size_in_bytes"`
	Count       int64  `json:"count"`
}

// TranslogStats represents translog statistics
type TranslogStats struct {
	Operations              int64 `json:"operations"`
	SizeInBytes            int64 `json:"size_in_bytes"`
	UncommittedOperations  int64 `json:"uncommitted_operations"`
	UncommittedSizeInBytes int64 `json:"uncommitted_size_in_bytes"`
	EarliestLastModifiedAge int64 `json:"earliest_last_modified_age"`
}

// WriteMetrics represents write-specific performance metrics
type WriteMetrics struct {
	IndexingRate          float64   `json:"indexing_rate"` // docs per second
	AverageDocSize        int64     `json:"average_doc_size"`
	WriteLatency          float64   `json:"write_latency_ms"`
	BulkLatency           float64   `json:"bulk_latency_ms"`
	SegmentCount          int64     `json:"segment_count"`
	MergeRate             float64   `json:"merge_rate"`
	RefreshRate           float64   `json:"refresh_rate"`
	TranslogSize          int64     `json:"translog_size"`
	WriteLoad             float64   `json:"write_load"`
	OptimizationScore     float64   `json:"optimization_score"`
	Recommendations       []string  `json:"recommendations"`
	LastOptimized         time.Time `json:"last_optimized"`
}

// BulkRequest represents a bulk operation request
type BulkRequest struct {
	IndexName         string                   `json:"index_name"`
	Operations        []BulkOperation          `json:"operations"`
	BatchSize         int                      `json:"batch_size,omitempty"`
	ParallelWorkers   int                      `json:"parallel_workers,omitempty"`
	OptimizeFor       string                   `json:"optimize_for,omitempty"` // write_throughput, consistency
	ErrorTolerance    string                   `json:"error_tolerance,omitempty"` // low, medium, high
	Settings          *BulkSettings            `json:"settings,omitempty"`
}

// BulkOperation represents a single operation in a bulk request
type BulkOperation struct {
	Action    string                 `json:"action"` // index, create, update, delete
	Index     string                 `json:"_index,omitempty"`
	ID        string                 `json:"_id,omitempty"`
	Document  map[string]interface{} `json:"doc,omitempty"`
	Source    map[string]interface{} `json:"_source,omitempty"`
	Version   *int64                 `json:"_version,omitempty"`
	Routing   string                 `json:"_routing,omitempty"`
}

// BulkSettings represents settings for bulk operations
type BulkSettings struct {
	RefreshPolicy    string        `json:"refresh,omitempty"` // true, false, wait_for
	Timeout          time.Duration `json:"timeout,omitempty"`
	WaitForActiveShards string     `json:"wait_for_active_shards,omitempty"`
	Pipeline         string        `json:"pipeline,omitempty"`
	Routing          string        `json:"routing,omitempty"`
}

// BulkResponse represents the response from a bulk operation
type BulkResponse struct {
	Took      int64              `json:"took"`
	Errors    bool               `json:"errors"`
	Items     []BulkResponseItem `json:"items"`
	Summary   *BulkSummary       `json:"summary"`
	RequestID string             `json:"request_id"`
	Timestamp time.Time          `json:"timestamp"`
}

// BulkResponseItem represents a single item response in bulk operation
type BulkResponseItem struct {
	Index  *BulkItemResponse `json:"index,omitempty"`
	Create *BulkItemResponse `json:"create,omitempty"`
	Update *BulkItemResponse `json:"update,omitempty"`
	Delete *BulkItemResponse `json:"delete,omitempty"`
}

// BulkItemResponse represents the response for a single bulk item
type BulkItemResponse struct {
	Index   string `json:"_index"`
	ID      string `json:"_id"`
	Version int64  `json:"_version"`
	Result  string `json:"result"`
	Status  int    `json:"status"`
	Error   *BulkError `json:"error,omitempty"`
	Shards  *ShardsInfo `json:"_shards,omitempty"`
	SeqNo   int64  `json:"_seq_no,omitempty"`
	PrimaryTerm int64 `json:"_primary_term,omitempty"`
}

// BulkError represents an error in bulk operation
type BulkError struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
	Index  string `json:"index,omitempty"`
	Shard  string `json:"shard,omitempty"`
	Status int    `json:"status,omitempty"`
}

// ShardsInfo represents shard information
type ShardsInfo struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Failed     int `json:"failed"`
}

// BulkSummary provides a summary of bulk operation results
type BulkSummary struct {
	TotalOperations     int64         `json:"total_operations"`
	SuccessfulOperations int64        `json:"successful_operations"`
	FailedOperations    int64         `json:"failed_operations"`
	IndexedDocuments    int64         `json:"indexed_documents"`
	UpdatedDocuments    int64         `json:"updated_documents"`
	DeletedDocuments    int64         `json:"deleted_documents"`
	ProcessingTime      time.Duration `json:"processing_time"`
	ThroughputPerSecond float64       `json:"throughput_per_second"`
	AverageLatency      time.Duration `json:"average_latency"`
	ErrorRate           float64       `json:"error_rate"`
}

// OptimizationRequest represents a request to optimize an index
type OptimizationRequest struct {
	IndexName    string   `json:"index_name"`
	OptimizeFor  string   `json:"optimize_for"` // write_throughput, read_performance, storage
	Workload     string   `json:"workload,omitempty"` // bulk_write, real_time_write, read_heavy
	CorpusSize   string   `json:"corpus_size,omitempty"` // small, medium, large, huge
	Priority     string   `json:"priority,omitempty"` // write_throughput, read_latency, storage_efficiency
	ApplyChanges bool     `json:"apply_changes"`
}

// OptimizationResponse represents the response from index optimization
type OptimizationResponse struct {
	IndexName         string                 `json:"index_name"`
	CurrentSettings   map[string]interface{} `json:"current_settings"`
	RecommendedSettings map[string]interface{} `json:"recommended_settings"`
	OptimizationsApplied []OptimizationChange `json:"optimizations_applied"`
	PerformanceImpact   *PerformanceImpact   `json:"performance_impact"`
	Applied             bool                   `json:"applied"`
	RequestID           string                 `json:"request_id"`
	Timestamp           time.Time              `json:"timestamp"`
}

// OptimizationChange represents a single optimization change
type OptimizationChange struct {
	Setting     string      `json:"setting"`
	OldValue    interface{} `json:"old_value"`
	NewValue    interface{} `json:"new_value"`
	Reason      string      `json:"reason"`
	Impact      string      `json:"impact"` // low, medium, high
	Category    string      `json:"category"` // write_performance, storage, reliability
}

// PerformanceImpact represents the expected impact of optimizations
type PerformanceImpact struct {
	WritePerformance  string `json:"write_performance"` // improved, degraded, neutral
	ReadPerformance   string `json:"read_performance"`
	StorageEfficiency string `json:"storage_efficiency"`
	ResourceUsage     string `json:"resource_usage"`
	EstimatedImprovementPercent float64 `json:"estimated_improvement_percent"`
}

// IndexTemplateRequest represents a request to create an index template
type IndexTemplateRequest struct {
	TemplateName string                 `json:"template_name" binding:"required"`
	IndexPatterns []string              `json:"index_patterns" binding:"required"`
	Settings     *IndexSettings         `json:"settings,omitempty"`
	Mappings     map[string]interface{} `json:"mappings,omitempty"`
	Aliases      map[string]interface{} `json:"aliases,omitempty"`
	Priority     int                    `json:"priority,omitempty"`
	Version      int                    `json:"version,omitempty"`
	Metadata     map[string]interface{} `json:"_meta,omitempty"`
	WriteOptimized bool                 `json:"write_optimized,omitempty"`
	TextHeavy      bool                 `json:"text_heavy,omitempty"`
}

// IndexTemplateResponse represents the response after creating an index template
type IndexTemplateResponse struct {
	TemplateName  string    `json:"template_name"`
	Acknowledged  bool      `json:"acknowledged"`
	IndexPatterns []string  `json:"index_patterns"`
	RequestID     string    `json:"request_id"`
	Timestamp     time.Time `json:"timestamp"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}