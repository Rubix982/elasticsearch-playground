#!/usr/bin/env python3
"""
Performance Testing Dataset Generator for Elasticsearch Write Optimization

Generates high-volume datasets specifically designed for performance testing
and write optimization benchmarking.
"""

import json
import random
import argparse
import uuid
from datetime import datetime, timedelta
from faker import Faker
import os
import threading
import time

fake = Faker()

class PerformanceGenerator:
    def __init__(self):
        self.batch_templates = {
            'small': self.generate_small_template,
            'medium': self.generate_medium_template,
            'large': self.generate_large_template,
            'huge': self.generate_huge_template
        }
        
        # Pre-generate common values for performance
        self.common_words = [fake.word() for _ in range(1000)]
        self.common_sentences = [fake.sentence() for _ in range(500)]
        self.common_paragraphs = [fake.paragraph() for _ in range(100)]
        self.common_names = [fake.name() for _ in range(200)]
        self.common_companies = [fake.company() for _ in range(100)]
        
    def generate_small_template(self, doc_id):
        """Generate small document template (~200-500 bytes)"""
        return {
            "id": doc_id,
            "timestamp": datetime.now().isoformat(),
            "level": random.choice(['INFO', 'WARN', 'ERROR']),
            "service": random.choice(['auth', 'api', 'web', 'mobile']),
            "message": random.choice(self.common_sentences),
            "user_id": random.randint(1, 100000),
            "session": fake.uuid4()[:8],
            "duration": random.randint(10, 1000),
            "status": random.choice([200, 404, 500])
        }
    
    def generate_medium_template(self, doc_id):
        """Generate medium document template (~2-5KB)"""
        return {
            "id": doc_id,
            "title": random.choice(self.common_sentences),
            "content": random.choice(self.common_paragraphs),
            "author": {
                "name": random.choice(self.common_names),
                "email": fake.email(),
                "company": random.choice(self.common_companies)
            },
            "metadata": {
                "category": random.choice(['tech', 'business', 'science']),
                "tags": random.sample(self.common_words, k=random.randint(3, 8)),
                "priority": random.choice(['low', 'medium', 'high']),
                "views": random.randint(0, 10000),
                "rating": round(random.uniform(1.0, 5.0), 1)
            },
            "timestamps": {
                "created": fake.date_time_between(start_date='-1y').isoformat(),
                "updated": fake.date_time_between(start_date='-30d').isoformat()
            },
            "location": {
                "country": fake.country(),
                "city": fake.city(),
                "coordinates": [float(fake.longitude()), float(fake.latitude())]
            },
            "stats": {
                "word_count": random.randint(100, 1000),
                "read_time": random.randint(60, 600),
                "engagement": round(random.uniform(0.1, 1.0), 2)
            }
        }
    
    def generate_large_template(self, doc_id):
        """Generate large document template (~10-50KB)"""
        return {
            "id": doc_id,
            "title": random.choice(self.common_sentences),
            "subtitle": random.choice(self.common_sentences),
            "content": {
                "introduction": ' '.join(random.sample(self.common_paragraphs, k=3)),
                "body": ' '.join(random.sample(self.common_paragraphs, k=8)),
                "conclusion": ' '.join(random.sample(self.common_paragraphs, k=2))
            },
            "authors": [
                {
                    "name": random.choice(self.common_names),
                    "email": fake.email(),
                    "bio": random.choice(self.common_paragraphs),
                    "expertise": random.sample(self.common_words, k=random.randint(3, 8))
                } for _ in range(random.randint(1, 4))
            ],
            "metadata": {
                "category": random.choice(['research', 'analysis', 'report']),
                "subcategory": random.choice(self.common_words),
                "tags": random.sample(self.common_words, k=random.randint(8, 15)),
                "keywords": random.sample(self.common_words, k=random.randint(10, 20)),
                "language": 'en',
                "word_count": random.randint(3000, 8000),
                "read_time": random.randint(900, 2400)
            },
            "publishing": {
                "publisher": random.choice(self.common_companies),
                "journal": random.choice(self.common_sentences),
                "volume": random.randint(1, 50),
                "issue": random.randint(1, 12),
                "pages": f"{random.randint(1, 100)}-{random.randint(101, 200)}"
            },
            "timestamps": {
                "submitted": fake.date_time_between(start_date='-2y').isoformat(),
                "accepted": fake.date_time_between(start_date='-1y').isoformat(),
                "published": fake.date_time_between(start_date='-6m').isoformat()
            },
            "metrics": {
                "downloads": random.randint(0, 10000),
                "citations": random.randint(0, 500),
                "shares": random.randint(0, 1000),
                "views": random.randint(0, 50000)
            },
            "references": [
                {
                    "title": random.choice(self.common_sentences),
                    "authors": random.sample(self.common_names, k=random.randint(1, 3)),
                    "year": random.randint(2010, 2023),
                    "journal": random.choice(self.common_companies)
                } for _ in range(random.randint(20, 50))
            ]
        }
    
    def generate_huge_template(self, doc_id):
        """Generate huge document template (~100KB+)"""
        return {
            "id": doc_id,
            "title": random.choice(self.common_sentences),
            "type": "comprehensive_document",
            "abstract": ' '.join(random.sample(self.common_paragraphs, k=2)),
            "table_of_contents": [
                {
                    "chapter": i+1,
                    "title": random.choice(self.common_sentences),
                    "page": random.randint(i*20, (i+1)*20),
                    "sections": [
                        {
                            "section": f"{i+1}.{j+1}",
                            "title": random.choice(self.common_sentences),
                            "page": random.randint(i*20+j*3, i*20+(j+1)*3)
                        } for j in range(random.randint(4, 8))
                    ]
                } for i in range(random.randint(10, 15))
            ],
            "content": {
                "executive_summary": ' '.join(random.sample(self.common_paragraphs, k=5)),
                "introduction": ' '.join(random.sample(self.common_paragraphs, k=8)),
                "methodology": ' '.join(random.sample(self.common_paragraphs, k=6)),
                "chapters": [
                    {
                        "number": i+1,
                        "title": random.choice(self.common_sentences),
                        "content": ' '.join(random.sample(self.common_paragraphs, k=15)),
                        "subsections": [
                            {
                                "title": random.choice(self.common_sentences),
                                "content": ' '.join(random.sample(self.common_paragraphs, k=5))
                            } for _ in range(random.randint(4, 8))
                        ]
                    } for i in range(random.randint(8, 12))
                ],
                "conclusion": ' '.join(random.sample(self.common_paragraphs, k=4)),
                "recommendations": [random.choice(self.common_sentences) for _ in range(random.randint(8, 15))],
                "appendices": [
                    {
                        "id": chr(65+i),
                        "title": random.choice(self.common_sentences),
                        "content": ' '.join(random.sample(self.common_paragraphs, k=8))
                    } for i in range(random.randint(3, 6))
                ]
            },
            "contributors": [
                {
                    "name": random.choice(self.common_names),
                    "role": random.choice(['lead_author', 'co_author', 'contributor', 'reviewer']),
                    "affiliation": random.choice(self.common_companies),
                    "bio": random.choice(self.common_paragraphs),
                    "expertise": random.sample(self.common_words, k=random.randint(5, 10))
                } for _ in range(random.randint(5, 10))
            ],
            "metadata": {
                "word_count": random.randint(50000, 150000),
                "page_count": random.randint(200, 800),
                "language": 'en',
                "version": f"{random.randint(1, 3)}.{random.randint(0, 9)}",
                "keywords": random.sample(self.common_words, k=random.randint(25, 50))
            },
            "bibliography": [
                {
                    "authors": random.sample(self.common_names, k=random.randint(1, 4)),
                    "title": random.choice(self.common_sentences),
                    "journal": random.choice(self.common_companies),
                    "year": random.randint(2010, 2023),
                    "volume": random.randint(1, 50)
                } for _ in range(random.randint(100, 200))
            ]
        }
    
    def generate_batch(self, doc_type, start_id, count):
        """Generate a batch of documents of specified type"""
        template_func = self.batch_templates[doc_type]
        documents = []
        
        for i in range(count):
            doc_id = f"{doc_type}_{start_id + i}"
            documents.append(template_func(doc_id))
        
        return documents
    
    def generate_performance_dataset(self, total_count, doc_type='mixed', batch_size=1000):
        """Generate high-performance dataset with batching"""
        print(f"🚧 Generating {total_count} {doc_type} documents in batches of {batch_size}...")
        
        documents = []
        batches_processed = 0
        
        if doc_type == 'mixed':
            # Mixed distribution for realistic performance testing
            type_distribution = {
                'small': int(total_count * 0.4),   # 40%
                'medium': int(total_count * 0.35), # 35%
                'large': int(total_count * 0.20),  # 20%
                'huge': int(total_count * 0.05)    # 5%
            }
            
            current_id = 0
            for dtype, count in type_distribution.items():
                batches = (count + batch_size - 1) // batch_size
                for batch_num in range(batches):
                    batch_count = min(batch_size, count - batch_num * batch_size)
                    if batch_count > 0:
                        batch_docs = self.generate_batch(dtype, current_id, batch_count)
                        documents.extend(batch_docs)
                        current_id += batch_count
                        batches_processed += 1
                        
                        if batches_processed % 10 == 0:
                            print(f"   📦 Processed {batches_processed} batches, {len(documents)} documents...")
            
            # Shuffle for realistic mixed workload
            random.shuffle(documents)
        else:
            # Single document type
            batches = (total_count + batch_size - 1) // batch_size
            current_id = 0
            
            for batch_num in range(batches):
                batch_count = min(batch_size, total_count - batch_num * batch_size)
                if batch_count > 0:
                    batch_docs = self.generate_batch(doc_type, current_id, batch_count)
                    documents.extend(batch_docs)
                    current_id += batch_count
                    batches_processed += 1
                    
                    if batches_processed % 10 == 0:
                        print(f"   📦 Processed {batches_processed} batches, {len(documents)} documents...")
        
        return documents
    
    def save_as_ndjson_streaming(self, documents, filepath, chunk_size=1000):
        """Save documents as NDJSON with streaming for memory efficiency"""
        os.makedirs(os.path.dirname(filepath), exist_ok=True)
        
        with open(filepath, 'w') as f:
            for i, doc in enumerate(documents):
                f.write(json.dumps(doc) + '\n')
                
                # Flush periodically for large files
                if (i + 1) % chunk_size == 0:
                    f.flush()
                    print(f"   💾 Written {i + 1} documents...")
    
    def generate_time_series_data(self, count, doc_type='small'):
        """Generate time-series data for performance testing"""
        print(f"📈 Generating {count} time-series {doc_type} documents...")
        
        documents = []
        start_time = datetime.now() - timedelta(days=7)
        
        for i in range(count):
            timestamp = start_time + timedelta(seconds=i * 60)  # 1 minute intervals
            
            base_doc = self.batch_templates[doc_type](f"ts_{i}")
            base_doc['timestamp'] = timestamp.isoformat()
            base_doc['sequence_id'] = i
            base_doc['time_bucket'] = timestamp.strftime('%Y-%m-%d-%H')
            
            # Add time-series specific fields
            base_doc['metrics'] = {
                'value': random.uniform(0, 100),
                'trend': random.choice(['up', 'down', 'stable']),
                'anomaly_score': random.uniform(0, 1)
            }
            
            documents.append(base_doc)
        
        return documents

