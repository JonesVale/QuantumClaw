package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// ---------- RSS XML structures ----------

type rssFeed struct {
	XMLName xml.Name `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string     `xml:"title"`
	Link        string     `xml:"link"`
	Description string     `xml:"description"`
	Items       []rssItem  `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Author      string `xml:"author"`
	PubDate     string `xml:"pubDate"`
}

// ---------- Atom XML structures ----------

type atomFeed struct {
	XMLName xml.Name   `xml:"feed"`
	Title   string     `xml:"title"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title     string    `xml:"title"`
	Link      []atomLink `xml:"link"`
	Published string    `xml:"published"`
	Updated   string    `xml:"updated"`
	Author    atomAuthor `xml:"author"`
	Summary   string    `xml:"summary"`
	Content   string    `xml:"content"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

// ---------- Source definition ----------

type RssSource struct {
	Name     string
	FeedURL  string
	Language string
	Enabled  bool
}

// RssSources returns the list of RSS/Atom feeds to fetch.
// Matches the hardcoded sources from news-sources.ts.
// NOTE: These are fallback defaults; the service reads from rss_sources DB table first.
// If no enabled sources exist in DB, these defaults are used and auto-seeded into the database.
func RssSources() []RssSource {
	return []RssSource{
		// ── 中文 ──
		{
			Name:     "机器之心",
			FeedURL:  "https://www.jiqizhixin.com/rss",
			Language: "zh",
			Enabled:  false, // RSS feed deprecated (returns HTML, not RSS)
		},
		{
			Name:     "量子位",
			FeedURL:  "https://www.qbitai.com/feed",
			Language: "zh",
			Enabled:  false, // HTTP 403 — 已被封禁，需换新地址
		},
		{
			Name:     "36氪 AI",
			FeedURL:  "https://36kr.com/feed",
			Language: "zh",
			Enabled:  true,
		},
		{
			Name:     "雷锋网",
			FeedURL:  "https://www.leiphone.com/feed",
			Language: "zh",
			Enabled:  true,
		},
		{
			Name:     "AI科技评论",
			FeedURL:  "https://www.aireviewweekly.com/rss.xml",
			Language: "zh",
			Enabled:  true,
		},
		{
			Name:     "InfoQ AI",
			FeedURL:  "https://www.infoq.cn/feed?tag=人工智能",
			Language: "zh",
			Enabled:  true,
		},
		{
			Name:     "新智元",
			FeedURL:  "https://www.jiqizhixin.com/rss", // 新智元与机器之心同源，备用
			Language: "zh",
			Enabled:  false, // 待验证
		},
		// ── English ──
		{
			Name:     "OpenAI Blog",
			FeedURL:  "https://openai.com/blog/feed.xml",
			Language: "en",
			Enabled:  false, // OpenAI no longer serves RSS (HTTP 403)
		},
		{
			Name:     "Anthropic",
			FeedURL:  "https://www.anthropic.com/blog/feed.xml",
			Language: "en",
			Enabled:  false, // Anthropic RSS feed not available (HTTP 404)
		},
		{
			Name:     "Google AI",
			FeedURL:  "https://feeds.feedburner.com/blogspot/gJZg",
			Language: "en",
			Enabled:  false, // 国内超时，需代理
		},
		{
			Name:     "MIT Tech Review",
			FeedURL:  "https://www.technologyreview.com/topic/artificial-intelligence/feed/",
			Language: "en",
			Enabled:  true,
		},
		{
			Name:     "ArXiv",
			FeedURL:  "https://export.arxiv.org/rss/cs.AI",
			Language: "en",
			Enabled:  false, // 国内可能超时
		},
		{
			Name:     "Hugging Face",
			FeedURL:  "https://huggingface.co/blog/feed.xml",
			Language: "en",
			Enabled:  false, // 国内超时
		},
		{
			Name:     "Reddit AI",
			FeedURL:  "https://www.reddit.com/r/artificial/.rss",
			Language: "en",
			Enabled:  false, // 国内超时
		},
		// ── Quantum Computing ──
		{
			Name:     "arXiv Quantum Physics",
			FeedURL:  "https://export.arxiv.org/rss/quant-ph",
			Language: "en",
			Enabled:  false, // 国内超时
		},
		{
			Name:     "Google Quantum AI",
			FeedURL:  "https://blog.research.google/feeds/posts/default/-/quantum",
			Language: "en",
			Enabled:  false, // 国内超时
		},
		{
			Name:     "Quanta Magazine",
			FeedURL:  "https://www.quantamagazine.org/feed",
			Language: "en",
			Enabled:  true,
		},
	}
}

// ---------- Rate limiter ----------

type sourceRateLimiter struct {
	mu      sync.Mutex
	lastRun map[string]time.Time
}

var (
	rssRateLimiter = &sourceRateLimiter{
		lastRun: make(map[string]time.Time),
	}
	rssClient = &http.Client{
		Timeout: 30 * time.Second,
	}
)

func (rl *sourceRateLimiter) canFetch(name string, minInterval time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	last, ok := rl.lastRun[name]
	if !ok || time.Since(last) >= minInterval {
		rl.lastRun[name] = time.Now()
		return true
	}
	return false
}

func (rl *sourceRateLimiter) markFetched(name string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.lastRun[name] = time.Now()
}

// ---------- RSS fetching logic ----------

// fetchFeed fetches a URL and returns the raw bytes
func fetchFeed(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "QuantumClaw-RSS/1.0")
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml")

	resp, err := rssClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // 5MB limit
	if err != nil {
		return nil, err
	}
	return body, nil
}

