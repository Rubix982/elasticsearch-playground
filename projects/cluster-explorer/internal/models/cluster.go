package models

import "time"

// ClusterInfo represents comprehensive cluster information
type ClusterInfo struct {
	Health      *ClusterHealth      `json:"health"`
	State       *ClusterState       `json:"state"`
	Stats       *ClusterStats       `json:"stats"`
	Nodes       []NodeInfo          `json:"nodes"`
	Indices     []IndexInfo         `json:"indices"`
	Shards      *ShardAllocation    `json:"shards"`
	Performance *PerformanceMetrics `json:"performance"`
	RequestID   string              `json:"request_id"`
	Timestamp   time.Time           `json:"timestamp"`
}

// ClusterHealth represents detailed cluster health information
type ClusterHealth struct {
	ClusterName                 string  `json:"cluster_name"`
	Status                      string  `json:"status"`
	TimedOut                    bool    `json:"timed_out"`
	NumberOfNodes               int     `json:"number_of_nodes"`
	NumberOfDataNodes           int     `json:"number_of_data_nodes"`
	ActivePrimaryShards         int     `json:"active_primary_shards"`
	ActiveShards                int     `json:"active_shards"`
	RelocatingShards           int     `json:"relocating_shards"`
	InitializingShards         int     `json:"initializing_shards"`
	UnassignedShards           int     `json:"unassigned_shards"`
	DelayedUnassignedShards    int     `json:"delayed_unassigned_shards"`
	NumberOfPendingTasks       int     `json:"number_of_pending_tasks"`
	NumberOfInFlightFetch      int     `json:"number_of_in_flight_fetch"`
	TaskMaxWaitingInQueueMillis int     `json:"task_max_waiting_in_queue_millis"`
	ActiveShardsPercentAsNumber float64 `json:"active_shards_percent_as_number"`
}

// ClusterState represents the cluster state information
type ClusterState struct {
	ClusterName    string                    `json:"cluster_name"`
	ClusterUUID    string                    `json:"cluster_uuid"`
	Version        int                       `json:"version"`
	StateUUID      string                    `json:"state_uuid"`
	MasterNode     string                    `json:"master_node"`
	Blocks         map[string]interface{}    `json:"blocks"`
	Nodes          map[string]NodeState      `json:"nodes"`
	Metadata       ClusterMetadata           `json:"metadata"`
	RoutingTable   map[string]interface{}    `json:"routing_table"`
	RoutingNodes   RoutingNodes              `json:"routing_nodes"`
}

// NodeState represents node state in cluster
type NodeState struct {
	Name             string            `json:"name"`
	EphemeralID      string            `json:"ephemeral_id"`
	TransportAddress string            `json:"transport_address"`
	Attributes       map[string]string `json:"attributes"`
	Roles            []string          `json:"roles"`
}

// ClusterMetadata represents cluster metadata
type ClusterMetadata struct {
	ClusterUUID          string                    `json:"cluster_uuid"`
	ClusterCoordination  ClusterCoordination       `json:"cluster_coordination"`
	Templates            map[string]interface{}    `json:"templates"`
	Indices              map[string]interface{}    `json:"indices"`
	IndexGraveyard       IndexGraveyard            `json:"index-graveyard"`
	ClusterUUIDCommitted bool                      `json:"cluster_uuid_committed"`
}

// ClusterCoordination represents cluster coordination settings
type ClusterCoordination struct {
	Term                 int      `json:"term"`
	LastCommittedConfig  []string `json:"last_committed_config"`
	LastAcceptedConfig   []string `json:"last_accepted_config"`
	VotingConfigExclusions []interface{} `json:"voting_config_exclusions"`
}

// IndexGraveyard represents deleted indices information
type IndexGraveyard struct {
	Tombstones []Tombstone `json:"tombstones"`
}

// Tombstone represents a deleted index
type Tombstone struct {
	Index       IndexTombstone `json:"index"`
	DeleteDateInMillis int64   `json:"delete_date_in_millis"`
}

