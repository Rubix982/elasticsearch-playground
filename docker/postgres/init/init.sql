-- Initialize PostgreSQL database for ES Playground
-- This database can be used for storing metadata, user management, or application state

-- Create application user
CREATE USER es_playground_user WITH PASSWORD 'playground123';

-- Create main database
CREATE DATABASE es_playground_main OWNER es_playground_user;

-- Connect to the main database
\c es_playground_main;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE es_playground_main TO es_playground_user;
GRANT ALL ON SCHEMA public TO es_playground_user;

-- Create sample tables for metadata storage
CREATE TABLE IF NOT EXISTS index_metadata (
    id SERIAL PRIMARY KEY,
    index_name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    document_count BIGINT DEFAULT 0,
    total_size_bytes BIGINT DEFAULT 0,
    optimization_score INTEGER DEFAULT 0,
    last_optimization TIMESTAMP,
    configuration JSONB,
    tags TEXT[]
);

CREATE TABLE IF NOT EXISTS performance_metrics (
    id SERIAL PRIMARY KEY,
    index_name VARCHAR(255) NOT NULL,
    metric_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    docs_per_second DECIMAL(10,2),
    avg_latency_ms INTEGER,
    error_count INTEGER DEFAULT 0,
    batch_size INTEGER,
    worker_count INTEGER,
    total_documents BIGINT,
    test_type VARCHAR(50),
    metadata JSONB
);

CREATE TABLE IF NOT EXISTS optimization_history (
    id SERIAL PRIMARY KEY,
    index_name VARCHAR(255) NOT NULL,
    optimization_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    optimization_type VARCHAR(100),
    before_score INTEGER,
    after_score INTEGER,
    changes_applied JSONB,
    performance_impact JSONB
);

-- Create indexes for better query performance
CREATE INDEX idx_index_metadata_name ON index_metadata(index_name);
CREATE INDEX idx_performance_metrics_index_timestamp ON performance_metrics(index_name, metric_timestamp);
CREATE INDEX idx_optimization_history_index ON optimization_history(index_name);

-- Insert sample data
INSERT INTO index_metadata (index_name, document_count, total_size_bytes, optimization_score, configuration, tags) VALUES
('sample-index', 1000, 5242880, 85, '{"refresh_interval": "30s", "number_of_replicas": 0}', ARRAY['write-optimized', 'test']),
('logs-index', 50000, 20971520, 78, '{"refresh_interval": "60s", "number_of_replicas": 1}', ARRAY['logs', 'time-series']),
('products-index', 10000, 15728640, 92, '{"refresh_interval": "5s", "number_of_replicas": 0}', ARRAY['ecommerce', 'catalog']);

-- Create a view for easy performance monitoring
CREATE VIEW performance_summary AS
SELECT 
    index_name,
    COUNT(*) as test_runs,
    AVG(docs_per_second) as avg_throughput,
    MAX(docs_per_second) as max_throughput,
    AVG(avg_latency_ms) as avg_latency,
    MIN(avg_latency_ms) as min_latency,
    SUM(error_count) as total_errors,
    MAX(metric_timestamp) as last_test
FROM performance_metrics 
GROUP BY index_name;

-- Grant permissions on all tables
GRANT ALL ON ALL TABLES IN SCHEMA public TO es_playground_user;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO es_playground_user;

-- Enable row level security if needed (commented out for development)
-- ALTER TABLE index_metadata ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE performance_metrics ENABLE ROW LEVEL SECURITY;

COMMENT ON DATABASE es_playground_main IS 'PostgreSQL database for ES Playground metadata and performance tracking';
COMMENT ON TABLE index_metadata IS 'Stores metadata about Elasticsearch indices for tracking and optimization';
COMMENT ON TABLE performance_metrics IS 'Stores performance test results and benchmarks';
COMMENT ON TABLE optimization_history IS 'Tracks optimization changes and their impact over time';