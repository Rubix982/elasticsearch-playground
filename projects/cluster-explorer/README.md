# Elasticsearch Cluster Explorer

A comprehensive learning project to understand Elasticsearch clusters in depth - their architecture, APIs, management, and performance characteristics.

## 🎯 Learning Objectives

By the end of this project, you'll understand:

- **Cluster Architecture**: Nodes, roles, and cluster topology
- **Cluster APIs**: Health, stats, settings, and node management
- **Shard Management**: Allocation, rebalancing, and routing
- **Performance Monitoring**: Metrics, bottlenecks, and optimization
- **Cluster Operations**: Scaling, maintenance, and troubleshooting

## 📚 What is an Elasticsearch Cluster?

An Elasticsearch cluster is a collection of one or more nodes (servers) that together hold your entire data and provide federated indexing and search capabilities across all nodes.

### Key Concepts

- **Cluster**: A collection of nodes identified by a unique cluster name
- **Node**: A single server that stores data and participates in indexing/searching
- **Shard**: A subset of an index that lives on a single node
- **Replica**: A copy of a shard for redundancy and performance

## 🏗️ Project Structure

```
cluster-explorer/
├── cmd/                    # Main application
├── internal/
│   ├── handlers/          # HTTP handlers for cluster APIs
│   ├── services/          # Business logic for cluster operations
│   └── models/           # Data structures for cluster info
├── examples/             # Practical examples and learning guides
├── configs/              # Configuration files
└── templates/            # Web UI templates (future)
```

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+
- Make (optional, for convenience)

### Setup Steps

1. **Start Elasticsearch cluster**:

   ```bash
   make docker-up
   ```

2. **Build and run the cluster explorer**:

   ```bash
   make build PROJECT=cluster-explorer
   make run PROJECT=cluster-explorer
   ```

3. **Access the explorer**:
   - **Web interface**: <http://localhost:8081>
   - **API endpoints**: <http://localhost:8081/api/v1/cluster/*>
   - **Health check**: <http://localhost:8081/health>

## 🔍 Complete Feature Set

### 1. Cluster Health & Status

- **Real-time health monitoring** with Green/Yellow/Red status
- **Node count and roles** analysis
- **Shard allocation status** tracking
- **Index health overview** with detailed metrics

### 2. Node Management

- **Detailed node information** including OS, JVM, and hardware specs
- **Hot threads analysis** for performance bottlenecks
- **Node roles and responsibilities** (master, data, ingest, etc.)
- **Resource usage monitoring** (CPU, memory, disk)

### 3. Shard Operations

- **Shard allocation tracking** across all nodes
- **Rebalancing operations** monitoring
- **Routing table inspection**
- **Unassigned shard analysis** with reasons

### 4. Performance Insights

- **Cluster performance metrics** aggregated from all nodes
- **Bottleneck identification** using hot threads
- **Resource utilization analysis**
- **Search and indexing performance** statistics

### 5. Settings Management

- **View current cluster settings** (persistent and transient)
- **Update cluster settings** with validation
- **Performance tuning** recommendations

## 🎓 Comprehensive Learning Path

### Level 1: Cluster Basics (Start Here!)

1. **Understanding Cluster Health**

   - Green, Yellow, Red states and what they mean
   - Why single-node clusters are typically Yellow
   - Health API exploration and interpretation
   - Common health issues and solutions

2. **Node Discovery and Roles**
   - Node types: master, data, ingest, coordinating
   - Node statistics and what they tell youFis
   - Hot threads analysis for performance debugging

### Level 2: Shard Management

1. **Shard Allocation Fundamentals**

   - Primary vs replica shards
   - Allocation awareness and rack awareness
   - Shard routing and request distribution

2. **Cluster Rebalancing**
   - Automatic rebalancing triggers
   - Manual shard movement
   - Allocation filters and exclusions

### Level 3: Advanced Operations

1. **Cluster Settings and Tuning**

   - Dynamic vs static settings
   - Persistent vs transient settings
   - Performance tuning for different workloads

2. **Monitoring & Alerting**
   - Key metrics to monitor continuously
   - Setting up effective alerts
   - Performance optimization strategies

## 🛠️ API Reference

### Core Cluster APIs

```bash
# Comprehensive cluster information
curl "http://localhost:8081/api/v1/cluster/info"

# Quick cluster overview
curl "http://localhost:8081/api/v1/cluster/overview"

# Individual components
curl "http://localhost:8081/api/v1/cluster/health"
curl "http://localhost:8081/api/v1/cluster/state"
curl "http://localhost:8081/api/v1/cluster/stats"
curl "http://localhost:8081/api/v1/cluster/nodes"
curl "http://localhost:8081/api/v1/cluster/indices"
curl "http://localhost:8081/api/v1/cluster/shards"
```

### Monitoring APIs

```bash
# Real-time health monitoring (Server-Sent Events)
curl "http://localhost:8081/api/v1/cluster/monitor/health?interval=5s"

# Performance metrics
curl "http://localhost:8081/api/v1/cluster/performance"

# Hot threads analysis
curl "http://localhost:8081/api/v1/cluster/nodes/_all/hot-threads"
```

### Settings Management

```bash
# View current settings
curl "http://localhost:8081/api/v1/cluster/settings"

# Update settings
curl -X PUT "http://localhost:8081/api/v1/cluster/settings" \
  -H "Content-Type: application/json" \
  -d '{
    "settings": {
      "cluster.routing.allocation.disk.watermark.low": "85%"
    },
    "persistent": true
  }'
```

## 📖 Step-by-Step Learning Guide

### Step 1: Your First Cluster Check

