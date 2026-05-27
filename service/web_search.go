package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/quantumclaw/quantumclaw/common/config"
)

// WebSearchConfig holds the global search configuration.
type WebSearchConfig struct {
	Enabled    bool   `json:"enabled"`
	Provider   string `json:"provider"`   // bing | searxng | serpapi
	APIKey     string `json:"-"`          // never serialized to frontend
	Endpoint   string `json:"endpoint"`   // custom endpoint for SearXNG
	MaxResults int    `json:"max_results"`
	TimeoutSec int    `json:"timeout_sec"`

	// AutoSearch: automatically trigger search when the user's message
	// appears to ask for real-time information.
	AutoSearch bool `json:"auto_search"`

	// Pricing: cost per search query (in quota units)
	CostPerSearch int64 `json:"cost_per_search"`
}

//  Web Search Service 
// Provides internet search capabilities that can be injected into AI requests.
// Supports multiple search backends: Bing, SearXNG, SerpAPI.

// SearchProvider defines the search backend type.
type SearchProvider string

const (
	SearchProviderBing    SearchProvider = "bing"
	SearchProviderSearXNG SearchProvider = "searxng"
	SearchProviderSerpAPI SearchProvider = "serpapi"
)

// SearchResult represents a single search result item.
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// SearchResponse is the standardized search response.
type SearchResponse struct {
	Results    []SearchResult `json:"results"`
	TotalCount int            `json:"total_count"`
	Source     string         `json:"source"`
	Error      string         `json:"error,omitempty"`
}

var DefaultWebSearchConfig = WebSearchConfig{
	Enabled:       false,
	Provider:      string(SearchProviderBing),
	MaxResults:    5,
	TimeoutSec:    10,
	AutoSearch:    false,
	CostPerSearch: 100, // default 100 quota units per search
}

var (
	webSearchConfig    = DefaultWebSearchConfig
	webSearchClient    = &http.Client{Timeout: 10 * time.Second}
	webSearchConfigMu  sync.RWMutex
)

// GetWebSearchConfig returns a copy of the current search config.
func GetWebSearchConfig() WebSearchConfig {
	webSearchConfigMu.RLock()
	defer webSearchConfigMu.RUnlock()
	return webSearchConfig
}

// UpdateWebSearchConfig updates the search config at runtime.
func UpdateWebSearchConfig(cfg WebSearchConfig) {
	webSearchConfigMu.Lock()
	defer webSearchConfigMu.Unlock()

	if cfg.Provider == "" {
		cfg.Provider = DefaultWebSearchConfig.Provider
	}
	if cfg.MaxResults <= 0 {
		cfg.MaxResults = DefaultWebSearchConfig.MaxResults
	}
	if cfg.TimeoutSec <= 0 {
		cfg.TimeoutSec = DefaultWebSearchConfig.TimeoutSec
	}
	if cfg.CostPerSearch <= 0 {
		cfg.CostPerSearch = DefaultWebSearchConfig.CostPerSearch
	}

	// Read API key from env if not set in config
	if cfg.APIKey == "" {
		switch SearchProvider(cfg.Provider) {
		case SearchProviderBing:
			cfg.APIKey = config.BingSearchAPIKey
		case SearchProviderSerpAPI:
			cfg.APIKey = config.SerpAPIKey
		}
	}

	// Update timeout
	webSearchClient.Timeout = time.Duration(cfg.TimeoutSec) * time.Second
	webSearchConfig = cfg
}

//  Auto-search trigger patterns 
// These are simple heuristics to detect if a user's message requires real-time info.

var autoSearchPatterns = []string{
	"weather", "today", "current", "latest", "recent", "news",
	"stock", "price", "price of", "what is the",
	"how to", "who is", "where is", "when did",
}

// ShouldAutoSearch checks if a message likely needs real-time info.
func ShouldAutoSearch(message string) bool {
	msg := strings.ToLower(message)
	for _, pat := range autoSearchPatterns {
		if strings.Contains(msg, strings.ToLower(pat)) {
			return true
		}
	}
	return false
}

//  Search execution 

// Search performs a web search using the configured provider.
func Search(ctx context.Context, query string) (*SearchResponse, error) {
	cfg := GetWebSearchConfig()
	if !cfg.Enabled {
		return &SearchResponse{Error: "web search is not enabled"}, fmt.Errorf("web search is not enabled")
	}

	switch SearchProvider(cfg.Provider) {
	case SearchProviderBing:
		return searchBing(ctx, query, cfg)
	case SearchProviderSearXNG:
		return searchSearXNG(ctx, query, cfg)
	case SearchProviderSerpAPI:
		return searchSerpAPI(ctx, query, cfg)
	default:
		return &SearchResponse{Error: fmt.Sprintf("unknown search provider: %s", cfg.Provider)},
			fmt.Errorf("unknown search provider: %s", cfg.Provider)
	}
}

