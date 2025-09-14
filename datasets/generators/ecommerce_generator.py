#!/usr/bin/env python3
"""
E-commerce Dataset Generator for Elasticsearch Write Optimization Testing

Generates realistic product catalog data with reviews and inventory information.
"""

import json
import random
import argparse
import uuid
from datetime import datetime, timedelta
from faker import Faker
import os

fake = Faker()

class EcommerceGenerator:
    def __init__(self):
        self.categories = [
            'Electronics', 'Clothing', 'Home & Garden', 'Sports', 'Books', 'Toys',
            'Beauty', 'Automotive', 'Health', 'Office', 'Jewelry', 'Shoes'
        ]
        
        self.brands = [
            'Apple', 'Samsung', 'Nike', 'Adidas', 'Sony', 'Canon', 'Dell', 'HP',
            'Microsoft', 'Google', 'Amazon', 'Zara', 'H&M', 'IKEA', 'Toyota'
        ]
        
        self.colors = [
            'Black', 'White', 'Red', 'Blue', 'Green', 'Yellow', 'Purple', 'Orange',
            'Pink', 'Brown', 'Gray', 'Silver', 'Gold', 'Navy', 'Beige'
        ]
        
        self.sizes = ['XS', 'S', 'M', 'L', 'XL', 'XXL', '6', '7', '8', '9', '10', '11', '12']
    
    def generate_product(self, product_id=None):
        """Generate a single product with full details"""
        category = random.choice(self.categories)
        brand = random.choice(self.brands)
        
        # Generate specifications based on category
        specs = self.generate_specifications(category)
        
        # Generate reviews
        review_count = random.randint(0, 500)
        reviews = [self.generate_review() for _ in range(min(review_count, 10))]  # Limit to 10 for size
        
        return {
            "id": product_id or str(uuid.uuid4()),
            "sku": f"{brand[:3].upper()}-{fake.random_int(100000, 999999)}",
            "name": f"{brand} {fake.catch_phrase()}",
            "description": fake.text(max_nb_chars=800),
            "short_description": fake.sentence(nb_words=12),
            "category": {
                "primary": category,
                "subcategory": fake.word().title(),
                "path": f"{category} > {fake.word().title()}"
            },
            "brand": brand,
            "price": {
                "current": round(random.uniform(9.99, 999.99), 2),
                "original": round(random.uniform(19.99, 1199.99), 2),
                "currency": "USD",
                "discount_percentage": random.randint(0, 50) if random.random() < 0.3 else 0
            },
            "inventory": {
                "stock_quantity": random.randint(0, 1000),
                "warehouse_location": fake.city(),
                "restock_date": fake.date_between(start_date='+1d', end_date='+30d').isoformat() if random.random() < 0.2 else None,
                "low_stock_threshold": random.randint(5, 50)
            },
            "attributes": {
                "color": random.choice(self.colors) if random.random() < 0.7 else None,
                "size": random.choice(self.sizes) if random.random() < 0.5 else None,
                "weight": f"{round(random.uniform(0.1, 50.0), 2)} lbs",
                "dimensions": {
                    "length": round(random.uniform(1, 100), 1),
                    "width": round(random.uniform(1, 100), 1),
                    "height": round(random.uniform(1, 100), 1),
                    "unit": "cm"
                },
                "material": fake.word().title() if random.random() < 0.6 else None
            },
            "specifications": specs,
            "images": [
                {
                    "url": fake.image_url(width=800, height=600),
                    "alt_text": f"{brand} product image",
                    "is_primary": i == 0
                } for i in range(random.randint(1, 6))
            ],
            "seo": {
                "meta_title": f"{brand} {fake.catch_phrase()} - Best Price",
                "meta_description": fake.text(max_nb_chars=160),
                "keywords": [fake.word() for _ in range(random.randint(5, 15))],
                "url_slug": fake.slug()
            },
            "reviews": {
                "average_rating": round(random.uniform(1.0, 5.0), 1),
                "total_reviews": review_count,
                "rating_distribution": {
                    "5_star": random.randint(0, review_count),
                    "4_star": random.randint(0, review_count),
                    "3_star": random.randint(0, review_count),
                    "2_star": random.randint(0, review_count),
                    "1_star": random.randint(0, review_count)
                },
                "recent_reviews": reviews
            },
            "shipping": {
                "free_shipping": random.random() < 0.4,
                "weight_class": random.choice(['light', 'medium', 'heavy']),
                "shipping_cost": round(random.uniform(2.99, 29.99), 2) if random.random() < 0.6 else 0,
                "estimated_delivery": f"{random.randint(1, 7)}-{random.randint(5, 14)} business days"
            },
            "vendor": {
                "id": str(uuid.uuid4()),
                "name": fake.company(),
                "rating": round(random.uniform(3.0, 5.0), 1),
                "location": fake.city(),
                "years_selling": random.randint(1, 20)
            },
            "tags": [fake.word() for _ in range(random.randint(3, 10))],
            "status": random.choice(['active', 'inactive', 'discontinued', 'pre_order']),
            "created_at": fake.date_time_between(start_date='-2y').isoformat(),
            "updated_at": fake.date_time_between(start_date='-30d').isoformat(),
            "analytics": {
                "page_views": random.randint(0, 10000),
                "unique_visitors": random.randint(0, 5000),
                "conversion_rate": round(random.uniform(0.01, 0.15), 3),
                "bounce_rate": round(random.uniform(0.2, 0.8), 2),
                "add_to_cart_rate": round(random.uniform(0.05, 0.3), 3)
            }
        }
    
    def generate_specifications(self, category):
        """Generate category-specific specifications"""
        base_specs = {
            "model_year": random.randint(2020, 2024),
            "warranty": f"{random.randint(1, 5)} years",
            "country_of_origin": fake.country()
        }
        
        if category == 'Electronics':
            base_specs.update({
                "screen_size": f"{random.randint(10, 65)} inches",
                "resolution": random.choice(['1080p', '4K', '8K']),
                "connectivity": random.choice(['WiFi', 'Bluetooth', 'WiFi + Bluetooth']),
                "power_consumption": f"{random.randint(50, 500)}W"
            })
        elif category == 'Clothing':
            base_specs.update({
                "fabric": random.choice(['Cotton', 'Polyester', 'Silk', 'Wool', 'Linen']),
                "care_instructions": "Machine wash cold",
                "fit_type": random.choice(['Regular', 'Slim', 'Loose', 'Athletic']),
                "season": random.choice(['Spring', 'Summer', 'Fall', 'Winter', 'All Season'])
            })
        elif category == 'Home & Garden':
            base_specs.update({
                "room_type": random.choice(['Living Room', 'Bedroom', 'Kitchen', 'Bathroom', 'Garden']),
                "assembly_required": random.choice([True, False]),
                "care_level": random.choice(['Low', 'Medium', 'High']),
                "indoor_outdoor": random.choice(['Indoor', 'Outdoor', 'Both'])
            })
        
        return base_specs
    
    def generate_review(self):
        """Generate a single product review"""
        return {
            "id": str(uuid.uuid4()),
            "customer_name": fake.name(),
            "rating": random.randint(1, 5),
            "title": fake.sentence(nb_words=6),
            "content": fake.text(max_nb_chars=300),
            "verified_purchase": random.random() < 0.8,
            "helpful_votes": random.randint(0, 50),
            "date": fake.date_between(start_date='-1y').isoformat(),
            "photos": [fake.image_url() for _ in range(random.randint(0, 3))] if random.random() < 0.2 else []
        }
    
    def generate_catalog(self, count):
        """Generate a complete product catalog"""
        return [self.generate_product(f"prod_{i}") for i in range(count)]
    
    def save_as_ndjson(self, products, filepath):
        """Save products as NDJSON format"""
        os.makedirs(os.path.dirname(filepath), exist_ok=True)
        with open(filepath, 'w') as f:
            for product in products:
                f.write(json.dumps(product) + '\n')
    
    def save_as_json(self, products, filepath):
        """Save products as JSON array"""
        os.makedirs(os.path.dirname(filepath), exist_ok=True)
        with open(filepath, 'w') as f:
            json.dump(products, f, indent=2)

