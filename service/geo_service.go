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

// ── Configuration ──

// GeoServiceConfig holds the global Geo service configuration.
type GeoServiceConfig struct {
	Enabled      bool   `json:"enabled"`
	Provider     string `json:"provider"`      // "amap" | "google" | "openstreetmap"
	APIKey       string `json:"-"`             // Amap key or Google Maps key; never serialized
	GeoCodeKey   string `json:"-"`             // optional separate geocode key (amap)
	MaxResults   int    `json:"max_results"`
	TimeoutSec   int    `json:"timeout_sec"`
	RegionFilter string `json:"region_filter"` // "china" | "global" | "all"
	CostPerQuery int64  `json:"cost_per_query"`
}

var DefaultGeoServiceConfig = GeoServiceConfig{
	Enabled:      false,
	Provider:     "amap",
	MaxResults:   5,
	TimeoutSec:   10,
	RegionFilter: "all",
	CostPerQuery: 50,
}

var (
	geoConfig    = DefaultGeoServiceConfig
	geoClient    = &http.Client{Timeout: 10 * time.Second}
	geoConfigMu  sync.RWMutex
)

// GetGeoConfig returns a copy of the current Geo config.
func GetGeoConfig() GeoServiceConfig {
	geoConfigMu.RLock()
	defer geoConfigMu.RUnlock()
	return geoConfig
}

// UpdateGeoConfig updates the Geo service configuration at runtime.
func UpdateGeoConfig(cfg GeoServiceConfig) {
	geoConfigMu.Lock()
	defer geoConfigMu.Unlock()

	if cfg.Provider == "" {
		cfg.Provider = DefaultGeoServiceConfig.Provider
	}
	if cfg.MaxResults <= 0 {
		cfg.MaxResults = DefaultGeoServiceConfig.MaxResults
	}
	if cfg.TimeoutSec <= 0 {
		cfg.TimeoutSec = DefaultGeoServiceConfig.TimeoutSec
	}
	if cfg.CostPerQuery <= 0 {
		cfg.CostPerQuery = DefaultGeoServiceConfig.CostPerQuery
	}
	if cfg.APIKey == "" {
		// Read from env
		switch cfg.Provider {
		case "amap":
			cfg.APIKey = config.AmapAPIKey
			cfg.GeoCodeKey = config.AmapGeoCodeKey
		case "google":
			cfg.APIKey = config.GoogleMapsAPIKey
		}
	}

	geoClient.Timeout = time.Duration(cfg.TimeoutSec) * time.Second
	geoConfig = cfg
}

// ── API Types ──

// GeoQueryType defines the type of geo query.
type GeoQueryType string

const (
	GeoTypeWeather  GeoQueryType = "weather"
	GeoTypePOI      GeoQueryType = "poi"
	GeoTypeRoute    GeoQueryType = "route"
	GeoTypeGeocode  GeoQueryType = "geocode"
)

// GeoCoord represents a geographic coordinate.
type GeoCoord struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// WeatherResult holds the current weather data.
type WeatherResult struct {
	City          string  `json:"city"`
	Temperature   float64 `json:"temperature"`
	Weather       string  `json:"weather"`       // e.g. "sunny", "cloudy", "rainy"
	Humidity      int     `json:"humidity"`
	WindSpeed     float64 `json:"wind_speed"`
	WindDirection string  `json:"wind_direction"`
	HumidityStr   string  `json:"humidity_str"`
	Description   string  `json:"description"`
}

// POIResult is a point of interest.
type POIResult struct {
	Name      string   `json:"name"`
	Address   string   `json:"address"`
	Location  GeoCoord `json:"location"`
	Distance  int      `json:"distance"`   // meters
	Category  string   `json:"category"`
	Phone     string   `json:"phone,omitempty"`
	Rating    float64  `json:"rating,omitempty"`
}

