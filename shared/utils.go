package shared

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DecodeJSONResponse decodes a JSON response from Elasticsearch
func DecodeJSONResponse(res *http.Response, v interface{}) error {
	if res.Body == nil {
		return fmt.Errorf("response body is nil")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// ESError represents an Elasticsearch error response
type ESError struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

// ESErrorResponse represents the structure of Elasticsearch error responses
type ESErrorResponse struct {
	Error ESError `json:"error"`
}

// ParseESError attempts to parse an Elasticsearch error from the response
func ParseESError(res *http.Response) error {
	if res.Body == nil {
		return fmt.Errorf("elasticsearch error: %s", res.Status())
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("elasticsearch error: %s (failed to read body: %v)", res.Status(), err)
	}

	var esErr ESErrorResponse
	if err := json.Unmarshal(body, &esErr); err != nil {
		return fmt.Errorf("elasticsearch error: %s (body: %s)", res.Status(), string(body))
	}

	return fmt.Errorf("elasticsearch error [%s]: %s", esErr.Error.Type, esErr.Error.Reason)
}

// FormatIndexName ensures index names follow Elasticsearch conventions
func FormatIndexName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)
	
	// Replace invalid characters with hyphens
	invalidChars := []string{" ", "_", "\\", "/", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "-")
	}
	
	// Remove leading/trailing hyphens and dots
	name = strings.Trim(name, "-.")
	
	// Ensure it doesn't start with hyphen, underscore, or plus
	if len(name) > 0 && (name[0] == '-' || name[0] == '_' || name[0] == '+') {
		name = "idx-" + name
	}
	
	return name
}

// BuildQuery is a helper to build Elasticsearch queries
type QueryBuilder struct {
	query map[string]interface{}
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		query: make(map[string]interface{}),
	}
}

// MatchAll adds a match_all query
func (qb *QueryBuilder) MatchAll() *QueryBuilder {
	qb.query["match_all"] = map[string]interface{}{}
	return qb
}

// Match adds a match query
func (qb *QueryBuilder) Match(field, value string) *QueryBuilder {
	qb.query["match"] = map[string]interface{}{
		field: value,
	}
	return qb
}

// Term adds a term query
func (qb *QueryBuilder) Term(field, value string) *QueryBuilder {
	qb.query["term"] = map[string]interface{}{
		field: value,
	}
	return qb
}

// Range adds a range query
func (qb *QueryBuilder) Range(field string, gte, lte interface{}) *QueryBuilder {
	rangeQuery := make(map[string]interface{})
	if gte != nil {
		rangeQuery["gte"] = gte
	}
	if lte != nil {
		rangeQuery["lte"] = lte
	}
	
	qb.query["range"] = map[string]interface{}{
		field: rangeQuery,
	}
	return qb
}

// Bool starts a bool query
func (qb *QueryBuilder) Bool() *BoolQueryBuilder {
	boolQuery := &BoolQueryBuilder{
		parent: qb,
		bool:   make(map[string]interface{}),
	}
	qb.query["bool"] = boolQuery.bool
	return boolQuery
}

// Build returns the final query
func (qb *QueryBuilder) Build() map[string]interface{} {
	return map[string]interface{}{
		"query": qb.query,
	}
}

// BoolQueryBuilder helps build bool queries
type BoolQueryBuilder struct {
	parent *QueryBuilder
	bool   map[string]interface{}
}

// Must adds must clauses
func (bqb *BoolQueryBuilder) Must(queries ...map[string]interface{}) *BoolQueryBuilder {
	if _, exists := bqb.bool["must"]; !exists {
		bqb.bool["must"] = make([]map[string]interface{}, 0)
	}
	bqb.bool["must"] = append(bqb.bool["must"].([]map[string]interface{}), queries...)
	return bqb
}

// Should adds should clauses
func (bqb *BoolQueryBuilder) Should(queries ...map[string]interface{}) *BoolQueryBuilder {
	if _, exists := bqb.bool["should"]; !exists {
		bqb.bool["should"] = make([]map[string]interface{}, 0)
	}
	bqb.bool["should"] = append(bqb.bool["should"].([]map[string]interface{}), queries...)
	return bqb
}

// MustNot adds must_not clauses
func (bqb *BoolQueryBuilder) MustNot(queries ...map[string]interface{}) *BoolQueryBuilder {
	if _, exists := bqb.bool["must_not"]; !exists {
		bqb.bool["must_not"] = make([]map[string]interface{}, 0)
	}
	bqb.bool["must_not"] = append(bqb.bool["must_not"].([]map[string]interface{}), queries...)
	return bqb
}

// Filter adds filter clauses
func (bqb *BoolQueryBuilder) Filter(queries ...map[string]interface{}) *BoolQueryBuilder {
	if _, exists := bqb.bool["filter"]; !exists {
		bqb.bool["filter"] = make([]map[string]interface{}, 0)
	}
	bqb.bool["filter"] = append(bqb.bool["filter"].([]map[string]interface{}), queries...)
	return bqb
}

// Build returns the final query
func (bqb *BoolQueryBuilder) Build() map[string]interface{} {
	return bqb.parent.Build()
}