#!/usr/bin/env python3
"""
Generic Document Generator for Elasticsearch Write Optimization Testing

This script generates various types of documents for testing write-optimized
Elasticsearch operations with different document sizes and structures.
"""

import json
import random
import argparse
import uuid
from datetime import datetime, timedelta
from faker import Faker
import os

fake = Faker()

class DocumentGenerator:
    def __init__(self):
        self.categories = ['technology', 'business', 'science', 'health', 'sports', 'entertainment']
        self.priorities = ['low', 'medium', 'high', 'critical']
        self.statuses = ['active', 'inactive', 'pending', 'archived']
        
    def generate_small_document(self, doc_id=None):
        """Generate small document (< 1KB) - suitable for logs, events, metrics"""
        return {
            "id": doc_id or str(uuid.uuid4()),
            "timestamp": fake.date_time_between(start_date='-30d').isoformat(),
            "level": random.choice(['INFO', 'WARN', 'ERROR', 'DEBUG']),
            "message": fake.sentence(nb_words=10),
            "service": fake.word(),
            "host": fake.ipv4(),
            "user_id": fake.random_int(1, 10000),
            "session_id": fake.uuid4()[:8],
            "response_time": fake.random_int(10, 500),
            "status_code": random.choice([200, 404, 500, 503])
        }
    
    def generate_medium_document(self, doc_id=None):
        """Generate medium document (1-10KB) - suitable for articles, emails, products"""
        return {
            "id": doc_id or str(uuid.uuid4()),
            "title": fake.sentence(nb_words=8),
            "content": fake.text(max_nb_chars=3000),
            "author": {
                "name": fake.name(),
                "email": fake.email(),
                "bio": fake.text(max_nb_chars=200)
            },
            "metadata": {
                "category": random.choice(self.categories),
                "tags": [fake.word() for _ in range(random.randint(3, 8))],
                "priority": random.choice(self.priorities),
                "status": random.choice(self.statuses),
                "views": fake.random_int(0, 10000),
                "likes": fake.random_int(0, 1000),
                "shares": fake.random_int(0, 500)
            },
            "timestamps": {
                "created": fake.date_time_between(start_date='-1y').isoformat(),
                "updated": fake.date_time_between(start_date='-30d').isoformat(),
                "published": fake.date_time_between(start_date='-6m').isoformat()
            },
            "location": {
                "country": fake.country(),
                "city": fake.city(),
                "coordinates": {
                    "lat": float(fake.latitude()),
                    "lon": float(fake.longitude())
                }
            },
            "analytics": {
                "bounce_rate": round(random.uniform(0.1, 0.8), 2),
                "time_on_page": fake.random_int(30, 600),
                "conversion_rate": round(random.uniform(0.01, 0.15), 3)
            }
        }
    
    def generate_large_document(self, doc_id=None):
        """Generate large document (10-100KB) - suitable for reports, documentation"""
        return {
            "id": doc_id or str(uuid.uuid4()),
            "title": fake.sentence(nb_words=12),
            "subtitle": fake.sentence(nb_words=8),
            "abstract": fake.text(max_nb_chars=500),
            "content": {
                "introduction": fake.text(max_nb_chars=2000),
                "body": fake.text(max_nb_chars=15000),
                "conclusion": fake.text(max_nb_chars=1000),
                "references": [fake.url() for _ in range(random.randint(10, 30))]
            },
            "authors": [
                {
                    "name": fake.name(),
                    "email": fake.email(),
                    "affiliation": fake.company(),
                    "bio": fake.text(max_nb_chars=300),
                    "expertise": [fake.word() for _ in range(random.randint(3, 8))]
                } for _ in range(random.randint(1, 5))
            ],
            "metadata": {
                "category": random.choice(self.categories),
                "subcategory": fake.word(),
                "tags": [fake.word() for _ in range(random.randint(8, 15))],
                "keywords": [fake.word() for _ in range(random.randint(10, 20))],
                "language": fake.language_code(),
                "word_count": fake.random_int(3000, 8000),
                "reading_time": fake.random_int(15, 45),
                "difficulty": random.choice(['beginner', 'intermediate', 'advanced', 'expert'])
            },
            "publishing": {
                "publisher": fake.company(),
                "journal": fake.catch_phrase(),
                "volume": fake.random_int(1, 50),
                "issue": fake.random_int(1, 12),
                "pages": f"{fake.random_int(1, 100)}-{fake.random_int(101, 200)}",
                "doi": f"10.{fake.random_int(1000, 9999)}/{fake.random_int(100000, 999999)}"
            },
            "timestamps": {
                "submitted": fake.date_time_between(start_date='-2y').isoformat(),
                "reviewed": fake.date_time_between(start_date='-18m').isoformat(),
                "accepted": fake.date_time_between(start_date='-1y').isoformat(),
                "published": fake.date_time_between(start_date='-6m').isoformat(),
                "updated": fake.date_time_between(start_date='-30d').isoformat()
            },
            "metrics": {
                "downloads": fake.random_int(0, 10000),
                "citations": fake.random_int(0, 500),
                "social_shares": fake.random_int(0, 1000),
                "altmetric_score": round(random.uniform(0, 100), 1)
            },
            "versions": [
                {
                    "version": f"v{i+1}.{random.randint(0, 9)}",
                    "date": fake.date_time_between(start_date='-1y').isoformat(),
                    "changes": fake.text(max_nb_chars=200)
                } for i in range(random.randint(1, 5))
            ]
        }
    
    def generate_huge_document(self, doc_id=None):
        """Generate huge document (> 100KB) - suitable for books, comprehensive reports"""
        return {
            "id": doc_id or str(uuid.uuid4()),
            "title": fake.sentence(nb_words=15),
            "subtitle": fake.sentence(nb_words=10),
            "type": "comprehensive_report",
            "abstract": fake.text(max_nb_chars=1000),
            "table_of_contents": [
                {
                    "chapter": i+1,
                    "title": fake.sentence(nb_words=6),
                    "page": fake.random_int(i*20, (i+1)*20),
                    "sections": [
                        {
                            "section": f"{i+1}.{j+1}",
                            "title": fake.sentence(nb_words=4),
                            "page": fake.random_int(i*20+j*3, i*20+(j+1)*3)
                        } for j in range(random.randint(3, 8))
                    ]
                } for i in range(random.randint(8, 15))
            ],
            "content": {
                "executive_summary": fake.text(max_nb_chars=2000),
                "introduction": fake.text(max_nb_chars=3000),
                "methodology": fake.text(max_nb_chars=2500),
                "chapters": [
                    {
                        "number": i+1,
                        "title": fake.sentence(nb_words=6),
                        "content": fake.text(max_nb_chars=8000),
                        "subsections": [
                            {
                                "title": fake.sentence(nb_words=4),
                                "content": fake.text(max_nb_chars=2000)
                            } for _ in range(random.randint(3, 6))
                        ],
                        "figures": [
                            {
                                "id": f"fig_{i+1}_{j+1}",
                                "caption": fake.sentence(nb_words=8),
                                "description": fake.text(max_nb_chars=200)
                            } for j in range(random.randint(2, 5))
                        ],
                        "tables": [
                            {
                                "id": f"table_{i+1}_{j+1}",
                                "caption": fake.sentence(nb_words=6),
                                "rows": fake.random_int(5, 20),
                                "columns": fake.random_int(3, 8)
                            } for j in range(random.randint(1, 3))
                        ]
                    } for i in range(random.randint(6, 12))
                ],
                "conclusion": fake.text(max_nb_chars=2000),
                "recommendations": [fake.sentence(nb_words=12) for _ in range(random.randint(5, 10))],
                "appendices": [
                    {
                        "id": chr(65+i),  # A, B, C, etc.
                        "title": fake.sentence(nb_words=5),
                        "content": fake.text(max_nb_chars=3000)
                    } for i in range(random.randint(2, 5))
                ],
                "glossary": {
                    fake.word(): fake.sentence(nb_words=8) 
                    for _ in range(random.randint(20, 50))
                },
                "bibliography": [
                    {
                        "authors": [fake.name() for _ in range(random.randint(1, 4))],
                        "title": fake.sentence(nb_words=8),
                        "journal": fake.catch_phrase(),
                        "year": fake.random_int(2010, 2023),
                        "volume": fake.random_int(1, 50),
                        "pages": f"{fake.random_int(1, 100)}-{fake.random_int(101, 200)}",
                        "doi": f"10.{fake.random_int(1000, 9999)}/{fake.random_int(100000, 999999)}"
                    } for _ in range(random.randint(50, 150))
                ]
            },
            "metadata": {
                "category": random.choice(self.categories),
                "subcategories": [fake.word() for _ in range(random.randint(3, 8))],
                "tags": [fake.word() for _ in range(random.randint(15, 30))],
                "keywords": [fake.word() for _ in range(random.randint(20, 40))],
                "language": fake.language_code(),
                "word_count": fake.random_int(25000, 100000),
                "page_count": fake.random_int(150, 500),
                "reading_time": fake.random_int(180, 600),  # minutes
                "complexity_score": round(random.uniform(0.3, 1.0), 2)
            },
            "contributors": [
                {
                    "name": fake.name(),
                    "role": random.choice(['lead_author', 'co_author', 'contributor', 'reviewer', 'editor']),
                    "affiliation": fake.company(),
                    "email": fake.email(),
                    "bio": fake.text(max_nb_chars=400),
                    "expertise": [fake.word() for _ in range(random.randint(5, 12))],
                    "publications": fake.random_int(5, 100)
                } for _ in range(random.randint(3, 10))
            ],
            "review_process": {
                "reviewers": [
                    {
                        "name": fake.name(),
                        "affiliation": fake.company(),
                        "expertise": [fake.word() for _ in range(random.randint(3, 8))],
                        "review_date": fake.date_time_between(start_date='-1y').isoformat(),
                        "recommendation": random.choice(['accept', 'minor_revision', 'major_revision', 'reject']),
                        "comments": fake.text(max_nb_chars=500)
                    } for _ in range(random.randint(2, 5))
                ],
                "rounds": fake.random_int(1, 3),
                "total_time": fake.random_int(90, 365)  # days
            },
            "analytics": {
                "downloads": fake.random_int(0, 50000),
                "views": fake.random_int(0, 100000),
                "citations": fake.random_int(0, 1000),
                "social_shares": fake.random_int(0, 5000),
                "altmetric_score": round(random.uniform(0, 200), 1),
                "geographic_distribution": {
                    fake.country(): round(random.uniform(0, 100), 1) 
                    for _ in range(random.randint(10, 25))
                }
            }
        }
    
    def generate_mixed_documents(self, total_count):
        """Generate mixed-size documents for realistic workload testing"""
        documents = []
        
        # 40% small, 35% medium, 20% large, 5% huge
        small_count = int(total_count * 0.4)
        medium_count = int(total_count * 0.35)
        large_count = int(total_count * 0.2)
        huge_count = total_count - small_count - medium_count - large_count
        
        for i in range(small_count):
            documents.append(self.generate_small_document(f"small_{i}"))
        
        for i in range(medium_count):
            documents.append(self.generate_medium_document(f"medium_{i}"))
        
        for i in range(large_count):
            documents.append(self.generate_large_document(f"large_{i}"))
        
        for i in range(huge_count):
            documents.append(self.generate_huge_document(f"huge_{i}"))
        
        # Shuffle to simulate realistic mixed workload
        random.shuffle(documents)
        return documents
    
    def generate_documents(self, doc_type, count):
        """Generate documents of specified type and count"""
        if doc_type == 'small':
            return [self.generate_small_document(f"small_{i}") for i in range(count)]
        elif doc_type == 'medium':
            return [self.generate_medium_document(f"medium_{i}") for i in range(count)]
        elif doc_type == 'large':
            return [self.generate_large_document(f"large_{i}") for i in range(count)]
        elif doc_type == 'huge':
            return [self.generate_huge_document(f"huge_{i}") for i in range(count)]
        elif doc_type == 'mixed':
            return self.generate_mixed_documents(count)
        else:
            raise ValueError(f"Unknown document type: {doc_type}")
    
    def save_as_ndjson(self, documents, filepath):
        """Save documents as NDJSON format"""
        os.makedirs(os.path.dirname(filepath), exist_ok=True)
        with open(filepath, 'w') as f:
            for doc in documents:
                f.write(json.dumps(doc) + '\n')
    
    def save_as_json(self, documents, filepath):
        """Save documents as JSON array"""
        os.makedirs(os.path.dirname(filepath), exist_ok=True)
        with open(filepath, 'w') as f:
            json.dump(documents, f, indent=2)

