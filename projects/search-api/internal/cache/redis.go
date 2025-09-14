package cache

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/saif-islam/es-playground/projects/search-api/internal/models"
)

// RedisCache provides Redis-based caching functionality
type RedisCache struct {
	client   *redis.Client
	logger   *zap.Logger
	config   models.CacheConfig
	prefix   string
	enabled  bool
}

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Data        interface{} `json:"data"`
	CreatedAt   time.Time   `json:"created_at"`
	AccessCount int64       `json:"access_count"`
	LastAccess  time.Time   `json:"last_access"`
	TTL         time.Duration `json:"ttl"`
	Version     string      `json:"version"`
	Compressed  bool        `json:"compressed"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	HitCount        int64   `json:"hit_count"`
	MissCount       int64   `json:"miss_count"`
	HitRate         float64 `json:"hit_rate"`
	TotalKeys       int64   `json:"total_keys"`
	MemoryUsage     int64   `json:"memory_usage_bytes"`
	EvictedKeys     int64   `json:"evicted_keys"`
	ExpiredKeys     int64   `json:"expired_keys"`
	AverageKeySize  float64 `json:"average_key_size"`
	PopularKeys     []string `json:"popular_keys"`
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(redisClient *redis.Client, config models.CacheConfig, logger *zap.Logger) *RedisCache {
	// Set defaults
	if config.Prefix == "" {
		config.Prefix = "es_playground"
	}
	if config.TTL == 0 {
		config.TTL = 5 * time.Minute
	}
	if config.MaxKeyLength == 0 {
		config.MaxKeyLength = 512
	}
	if config.MaxValueSize == 0 {
		config.MaxValueSize = 10 * 1024 * 1024 // 10MB
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 3 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 3 * time.Second
	}

	return &RedisCache{
		client:  redisClient,
		logger:  logger,
		config:  config,
		prefix:  config.Prefix,
		enabled: config.Enabled,
	}
}

// Get retrieves a value from the cache
func (c *RedisCache) Get(ctx context.Context, key string) (interface{}, bool) {
	if !c.enabled {
		return nil, false
	}

	fullKey := c.buildKey(key)
	
	// Get the cached entry
	data, err := c.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err != redis.Nil {
			c.logger.Warn("Cache get error", zap.String("key", key), zap.Error(err))
		}
		c.incrementMissCount(ctx)
		return nil, false
	}

	// Deserialize the cache entry
	var entry CacheEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		c.logger.Error("Failed to deserialize cache entry", zap.String("key", key), zap.Error(err))
		c.incrementMissCount(ctx)
		return nil, false
	}

	// Update access statistics in background
	go c.updateAccessStats(ctx, fullKey)

	c.incrementHitCount(ctx)
	return entry.Data, true
}

// Set stores a value in the cache
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !c.enabled {
		return nil
	}

	fullKey := c.buildKey(key)
	
	// Create cache entry with metadata
	entry := CacheEntry{
		Data:        value,
		CreatedAt:   time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
		TTL:         ttl,
		Version:     "1.0",
		Compressed:  false,
	}

	// Serialize the entry
	data, err := json.Marshal(entry)
	if err != nil {
		c.logger.Error("Failed to serialize cache entry", zap.String("key", key), zap.Error(err))
		return err
	}

	// Check value size limit
	if len(data) > c.config.MaxValueSize {
		c.logger.Warn("Cache value too large", 
			zap.String("key", key), 
			zap.Int("size", len(data)), 
			zap.Int("max_size", c.config.MaxValueSize))
		return fmt.Errorf("cache value too large: %d bytes", len(data))
	}

	// Use adaptive TTL if enabled
	if c.config.AdaptiveTTL {
		ttl = c.calculateAdaptiveTTL(key, value, ttl)
	}

	// Store in Redis
	if err := c.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		c.logger.Error("Failed to set cache entry", zap.String("key", key), zap.Error(err))
		return err
	}

	c.logger.Debug("Cache entry stored", 
		zap.String("key", key), 
		zap.Duration("ttl", ttl),
		zap.Int("size", len(data)))

	return nil
}

// Delete removes a value from the cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if !c.enabled {
		return nil
	}

	fullKey := c.buildKey(key)
	return c.client.Del(ctx, fullKey).Err()
}

// Exists checks if a key exists in the cache
func (c *RedisCache) Exists(ctx context.Context, key string) bool {
	if !c.enabled {
		return false
	}

	fullKey := c.buildKey(key)
	count, err := c.client.Exists(ctx, fullKey).Result()
	return err == nil && count > 0
}

// GetSearchResult retrieves a cached search result
func (c *RedisCache) GetSearchResult(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, bool) {
	key := c.generateSearchKey(req)
	
	if data, found := c.Get(ctx, key); found {
		if response, ok := data.(*models.SearchResponse); ok {
			// Add cache hit indicator
			response.CacheHit = true
			return response, true
		}
	}
	
	return nil, false
}

// SetSearchResult caches a search result
func (c *RedisCache) SetSearchResult(ctx context.Context, req *models.SearchRequest, response *models.SearchResponse) error {
	key := c.generateSearchKey(req)
	ttl := c.config.TTL
	
	// Calculate adaptive TTL based on query characteristics
	if c.config.AdaptiveTTL {
		ttl = c.calculateSearchTTL(req, response)
	}
	
	// Clone response to avoid cache hit flag in cached version
	cachedResponse := *response
	cachedResponse.CacheHit = false
	
	return c.Set(ctx, key, &cachedResponse, ttl)
}

// InvalidatePattern removes all keys matching a pattern
func (c *RedisCache) InvalidatePattern(ctx context.Context, pattern string) error {
	if !c.enabled {
		return nil
	}

	fullPattern := c.buildKey(pattern)
	
	// Find matching keys
	keys, err := c.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		// Delete matching keys
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return err
		}
		
		c.logger.Info("Invalidated cache keys", 
			zap.String("pattern", pattern), 
			zap.Int("count", len(keys)))
	}

	return nil
}

// GetStats returns cache statistics
func (c *RedisCache) GetStats(ctx context.Context) (*CacheStats, error) {
	if !c.enabled {
		return &CacheStats{}, nil
	}

	// Get Redis info
	info, err := c.client.Info(ctx, "memory", "stats").Result()
	if err != nil {
		return nil, err
	}

	// Get cache-specific stats
	hitCount := c.getStatCounter(ctx, "hits")
	missCount := c.getStatCounter(ctx, "misses")
	total := hitCount + missCount
	
	var hitRate float64
	if total > 0 {
		hitRate = float64(hitCount) / float64(total) * 100
	}

	// Get key count
	keyCount, _ := c.client.DBSize(ctx).Result()

	// Get popular keys
	popularKeys := c.getPopularKeys(ctx, 10)

	stats := &CacheStats{
		HitCount:       hitCount,
		MissCount:      missCount,
		HitRate:        hitRate,
		TotalKeys:      keyCount,
		PopularKeys:    popularKeys,
	}

	// Parse memory usage from Redis info (simplified)
	// In production, you'd want more sophisticated parsing
	if len(info) > 0 {
		stats.MemoryUsage = 1024 * 1024 // Placeholder
	}

	return stats, nil
}

// WarmUp pre-loads frequently accessed data
func (c *RedisCache) WarmUp(ctx context.Context, keys []string) error {
	if !c.enabled || len(keys) == 0 {
		return nil
	}

	c.logger.Info("Starting cache warm-up", zap.Int("keys", len(keys)))
	
	// This would typically involve pre-loading data from the primary data source
	// For now, we'll just log the intent
	for _, key := range keys {
		c.logger.Debug("Warming up cache key", zap.String("key", key))
	}

	return nil
}

// Clear removes all cache entries
func (c *RedisCache) Clear(ctx context.Context) error {
	if !c.enabled {
		return nil
	}

	pattern := c.buildKey("*")
	return c.InvalidatePattern(ctx, pattern)
}

// Helper methods

func (c *RedisCache) buildKey(key string) string {
	if len(key) > c.config.MaxKeyLength {
		// Use hash for long keys
		hash := md5.Sum([]byte(key))
		key = hex.EncodeToString(hash[:])
	}
	return fmt.Sprintf("%s:%s", c.prefix, key)
}

func (c *RedisCache) generateSearchKey(req *models.SearchRequest) string {
	// Create a deterministic key based on search parameters
	keyData := map[string]interface{}{
		"query":      req.Query,
		"index":      req.Index,
		"size":       req.Size,
		"from":       req.From,
		"query_type": req.QueryType,
		"fields":     req.Fields,
		"sort":       req.Sort,
		"filters":    req.Filters,
	}
	
	keyBytes, _ := json.Marshal(keyData)
	hash := md5.Sum(keyBytes)
	return fmt.Sprintf("search:%s", hex.EncodeToString(hash[:]))
}

func (c *RedisCache) calculateSearchTTL(req *models.SearchRequest, response *models.SearchResponse) time.Duration {
	baseTTL := c.config.TTL
	
	// Longer TTL for popular queries
	if response.Total.Value > 100 {
		baseTTL = baseTTL * 2
	}
	
	// Shorter TTL for real-time data
	if req.Index == "realtime" || req.Index == "live" {
		baseTTL = baseTTL / 4
	}
	
	// Adjust based on result count
	if response.Total.Value == 0 {
		baseTTL = baseTTL / 2 // Cache empty results for shorter time
	}
	
	return baseTTL
}

func (c *RedisCache) calculateAdaptiveTTL(key string, value interface{}, defaultTTL time.Duration) time.Duration {
	// Simple adaptive TTL logic - in production, this would be more sophisticated
	return defaultTTL
}

func (c *RedisCache) updateAccessStats(ctx context.Context, key string) {
	// Update access count and last access time
	pipeline := c.client.Pipeline()
	pipeline.HIncrBy(ctx, key+":stats", "access_count", 1)
	pipeline.HSet(ctx, key+":stats", "last_access", time.Now().Unix())
	pipeline.Exec(ctx)
}

func (c *RedisCache) incrementHitCount(ctx context.Context) {
	c.client.Incr(ctx, c.buildKey("stats:hits"))
}

func (c *RedisCache) incrementMissCount(ctx context.Context) {
	c.client.Incr(ctx, c.buildKey("stats:misses"))
}

func (c *RedisCache) getStatCounter(ctx context.Context, stat string) int64 {
	count, _ := c.client.Get(ctx, c.buildKey("stats:"+stat)).Int64()
	return count
}

func (c *RedisCache) getPopularKeys(ctx context.Context, limit int) []string {
	// This would typically query the most accessed keys
	// For now, return empty slice
	return []string{}
}

// CacheManager provides high-level cache operations
type CacheManager struct {
	cache  *RedisCache
	logger *zap.Logger
}

// NewCacheManager creates a new cache manager
func NewCacheManager(cache *RedisCache, logger *zap.Logger) *CacheManager {
	return &CacheManager{
		cache:  cache,
		logger: logger,
	}
}

// GetCache returns the underlying cache instance
func (cm *CacheManager) GetCache() *RedisCache {
	return cm.cache
}