# Basic Elasticsearch Cluster Operations

This guide demonstrates fundamental cluster operations using the Cluster Explorer API.

## Prerequisites

1. Start Elasticsearch:
   ```bash
   make docker-up
   ```

2. Start the Cluster Explorer:
   ```bash
   make build PROJECT=cluster-explorer
   make run PROJECT=cluster-explorer
   ```

3. The explorer will be available at http://localhost:8081

## 1. Cluster Health - Your First Check

Cluster health is the most important metric in Elasticsearch. It tells you the overall status of your cluster.

### Understanding Health Status

- **Green**: All primary and replica shards are allocated
- **Yellow**: All primary shards are allocated, but some replicas are missing
- **Red**: Some primary shards are not allocated

### API Calls

```bash
# Get cluster health
curl "http://localhost:8081/api/v1/cluster/health"

# Response example:
{
  "health": {
    "cluster_name": "es-playground-cluster",
    "status": "yellow",
    "timed_out": false,
    "number_of_nodes": 1,
    "number_of_data_nodes": 1,
    "active_primary_shards": 0,
    "active_shards": 0,
    "relocating_shards": 0,
    "initializing_shards": 0,
    "unassigned_shards": 0,
    "delayed_unassigned_shards": 0,
    "number_of_pending_tasks": 0,
    "number_of_in_flight_fetch": 0,
    "task_max_waiting_in_queue_millis": 0,
    "active_shards_percent_as_number": 100.0
  }
}
```

### Why Single Node Clusters are Yellow

In our single-node setup, the cluster status is typically **yellow** because:
- Primary shards are allocated (good!)
- Replica shards cannot be allocated (they need a different node)
- This is normal for development environments

## 2. Cluster Overview - Quick Summary

Get a high-level view of your cluster:

```bash
curl "http://localhost:8081/api/v1/cluster/overview"
```

This provides a summarized view perfect for dashboards and monitoring.

## 3. Node Information - Understanding Your Infrastructure

### Get All Nodes
```bash
curl "http://localhost:8081/api/v1/cluster/nodes"
```

### Understanding Node Roles

Each node can have multiple roles:
- **master**: Can be elected as the master node
- **data**: Stores data and performs data-related operations
- **ingest**: Preprocesses documents before indexing
- **coordinating_only**: Routes requests and merges results

### Single Node Setup
In our development setup, one node typically has all roles:
```json
{
  "roles": ["data", "ingest", "master", "remote_cluster_client"]
}
```

## 4. Cluster State - The Brain of Your Cluster

The cluster state contains metadata about your cluster:

```bash
curl "http://localhost:8081/api/v1/cluster/state"
```

Key components:
- **Master node**: Who's in charge
- **Nodes**: All nodes in the cluster
- **Routing table**: How shards are distributed
- **Metadata**: Index settings, mappings, templates

## 5. Cluster Statistics - Performance Metrics

Get detailed statistics about your cluster:

```bash
curl "http://localhost:8081/api/v1/cluster/stats"
```

This includes:
- **Indices stats**: Document counts, storage sizes
- **Nodes stats**: JVM, OS, process information
- **Shard information**: Total shards, primaries, replicas

## 6. Monitoring Cluster Health in Real-time

Monitor health continuously using Server-Sent Events:

```bash
curl "http://localhost:8081/api/v1/cluster/monitor/health?interval=5s"
```

This streams health updates every 5 seconds. Perfect for:
- Dashboard implementations
- Alerting systems
- Real-time monitoring

## 7. Working with Cluster Settings

### View Current Settings
```bash
curl "http://localhost:8081/api/v1/cluster/settings"
```

### Update Settings
```bash
curl -X PUT "http://localhost:8081/api/v1/cluster/settings" \
  -H "Content-Type: application/json" \
  -d '{
    "settings": {
      "cluster.routing.allocation.disk.watermark.low": "85%"
    },
    "persistent": true
  }'
```

### Setting Types
- **Persistent**: Survive cluster restarts
- **Transient**: Reset on cluster restart

## 8. Creating Your First Index

Let's create an index to see cluster changes:

```bash
# Create an index with 2 primary shards and 1 replica
curl -X PUT "http://localhost:9200/my-first-index" \
  -H "Content-Type: application/json" \
  -d '{
    "settings": {
      "number_of_shards": 2,
      "number_of_replicas": 1
    }
  }'
```

Now check cluster health again:
```bash
curl "http://localhost:8081/api/v1/cluster/health"
```

You'll see:
- **active_primary_shards**: 2 (our primary shards)
- **unassigned_shards**: 2 (replica shards can't be allocated)
- **status**: Still yellow (because of unassigned replicas)

## 9. Understanding Shard Allocation

View how shards are distributed:

```bash
curl "http://localhost:8081/api/v1/cluster/shards"
```

This shows:
- **Assigned shards**: Which node they're on
- **Unassigned shards**: Why they can't be allocated
- **Allocation summary**: Overall shard distribution

## 10. Performance Monitoring

Get performance metrics:

```bash
curl "http://localhost:8081/api/v1/cluster/performance"
```

Key metrics include:
- **CPU usage**: Cluster-wide CPU utilization
- **Memory**: Heap usage, garbage collection
- **Disk I/O**: Read/write operations
- **Network**: Data transfer rates

## Common Scenarios

### Scenario 1: New Empty Cluster
- Status: **Green** (no indices yet)
- Nodes: 1
- Active shards: 0

### Scenario 2: Single Node with Indices
- Status: **Yellow** (replica shards unassigned)
- Nodes: 1
- Active shards: Number of primary shards
- Unassigned shards: Number of replica shards

### Scenario 3: Multi-Node Cluster
- Status: **Green** (if all shards allocated)
- Nodes: Multiple
- Active shards: Primaries + replicas
- Unassigned shards: 0 (ideally)

## Best Practices

1. **Monitor Health Regularly**: Check cluster health frequently
2. **Understand Yellow Status**: Normal for single-node development
3. **Watch Unassigned Shards**: Investigate if they appear in multi-node clusters
4. **Monitor Resource Usage**: Keep an eye on CPU, memory, and disk
5. **Use Appropriate Shard Counts**: Don't over-shard small datasets

## Next Steps

1. **Add More Nodes**: Try multi-node clustering
2. **Index Real Data**: Create indices with actual documents
3. **Experiment with Settings**: Try different cluster settings
4. **Monitor Performance**: Watch how metrics change under load
5. **Learn Shard Management**: Understand allocation and rebalancing

## Troubleshooting

### Cluster Not Responding
```bash
# Check if Elasticsearch is running
curl "http://localhost:9200"

# Check cluster explorer health
curl "http://localhost:8081/health"
```

### High Memory Usage
```bash
# Check JVM stats
curl "http://localhost:8081/api/v1/cluster/performance"
```

### Slow Queries
```bash
# Get hot threads (performance bottlenecks)
curl "http://localhost:8081/api/v1/cluster/nodes/_all/hot-threads"
```

Remember: Understanding clusters is fundamental to mastering Elasticsearch. Start with these basics and gradually explore more advanced topics!