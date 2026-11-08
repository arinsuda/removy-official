package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port       string
	AppBaseURL string
	CORS       CORSConfig
	DB         DBConfig
	TMDB       TMDBConfig
	OMDb       OMDbConfig
	JWT        JWTConfig
	Brevo      BrevoConfig
	Cookie     CookieConfig
	MinIO      MinIOConfig
	Google     GoogleConfig
	Link       LinkConfig

	// TrustedProxyHops is how many proxies append to X-Forwarded-For (Render = 1).
	TrustedProxyHops int
}

// LinkConfig configures PLANZI <-> POPCORN account linking. Empty Secrets disables the feature.
type LinkConfig struct {
	Secrets      []string // first signs, all verify
	RemoteAPIURL string
	RemoteWebURL string
	SelfPlatform string
}

func (l LinkConfig) Enabled() bool { return len(l.Secrets) > 0 }

type GoogleConfig struct {
	Enabled            bool
	ClientID           string
	ClientSecret       string
	RedirectURL        string
	FrontendSuccessURL string
	FrontendErrorURL   string
}

type CORSConfig struct {
	AllowedOrigins []string
}

// IsAllowed reports whether origin exactly matches the allowlist.
func (c CORSConfig) IsAllowed(origin string) bool {
	return slices.Contains(c.AllowedOrigins, origin)
}

type DBConfig struct {
	Host                        string
	User                        string
	Password                    string
	Name                        string
	Port                        string
	SSLMode                     string
	MigrationLockTimeoutMs      int
	MigrationStatementTimeoutMs int
}

type TMDBConfig struct {
	BaseURL string
	APIKey  string
}

// OMDbConfig powers IMDb and Rotten Tomatoes ratings. An empty APIKey disables the feature.
type OMDbConfig struct {
	APIKey string
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type BrevoConfig struct {
	APIKey      string
	SenderEmail string
	SenderName  string
	Enabled     bool
	// DevLog prints emails to stdout instead of sending them. Local dev only; see DevMailLogAllowed.
	DevLog bool
}

type CookieConfig struct {
	Domain   string
	Secure   bool
	SameSite string
}

type MinIOConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool
}

func Load() (*Config, error) {
	migrationLockTimeoutMs, _ := strconv.Atoi(getEnv("MIGRATION_LOCK_TIMEOUT_MS", "10000"))
	migrationStatementTimeoutMs, _ := strconv.Atoi(getEnv("MIGRATION_STATEMENT_TIMEOUT_MS", "15000"))

	brevoAPIKey := getEnv("BREVO_API_KEY", "")

	allowedOrigins, err := parseAllowedOrigins(getEnv("CORS_ALLOWED_ORIGIN", defaultAllowedOrigins))
	if err != nil {
		return nil, err
	}

	devMailLog := false
	if getEnv("DEV_MAIL_LOG", "") == "true" {
		// Secure cookies mean HTTPS, i.e. a deployed environment: never print mail there.
		if getEnv("COOKIE_SECURE", "false") == "true" {
			log.Printf("WARN DEV_MAIL_LOG ignored: COOKIE_SECURE=true")
		} else if DevMailLogAllowed(allowedOrigins, getEnv("APP_BASE_URL", "http://localhost:3000"), getEnv("FRONTEND_BASE_URL", "http://localhost:3000")) {
			devMailLog = true
			log.Printf("DEV_MAIL_LOG active — emails are printed, not sent")
		} else {
			log.Printf("WARN DEV_MAIL_LOG ignored: non-local origin")
		}
	}

	trustedProxyHops, _ := strconv.Atoi(getEnv("TRUSTED_PROXY_HOPS", "1"))
	if trustedProxyHops < 1 {
		trustedProxyHops = 1
	}

	cfg := &Config{
		TrustedProxyHops: trustedProxyHops,
		Port:             getEnv("PORT", "8080"),
		AppBaseURL:       getEnv("APP_BASE_URL", "http://localhost:3000"),
		CORS: CORSConfig{
			AllowedOrigins: allowedOrigins,
		},
		DB: DBConfig{
			Host:                        requireEnv("POSTGRES_HOST"),
			User:                        requireEnv("POSTGRES_USER"),
			Password:                    requireEnv("POSTGRES_PASSWORD"),
			Name:                        requireEnv("POSTGRES_DB"),
			Port:                        getEnv("POSTGRES_PORT", "5432"),
			SSLMode:                     getEnv("POSTGRES_SSLMODE", ""),
			MigrationLockTimeoutMs:      migrationLockTimeoutMs,
			MigrationStatementTimeoutMs: migrationStatementTimeoutMs,
		},
		TMDB: TMDBConfig{
			BaseURL: requireEnv("THE_MOVIE_BASE_API"),
			APIKey:  requireEnv("THE_MOVIE_API_KEY"),
		},
		OMDb: OMDbConfig{APIKey: os.Getenv("OMDB_API_KEY")},
		JWT: JWTConfig{
			AccessSecret:  requireEnv("JWT_ACCESS_SECRET"),
			RefreshSecret: requireEnv("JWT_REFRESH_SECRET"),
			AccessTTL:     15 * time.Minute,
			RefreshTTL:    7 * 24 * time.Hour,
		},
		Brevo: BrevoConfig{
			APIKey:      brevoAPIKey,
			SenderEmail: getEnv("BREVO_SENDER_EMAIL", getEnv("ADMIN_EMAIL", "removy.official@gmail.com")),
			SenderName:  getEnv("BREVO_SENDER_NAME", "POPCORN"),
			Enabled:     brevoAPIKey != "" || devMailLog,
			DevLog:      devMailLog,
		},
		Cookie: CookieConfig{
			Domain:   getEnv("COOKIE_DOMAIN", ""),
			Secure:   getEnv("COOKIE_SECURE", "false") == "true",
			SameSite: getEnv("COOKIE_SAMESITE", "Lax"),
		},
		MinIO: MinIOConfig{
			Endpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey:  getEnv("MINIO_ROOT_USER", "minioadmin"),
			SecretKey:  getEnv("MINIO_ROOT_PASSWORD", "minioadmin"),
			BucketName: getEnv("MINIO_BUCKET_NAME", "removy-private"),
			UseSSL:     getEnv("MINIO_USE_SSL", "false") == "true",
		},
		Google: GoogleConfig{
			Enabled:            getEnv("GOOGLE_OAUTH_ENABLED", "false") == "true",
			ClientID:           getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret:       getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:        getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback"),
			FrontendSuccessURL: getEnv("GOOGLE_FRONTEND_SUCCESS_URL", "http://localhost:5173/"),
			FrontendErrorURL:   getEnv("GOOGLE_FRONTEND_ERROR_URL", "http://localhost:5173/login"),
		},
		Link: LinkConfig{
			Secrets:      splitSecrets(getEnv("LINK_SHARED_SECRET", "")),
			RemoteAPIURL: strings.TrimRight(getEnv("LINK_REMOTE_API_URL", ""), "/"),
			RemoteWebURL: strings.TrimRight(getEnv("LINK_REMOTE_WEB_URL", ""), "/"),
			SelfPlatform: getEnv("LINK_SELF_PLATFORM", "popcorn"),
		},
	}

	if cfg.Google.Enabled {
		if cfg.Google.ClientID == "" || cfg.Google.ClientSecret == "" || cfg.Google.RedirectURL == "" || cfg.Google.FrontendSuccessURL == "" || cfg.Google.FrontendErrorURL == "" {
			return nil, fmt.Errorf("GOOGLE_OAUTH_ENABLED is true but missing required Google OAuth configuration variables")
		}
	}

	return cfg, nil
}