// IndexTombstone represents the deleted index info
type IndexTombstone struct {
	IndexName string `json:"index_name"`
	IndexUUID string `json:"index_uuid"`
}

// RoutingNodes represents routing node information
type RoutingNodes struct {
	Unassigned []UnassignedShard `json:"unassigned"`
	Nodes      map[string][]Shard `json:"nodes"`
}

// UnassignedShard represents an unassigned shard
type UnassignedShard struct {
	Index                    string `json:"index"`
	Shard                    int    `json:"shard"`
	Primary                  bool   `json:"primary"`
	CurrentState             string `json:"current_state"`
	UnassignedInfo           UnassignedInfo `json:"unassigned_info"`
	AllocationID             AllocationID `json:"allocation_id,omitempty"`
}

// UnassignedInfo represents why a shard is unassigned
type UnassignedInfo struct {
	Reason               string `json:"reason"`
	At                   string `json:"at"`
	FailedAttempts       int    `json:"failed_attempts"`
	Delayed              bool   `json:"delayed"`
	Details              string `json:"details,omitempty"`
	AllocationStatus     string `json:"allocation_status"`
}

// AllocationID represents shard allocation ID
type AllocationID struct {
	ID           string `json:"id"`
	RelocationID string `json:"relocation_id,omitempty"`
}

// Shard represents a shard in the cluster
type Shard struct {
	Index        string       `json:"index"`
	Shard        int          `json:"shard"`
	Primary      bool         `json:"primary"`
	CurrentState string       `json:"current_state"`
	Node         string       `json:"node"`
	RelocatingNode string     `json:"relocating_node,omitempty"`
	AllocationID AllocationID `json:"allocation_id"`
}

// ClusterStats represents cluster statistics
type ClusterStats struct {
	Timestamp     int64       `json:"timestamp"`
	ClusterName   string      `json:"cluster_name"`
	ClusterUUID   string      `json:"cluster_uuid"`
	Status        string      `json:"status"`
	Indices       IndicesStats `json:"indices"`
	Nodes         NodesStats   `json:"nodes"`
}

// IndicesStats represents indices statistics
type IndicesStats struct {
	Count       int         `json:"count"`
	Shards      ShardsStats `json:"shards"`
	Docs        DocsStats   `json:"docs"`
	Store       StoreStats  `json:"store"`
	Fielddata   FielddataStats `json:"fielddata"`
	QueryCache  QueryCacheStats `json:"query_cache"`
	Completion  CompletionStats `json:"completion"`
	Segments    SegmentsStats `json:"segments"`
}

// ShardsStats represents shard statistics
type ShardsStats struct {
	Total       int     `json:"total"`
	Primaries   int     `json:"primaries"`
	Replication float64 `json:"replication"`
	Index       ShardIndex `json:"index"`
}

// ShardIndex represents shard index statistics
type ShardIndex struct {
	Shards      ShardCounts `json:"shards"`
	Primaries   ShardCounts `json:"primaries"`
	Replication float64     `json:"replication"`
}

// ShardCounts represents shard count statistics
type ShardCounts struct {
	Min int     `json:"min"`
	Max int     `json:"max"`
	Avg float64 `json:"avg"`
}

// DocsStats represents document statistics
type DocsStats struct {
	Count   int64 `json:"count"`
	Deleted int64 `json:"deleted"`
}

// StoreStats represents storage statistics
type StoreStats struct {
	SizeInBytes          int64 `json:"size_in_bytes"`
	ReservedInBytes      int64 `json:"reserved_in_bytes"`
	TotalDataSetSizeInBytes int64 `json:"total_data_set_size_in_bytes,omitempty"`
}

// FielddataStats represents fielddata statistics
type FielddataStats struct {
	MemorySizeInBytes int64 `json:"memory_size_in_bytes"`
	Evictions         int64 `json:"evictions"`
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
	FileSizes                 map[string]interface{} `json:"file_sizes"`
}

