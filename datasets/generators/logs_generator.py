#!/usr/bin/env python3
"""
Log Events Dataset Generator for Elasticsearch Write Optimization Testing

Generates realistic application logs, system events, and metrics data.
"""

import json
import random
import argparse
import uuid
from datetime import datetime, timedelta
from faker import Faker
import os

fake = Faker()

class LogsGenerator:
    def __init__(self):
        self.log_levels = ['DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL']
        self.services = [
            'auth-service', 'user-service', 'payment-service', 'order-service',
            'inventory-service', 'notification-service', 'api-gateway', 'web-frontend',
            'mobile-backend', 'analytics-service', 'search-service', 'recommendations'
        ]
        
        self.hosts = [f"server-{i:03d}" for i in range(1, 21)]  # 20 hosts
        self.environments = ['production', 'staging', 'development', 'test']
        
        self.http_methods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS']
        self.status_codes = [200, 201, 204, 301, 302, 400, 401, 403, 404, 500, 502, 503]
        
        self.user_agents = [
            'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
            'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
            'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36',
            'Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X)',
            'Mozilla/5.0 (Android 11; Mobile; rv:68.0) Gecko/68.0'
        ]
    
    def generate_application_log(self, log_id=None):
        """Generate application log entry"""
        timestamp = fake.date_time_between(start_date='-7d')
        level = random.choice(self.log_levels)
        service = random.choice(self.services)
        
        # Generate message based on log level
        if level == 'ERROR':
            message = f"Error processing request: {fake.sentence()}"
            exception = {
                "type": random.choice(['NullPointerException', 'SQLException', 'TimeoutException', 'ValidationException']),
                "message": fake.sentence(),
                "stack_trace": [f"at {fake.word()}.{fake.word()}({fake.word()}.java:{random.randint(1, 500)})" for _ in range(random.randint(3, 8))]
            } if random.random() < 0.7 else None
        elif level == 'WARN':
            message = f"Warning: {fake.sentence()}"
            exception = None
        elif level == 'INFO':
            message = f"Successfully processed: {fake.sentence()}"
            exception = None
        elif level == 'DEBUG':
            message = f"Debug info: {fake.sentence()}"
            exception = None
        else:  # FATAL
            message = f"Critical system failure: {fake.sentence()}"
            exception = {
                "type": "SystemException",
                "message": fake.sentence(),
                "stack_trace": [f"at {fake.word()}.{fake.word()}({fake.word()}.java:{random.randint(1, 500)})" for _ in range(random.randint(5, 15))]
            }
        
        return {
            "id": log_id or str(uuid.uuid4()),
            "timestamp": timestamp.isoformat(),
            "level": level,
            "message": message,
            "service": service,
            "host": random.choice(self.hosts),
            "environment": random.choice(self.environments),
            "thread": f"thread-{random.randint(1, 100)}",
            "logger": f"{service}.{fake.word()}",
            "correlation_id": str(uuid.uuid4())[:8],
            "session_id": str(uuid.uuid4())[:16] if random.random() < 0.6 else None,
            "user_id": fake.random_int(1, 100000) if random.random() < 0.4 else None,
            "request_id": str(uuid.uuid4()) if random.random() < 0.8 else None,
            "duration_ms": random.randint(1, 5000) if level in ['INFO', 'WARN'] else None,
            "memory_usage": {
                "heap": f"{random.randint(100, 2048)}MB",
                "non_heap": f"{random.randint(50, 512)}MB",
                "used": f"{random.randint(200, 1500)}MB"
            } if random.random() < 0.2 else None,
            "exception": exception,
            "tags": [fake.word() for _ in range(random.randint(1, 4))],
            "metadata": {
                "version": f"{random.randint(1, 3)}.{random.randint(0, 9)}.{random.randint(0, 9)}",
                "build": f"build-{random.randint(1000, 9999)}",
                "region": fake.country_code(),
                "datacenter": f"dc-{random.choice(['us-east', 'us-west', 'eu-central', 'ap-south'])}"
            }
        }
    
    def generate_access_log(self, log_id=None):
        """Generate HTTP access log entry"""
        timestamp = fake.date_time_between(start_date='-7d')
        method = random.choice(self.http_methods)
        status_code = random.choice(self.status_codes)
        
        return {
            "id": log_id or str(uuid.uuid4()),
            "timestamp": timestamp.isoformat(),
            "type": "access",
            "remote_addr": fake.ipv4(),
            "remote_user": fake.user_name() if random.random() < 0.3 else None,
            "request": {
                "method": method,
                "url": fake.uri_path(),
                "query_string": fake.uri_path() if random.random() < 0.4 else None,
                "protocol": "HTTP/1.1",
                "headers": {
                    "user_agent": random.choice(self.user_agents),
                    "referer": fake.url() if random.random() < 0.6 else None,
                    "accept_language": random.choice(['en-US', 'en-GB', 'es-ES', 'fr-FR', 'de-DE']),
                    "x_forwarded_for": fake.ipv4() if random.random() < 0.3 else None
                }
            },
            "response": {
                "status_code": status_code,
                "content_length": random.randint(100, 50000) if status_code == 200 else random.randint(0, 1000),
                "content_type": random.choice(['text/html', 'application/json', 'text/css', 'application/javascript', 'image/png']),
                "cache_status": random.choice(['HIT', 'MISS', 'BYPASS']) if random.random() < 0.7 else None
            },
            "timing": {
                "response_time_ms": random.randint(1, 3000),
                "upstream_time_ms": random.randint(1, 2000) if random.random() < 0.6 else None
            },
            "geo": {
                "country": fake.country_code(),
                "city": fake.city(),
                "coordinates": {
                    "lat": float(fake.latitude()),
                    "lon": float(fake.longitude())
                }
            } if random.random() < 0.8 else None,
            "session_id": str(uuid.uuid4())[:16] if random.random() < 0.5 else None,
            "user_id": fake.random_int(1, 100000) if random.random() < 0.3 else None,
            "host": random.choice(self.hosts),
            "server_name": fake.domain_name(),
            "ssl": random.random() < 0.8,
            "bot_detected": random.random() < 0.1
        }
    
    def generate_system_metric(self, metric_id=None):
        """Generate system metrics log entry"""
        timestamp = fake.date_time_between(start_date='-7d')
        
        return {
            "id": metric_id or str(uuid.uuid4()),
            "timestamp": timestamp.isoformat(),
            "type": "metric",
            "host": random.choice(self.hosts),
            "service": random.choice(self.services),
            "metrics": {
                "cpu": {
                    "usage_percent": round(random.uniform(0, 100), 2),
                    "load_1m": round(random.uniform(0, 8), 2),
                    "load_5m": round(random.uniform(0, 6), 2),
                    "load_15m": round(random.uniform(0, 4), 2)
                },
                "memory": {
                    "usage_percent": round(random.uniform(20, 95), 2),
                    "used_mb": random.randint(1000, 8000),
                    "available_mb": random.randint(2000, 16000),
                    "swap_used_mb": random.randint(0, 2000)
                },
                "disk": {
                    "usage_percent": round(random.uniform(10, 90), 2),
                    "read_iops": random.randint(0, 1000),
                    "write_iops": random.randint(0, 500),
                    "read_throughput_mb": round(random.uniform(0, 100), 2),
                    "write_throughput_mb": round(random.uniform(0, 50), 2)
                },
                "network": {
                    "rx_bytes": random.randint(1000000, 100000000),
                    "tx_bytes": random.randint(500000, 50000000),
                    "rx_packets": random.randint(1000, 100000),
                    "tx_packets": random.randint(500, 50000),
                    "errors": random.randint(0, 10)
                },
                "application": {
                    "active_connections": random.randint(0, 1000),
                    "queue_size": random.randint(0, 100),
                    "threads_active": random.randint(5, 200),
                    "heap_usage_mb": random.randint(100, 2048),
                    "gc_collections": random.randint(0, 20),
                    "gc_time_ms": random.randint(0, 500)
                }
            },
            "alerts": [
                {
                    "type": random.choice(['cpu_high', 'memory_high', 'disk_full', 'service_down']),
                    "severity": random.choice(['warning', 'critical']),
                    "message": fake.sentence()
                } for _ in range(random.randint(0, 2))
            ] if random.random() < 0.1 else [],
            "environment": random.choice(self.environments)
        }
    
    def generate_security_event(self, event_id=None):
        """Generate security event log entry"""
        timestamp = fake.date_time_between(start_date='-7d')
        event_types = ['login_attempt', 'permission_denied', 'suspicious_activity', 'data_access', 'config_change']
        
        return {
            "id": event_id or str(uuid.uuid4()),
            "timestamp": timestamp.isoformat(),
            "type": "security",
            "event_type": random.choice(event_types),
            "severity": random.choice(['low', 'medium', 'high', 'critical']),
            "source_ip": fake.ipv4(),
            "user": {
                "id": fake.random_int(1, 100000) if random.random() < 0.8 else None,
                "username": fake.user_name() if random.random() < 0.8 else None,
                "role": random.choice(['admin', 'user', 'guest', 'service']) if random.random() < 0.8 else None
            },
            "resource": {
                "type": random.choice(['file', 'database', 'api_endpoint', 'configuration']),
                "path": fake.file_path(),
                "permissions": random.choice(['read', 'write', 'execute', 'admin'])
            },
            "action": {
                "attempted": random.choice(['access', 'modify', 'delete', 'create', 'login']),
                "result": random.choice(['success', 'failure', 'blocked']),
                "details": fake.sentence()
            },
            "geo_location": {
                "country": fake.country_code(),
                "city": fake.city(),
                "coordinates": {
                    "lat": float(fake.latitude()),
                    "lon": float(fake.longitude())
                }
            } if random.random() < 0.7 else None,
            "device_info": {
                "user_agent": random.choice(self.user_agents),
                "device_type": random.choice(['desktop', 'mobile', 'tablet', 'server']),
                "os": random.choice(['Windows', 'macOS', 'Linux', 'iOS', 'Android'])
            } if random.random() < 0.6 else None,
            "risk_score": round(random.uniform(0, 100), 1),
            "tags": [fake.word() for _ in range(random.randint(1, 4))],
            "correlation_id": str(uuid.uuid4())[:8]
        }
    
    def generate_mixed_logs(self, count):
        """Generate mixed log types for realistic workload"""
        logs = []
        
        # Distribution: 50% app logs, 30% access logs, 15% metrics, 5% security
        app_count = int(count * 0.5)
        access_count = int(count * 0.3)
        metric_count = int(count * 0.15)
        security_count = count - app_count - access_count - metric_count
        
        for i in range(app_count):
            logs.append(self.generate_application_log(f"app_{i}"))
        
        for i in range(access_count):
            logs.append(self.generate_access_log(f"access_{i}"))
        
        for i in range(metric_count):
            logs.append(self.generate_system_metric(f"metric_{i}"))
        
        for i in range(security_count):
            logs.append(self.generate_security_event(f"security_{i}"))
        
        # Shuffle to simulate realistic log stream
        random.shuffle(logs)
        return logs
    
    def save_as_ndjson(self, logs, filepath):
        """Save logs as NDJSON format"""
        os.makedirs(os.path.dirname(filepath), exist_ok=True)
        with open(filepath, 'w') as f:
            for log in logs:
                f.write(json.dumps(log) + '\n')
    
    def save_as_json(self, logs, filepath):
        """Save logs as JSON array"""
        os.makedirs(os.path.dirname(filepath), exist_ok=True)
        with open(filepath, 'w') as f:
            json.dump(logs, f, indent=2)