// RouteResult holds route planning data.
type RouteResult struct {
	Distance     int     `json:"distance"`       // meters
	Duration     int     `json:"duration"`       // seconds
	Steps        int     `json:"steps"`
	Polyline     string  `json:"polyline,omitempty"`
	Description  string  `json:"description"`
}

// GeoResponse is the standardized Geo service response.
type GeoResponse struct {
	QueryType GeoQueryType    `json:"query_type"`
	Query     string          `json:"query"`
	Weather   *WeatherResult  `json:"weather,omitempty"`
	POIs      []POIResult     `json:"pois,omitempty"`
	Route     *RouteResult    `json:"route,omitempty"`
	Coords    *GeoCoord       `json:"coords,omitempty"`
	Source    string          `json:"source"`
	Error     string          `json:"error,omitempty"`
}

// ── Intent Detection ──

// geoIntentPatterns are simple heuristics to detect geo-related queries.
type geoIntent struct {
	Type     GeoQueryType
	Keywords []string
}

var geoIntentPatterns = []geoIntent{
	{Type: GeoTypeWeather, Keywords: []string{"weather", "temperature", "forecast", "humidity", "风", "雨", "雪", "天气", "温度", "气温", "气候"}},
	{Type: GeoTypePOI, Keywords: []string{"restaurant near", "hotel near", "附近", "周边", "商圈", "搜索附近的", "where is", "find", "near me"}},
	{Type: GeoTypeGeocode, Keywords: []string{"latitude of", "longitude of", "coordinates of", "where is", "location of"}},
}

// DetectGeoIntent checks if a message has geo-related intent and returns the type and extracted location.
func DetectGeoIntent(message string) (GeoQueryType, string) {
	msg := strings.ToLower(message)
	for _, intent := range geoIntentPatterns {
		for _, kw := range intent.Keywords {
			if strings.Contains(msg, kw) {
				return intent.Type, message
			}
		}
	}
	return "", ""
}

// ── Search Execution ──

// GeoQuery executes a geo service query.
func GeoQuery(ctx context.Context, queryType GeoQueryType, query string) (*GeoResponse, error) {
	cfg := GetGeoConfig()
	if !cfg.Enabled {
		return &GeoResponse{Error: "geo service is not enabled"}, fmt.Errorf("geo service is not enabled")
	}
	if cfg.APIKey == "" {
		return &GeoResponse{Error: "geo API key not configured"}, fmt.Errorf("geo API key not configured")
	}

	switch cfg.Provider {
	case "amap":
		return geoAmap(ctx, queryType, query, cfg)
	case "google":
		return geoGoogle(ctx, queryType, query, cfg)
	default:
		return geoAmap(ctx, queryType, query, cfg)
	}
}

// FormatGeoResultsForPrompt formats geo results for injection into AI context.
func FormatGeoResultsForPrompt(resp *GeoResponse) string {
	if resp == nil {
		return ""
	}
	var b strings.Builder

	switch resp.QueryType {
	case GeoTypeWeather:
		if resp.Weather != nil {
			w := resp.Weather
			b.WriteString(fmt.Sprintf("[Real-time Weather: %s]\n", w.City))
			b.WriteString(fmt.Sprintf("  Weather: %s\n", w.Weather))
			b.WriteString(fmt.Sprintf("  Temperature: %.1f°C\n", w.Temperature))
			b.WriteString(fmt.Sprintf("  Humidity: %d%%\n", w.Humidity))
			b.WriteString(fmt.Sprintf("  Wind: %.1f m/s (%s)\n", w.WindSpeed, w.WindDirection))
			if w.Description != "" {
				b.WriteString(fmt.Sprintf("  Description: %s\n", w.Description))
			}
		}
	case GeoTypePOI:
		b.WriteString(fmt.Sprintf("[Points of Interest: %s]\n", resp.Query))
		for i, poi := range resp.POIs {
			if i >= 5 {
				break
			}
			dist := ""
			if poi.Distance > 0 {
				if poi.Distance > 1000 {
					dist = fmt.Sprintf(" (%.1fkm)", float64(poi.Distance)/1000.0)
				} else {
					dist = fmt.Sprintf(" (%dm)", poi.Distance)
				}
			}
			b.WriteString(fmt.Sprintf("  %d. %s%s - %s\n", i+1, poi.Name, dist, poi.Address))
		}
	case GeoTypeRoute:
		if resp.Route != nil {
			r := resp.Route
			b.WriteString(fmt.Sprintf("[Route: %s]\n", resp.Query))
			durHours := r.Duration / 3600
			durMin := (r.Duration % 3600) / 60
			distKm := float64(r.Distance) / 1000.0
			b.WriteString(fmt.Sprintf("  Total distance: %.1f km\n", distKm))
			b.WriteString(fmt.Sprintf("  Estimated time: %dh %dm\n", durHours, durMin))
			b.WriteString(fmt.Sprintf("  Description: %s\n", r.Description))
		}
	case GeoTypeGeocode:
		if resp.Coords != nil {
			b.WriteString(fmt.Sprintf("  Coordinates: %.6f, %.6f\n", resp.Coords.Lat, resp.Coords.Lng))
		}
	}

	return b.String()
}