// NodesStats represents nodes statistics
type NodesStats struct {
	Count        NodeCounts            `json:"count"`
	Versions     []string              `json:"versions"`
	OS           OSStats               `json:"os"`
	Process      ProcessStats          `json:"process"`
	JVM          JVMStats              `json:"jvm"`
	FS           FSStats               `json:"fs"`
	Plugins      []PluginInfo          `json:"plugins"`
	NetworkTypes NetworkTypesStats     `json:"network_types"`
	DiscoveryTypes DiscoveryTypesStats `json:"discovery_types"`
	PackagingTypes []PackagingTypeStats `json:"packaging_types"`
	Ingest       NodeIngestStats       `json:"ingest"`
}

// NodeCounts represents node count statistics
type NodeCounts struct {
	Total            int `json:"total"`
	CoordinatingOnly int `json:"coordinating_only"`
	Data             int `json:"data"`
	DataCold         int `json:"data_cold"`
	DataContent      int `json:"data_content"`
	DataFrozen       int `json:"data_frozen"`
	DataHot          int `json:"data_hot"`
	DataWarm         int `json:"data_warm"`
	Ingest           int `json:"ingest"`
	Master           int `json:"master"`
	ML               int `json:"ml"`
	RemoteClusterClient int `json:"remote_cluster_client"`
	Transform        int `json:"transform"`
	VotingOnly       int `json:"voting_only"`
}

// OSStats represents operating system statistics
type OSStats struct {
	AvailableProcessors int                   `json:"available_processors"`
	AllocatedProcessors int                   `json:"allocated_processors"`
	Names               []OSNameStats         `json:"names"`
	PrettyNames         []OSPrettyNameStats   `json:"pretty_names"`
	Architectures       []OSArchitectureStats `json:"architectures"`
	Mem                 OSMemStats            `json:"mem"`
}

// OSNameStats represents OS name statistics
type OSNameStats struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// OSPrettyNameStats represents OS pretty name statistics
type OSPrettyNameStats struct {
	PrettyName string `json:"pretty_name"`
	Count      int    `json:"count"`
}

// OSArchitectureStats represents OS architecture statistics
type OSArchitectureStats struct {
	Arch  string `json:"arch"`
	Count int    `json:"count"`
}

// OSMemStats represents OS memory statistics
type OSMemStats struct {
	TotalInBytes int64 `json:"total_in_bytes"`
	FreeInBytes  int64 `json:"free_in_bytes"`
	UsedInBytes  int64 `json:"used_in_bytes"`
	FreePercent  int   `json:"free_percent"`
	UsedPercent  int   `json:"used_percent"`
}

// ProcessStats represents process statistics
type ProcessStats struct {
	CPU                 ProcessCPUStats `json:"cpu"`
	OpenFileDescriptors ProcessFDStats  `json:"open_file_descriptors"`
}

// ProcessCPUStats represents process CPU statistics
type ProcessCPUStats struct {
	Percent int `json:"percent"`
}

// ProcessFDStats represents process file descriptor statistics
type ProcessFDStats struct {
	Min int `json:"min"`
	Max int `json:"max"`
	Avg int `json:"avg"`
}

// JVMStats represents JVM statistics
type JVMStats struct {
	MaxUptimeInMillis int64        `json:"max_uptime_in_millis"`
	Versions          []JVMVersion `json:"versions"`
	Mem               JVMMemStats  `json:"mem"`
	Threads           int64        `json:"threads"`
}

// JVMVersion represents JVM version information
type JVMVersion struct {
	Version   string `json:"version"`
	VMName    string `json:"vm_name"`
	VMVersion string `json:"vm_version"`
	VMVendor  string `json:"vm_vendor"`
	BundledJDK bool   `json:"bundled_jdk"`
	UsingBundledJDK bool `json:"using_bundled_jdk"`
	Count     int    `json:"count"`
}