// parseFeed attempts to parse the body as RSS 2.0, then as Atom
func parseFeed(body []byte) ([]model.RssArticle, string, error) {
	// Try RSS 2.0 first
	var feed rssFeed
	if err := xml.Unmarshal(body, &feed); err == nil && len(feed.Channel.Items) > 0 {
		articles := make([]model.RssArticle, 0, len(feed.Channel.Items))
		for _, item := range feed.Channel.Items {
			pubDate := parseDate(item.PubDate)
			articles = append(articles, model.RssArticle{
				Title:       strings.TrimSpace(item.Title),
				Link:        strings.TrimSpace(item.Link),
				Description: stripHTMLTags(strings.TrimSpace(item.Description)),
				Author:      strings.TrimSpace(item.Author),
				PublishedAt: pubDate,
			})
		}
		return articles, feed.Channel.Title, nil
	}

	// Try Atom
	var atom atomFeed
	if err := xml.Unmarshal(body, &atom); err == nil && len(atom.Entries) > 0 {
		articles := make([]model.RssArticle, 0, len(atom.Entries))
		for _, entry := range atom.Entries {
			var link string
			for _, l := range entry.Link {
				if l.Rel == "alternate" || (l.Rel == "" && l.Href != "") {
					link = l.Href
					break
				}
			}
			if link == "" && len(entry.Link) > 0 {
				link = entry.Link[0].Href
			}

			pubDate := parseDate(entry.Published)
			if pubDate.IsZero() {
				pubDate = parseDate(entry.Updated)
			}

			summary := entry.Summary
			if summary == "" {
				summary = entry.Content
			}

			articles = append(articles, model.RssArticle{
				Title:       strings.TrimSpace(entry.Title),
				Link:        strings.TrimSpace(link),
				Description: stripHTMLTags(strings.TrimSpace(summary)),
				Author:      strings.TrimSpace(entry.Author.Name),
				PublishedAt: pubDate,
			})
		}
		return articles, atom.Title, nil
	}

	return nil, "", fmt.Errorf("unable to parse as RSS or Atom")
}

// parseDate tries common date formats
func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Now()
	}
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 MST",
		"02 Jan 2006 15:04:05 -0700",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, dateStr); err == nil {
			return t
		}
	}
	return time.Now()
}

// stripHTMLTags removes basic HTML tags from a string
func stripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ---------- Core service ----------

// FetchSingleSource is a public wrapper for fetchAndStoreSource, used by admin API to manually trigger a fetch.
func FetchSingleSource(name, feedURL, language string) error {
	return fetchAndStoreSource(RssSource{Name: name, FeedURL: feedURL, Language: language})
}

// FetchAllEnabledSources fetches all enabled sources from the database (used for manual trigger).
func FetchAllEnabledSources() (int, int) {
	sources, err := model.GetEnabledRssSources()
	if err != nil || len(sources) == 0 {
		return 0, 0
	}
	success := 0
	for _, s := range sources {
		if err := fetchAndStoreSource(RssSource{Name: s.Name, FeedURL: s.FeedURL, Language: s.Language}); err != nil {
			model.UpdateRssSourceFetchStatus(s.Id, time.Now(), err.Error())
		} else {
			model.UpdateRssSourceFetchStatus(s.Id, time.Now(), "")
			success++
		}
	}
	return len(sources), success
}