// ── Amap (高德地图) Provider ──

func geoAmap(ctx context.Context, queryType GeoQueryType, query string, cfg GeoServiceConfig) (*GeoResponse, error) {
	apiKey := cfg.APIKey
	if apiKey == "" {
		return nil, fmt.Errorf("amap API key not configured")
	}
	geoKey := cfg.GeoCodeKey
	if geoKey == "" {
		geoKey = apiKey
	}

	switch queryType {
	case GeoTypeWeather:
		return amapWeather(ctx, query, apiKey)
	case GeoTypePOI:
		return amapPOI(ctx, query, apiKey, cfg.MaxResults)
	case GeoTypeGeocode:
		return amapGeocode(ctx, query, geoKey)
	default:
		return &GeoResponse{QueryType: queryType, Error: fmt.Sprintf("unsupported query type: %s", queryType)},
			fmt.Errorf("unsupported query type: %s", queryType)
	}
}

func amapWeather(ctx context.Context, city string, key string) (*GeoResponse, error) {
	// First geocode to get city code
	geoResp, err := amapGeocode(ctx, city, key)
	if err != nil || geoResp.Coords == nil {
		// Fallback: try using city name directly
		geoResp, err = amapGeocode(ctx, city+"市", key)
		if err != nil || geoResp.Coords == nil {
			return &GeoResponse{QueryType: GeoTypeWeather, Query: city, Error: "city not found"},
				fmt.Errorf("amap: city %q not found", city)
		}
	}

	// Amap weather API uses adcode (行政区域编码), which we can't easily get from geocode.
	// Use the city name directly with weather API.
	adcodeURL := fmt.Sprintf("https://restapi.amap.com/v3/config/district?keywords=%s&subdistrict=0&key=%s",
		url.QueryEscape(city), key)

	req, err := http.NewRequestWithContext(ctx, "GET", adcodeURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create adcode request: %w", err)
	}
	resp, err := geoClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("amap adcode: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var districtResp struct {
		Status string `json:"status"`
		Districts []struct {
			Adcode string `json:"adcode"`
			Name   string `json:"name"`
		} `json:"districts"`
	}
	if err := json.Unmarshal(body, &districtResp); err != nil || districtResp.Status != "1" || len(districtResp.Districts) == 0 {
		return &GeoResponse{QueryType: GeoTypeWeather, Query: city, Error: "city adcode not found"}, nil
	}
	adcode := districtResp.Districts[0].Adcode

	// Now query weather
	weatherURL := fmt.Sprintf("https://restapi.amap.com/v3/weather/weatherInfo?city=%s&key=%s&extensions=base",
		adcode, key)
	req2, err := http.NewRequestWithContext(ctx, "GET", weatherURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create weather request: %w", err)
	}
	resp2, err := geoClient.Do(req2)
	if err != nil {
		return nil, fmt.Errorf("amap weather: %w", err)
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)
	var weatherResp struct {
		Status    string `json:"status"`
		Count     string `json:"count"`
		Lives []struct {
			Province      string `json:"province"`
			City          string `json:"city"`
			Adcode        string `json:"adcode"`
			Weather       string `json:"weather"`
			Temperature   string `json:"temperature"`
			WindDirection string `json:"winddirection"`
			WindPower     string `json:"windpower"`
			Humidity      string `json:"humidity"`
		} `json:"lives"`
	}
	if err := json.Unmarshal(body2, &weatherResp); err != nil || weatherResp.Status != "1" || len(weatherResp.Lives) == 0 {
		return &GeoResponse{QueryType: GeoTypeWeather, Query: city, Error: "weather data not available"}, nil
	}

	live := weatherResp.Lives[0]
	temp := 0.0
	fmt.Sscanf(live.Temperature, "%f", &temp)
	humid := 0
	fmt.Sscanf(live.Humidity, "%d", &humid)
	windSpeed := 0.0
	fmt.Sscanf(live.WindPower, "%f", &windSpeed)

	weather := &WeatherResult{
		City:          live.City,
		Temperature:   temp,
		Weather:       live.Weather,
		Humidity:      humid,
		WindSpeed:     windSpeed,
		WindDirection: live.WindDirection,
	}
	return &GeoResponse{
		QueryType: GeoTypeWeather,
		Query:     city,
		Weather:   weather,
		Source:    "amap",
	}, nil
}