// JVMMemStats represents JVM memory statistics
type JVMMemStats struct {
	HeapUsedInBytes int64 `json:"heap_used_in_bytes"`
	HeapMaxInBytes  int64 `json:"heap_max_in_bytes"`
}

// FSStats represents filesystem statistics
type FSStats struct {
	TotalInBytes     int64 `json:"total_in_bytes"`
	FreeInBytes      int64 `json:"free_in_bytes"`
	AvailableInBytes int64 `json:"available_in_bytes"`
}

// PluginInfo represents plugin information
type PluginInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Classname   string `json:"classname"`
	HasNativeController bool `json:"has_native_controller"`
}

// NetworkTypesStats represents network types statistics
type NetworkTypesStats struct {
	TransportTypes map[string]int `json:"transport_types"`
	HTTPTypes      map[string]int `json:"http_types"`
}

// DiscoveryTypesStats represents discovery types statistics
type DiscoveryTypesStats struct {
	Types map[string]int `json:"types"`
}

// PackagingTypeStats represents packaging type statistics
type PackagingTypeStats struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// NodeIngestStats represents node ingest statistics
type NodeIngestStats struct {
	NumberOfPipelines int                    `json:"number_of_pipelines"`
	ProcessorStats    map[string]interface{} `json:"processor_stats"`
}

// NodeInfo represents detailed information about a single node
type NodeInfo struct {
	Name            string            `json:"name"`
	TransportAddress string           `json:"transport_address"`
	Host            string            `json:"host"`
	IP              string            `json:"ip"`
	Version         string            `json:"version"`
	BuildFlavor     string            `json:"build_flavor"`
	BuildType       string            `json:"build_type"`
	BuildHash       string            `json:"build_hash"`
	Roles           []string          `json:"roles"`
	Attributes      map[string]string `json:"attributes"`
	Settings        NodeSettings      `json:"settings"`
	OS              NodeOSInfo        `json:"os"`
	Process         NodeProcessInfo   `json:"process"`
	JVM             NodeJVMInfo       `json:"jvm"`
	ThreadPool      map[string]ThreadPoolInfo `json:"thread_pool"`
	Transport       NodeTransportInfo `json:"transport"`
	HTTP            NodeHTTPInfo      `json:"http"`
	Plugins         []PluginInfo      `json:"plugins"`
	Modules         []ModuleInfo      `json:"modules"`
	Ingest          NodeIngestInfo    `json:"ingest"`
	Aggregations    map[string]interface{} `json:"aggregations"`
}

// NodeSettings represents node settings
type NodeSettings struct {
	Path     PathSettings     `json:"path"`
	Network  NetworkSettings  `json:"network,omitempty"`
	HTTP     HTTPSettings     `json:"http,omitempty"`
	Cluster  ClusterSettings  `json:"cluster"`
	Node     NodeIdentity     `json:"node"`
	Discovery DiscoverySettings `json:"discovery,omitempty"`
}

// PathSettings represents path settings
type PathSettings struct {
	Logs string   `json:"logs"`
	Home string   `json:"home"`
	Repo []string `json:"repo,omitempty"`
	Data []string `json:"data"`
}

// NetworkSettings represents network settings
type NetworkSettings struct {
	Host string `json:"host,omitempty"`
}

// HTTPSettings represents HTTP settings
type HTTPSettings struct {
	Port string `json:"port,omitempty"`
}

// ClusterSettings represents cluster settings in node config
type ClusterSettings struct {
	Name string `json:"name"`
}

// NodeIdentity represents node identity settings
type NodeIdentity struct {
	Name string   `json:"name"`
	Roles []string `json:"roles"`
}

// DiscoverySettings represents discovery settings
type DiscoverySettings struct {
	SeedHosts []string `json:"seed_hosts,omitempty"`
}

// NodeOSInfo represents OS information for a node
type NodeOSInfo struct {
	RefreshIntervalInMillis int64  `json:"refresh_interval_in_millis"`
	Name                    string `json:"name"`
	PrettyName              string `json:"pretty_name"`
	Arch                    string `json:"arch"`
	Version                 string `json:"version"`
	AvailableProcessors     int    `json:"available_processors"`
	AllocatedProcessors     int    `json:"allocated_processors"`
}