func (d DBConfig) DSN() string {
	sslMode := d.SSLMode
	if sslMode == "" {
		if d.Host == "localhost" || d.Host == "127.0.0.1" {
			sslMode = "disable"
		} else {
			sslMode = "require"
		}
	}
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Bangkok",
		d.Host, d.User, d.Password, d.Name, d.Port, sslMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

// defaultAllowedOrigins covers the local Vite dev server (:3000) and the legacy :5173 setup.
const defaultAllowedOrigins = "http://localhost:3000,http://localhost:5173"

// parseAllowedOrigins parses a comma-separated allowlist. Wildcards are rejected
// because every CORS consumer here allows credentials.
func parseAllowedOrigins(raw string) ([]string, error) {
	var origins []string
	for _, entry := range strings.Split(raw, ",") {
		origin := strings.TrimRight(strings.TrimSpace(entry), "/")
		if origin == "" {
			continue
		}
		u, err := url.Parse(origin)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
			return nil, fmt.Errorf("CORS_ALLOWED_ORIGIN: invalid origin %q (want scheme://host[:port], no wildcard)", entry)
		}
		if !slices.Contains(origins, origin) {
			origins = append(origins, origin)
		}
	}
	if len(origins) == 0 {
		return nil, fmt.Errorf("CORS_ALLOWED_ORIGIN: no origins configured")
	}
	return origins, nil
}

func splitSecrets(raw string) []string {
	var out []string
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if len(p) < 32 {
			// secrets shorter than 32 bytes are rejected; none left => feature disabled
			log.Println("WARN: LINK_SHARED_SECRET entry shorter than 32 bytes ignored")
			continue
		}
		out = append(out, p)
	}
	return out
}

// DevMailLogAllowed is the safety gate for DEV_MAIL_LOG: true only when there is at least one
// origin and every CORS origin and every link base URL is a loopback host. Any empty,
// unparsable or non-local value disables it.
func DevMailLogAllowed(origins []string, baseURLs ...string) bool {
	all := append(append([]string{}, origins...), baseURLs...)
	if len(origins) == 0 || len(baseURLs) == 0 {
		return false
	}
	for _, raw := range all {
		u, err := url.Parse(strings.TrimSpace(raw))
		if err != nil || u.Host == "" {
			return false
		}
		switch u.Hostname() {
		case "localhost", "127.0.0.1", "::1":
		default:
			return false
		}
	}
	return true
}