```bash
# Check if everything is running
curl "http://localhost:8081/health"

# Get cluster overview
curl "http://localhost:8081/api/v1/cluster/overview"
```

**What you'll see:** Basic cluster information including name, status, node count, and shard distribution.

### Step 2: Understanding Health Status

```bash
# Get detailed health information
curl "http://localhost:8081/api/v1/cluster/health"
```

**Key insights:**

- **Green**: All shards allocated (rare in single-node setup)
- **Yellow**: Primary shards allocated, replicas unassigned (normal for single-node)
- **Red**: Some primary shards unallocated (needs attention)

### Step 3: Exploring Nodes

```bash
# Get all node information
curl "http://localhost:8081/api/v1/cluster/nodes"
```

**What to look for:**

- Node roles (master, data, ingest)
- JVM heap settings
- Operating system information
- Available processors

### Step 4: Creating Your First Index

```bash
# Create an index to see cluster changes
curl -X PUT "http://localhost:9200/my-learning-index" \
  -H "Content-Type: application/json" \
  -d '{
    "settings": {
      "number_of_shards": 2,
      "number_of_replicas": 1
    }
  }'

# Now check cluster health again
curl "http://localhost:8081/api/v1/cluster/health"
```

**Expected changes:**

- `active_primary_shards`: 2
- `unassigned_shards`: 2 (replicas can't be allocated in single-node)
- Status remains Yellow

### Step 5: Monitoring in Real-Time

```bash
# Stream health updates every 5 seconds
curl "http://localhost:8081/api/v1/cluster/monitor/health?interval=5s"
```

This is perfect for understanding how cluster metrics change over time.

## 🎯 Common Scenarios Explained

### Scenario 1: Fresh Empty Cluster

- **Status**: Green (no indices = no unallocated shards)
- **Nodes**: 1
- **Active shards**: 0
- **What it means**: Perfect starting state

### Scenario 2: Single Node with Data

- **Status**: Yellow (replica shards can't be allocated)
- **Nodes**: 1
- **Active shards**: Number of primary shards
- **Unassigned shards**: Number of replica shards
- **What it means**: Normal for development, data is safe

### Scenario 3: Multi-Node Production

- **Status**: Green (all shards allocated)
- **Nodes**: Multiple
- **Active shards**: Primaries + replicas
- **Unassigned shards**: 0
- **What it means**: Healthy production cluster

## 📊 Performance Monitoring

### Key Metrics to Watch

1. **Cluster Health Status**

   - Should be Green in production
   - Yellow acceptable for development
   - Red requires immediate attention

2. **Shard Allocation**

   - Unassigned shards indicate problems
   - Relocating shards show rebalancing
   - Even distribution across nodes

3. **Node Resources**

   - JVM heap usage (< 85%)
   - CPU utilization
   - Disk space and I/O
   - Network throughput

4. **Performance Counters**
   - Search latency
   - Indexing throughput
   - Query cache hit ratio
   - Thread pool queue sizes

## 🔧 Advanced Features

### Hot Threads Analysis

```bash
# Identify performance bottlenecks
curl "http://localhost:8081/api/v1/cluster/nodes/_all/hot-threads"
```

### Performance Profiling

```bash
# Get comprehensive performance metrics
curl "http://localhost:8081/api/v1/cluster/performance"
```

### Shard Distribution Analysis

```bash
# Understand how shards are distributed
curl "http://localhost:8081/api/v1/cluster/shards"
```

## 🚨 Troubleshooting Guide

### Cluster Not Responding

```bash
# Check if Elasticsearch is running
curl "http://localhost:9200"

# Check cluster explorer health
curl "http://localhost:8081/health"
```

### High Memory Usage

```bash
# Check JVM statistics
curl "http://localhost:8081/api/v1/cluster/performance"
```

### Slow Performance

```bash
# Analyze hot threads
curl "http://localhost:8081/api/v1/cluster/nodes/_all/hot-threads"
```

### Unassigned Shards

```bash
# Investigate shard allocation issues
curl "http://localhost:8081/api/v1/cluster/shards"
```

## 💡 Best Practices

1. **Monitor Health Regularly**: Set up automated health checks
2. **Understand Yellow Status**: Normal for single-node development
3. **Watch Resource Usage**: Keep heap usage below 85%
4. **Plan Shard Strategy**: Don't over-shard small datasets
5. **Use Appropriate Settings**: Tune for your specific workload

## 🎓 Practical Exercises

### Exercise 1: Health Monitoring

Set up continuous health monitoring and observe how metrics change when you:

- Create new indices
- Index documents
- Perform searches
- Restart Elasticsearch

### Exercise 2: Performance Analysis

Use the performance APIs to:

- Identify memory bottlenecks
- Analyze thread pool usage
- Monitor disk I/O patterns

### Exercise 3: Settings Tuning

Experiment with cluster settings:

- Adjust allocation watermarks
- Modify rebalancing settings
- Tune performance parameters

## 📚 Additional Resources

- **Examples Directory**: Check `examples/basic-cluster-operations.md` for detailed walkthroughs
- **Elasticsearch Docs**: [Official Cluster APIs](https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster.html)
- **Performance Guide**: [Elasticsearch Performance Tips](https://www.elastic.co/guide/en/elasticsearch/reference/current/tune-for-search-speed.html)

## 🎉 What's Next?

After mastering clusters, explore:

1. **Index Management**: Deep dive into index lifecycle
2. **Search APIs**: Advanced query techniques
3. **Aggregations**: Analytics and reporting
4. **Machine Learning**: Anomaly detection
5. **Security**: Authentication and authorization

**Ready to become an Elasticsearch cluster expert? Start with the health API and work your way up!** 🚀