def main():
    parser = argparse.ArgumentParser(description='Generate documents for Elasticsearch write optimization testing')
    parser.add_argument('--type', choices=['small', 'medium', 'large', 'huge', 'mixed'], 
                       default='mixed', help='Type of documents to generate')
    parser.add_argument('--count', type=int, default=1000, help='Number of documents to generate')
    parser.add_argument('--output', default='datasets/samples/generated-documents.ndjson', 
                       help='Output file path')
    parser.add_argument('--format', choices=['ndjson', 'json'], default='ndjson', 
                       help='Output format')
    parser.add_argument('--seed', type=int, help='Random seed for reproducible generation')
    
    args = parser.parse_args()
    
    if args.seed:
        random.seed(args.seed)
        fake.seed_instance(args.seed)
    
    print(f"🔄 Generating {args.count} {args.type} documents...")
    
    generator = DocumentGenerator()
    documents = generator.generate_documents(args.type, args.count)
    
    if args.format == 'ndjson':
        generator.save_as_ndjson(documents, args.output)
    else:
        generator.save_as_json(documents, args.output)
    
    # Calculate and display statistics
    total_size = sum(len(json.dumps(doc).encode('utf-8')) for doc in documents)
    avg_size = total_size / len(documents)
    
    print(f"✅ Generated {len(documents)} documents")
    print(f"📊 Statistics:")
    print(f"   • Total size: {total_size / 1024 / 1024:.2f} MB")
    print(f"   • Average size: {avg_size / 1024:.2f} KB")
    print(f"   • Output file: {args.output}")
    print(f"   • Format: {args.format.upper()}")
    
    # Provide usage examples
    print(f"\n💡 Usage Examples:")
    print(f"   # Import via curl:")
    print(f"   curl -X POST 'http://localhost:8082/api/v1/indices/test-index/import/ndjson?batch_size=500' \\")
    print(f"     --data-binary @{args.output}")
    print(f"   ")
    print(f"   # Run performance test:")
    print(f"   cd projects/index-explorer && go run cmd/perf-test/main.go")

if __name__ == '__main__':
    main()