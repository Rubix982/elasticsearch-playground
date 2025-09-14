#!/usr/bin/env python3
"""
News & Media Dataset Generator for Elasticsearch Write Optimization Testing

Generates realistic news articles with metadata, comments, and analytics.
"""

import json
import random
import argparse
import uuid
from datetime import datetime, timedelta
from faker import Faker
import os

fake = Faker()

class NewsGenerator:
    def __init__(self):
        self.categories = [
            'Politics', 'Technology', 'Business', 'Sports', 'Entertainment', 'Health',
            'Science', 'World News', 'Local News', 'Opinion', 'Weather', 'Finance'
        ]
        
        self.sources = [
            'Reuters', 'Associated Press', 'BBC', 'CNN', 'Fox News', 'The Guardian',
            'New York Times', 'Washington Post', 'Wall Street Journal', 'USA Today'
        ]
        
        self.authors = [fake.name() for _ in range(50)]  # Pool of authors
        
        self.sentiments = ['positive', 'negative', 'neutral']
        
    def generate_article(self, article_id=None):
        """Generate a single news article with full metadata"""
        category = random.choice(self.categories)
        source = random.choice(self.sources)
        author = random.choice(self.authors)
        
        # Generate comments
        comment_count = random.randint(0, 200)
        comments = [self.generate_comment() for _ in range(min(comment_count, 15))]  # Limit for size
        
        # Generate related articles
        related_articles = [self.generate_related_article() for _ in range(random.randint(2, 6))]
        
        published_date = fake.date_time_between(start_date='-1y')
        
        return {
            "id": article_id or str(uuid.uuid4()),
            "headline": fake.sentence(nb_words=10).rstrip('.'),
            "subheadline": fake.sentence(nb_words=15) if random.random() < 0.7 else None,
            "content": {
                "lead": fake.text(max_nb_chars=300),
                "body": fake.text(max_nb_chars=3000),
                "summary": fake.text(max_nb_chars=200),
                "word_count": random.randint(300, 2000)
            },
            "author": {
                "name": author,
                "email": fake.email(),
                "bio": fake.text(max_nb_chars=150),
                "twitter": f"@{fake.user_name()}",
                "experience_years": random.randint(1, 25),
                "specializations": [fake.word() for _ in range(random.randint(2, 5))]
            },
            "publication": {
                "source": source,
                "section": category,
                "subsection": fake.word().title(),
                "edition": random.choice(['Morning', 'Evening', 'Weekend', 'Online']),
                "page_number": random.randint(1, 50) if random.random() < 0.3 else None
            },
            "timestamps": {
                "created": (published_date - timedelta(hours=random.randint(1, 48))).isoformat(),
                "published": published_date.isoformat(),
                "last_updated": (published_date + timedelta(hours=random.randint(0, 24))).isoformat(),
                "embargo_until": None if random.random() < 0.9 else fake.future_datetime().isoformat()
            },
            "classification": {
                "category": category,
                "tags": [fake.word() for _ in range(random.randint(3, 8))],
                "topics": [fake.catch_phrase() for _ in range(random.randint(2, 5))],
                "urgency": random.choice(['low', 'medium', 'high', 'breaking']),
                "content_type": random.choice(['news', 'analysis', 'opinion', 'feature', 'breaking'])
            },
            "location": {
                "dateline": fake.city(),
                "country": fake.country(),
                "region": fake.state(),
                "coordinates": {
                    "lat": float(fake.latitude()),
                    "lon": float(fake.longitude())
                } if random.random() < 0.6 else None
            },
            "media": {
                "featured_image": {
                    "url": fake.image_url(width=1200, height=800),
                    "caption": fake.sentence(nb_words=8),
                    "credit": fake.name(),
                    "alt_text": fake.sentence(nb_words=6)
                } if random.random() < 0.8 else None,
                "gallery": [
                    {
                        "url": fake.image_url(width=800, height=600),
                        "caption": fake.sentence(nb_words=6),
                        "credit": fake.name()
                    } for _ in range(random.randint(0, 5))
                ] if random.random() < 0.3 else [],
                "video": {
                    "url": fake.url(),
                    "duration": f"{random.randint(1, 10)}:{random.randint(10, 59)}",
                    "thumbnail": fake.image_url()
                } if random.random() < 0.2 else None
            },
            "seo": {
                "meta_title": fake.sentence(nb_words=8).rstrip('.'),
                "meta_description": fake.text(max_nb_chars=160),
                "keywords": [fake.word() for _ in range(random.randint(5, 12))],
                "url_slug": fake.slug(),
                "canonical_url": fake.url()
            },
            "engagement": {
                "views": random.randint(100, 50000),
                "unique_views": random.randint(50, 25000),
                "shares": {
                    "facebook": random.randint(0, 1000),
                    "twitter": random.randint(0, 2000),
                    "linkedin": random.randint(0, 500),
                    "reddit": random.randint(0, 300),
                    "email": random.randint(0, 200)
                },
                "comments_count": comment_count,
                "reactions": {
                    "likes": random.randint(0, 1000),
                    "dislikes": random.randint(0, 100),
                    "love": random.randint(0, 200),
                    "angry": random.randint(0, 150),
                    "sad": random.randint(0, 50)
                },
                "read_time": f"{random.randint(2, 15)} minutes"
            },
            "comments": {
                "total": comment_count,
                "moderated": random.random() < 0.8,
                "recent": comments
            },
            "related_content": {
                "related_articles": related_articles,
                "trending_topics": [fake.catch_phrase() for _ in range(random.randint(3, 7))],
                "recommended_reads": [fake.sentence(nb_words=8) for _ in range(random.randint(2, 5))]
            },
            "analytics": {
                "sentiment": random.choice(self.sentiments),
                "readability_score": round(random.uniform(5.0, 15.0), 1),
                "engagement_score": round(random.uniform(0.1, 10.0), 2),
                "virality_coefficient": round(random.uniform(0.01, 5.0), 3),
                "bounce_rate": round(random.uniform(0.2, 0.8), 2),
                "time_on_page": random.randint(30, 600)  # seconds
            },
            "editorial": {
                "editor": fake.name(),
                "fact_checked": random.random() < 0.85,
                "fact_checker": fake.name() if random.random() < 0.85 else None,
                "editorial_notes": fake.text(max_nb_chars=100) if random.random() < 0.3 else None,
                "corrections": [
                    {
                        "date": fake.date_time_between(start_date=published_date).isoformat(),
                        "correction": fake.sentence(nb_words=10)
                    } for _ in range(random.randint(0, 2))
                ] if random.random() < 0.1 else []
            },
            "status": random.choice(['published', 'draft', 'archived', 'scheduled']),
            "visibility": {
                "public": random.random() < 0.95,
                "paywall": random.random() < 0.2,
                "subscriber_only": random.random() < 0.15,
                "region_restricted": random.random() < 0.05
            }
        }
    
    def generate_comment(self):
        """Generate a single comment"""
        return {
            "id": str(uuid.uuid4()),
            "author": fake.name(),
            "content": fake.text(max_nb_chars=200),
            "timestamp": fake.date_time_between(start_date='-30d').isoformat(),
            "likes": random.randint(0, 50),
            "replies": [
                {
                    "id": str(uuid.uuid4()),
                    "author": fake.name(),
                    "content": fake.text(max_nb_chars=100),
                    "timestamp": fake.date_time_between(start_date='-30d').isoformat(),
                    "likes": random.randint(0, 10)
                } for _ in range(random.randint(0, 3))
            ] if random.random() < 0.3 else [],
            "flagged": random.random() < 0.05,
            "verified_reader": random.random() < 0.7
        }
    
    def generate_related_article(self):
        """Generate a related article reference"""
        return {
            "id": str(uuid.uuid4()),
            "headline": fake.sentence(nb_words=8).rstrip('.'),
            "url": fake.url(),
            "category": random.choice(self.categories),
            "published": fake.date_time_between(start_date='-6m').isoformat(),
            "relevance_score": round(random.uniform(0.1, 1.0), 2)
        }
    
    def generate_news_feed(self, count):
        """Generate a collection of news articles"""
        return [self.generate_article(f"article_{i}") for i in range(count)]
    
    def save_as_ndjson(self, articles, filepath):
        """Save articles as NDJSON format"""
        os.makedirs(os.path.dirname(filepath), exist_ok=True)
        with open(filepath, 'w') as f:
            for article in articles:
                f.write(json.dumps(article) + '\n')
    
    def save_as_json(self, articles, filepath):
        """Save articles as JSON array"""
        os.makedirs(os.path.dirname(filepath), exist_ok=True)
        with open(filepath, 'w') as f:
            json.dump(articles, f, indent=2)