def main():
    parser = argparse.ArgumentParser(description='Generate high-performance datasets for write optimization testing')
    parser.add_argument('--count', type=int, default=10000, help='Number of documents to generate')
    parser.add_argument('--type', choices=['small', 'medium', 'large', 'huge', 'mixed', 'timeseries'], 
                       default='mixed', help='Type of documents to generate')
    parser.add_argument('--output', default='datasets/samples/performance-test.ndjson', 
                       help='Output file path')
    parser.add_argument('--batch-size', type=int, default=1000, 
                       help='Batch size for processing (affects memory usage)')
    parser.add_argument('--seed', type=int, help='Random seed for reproducible generation')
    parser.add_argument('--streaming', action='store_true', 
                       help='Use streaming save for large datasets (memory efficient)')
    
    args = parser.parse_args()
    
    if args.seed:
        random.seed(args.seed)
        fake.seed_instance(args.seed)
    
    print(f"⚡ Performance Dataset Generator")
    print(f"   • Documents: {args.count:,}")
    print(f"   • Type: {args.type}")
    print(f"   • Batch size: {args.batch_size}")
    print(f"   • Output: {args.output}")
    print()
    
    generator = PerformanceGenerator()
    start_time = time.time()
    
    if args.type == 'timeseries':
        documents = generator.generate_time_series_data(args.count, 'small')
    else:
        documents = generator.generate_performance_dataset(args.count, args.type, args.batch_size)
    
    generation_time = time.time() - start_time
    
    print(f"✅ Generated {len(documents):,} documents in {generation_time:.2f} seconds")
    print(f"   📊 Generation rate: {len(documents) / generation_time:.0f} docs/sec")
    
    # Save documents
    print(f"💾 Saving to {args.output}...")
    save_start = time.time()
    
    if args.streaming or len(documents) > 50000:
        generator.save_as_ndjson_streaming(documents, args.output)
    else:
        os.makedirs(os.path.dirname(args.output), exist_ok=True)
        with open(args.output, 'w') as f:
            for doc in documents:
                f.write(json.dumps(doc) + '\n')
    
    save_time = time.time() - save_start
    
    # Calculate final statistics
    total_size = os.path.getsize(args.output)
    avg_size = total_size / len(documents)
    
    print(f"✅ Dataset created successfully!")
    print(f"📊 Final Statistics:")
    print(f"   • Documents: {len(documents):,}")
    print(f"   • Total size: {total_size / 1024 / 1024:.2f} MB")
    print(f"   • Average size: {avg_size / 1024:.2f} KB")
    print(f"   • Generation time: {generation_time:.2f}s ({len(documents) / generation_time:.0f} docs/sec)")
    print(f"   • Save time: {save_time:.2f}s ({len(documents) / save_time:.0f} docs/sec)")
    print(f"   • Total time: {generation_time + save_time:.2f}s")
    
    if args.type == 'mixed':
        print(f"   • Distribution: 40% small, 35% medium, 20% large, 5% huge")
    
    print(f"\n💡 Performance Testing Usage:")
    print(f"   # Quick test (local):")
    print(f"   curl -X POST 'http://localhost:8082/api/v1/indices/perf-test/import/ndjson?batch_size=1000' \\")
    print(f"     --data-binary @{args.output}")
    print(f"   ")
    print(f"   # Benchmark with Index Explorer:")
    print(f"   cd projects/index-explorer && go run cmd/perf-test/main.go")

if __name__ == '__main__':
    main()