func amapPOI(ctx context.Context, query string, key string, maxResults int) (*GeoResponse, error) {
	poiURL := fmt.Sprintf("https://restapi.amap.com/v3/place/text?keywords=%s&offset=%d&key=%s",
		url.QueryEscape(query), maxResults, key)

	req, err := http.NewRequestWithContext(ctx, "GET", poiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create POI request: %w", err)
	}
	resp, err := geoClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("amap POI: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var poiResp struct {
		Status  string `json:"status"`
		Count   string `json:"count"`
		Results []struct {
			Name     string `json:"name"`
			Address  string `json:"address"`
			Location string `json:"location"`
			Distance string `json:"distance"`
			Type     string `json:"type"`
			Tel      string `json:"tel"`
		} `json:"pois"`
	}
	if err := json.Unmarshal(body, &poiResp); err != nil || poiResp.Status != "1" {
		return &GeoResponse{QueryType: GeoTypePOI, Query: query, Error: "no results found"}, nil
	}

	pois := make([]POIResult, 0, len(poiResp.Results))
	for _, r := range poiResp.Results {
		loc := GeoCoord{}
		if parts := strings.Split(r.Location, ","); len(parts) == 2 {
			fmt.Sscanf(parts[0], "%f", &loc.Lng)
			fmt.Sscanf(parts[1], "%f", &loc.Lat)
		}
		dist := 0
		fmt.Sscanf(r.Distance, "%d", &dist)
		pois = append(pois, POIResult{
			Name:     r.Name,
			Address:  r.Address,
			Location: loc,
			Distance: dist,
			Category: r.Type,
			Phone:    r.Tel,
		})
	}

	return &GeoResponse{
		QueryType: GeoTypePOI,
		Query:     query,
		POIs:      pois,
		Source:    "amap",
	}, nil
}