// NodeProcessInfo represents process information for a node
type NodeProcessInfo struct {
	RefreshIntervalInMillis int64 `json:"refresh_interval_in_millis"`
	ID                      int64 `json:"id"`
	MLockall                bool  `json:"mlockall"`
}

// NodeJVMInfo represents JVM information for a node
type NodeJVMInfo struct {
	PID               int64    `json:"pid"`
	Version           string   `json:"version"`
	VMName            string   `json:"vm_name"`
	VMVersion         string   `json:"vm_version"`
	VMVendor          string   `json:"vm_vendor"`
	BundledJDK        bool     `json:"bundled_jdk"`
	UsingBundledJDK   bool     `json:"using_bundled_jdk"`
	StartTimeInMillis int64    `json:"start_time_in_millis"`
	Mem               JVMMemInfo `json:"mem"`
	GCCollectors      []string `json:"gc_collectors"`
	MemoryPools       []string `json:"memory_pools"`
	UsingCompressedOrdinaryObjectPointers string `json:"using_compressed_ordinary_object_pointers"`
	InputArguments    []string `json:"input_arguments"`
}

// JVMMemInfo represents JVM memory information
type JVMMemInfo struct {
	HeapInitInBytes    int64 `json:"heap_init_in_bytes"`
	HeapMaxInBytes     int64 `json:"heap_max_in_bytes"`
	NonHeapInitInBytes int64 `json:"non_heap_init_in_bytes"`
	NonHeapMaxInBytes  int64 `json:"non_heap_max_in_bytes"`
	DirectMaxInBytes   int64 `json:"direct_max_in_bytes"`
}

// ThreadPoolInfo represents thread pool information
type ThreadPoolInfo struct {
	Type      string `json:"type"`
	Min       int    `json:"min,omitempty"`
	Max       int    `json:"max,omitempty"`
	KeepAlive string `json:"keep_alive,omitempty"`
	QueueSize int    `json:"queue_size,omitempty"`
}

// NodeTransportInfo represents transport information for a node
type NodeTransportInfo struct {
	BoundAddress      []string          `json:"bound_address"`
	PublishAddress    string            `json:"publish_address"`
	Profiles          map[string]interface{} `json:"profiles"`
}

// NodeHTTPInfo represents HTTP information for a node
type NodeHTTPInfo struct {
	BoundAddress      []string `json:"bound_address"`
	PublishAddress    string   `json:"publish_address"`
	MaxContentLength  string   `json:"max_content_length"`
}

// ModuleInfo represents module information
type ModuleInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Classname   string `json:"classname"`
	HasNativeController bool `json:"has_native_controller"`
}

// NodeIngestInfo represents ingest information for a node
type NodeIngestInfo struct {
	Processors []ProcessorInfo `json:"processors"`
}

// ProcessorInfo represents processor information
type ProcessorInfo struct {
	Type string `json:"type"`
}

// IndexInfo represents information about an index
type IndexInfo struct {
	Index    string          `json:"index"`
	UUID     string          `json:"uuid"`
	Health   string          `json:"health"`
	Status   string          `json:"status"`
	Primary  int             `json:"pri"`
	Replica  int             `json:"rep"`
	DocsCount int64          `json:"docs.count"`
	DocsDeleted int64        `json:"docs.deleted"`
	StoreSize string         `json:"store.size"`
	PrimaryStoreSize string  `json:"pri.store.size"`
	Settings IndexSettings   `json:"settings"`
	Mappings interface{}     `json:"mappings"`
	Aliases  map[string]interface{} `json:"aliases"`
}

// IndexSettings represents index settings
type IndexSettings struct {
	Index IndexConfig `json:"index"`
}

