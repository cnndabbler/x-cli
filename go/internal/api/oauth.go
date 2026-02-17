package api

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"x-cli/internal/config"
)

// percentEncode applies RFC 5849 percent-encoding.
func percentEncode(s string) string {
	return url.QueryEscape(s)
}

// generateOAuthHeader builds an OAuth 1.0a Authorization header (HMAC-SHA1).
func generateOAuthHeader(method, rawURL string, creds *config.Credentials) string {
	nonce := make([]byte, 16)
	_, _ = rand.Read(nonce)

	oauthParams := map[string]string{
		"oauth_consumer_key":     creds.APIKey,
		"oauth_nonce":            fmt.Sprintf("%x", nonce),
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        fmt.Sprintf("%d", time.Now().Unix()),
		"oauth_token":            creds.AccessToken,
		"oauth_version":          "1.0",
	}

	// Combine oauth params with query string params for signature base
	allParams := make(map[string]string)
	for k, v := range oauthParams {
		allParams[k] = v
	}

	parsed, _ := url.Parse(rawURL)
	if parsed.RawQuery != "" {
		qs := parsed.Query()
		for k := range qs {
			allParams[k] = qs.Get(k)
		}
	}

	// Sort and encode
	keys := make([]string, 0, len(allParams))
	for k := range allParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var paramParts []string
	for _, k := range keys {
		paramParts = append(paramParts, percentEncode(k)+"="+percentEncode(allParams[k]))
	}
	paramString := strings.Join(paramParts, "&")

	// Base URL (no query string)
	baseURL := fmt.Sprintf("%s://%s%s", parsed.Scheme, parsed.Host, parsed.Path)

	// Signature base string
	baseString := fmt.Sprintf("%s&%s&%s",
		strings.ToUpper(method),
		percentEncode(baseURL),
		percentEncode(paramString),
	)

	// Signing key
	signingKey := percentEncode(creds.APISecret) + "&" + percentEncode(creds.AccessTokenSecret)

	// HMAC-SHA1
	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(baseString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	oauthParams["oauth_signature"] = signature

	// Build header
	oauthKeys := make([]string, 0, len(oauthParams))
	for k := range oauthParams {
		oauthKeys = append(oauthKeys, k)
	}
	sort.Strings(oauthKeys)

	var headerParts []string
	for _, k := range oauthKeys {
		headerParts = append(headerParts, fmt.Sprintf(`%s="%s"`, percentEncode(k), percentEncode(oauthParams[k])))
	}

	return "OAuth " + strings.Join(headerParts, ", ")
}