func amapGeocode(ctx context.Context, address string, key string) (*GeoResponse, error) {
	geoURL := fmt.Sprintf("https://restapi.amap.com/v3/geocode/geo?address=%s&key=%s",
		url.QueryEscape(address), key)

	req, err := http.NewRequestWithContext(ctx, "GET", geoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create geocode request: %w", err)
	}
	resp, err := geoClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("amap geocode: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var geoResp struct {
		Status   string `json:"status"`
		Geocodes []struct {
			Location string `json:"location"`
		} `json:"geocodes"`
	}
	if err := json.Unmarshal(body, &geoResp); err != nil || geoResp.Status != "1" || len(geoResp.Geocodes) == 0 {
		return &GeoResponse{QueryType: GeoTypeGeocode, Query: address, Error: "address not found"}, nil
	}

	loc := GeoCoord{}
	if parts := strings.Split(geoResp.Geocodes[0].Location, ","); len(parts) == 2 {
		fmt.Sscanf(parts[0], "%f", &loc.Lng)
		fmt.Sscanf(parts[1], "%f", &loc.Lat)
	}

	return &GeoResponse{
		QueryType: GeoTypeGeocode,
		Query:     address,
		Coords:    &loc,
		Source:    "amap",
	}, nil
}

// ── Google Maps Provider ──

func geoGoogle(ctx context.Context, queryType GeoQueryType, query string, cfg GeoServiceConfig) (*GeoResponse, error) {
	key := cfg.APIKey
	if key == "" {
		return nil, fmt.Errorf("Google Maps API key not configured")
	}

	switch queryType {
	case GeoTypeGeocode:
		return googleGeocode(ctx, query, key)
	case GeoTypePOI:
		return googlePlaces(ctx, query, key, cfg.MaxResults)
	default:
		return &GeoResponse{QueryType: queryType, Error: fmt.Sprintf("google: unsupported query type: %s", queryType)},
			fmt.Errorf("google: unsupported query type: %s", queryType)
	}
}

func googleGeocode(ctx context.Context, address string, key string) (*GeoResponse, error) {
	geoURL := fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?address=%s&key=%s",
		url.QueryEscape(address), key)

	req, err := http.NewRequestWithContext(ctx, "GET", geoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create geocode request: %w", err)
	}
	resp, err := geoClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google geocode: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var geoResp struct {
		Status string `json:"status"`
		Results []struct {
			Geometry struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &geoResp); err != nil || geoResp.Status != "OK" || len(geoResp.Results) == 0 {
		return &GeoResponse{QueryType: GeoTypeGeocode, Query: address, Error: "address not found"}, nil
	}

	loc := geoResp.Results[0].Geometry.Location
	return &GeoResponse{
		QueryType: GeoTypeGeocode,
		Query:     address,
		Coords:    &GeoCoord{Lat: loc.Lat, Lng: loc.Lng},
		Source:    "google",
	}, nil
}

func googlePlaces(ctx context.Context, query string, key string, maxResults int) (*GeoResponse, error) {
	// Use Text Search API for general POI search
	placesURL := fmt.Sprintf("https://maps.googleapis.com/maps/api/place/textsearch/json?query=%s&key=%s",
		url.QueryEscape(query), key)

	req, err := http.NewRequestWithContext(ctx, "GET", placesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create places request: %w", err)
	}
	resp, err := geoClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google places: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var placesResp struct {
		Status  string `json:"status"`
		Results []struct {
			Name     string `json:"name"`
			Vicinity string `json:"formatted_address"`
			Geometry struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
			Rating     float64 `json:"rating"`
			Types      []string `json:"types"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &placesResp); err != nil || placesResp.Status != "OK" || len(placesResp.Results) == 0 {
		return &GeoResponse{QueryType: GeoTypePOI, Query: query, Error: "no results found"}, nil
	}

	pois := make([]POIResult, 0, len(placesResp.Results))
	for _, r := range placesResp.Results {
		category := ""
		if len(r.Types) > 0 {
			category = r.Types[0]
		}
		pois = append(pois, POIResult{
			Name:     r.Name,
			Address:  r.Vicinity,
			Location: GeoCoord{Lat: r.Geometry.Location.Lat, Lng: r.Geometry.Location.Lng},
			Category: category,
			Rating:   r.Rating,
		})
	}

	return &GeoResponse{
		QueryType: GeoTypePOI,
		Query:     query,
		POIs:      pois,
		Source:    "google",
	}, nil
}
