# Elasticsearch Painless & Query Cookbook

## Table of Contents

1. [Painless Scripting Fundamentals](#painless-scripting-fundamentals)
2. [Common Painless Patterns](#common-painless-patterns)
3. [Fuzzy Search Guide](#fuzzy-search-guide)
4. [Query Structure Patterns](#query-structure-patterns)
5. [Real-World Examples](#real-world-examples)
6. [Performance Tips](#performance-tips)
7. [Troubleshooting Guide](#troubleshooting-guide)

---

## Painless Scripting Fundamentals

### Basic Data Types and Variables

```painless
// Strings
String name = "John Doe";
String email = doc['email'].value;

// Numbers
int age = 25;
double price = 99.99;
long timestamp = System.currentTimeMillis();

// Booleans
boolean isActive = true;
boolean hasPermission = doc['permissions'].size() > 0;

// Collections
List<String> tags = ["urgent", "important"];
Map<String, Object> data = ["key": "value"];

// Arrays from document fields
def categories = doc['categories'];
def scores = doc['scores'];
```

### Accessing Document Fields

```painless
// In script queries (read-only)
doc['field_name'].value          // Single value field
doc['field_name'].values         // Multi-value field
doc['field_name'].size()         // Check if field has values

// In update/ingest scripts (read-write)
ctx._source.field_name           // Access source field
ctx._source['field_name']        // Alternative syntax
ctx._source.nested.field         // Nested field access
```

### Null Safety Patterns

```painless
// Check if field exists and has values
if (doc['optional_field'].size() > 0) {
    return doc['optional_field'].value;
}

// Safe null check for source fields
if (ctx._source.field != null) {
    // Process field
}

// Default value pattern
def value = doc['score'].size() > 0 ? doc['score'].value : 0;

// Elvis operator alternative
def result = ctx._source.status ?: 'unknown';
```

---

## Common Painless Patterns

### String Operations

```painless
// Basic string operations
String text = doc['description'].value;
text.length()                    // Length
text.toLowerCase()               // To lowercase
text.toUpperCase()               // To uppercase
text.contains('search')          // Contains substring
text.startsWith('prefix')        // Starts with
text.endsWith('suffix')          // Ends with
text.substring(0, 5)            // Substring
text.indexOf('word')            // Find index
text.split(' ')                 // Split to array

// String formatting
"User: ${doc['name'].value}, Age: ${doc['age'].value}"

// Regular expressions
text.matches('\\d{3}-\\d{2}-\\d{4}')  // Pattern matching
```

### Date and Time Operations

```painless
// Current time
long now = System.currentTimeMillis();
ZonedDateTime nowZoned = ZonedDateTime.now();

// Document date fields
ZonedDateTime docDate = doc['timestamp'].value;
long docMillis = doc['timestamp'].value.millis;

// Date comparisons
doc['created'].value.isAfter(ZonedDateTime.now().minusDays(7))
doc['created'].value.isBefore(ZonedDateTime.now())

// Date arithmetic
docDate.plusDays(30)
docDate.minusHours(2)
docDate.withHour(0)

// Format dates
doc['timestamp'].value.toString()
doc['timestamp'].value.format(DateTimeFormatter.ofPattern('yyyy-MM-dd'))
```

### Mathematical Operations

```painless
// Basic math
Math.max(doc['score1'].value, doc['score2'].value)
Math.min(doc['price'].value, 100.0)
Math.abs(doc['difference'].value)
Math.round(doc['rating'].value * 100.0) / 100.0

// Statistical operations
def scores = doc['scores'];
double sum = 0;
for (score in scores) {
    sum += score;
}
double average = sum / scores.size();

// Logarithmic and exponential
Math.log(doc['value'].value)
Math.exp(doc['exponent'].value)
Math.pow(doc['base'].value, 2)
Math.sqrt(doc['number'].value)
```

### Collection Operations

```painless
// List operations
List items = doc['tags'];
items.size()                     // Size
items.contains('important')      // Contains check
items.get(0)                    // Get by index
items.isEmpty()                 // Is empty check

// Iteration
for (String tag : doc['tags']) {
    if (tag.startsWith('prio-')) {
        return tag;
    }
}

// Stream-like operations (using loops)
def filtered = [];
for (item in doc['items']) {
    if (item.contains('filter')) {
        filtered.add(item);
    }
}

// Map operations
Map data = ctx._source.metadata;
data.containsKey('status')       // Check if key exists
data.get('status')              // Get value
data.keySet()                   // Get all keys
data.values()                   // Get all values
```

### Conditional Logic Patterns

```painless
// Simple if-else
if (doc['status'].value == 'active') {
    return doc['priority_score'].value;
} else if (doc['status'].value == 'pending') {
    return doc['priority_score'].value * 0.5;
} else {
    return 0;
}

// Ternary operator
def score = doc['premium'].value ? doc['base_score'].value * 2 : doc['base_score'].value;

// Switch-like pattern
def status = doc['status'].value;
if (status == 'gold') return 100;
if (status == 'silver') return 75;
if (status == 'bronze') return 50;
return 25;

// Multi-condition checks
if (doc['age'].value >= 18 && doc['verified'].value && doc['active'].value) {
    // Process adult verified active user
}
```

---

## Fuzzy Search Guide

### Basic Fuzzy Queries

#### Simple Fuzzy Query

```json
{
  "query": {
    "fuzzy": {
      "name": {
        "value": "john",
        "fuzziness": "AUTO"
      }
    }
  }
}
```

#### Fuzzy with Filters

```json
{
  "query": {
    "bool": {
      "must": [
        {
          "fuzzy": {
            "product_name": {
              "value": "iphone",
              "fuzziness": "AUTO"
            }
          }
        }
      ],
      "filter": [
        {
          "term": { "status": "active" }
        },
        {
          "range": { "price": { "gte": 100 } }
        }
      ]
    }
  }
}
```

### Match Query with Fuzziness (Recommended)

```json
{
  "query": {
    "bool": {
      "must": [
        {
          "match": {
            "title": {
              "query": "elasticsarch tutorial",
              "fuzziness": "AUTO",
              "operator": "and",
              "minimum_should_match": "75%"
            }
          }
        }
      ],
      "filter": [
        { "term": { "published": true } },
        { "range": { "created_date": { "gte": "2023-01-01" } } }
      ]
    }
  }
}
```

### Multi-Field Fuzzy Search

```json
{
  "query": {
    "bool": {
      "must": [
        {
          "multi_match": {
            "query": "johm smith",
            "fields": ["first_name^2", "last_name^2", "email"],
            "fuzziness": "AUTO",
            "type": "best_fields",
            "tie_breaker": 0.3
          }
        }
      ],
      "filter": [{ "term": { "department": "engineering" } }]
    }
  }
}
```

### Advanced Fuzzy Parameters

```json
{
  "query": {
    "fuzzy": {
      "product_name": {
        "value": "macbook",
        "fuzziness": 2, // Max 2 character changes
        "max_expansions": 50, // Limit term expansions
        "prefix_length": 2, // First 2 chars must match exactly
        "transpositions": true, // Allow character swaps
        "rewrite": "constant_score" // Scoring method
      }
    }
  }
}
```

### Fuzziness Options Reference

```json
{
  "fuzziness": "AUTO", // Automatic based on term length
  "fuzziness": 0, // No fuzziness (exact match)
  "fuzziness": 1, // Allow 1 character difference
  "fuzziness": 2, // Allow 2 character differences
  "fuzziness": "AUTO:3,6" // Custom AUTO settings
}
```

---

## Query Structure Patterns

### Bool Query Template

```json
{
  "query": {
    "bool": {
      "must": [
        // Documents MUST match these queries (affects score)
        { "match": { "field": "value" } }
      ],
      "filter": [
        // Documents MUST match these queries (no scoring)
        { "term": { "status": "published" } },
        { "range": { "date": { "gte": "2023-01-01" } } }
      ],
      "should": [
        // Documents SHOULD match these queries (boosts score)
        { "match": { "category": "preferred" } },
        { "term": { "featured": true } }
      ],
      "must_not": [
        // Documents MUST NOT match these queries
        { "term": { "hidden": true } }
      ],
      "minimum_should_match": 1 // At least 1 should clause must match
    }
  }
}
```

### Common Filter Types

```json
{
  "filter": [
    // Exact term match
    { "term": { "status": "active" } },

    // Multiple exact values
    { "terms": { "category": ["tech", "science", "education"] } },

    // Range queries
    { "range": { "age": { "gte": 18, "lte": 65 } } },
    { "range": { "price": { "gt": 10, "lt": 100 } } },
    { "range": { "date": { "gte": "now-7d", "lte": "now" } } },

    // Field existence
    { "exists": { "field": "email" } },

    // Missing field (opposite of exists)
    { "bool": { "must_not": { "exists": { "field": "deleted_at" } } } },

    // Prefix matching
    { "prefix": { "product_code": "PROD-" } },

    // Wildcard matching
    { "wildcard": { "filename": "*.pdf" } },

    // Regular expression
    { "regexp": { "phone": "\\d{3}-\\d{3}-\\d{4}" } },

    // Nested object queries
    {
      "nested": {
        "path": "comments",
        "query": { "match": { "comments.text": "great product" } }
      }
    }
  ]
}
```

### Script Query Patterns

```json
{
  "query": {
    "bool": {
      "filter": [
        {
          "script": {
            "script": {
              "source": "doc['field1'].value * params.multiplier > doc['field2'].value",
              "params": { "multiplier": 1.5 }
            }
          }
        }
      ]
    }
  }
}
```

---

## Real-World Examples

### E-commerce Product Search

```json
{
  "size": 20,
  "from": 0,
  "query": {
    "bool": {
      "must": [
        {
          "multi_match": {
            "query": "wireless headphones",
            "fields": ["name^3", "description^2", "brand", "category"],
            "fuzziness": "AUTO",
            "type": "best_fields",
            "tie_breaker": 0.3
          }
        }
      ],
      "filter": [
        { "term": { "status": "available" } },
        { "range": { "price": { "gte": 20, "lte": 500 } } },
        { "terms": { "brand": ["sony", "bose", "apple", "samsung"] } },
        { "range": { "rating": { "gte": 4.0 } } }
      ],
      "should": [
        { "term": { "featured": { "value": true, "boost": 2.0 } } },
        { "term": { "on_sale": { "value": true, "boost": 1.5 } } }
      ]
    }
  },
  "sort": [
    { "_score": { "order": "desc" } },
    { "popularity": { "order": "desc" } },
    { "price": { "order": "asc" } }
  ],
  "aggs": {
    "brands": {
      "terms": { "field": "brand", "size": 10 }
    },
    "price_ranges": {
      "range": {
        "field": "price",
        "ranges": [
          { "to": 50 },
          { "from": 50, "to": 100 },
          { "from": 100, "to": 200 },
          { "from": 200 }
        ]
      }
    }
  }
}
```

### User Search with Typo Tolerance

```json
{
  "query": {
    "bool": {
      "should": [
        {
          "match": {
            "full_name": {
              "query": "john doe",
              "fuzziness": "AUTO",
              "boost": 3.0
            }
          }
        },
        {
          "match": {
            "email": {
              "query": "john.doe@company.com",
              "fuzziness": 1,
              "boost": 2.0
            }
          }
        },
        {
          "match": {
            "username": {
              "query": "jdoe",
              "boost": 1.5
            }
          }
        }
      ],
      "filter": [
        { "term": { "active": true } },
        { "range": { "last_login": { "gte": "now-90d" } } },
        { "terms": { "role": ["admin", "user", "moderator"] } }
      ],
      "minimum_should_match": 1
    }
  }
}
```

### Content Search with Relevance Scoring

```json
{
  "query": {
    "function_score": {
      "query": {
        "bool": {
          "must": [
            {
              "multi_match": {
                "query": "elasticsearch tutorial",
                "fields": ["title^4", "content^2", "tags^3"],
                "fuzziness": "AUTO",
                "minimum_should_match": "75%"
              }
            }
          ],
          "filter": [
            { "term": { "published": true } },
            { "range": { "publish_date": { "gte": "now-2y" } } }
          ]
        }
      },
      "functions": [
        {
          "filter": { "term": { "featured": true } },
          "weight": 2.0
        },
        {
          "field_value_factor": {
            "field": "view_count",
            "factor": 0.1,
            "modifier": "log1p",
            "missing": 1
          }
        },
        {
          "gauss": {
            "publish_date": {
              "origin": "now",
              "scale": "30d",
              "decay": 0.5
            }
          }
        }
      ],
      "score_mode": "multiply",
      "boost_mode": "multiply"
    }
  }
}
```

### Complex Update Script

```json
{
  "script": {
    "source": """
      // Update status based on multiple conditions
      if (ctx._source.status == 'pending') {
        ctx._source.status = 'processing';
        ctx._source.processing_started = System.currentTimeMillis();

        // Initialize counters if they don't exist
        if (ctx._source.attempt_count == null) {
          ctx._source.attempt_count = 0;
        }
        ctx._source.attempt_count++;

        // Set priority based on age
        long age = System.currentTimeMillis() - ctx._source.created_timestamp;
        if (age > params.urgent_threshold) {
          ctx._source.priority = 'high';
        } else if (age > params.normal_threshold) {
          ctx._source.priority = 'normal';
        } else {
          ctx._source.priority = 'low';
        }

        // Add to processing queue if not already there
        if (!ctx._source.queues.contains('processing')) {
          ctx._source.queues.add('processing');
        }
      } else {
        // Skip update if not in pending status
        ctx.op = 'none';
      }
    """,
    "params": {
      "urgent_threshold": 86400000,  // 24 hours in milliseconds
      "normal_threshold": 3600000    // 1 hour in milliseconds
    }
  }
}
```

### Advanced Aggregation with Scripts

```json
{
  "size": 0,
  "aggs": {
    "price_stats_by_category": {
      "terms": {"field": "category"},
      "aggs": {
        "price_stats": {
          "stats": {"field": "price"}
        },
        "discounted_average": {
          "avg": {
            "script": {
              "source": "doc['price'].value * (1 - doc['discount_percent'].value / 100)",
              "lang": "painless"
            }
          }
        },
        "custom_score": {
          "sum": {
            "script": {
              "source": """
                double base_score = doc['rating'].value * doc['review_count'].value;
                if (doc['featured'].value) {
                  base_score *= 1.5;
                }
                return base_score;
              """
            }
          }
        }
      }
    }
  }
}
```

---

## Performance Tips

### Query Optimization

1. **Use Filters Instead of Queries When Possible**

   ```json
   // Good - uses filter (cached, no scoring)
   {"filter": [{"term": {"status": "published"}}]}

   // Less optimal - uses query (scored, not cached)
   {"must": [{"term": {"status": "published"}}]}
   ```

2. **Avoid Scripts on Large Datasets**

   ```json
   // Bad - script on every document
   {"script": {"source": "doc['price'].value > 100"}}

   // Good - use range query instead
   {"range": {"price": {"gt": 100}}}
   ```

3. **Use Parameters in Scripts**

   ```json
   // Good - uses parameters (can be cached)
   {
     "script": {
       "source": "doc['age'].value > params.min_age",
       "params": {"min_age": 18}
     }
   }

   // Bad - hardcoded values (cannot be cached)
   {"script": {"source": "doc['age'].value > 18"}}
   ```

4. **Optimize Field Access**

   ```json
   // Good - check field exists first
   if (doc['optional_field'].size() > 0) {
     return doc['optional_field'].value;
   }

   // Bad - may cause null pointer exceptions
   return doc['optional_field'].value;
   ```

### Script Performance Tips

1. **Minimize Object Creation**

   ```painless
   // Good - reuse variables
   def result = 0;
   for (item in doc['items']) {
     result += item;
   }

   // Bad - creates new objects in loop
   for (item in doc['items']) {
     def temp = new ArrayList();
     // ... operations
   }
   ```

2. **Use Early Returns**

   ```painless
   // Good - early return
   if (doc['status'].value != 'active') {
     return 0;
   }
   // ... expensive operations

   // Bad - unnecessary processing
   if (doc['status'].value == 'active') {
     // ... expensive operations
   } else {
     return 0;
   }
   ```

3. **Cache Expensive Calculations**

   ```painless
   // Good - calculate once
   double baseScore = doc['rating'].value * doc['popularity'].value;
   if (doc['featured'].value) {
     return baseScore * 2.0;
   }
   return baseScore;

   // Bad - calculate multiple times
   if (doc['featured'].value) {
     return doc['rating'].value * doc['popularity'].value * 2.0;
   }
   return doc['rating'].value * doc['popularity'].value;
   ```

---

## Troubleshooting Guide

### Common Errors and Solutions

#### 1. NullPointerException

```painless
// Problem
doc['missing_field'].value  // Throws NPE if field doesn't exist

// Solution
if (doc['missing_field'].size() > 0) {
  return doc['missing_field'].value;
}
return null;  // or default value
```

#### 2. ClassCastException

```painless
// Problem
String value = doc['numeric_field'].value;  // Trying to cast number to string

// Solution
String value = doc['numeric_field'].value.toString();
```

#### 3. Script Compilation Errors

```json
// Problem - syntax error
{"script": {"source": "doc[field].value"}}  // Missing quotes

// Solution
{"script": {"source": "doc['field'].value"}}
```

#### 4. Performance Issues

```painless
// Problem - inefficient loop
for (int i = 0; i < doc['large_array'].size(); i++) {
  if (doc['large_array'][i].contains('target')) {
    return true;
  }
}

// Solution - early exit
for (item in doc['large_array']) {
  if (item.contains('target')) {
    return true;
  }
}
```

### Debug Techniques

#### 1. Add Logging to Scripts

```painless
// Add debug information
Debug.explain('Field value: ' + doc['field'].value);
```

#### 2. Use Simple Test Cases

```json
// Test script with simple data first
{
  "query": {
    "script": {
      "script": {
        "source": "Debug.explain(doc); return true;"
      }
    }
  }
}
```

#### 3. Validate Field Types

```painless
// Check field type before processing
if (doc['field'].value instanceof String) {
  // String operations
} else if (doc['field'].value instanceof Number) {
  // Numeric operations
}
```

### Best Practices Checklist

- [ ] Use parameters instead of hardcoded values in scripts
- [ ] Check field existence before accessing values
- [ ] Use filters instead of queries when scoring isn't needed
- [ ] Test scripts with various data scenarios
- [ ] Monitor query performance and optimize slow queries
- [ ] Use appropriate fuzziness levels (usually "AUTO")
- [ ] Implement proper null handling
- [ ] Cache expensive calculations in scripts
- [ ] Use early returns to avoid unnecessary processing
- [ ] Validate input parameters in scripts

---

## Quick Reference

### Painless Syntax Cheat Sheet

```painless
// Variables
String str = "text";
int num = 42;
boolean flag = true;

// Document access
doc['field'].value           // Single value
doc['field'].values          // Multiple values
doc['field'].size()          // Count values
ctx._source.field            // Update scripts

// String operations
str.length()
str.toLowerCase()
str.contains('substring')
str.substring(0, 5)

// Math operations
Math.max(a, b)
Math.min(a, b)
Math.abs(value)

// Date operations
ZonedDateTime.now()
doc['date'].value.isAfter(otherDate)

// Collections
list.size()
list.contains(item)
list.get(index)
```

### Common Query Patterns

```json
// Basic search
{"match": {"field": "value"}}

// Fuzzy search
{"match": {"field": {"query": "value", "fuzziness": "AUTO"}}}

// Range
{"range": {"field": {"gte": 10, "lte": 100}}}

// Multiple values
{"terms": {"field": ["value1", "value2"]}}

// Boolean combination
{
  "bool": {
    "must": [...],
    "filter": [...],
    "should": [...],
    "must_not": [...]
  }
}
```

This cookbook should serve as your complete reference guide for Elasticsearch Painless scripting and fuzzy search queries!