def main():
    parser = argparse.ArgumentParser(description='Generate news and media articles data')
    parser.add_argument('--count', type=int, default=1000, help='Number of articles to generate')
    parser.add_argument('--output', default='datasets/samples/news-articles.ndjson', 
                       help='Output file path')
    parser.add_argument('--format', choices=['ndjson', 'json'], default='ndjson', 
                       help='Output format')
    parser.add_argument('--seed', type=int, help='Random seed for reproducible generation')
    
    args = parser.parse_args()
    
    if args.seed:
        random.seed(args.seed)
        fake.seed_instance(args.seed)
    
    print(f"📰 Generating {args.count} news articles...")
    
    generator = NewsGenerator()
    articles = generator.generate_news_feed(args.count)
    
    if args.format == 'ndjson':
        generator.save_as_ndjson(articles, args.output)
    else:
        generator.save_as_json(articles, args.output)
    
    # Calculate statistics
    total_size = sum(len(json.dumps(article).encode('utf-8')) for article in articles)
    avg_size = total_size / len(articles)
    
    print(f"✅ Generated {len(articles)} articles")
    print(f"📊 Statistics:")
    print(f"   • Total size: {total_size / 1024 / 1024:.2f} MB")
    print(f"   • Average size: {avg_size / 1024:.2f} KB")
    print(f"   • Output file: {args.output}")
    print(f"   • Categories: {len(set(a['classification']['category'] for a in articles))}")
    print(f"   • Sources: {len(set(a['publication']['source'] for a in articles))}")
    
    print(f"\n💡 Usage Examples:")
    print(f"   # Import via Index Explorer:")
    print(f"   curl -X POST 'http://localhost:8082/api/v1/indices/news/import/ndjson' \\")
    print(f"     --data-binary @{args.output}")

if __name__ == '__main__':
    main()