def main():
    parser = argparse.ArgumentParser(description='Generate log events and system metrics data')
    parser.add_argument('--count', type=int, default=10000, help='Number of log entries to generate')
    parser.add_argument('--type', choices=['app', 'access', 'metric', 'security', 'mixed'], 
                       default='mixed', help='Type of logs to generate')
    parser.add_argument('--output', default='datasets/samples/log-events.ndjson', 
                       help='Output file path')
    parser.add_argument('--format', choices=['ndjson', 'json'], default='ndjson', 
                       help='Output format')
    parser.add_argument('--seed', type=int, help='Random seed for reproducible generation')
    
    args = parser.parse_args()
    
    if args.seed:
        random.seed(args.seed)
        fake.seed_instance(args.seed)
    
    print(f"📊 Generating {args.count} {args.type} log entries...")
    
    generator = LogsGenerator()
    
    if args.type == 'app':
        logs = [generator.generate_application_log(f"app_{i}") for i in range(args.count)]
    elif args.type == 'access':
        logs = [generator.generate_access_log(f"access_{i}") for i in range(args.count)]
    elif args.type == 'metric':
        logs = [generator.generate_system_metric(f"metric_{i}") for i in range(args.count)]
    elif args.type == 'security':
        logs = [generator.generate_security_event(f"security_{i}") for i in range(args.count)]
    else:  # mixed
        logs = generator.generate_mixed_logs(args.count)
    
    if args.format == 'ndjson':
        generator.save_as_ndjson(logs, args.output)
    else:
        generator.save_as_json(logs, args.output)
    
    # Calculate statistics
    total_size = sum(len(json.dumps(log).encode('utf-8')) for log in logs)
    avg_size = total_size / len(logs)
    
    print(f"✅ Generated {len(logs)} log entries")
    print(f"📊 Statistics:")
    print(f"   • Total size: {total_size / 1024 / 1024:.2f} MB")
    print(f"   • Average size: {avg_size:.0f} bytes")
    print(f"   • Output file: {args.output}")
    if args.type == 'mixed':
        type_counts = {}
        for log in logs:
            log_type = log.get('type', 'application')
            type_counts[log_type] = type_counts.get(log_type, 0) + 1
        print(f"   • Log types: {type_counts}")
    
    print(f"\n💡 Usage Examples:")
    print(f"   # Import via Index Explorer:")
    print(f"   curl -X POST 'http://localhost:8082/api/v1/indices/logs/import/ndjson?batch_size=1000' \\")
    print(f"     --data-binary @{args.output}")

if __name__ == '__main__':
    main()