// IndexConfig represents index configuration
type IndexConfig struct {
	CreationDate           string `json:"creation_date"`
	NumberOfShards         string `json:"number_of_shards"`
	NumberOfReplicas       string `json:"number_of_replicas"`
	UUID                   string `json:"uuid"`
	Version                map[string]interface{} `json:"version"`
	ProvidedName           string `json:"provided_name"`
	RoutingPartitionSize   string `json:"routing_partition_size,omitempty"`
	MaxResultWindow        string `json:"max_result_window,omitempty"`
	BlocksReadOnlyAllowDelete string `json:"blocks.read_only_allow_delete,omitempty"`
}

// ShardAllocation represents shard allocation information
type ShardAllocation struct {
	Indices     map[string]IndexAllocation `json:"indices"`
	Unassigned  []UnassignedShardDetails   `json:"unassigned"`
	Summary     AllocationSummary          `json:"summary"`
}

// IndexAllocation represents allocation for a specific index
type IndexAllocation struct {
	Shards map[string][]ShardDetails `json:"shards"`
}

// ShardDetails represents detailed shard information
type ShardDetails struct {
	State          string `json:"state"`
	Primary        bool   `json:"primary"`
	Node           string `json:"node"`
	RelocatingNode string `json:"relocating_node,omitempty"`
	Index          string `json:"index"`
	Shard          int    `json:"shard"`
	PriraryTerm    int64  `json:"primary_term"`
	GlobalCheckpoint int64 `json:"global_checkpoint"`
	LocalCheckpoint  int64 `json:"local_checkpoint"`
	Docs           int64  `json:"docs"`
	Store          string `json:"store"`
	Segments       SegmentDetails `json:"segments"`
}

// SegmentDetails represents segment details for a shard
type SegmentDetails struct {
	Count   int64  `json:"count"`
	Memory  string `json:"memory"`
}

// UnassignedShardDetails represents detailed unassigned shard information
type UnassignedShardDetails struct {
	Index        string `json:"index"`
	Shard        int    `json:"shard"`
	Primary      bool   `json:"primary"`
	CurrentState string `json:"current_state"`
	Reason       string `json:"unassigned_reason"`
	Since        string `json:"unassigned_since"`
	Details      string `json:"details,omitempty"`
	NodeDecisions []NodeDecision `json:"node_decisions,omitempty"`
}

