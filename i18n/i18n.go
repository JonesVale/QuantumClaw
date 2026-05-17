package i18n

import (
	"embed"
	"encoding/json"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	LangZhCN    = "zh-CN"
	LangZhTW    = "zh-TW"
	LangEn      = "en"
	LangFr      = "fr"
	LangJa      = "ja"
	LangRu      = "ru"
	LangVi      = "vi"
	DefaultLang = LangEn
)

//go:embed locales/*.json
var localesFS embed.FS

var (
	translations = make(map[string]map[string]string)
	mu           sync.RWMutex
)

func Init() error {
	mu.Lock()
	defer mu.Unlock()

	entries, err := localesFS.ReadDir("locales")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		langCode := strings.TrimSuffix(entry.Name(), ".json")
		content, err := localesFS.ReadFile("locales/" + entry.Name())
		if err != nil {
			return err
		}

		var translation map[string]string
		if err := json.Unmarshal(content, &translation); err != nil {
			return err
		}
		translations[langCode] = translation
	}

	return nil
}

func GetLocalizer(lang string) string {
	return normalizeLang(lang)
}

func T(c *gin.Context, key string, args ...map[string]any) string {
	lang := GetLangFromContext(c)
	return Translate(lang, key, args...)
}

func Translate(lang, key string, args ...map[string]any) string {
	lang = normalizeLang(lang)

	mu.RLock()
	defer mu.RUnlock()

	if trans, ok := translations[lang]; ok {
		if translated, exists := trans[key]; exists {
			return translated
		}
	}

	if trans, ok := translations[DefaultLang]; ok {
		if translated, exists := trans[key]; exists {
			return translated
		}
	}

	return key
}

var userLangLoaderFunc func(userId int) string

func SetUserLangLoader(loader func(userId int) string) {
	userLangLoaderFunc = loader
}

func GetLangFromContext(c *gin.Context) string {
	if c == nil {
		return DefaultLang
	}

	if userLangLoaderFunc != nil {
		if userId, exists := c.Get("id"); exists {
			if uid, ok := userId.(int); ok && uid > 0 {
				lang := userLangLoaderFunc(uid)
				if lang != "" {
					normalized := normalizeLang(lang)
					if IsSupported(normalized) {
						return normalized
					}
				}
			}
		}
	}

	if lang := c.GetString("language"); lang != "" {
		normalized := normalizeLang(lang)
		if IsSupported(normalized) {
			return normalized
		}
	}

	if acceptLang := c.GetHeader("Accept-Language"); acceptLang != "" {
		lang := ParseAcceptLanguage(acceptLang)
		if IsSupported(lang) {
			return lang
		}
	}

	return DefaultLang
}

func ParseAcceptLanguage(header string) string {
	if header == "" {
		return DefaultLang
	}

	parts := strings.Split(header, ",")
	if len(parts) == 0 {
		return DefaultLang
	}

	firstLang := strings.TrimSpace(parts[0])
	if idx := strings.Index(firstLang, ";"); idx > 0 {
		firstLang = firstLang[:idx]
	}

	return normalizeLang(firstLang)
}

func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))

	switch {
	case strings.HasPrefix(lang, "zh-tw"):
		return LangZhTW
	case strings.HasPrefix(lang, "zh"):
		return LangZhCN
	case strings.HasPrefix(lang, "en"):
		return LangEn
	case strings.HasPrefix(lang, "fr"):
		return LangFr
	case strings.HasPrefix(lang, "ja"):
		return LangJa
	case strings.HasPrefix(lang, "ru"):
		return LangRu
	case strings.HasPrefix(lang, "vi"):
		return LangVi
	default:
		return DefaultLang
	}
}

func SupportedLanguages() []string {
	return []string{LangZhCN, LangZhTW, LangEn, LangFr, LangJa, LangRu, LangVi}
}

func IsSupported(lang string) bool {
	lang = normalizeLang(lang)
	for _, supported := range SupportedLanguages() {
		if lang == supported {
			return true
		}
	}
	return false
}