//  Bing Search 

func searchBing(ctx context.Context, query string, cfg WebSearchConfig) (*SearchResponse, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("Bing Search API key not configured")
	}

	reqURL := fmt.Sprintf("https://api.bing.microsoft.com/v7.0/search?q=%s&count=%d&mkt=zh-CN",
		url.QueryEscape(query), cfg.MaxResults)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", cfg.APIKey)

	resp, err := webSearchClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bing search: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bing search returned status %d: %s", resp.StatusCode, string(body))
	}

	var bingResp struct {
		WebPages struct {
			Value []struct {
				Name string `json:"name"`
				URL  string `json:"url"`
				Snippet string `json:"snippet"`
			} `json:"value"`
			TotalEstimatedMatches int `json:"totalEstimatedMatches"`
		} `json:"webPages"`
	}
	if err := json.Unmarshal(body, &bingResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	results := make([]SearchResult, 0, len(bingResp.WebPages.Value))
	for _, v := range bingResp.WebPages.Value {
		results = append(results, SearchResult{
			Title:   v.Name,
			URL:     v.URL,
			Snippet: v.Snippet,
		})
	}

	return &SearchResponse{
		Results:    results,
		TotalCount: bingResp.WebPages.TotalEstimatedMatches,
		Source:     "Bing",
	}, nil
}

//  SearXNG Search (self-hosted) 

func searchSearXNG(ctx context.Context, query string, cfg WebSearchConfig) (*SearchResponse, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("SearXNG endpoint not configured")
	}

	reqURL := fmt.Sprintf("%s/search?q=%s&format=json&language=zh-CN&categories=general",
		strings.TrimRight(cfg.Endpoint, "/"), url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := webSearchClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("searxng search: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("searxng returned status %d", resp.StatusCode)
	}

	var searxngResp struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &searxngResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	results := make([]SearchResult, 0, len(searxngResp.Results))
	for _, v := range searxngResp.Results {
		results = append(results, SearchResult{
			Title:   v.Title,
			URL:     v.URL,
			Snippet: v.Content,
		})
	}

	return &SearchResponse{
		Results:    results,
		TotalCount: len(results),
		Source:     "SearXNG",
	}, nil
}

//  SerpAPI Search 

func searchSerpAPI(ctx context.Context, query string, cfg WebSearchConfig) (*SearchResponse, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("SerpAPI key not configured")
	}

	reqURL := fmt.Sprintf("https://serpapi.com/search?q=%s&num=%d&engine=google&hl=zh-CN&api_key=%s",
		url.QueryEscape(query), cfg.MaxResults, cfg.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := webSearchClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("serpapi search: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("serpapi returned status %d: %s", resp.StatusCode, string(body))
	}

	var serpResp struct {
		OrganicResults []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"organic_results"`
	}
	if err := json.Unmarshal(body, &serpResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	results := make([]SearchResult, 0, len(serpResp.OrganicResults))
	for _, v := range serpResp.OrganicResults {
		results = append(results, SearchResult{
			Title:   v.Title,
			URL:     v.Link,
			Snippet: v.Snippet,
		})
	}

	return &SearchResponse{
		Results:    results,
		TotalCount: len(results),
		Source:     "Google (SerpAPI)",
	}, nil
}

// FormatResultsForPrompt formats search results into a text block suitable
// for injection into an AI system prompt or user message.
func FormatResultsForPrompt(results *SearchResponse) string {
	if results == nil || len(results.Results) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- Web Search Results (source: %s) ---\n", results.Source))

	for i, r := range results.Results {
		sb.WriteString(fmt.Sprintf("\n[%d] %s\n", i+1, r.Title))
		sb.WriteString(fmt.Sprintf("    URL: %s\n", r.URL))
		sb.WriteString(fmt.Sprintf("    %s\n", r.Snippet))
	}

	sb.WriteString("\n--- End of Web Search Results ---")
	return sb.String()
}

//  Init: read API keys from env 

func init() {
	cfg := GetWebSearchConfig()
	if cfg.APIKey == "" && config.BingSearchAPIKey != "" {
		cfg.APIKey = config.BingSearchAPIKey
		cfg.Enabled = true
		UpdateWebSearchConfig(cfg)
	}
}