// fetchAndStoreSource fetches a single RSS source and stores articles
func fetchAndStoreSource(source RssSource) error {
	body, err := fetchFeed(source.FeedURL)
	if err != nil {
		logger.SysError(fmt.Sprintf("RSS fetch error [%s]: %v", source.Name, err))
		return err
	}

	articles, feedTitle, err := parseFeed(body)
	if err != nil {
		logger.SysError(fmt.Sprintf("RSS parse error [%s]: %v", source.Name, err))
		return err
	}

	logger.SysLog(fmt.Sprintf("RSS [%s] fetched %d articles from %q", source.Name, len(articles), feedTitle))

	storeCount := 0
	for _, art := range articles {
		art.Source = source.Name
		art.Language = source.Language
		// Set CreatedAt if zero
		if art.CreatedAt.IsZero() {
			art.CreatedAt = time.Now()
		}
		if err := model.AddRssArticle(&art); err != nil {
			logger.SysError(fmt.Sprintf("RSS store error [%s]: %v", source.Name, err))
		} else {
			storeCount++
		}
	}

	logger.SysLog(fmt.Sprintf("RSS [%s] stored %d new articles", source.Name, storeCount))
	return nil
}

// StartRssService runs the RSS fetch loop in the background.
// It reads enabled sources from the rss_sources database table (preferred),
// falling back to hardcoded defaults if the DB is empty (auto-seeds defaults into DB).
func StartRssService(ctx context.Context) {
	logger.SysLog("RSS fetch service started")

	// Ensure default sources exist in database
	seeded, err := model.SeedDefaultRssSources()
	if err != nil {
		logger.SysError(fmt.Sprintf("RSS seed error: %v", err))
	} else if seeded > 0 {
		logger.SysLog(fmt.Sprintf("RSS: Seeded %d default sources into database", seeded))
	}

	// Load enabled sources from database
	dbSources, dbErr := model.GetEnabledRssSources()
	var sources []RssSource
	if dbErr == nil && len(dbSources) > 0 {
		// Database has sources — use them
		for _, s := range dbSources {
			sources = append(sources, RssSource{
				Name:     s.Name,
				FeedURL:  s.FeedURL,
				Language: s.Language,
				Enabled:  s.Enabled,
			})
		}
		logger.SysLog(fmt.Sprintf("RSS: Loaded %d enabled sources from database", len(sources)))
	} else {
		// Fallback to hardcoded defaults (DB empty or error)
		sources = RssSources()
		logger.SysLog(fmt.Sprintf("RSS: Using %d fallback hardcoded sources (DB error or empty)", len(sources)))
	}

	// Do an initial fetch on startup
	for _, source := range sources {
		if !source.Enabled {
			logger.SysLog(fmt.Sprintf("RSS source [%s] is disabled, skipping", source.Name))
			continue
		}
		select {
		case <-ctx.Done():
			logger.SysLog("RSS fetch service stopped (context cancelled)")
			return
		default:
		}

		if err := fetchAndStoreSource(source); err != nil {
			// Error already logged in fetchAndStoreSource
		}
		rssRateLimiter.markFetched(source.Name)
	}

	// Then fetch periodically every 10 minutes
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.SysLog("RSS fetch service stopped (context cancelled)")
			return
		case <-ticker.C:
			// Re-load from database each cycle to pick up enable/disable changes
			dbSources, dbErr = model.GetEnabledRssSources()
			if dbErr == nil && len(dbSources) > 0 {
				sources = make([]RssSource, 0, len(dbSources))
				for _, s := range dbSources {
					sources = append(sources, RssSource{
						Name:     s.Name,
						FeedURL:  s.FeedURL,
						Language: s.Language,
						Enabled:  s.Enabled,
					})
				}
			}

			for _, source := range sources {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if !source.Enabled {
					continue
				}

				if !rssRateLimiter.canFetch(source.Name, 10*time.Minute) {
					continue
				}

				if err := fetchAndStoreSource(source); err != nil {
					// Error already logged
				}
			}
		}
	}
}