// NodeDecision represents allocation decision for a node
type NodeDecision struct {
	NodeName string `json:"node_name"`
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// AllocationSummary represents allocation summary
type AllocationSummary struct {
	TotalShards        int `json:"total_shards"`
	AssignedShards     int `json:"assigned_shards"`
	UnassignedShards   int `json:"unassigned_shards"`
	RelocatingShards   int `json:"relocating_shards"`
	InitializingShards int `json:"initializing_shards"`
}

// PerformanceMetrics represents cluster performance metrics
type PerformanceMetrics struct {
	CPU              CPUMetrics      `json:"cpu"`
	Memory           MemoryMetrics   `json:"memory"`
	Disk             DiskMetrics     `json:"disk"`
	Network          NetworkMetrics  `json:"network"`
	GarbageCollection GCMetrics      `json:"gc"`
	ThreadPools       ThreadPoolMetrics `json:"thread_pools"`
	Search            SearchMetrics   `json:"search"`
	Indexing          IndexingMetrics `json:"indexing"`
}

// CPUMetrics represents CPU performance metrics
type CPUMetrics struct {
	LoadAverage  LoadAverageMetrics `json:"load_average"`
	UsagePercent float64            `json:"usage_percent"`
}

// LoadAverageMetrics represents load average metrics
type LoadAverageMetrics struct {
	OneMinute     float64 `json:"1m"`
	FiveMinutes   float64 `json:"5m"`
	FifteenMinutes float64 `json:"15m"`
}

// MemoryMetrics represents memory performance metrics
type MemoryMetrics struct {
	HeapUsedPercent    float64 `json:"heap_used_percent"`
	HeapUsedBytes      int64   `json:"heap_used_bytes"`
	HeapMaxBytes       int64   `json:"heap_max_bytes"`
	NonHeapUsedBytes   int64   `json:"non_heap_used_bytes"`
	DirectMemoryUsed   int64   `json:"direct_memory_used"`
}

// DiskMetrics represents disk performance metrics
type DiskMetrics struct {
	TotalBytes     int64   `json:"total_bytes"`
	FreeBytes      int64   `json:"free_bytes"`
	UsedBytes      int64   `json:"used_bytes"`
	UsedPercent    float64 `json:"used_percent"`
	IOOperations   IOMetrics `json:"io_operations"`
}

// IOMetrics represents I/O performance metrics
type IOMetrics struct {
	ReadOpsPerSec  float64 `json:"read_ops_per_sec"`
	WriteOpsPerSec float64 `json:"write_ops_per_sec"`
	ReadBytesPerSec  int64 `json:"read_bytes_per_sec"`
	WriteBytesPerSec int64 `json:"write_bytes_per_sec"`
}

// NetworkMetrics represents network performance metrics
type NetworkMetrics struct {
	BytesReceived    int64 `json:"bytes_received"`
	BytesSent        int64 `json:"bytes_sent"`
	PacketsReceived  int64 `json:"packets_received"`
	PacketsSent      int64 `json:"packets_sent"`
}

// GCMetrics represents garbage collection metrics
type GCMetrics struct {
	YoungGenCollections int64         `json:"young_gen_collections"`
	YoungGenTime        time.Duration `json:"young_gen_time"`
	OldGenCollections   int64         `json:"old_gen_collections"`
	OldGenTime          time.Duration `json:"old_gen_time"`
}

// ThreadPoolMetrics represents thread pool metrics
type ThreadPoolMetrics struct {
	Search    ThreadPoolStats `json:"search"`
	Index     ThreadPoolStats `json:"index"`
	Bulk      ThreadPoolStats `json:"bulk"`
	Get       ThreadPoolStats `json:"get"`
	Management ThreadPoolStats `json:"management"`
}

// ThreadPoolStats represents thread pool statistics
type ThreadPoolStats struct {
	Threads   int   `json:"threads"`
	Queue     int   `json:"queue"`
	Active    int   `json:"active"`
	Rejected  int64 `json:"rejected"`
	Largest   int   `json:"largest"`
	Completed int64 `json:"completed"`
}

// SearchMetrics represents search performance metrics
type SearchMetrics struct {
	QueryTotal        int64         `json:"query_total"`
	QueryTime         time.Duration `json:"query_time_in_millis"`
	QueryCurrent      int64         `json:"query_current"`
	FetchTotal        int64         `json:"fetch_total"`
	FetchTime         time.Duration `json:"fetch_time_in_millis"`
	FetchCurrent      int64         `json:"fetch_current"`
	ScrollTotal       int64         `json:"scroll_total"`
	ScrollTime        time.Duration `json:"scroll_time_in_millis"`
	ScrollCurrent     int64         `json:"scroll_current"`
	SuggestTotal      int64         `json:"suggest_total"`
	SuggestTime       time.Duration `json:"suggest_time_in_millis"`
	SuggestCurrent    int64         `json:"suggest_current"`
}

// IndexingMetrics represents indexing performance metrics
type IndexingMetrics struct {
	IndexTotal         int64         `json:"index_total"`
	IndexTime          time.Duration `json:"index_time_in_millis"`
	IndexCurrent       int64         `json:"index_current"`
	IndexFailed        int64         `json:"index_failed"`
	DeleteTotal        int64         `json:"delete_total"`
	DeleteTime         time.Duration `json:"delete_time_in_millis"`
	DeleteCurrent      int64         `json:"delete_current"`
	NoopUpdateTotal    int64         `json:"noop_update_total"`
	IsThrottled        bool          `json:"is_throttled"`
	ThrottleTime       time.Duration `json:"throttle_time_in_millis"`
}