def main():
    parser = argparse.ArgumentParser(description='Generate e-commerce product catalog data')
    parser.add_argument('--count', type=int, default=1000, help='Number of products to generate')
    parser.add_argument('--output', default='datasets/samples/ecommerce-catalog.ndjson', 
                       help='Output file path')
    parser.add_argument('--format', choices=['ndjson', 'json'], default='ndjson', 
                       help='Output format')
    parser.add_argument('--seed', type=int, help='Random seed for reproducible generation')
    
    args = parser.parse_args()
    
    if args.seed:
        random.seed(args.seed)
        fake.seed_instance(args.seed)
    
    print(f"🛒 Generating {args.count} e-commerce products...")
    
    generator = EcommerceGenerator()
    products = generator.generate_catalog(args.count)
    
    if args.format == 'ndjson':
        generator.save_as_ndjson(products, args.output)
    else:
        generator.save_as_json(products, args.output)
    
    # Calculate statistics
    total_size = sum(len(json.dumps(product).encode('utf-8')) for product in products)
    avg_size = total_size / len(products)
    
    print(f"✅ Generated {len(products)} products")
    print(f"📊 Statistics:")
    print(f"   • Total size: {total_size / 1024 / 1024:.2f} MB")
    print(f"   • Average size: {avg_size / 1024:.2f} KB")
    print(f"   • Output file: {args.output}")
    print(f"   • Categories: {len(set(p['category']['primary'] for p in products))}")
    print(f"   • Brands: {len(set(p['brand'] for p in products))}")
    
    print(f"\n💡 Usage Examples:")
    print(f"   # Import via Index Explorer:")
    print(f"   curl -X POST 'http://localhost:8082/api/v1/indices/ecommerce/import/ndjson' \\")
    print(f"     --data-binary @{args.output}")

if __name__ == '__main__':
    main()