package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"x-cli/internal/config"
)

const apiBase = "https://api.x.com/2"

// RateLimitError is returned when the API returns HTTP 429.
type RateLimitError struct {
	ResetAt string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limited, resets at %s", e.ResetAt)
}

// Client wraps HTTP calls to the X API v2.
type Client struct {
	creds  *config.Credentials
	http   *http.Client
	userID string
}

// NewClient creates a new API client.
func NewClient(creds *config.Credentials) *Client {
	return &Client{
		creds: creds,
		http:  &http.Client{Timeout: 30 * time.Second},
	}
}

// --- internal ---

func (c *Client) bearerGet(rawURL string) (map[string]any, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.creds.BearerToken)
	return c.doRequest(req)
}

func (c *Client) oauthRequest(method, rawURL string, jsonBody map[string]any) (map[string]any, error) {
	var body io.Reader
	if jsonBody != nil {
		b, err := json.Marshal(jsonBody)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, rawURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", generateOAuthHeader(method, rawURL, c.creds))
	if jsonBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.doRequest(req)
}

func (c *Client) doRequest(req *http.Request) (map[string]any, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == 429 {
		reset := resp.Header.Get("x-rate-limit-reset")
		if reset == "" {
			reset = "unknown"
		}
		return nil, &RateLimitError{ResetAt: reset}
	}

	var data map[string]any
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return nil, fmt.Errorf("JSON decode error: %w (body: %s)", err, string(bodyBytes[:min(len(bodyBytes), 500)]))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, extractAPIError(data, bodyBytes, resp.StatusCode)
	}

	// X API sometimes returns 200 with errors and no data
	if _, hasErrors := data["errors"]; hasErrors {
		if _, hasData := data["data"]; !hasData {
			return nil, extractAPIError(data, bodyBytes, resp.StatusCode)
		}
	}

	return data, nil
}

func extractAPIError(data map[string]any, bodyBytes []byte, statusCode int) error {
	errors, _ := data["errors"].([]any)
	var msgs []string
	for _, e := range errors {
		em, _ := e.(map[string]any)
		if d, ok := em["detail"].(string); ok && d != "" {
			msgs = append(msgs, d)
		} else if m, ok := em["message"].(string); ok && m != "" {
			msgs = append(msgs, m)
		}
	}
	msg := strings.Join(msgs, "; ")
	if msg == "" {
		msg = string(bodyBytes[:min(len(bodyBytes), 500)])
	}
	if statusCode == 0 || statusCode == 200 {
		return fmt.Errorf("API error: %s", msg)
	}
	return fmt.Errorf("API error (HTTP %d): %s", statusCode, msg)
}

// GetAuthenticatedUserID returns the authenticated user's ID (cached).
func (c *Client) GetAuthenticatedUserID() (string, error) {
	if c.userID != "" {
		return c.userID, nil
	}
	data, err := c.oauthRequest("GET", apiBase+"/users/me", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get authenticated user: %w", err)
	}
	d, _ := data["data"].(map[string]any)
	id, _ := d["id"].(string)
	c.userID = id
	return id, nil
}

// --- tweets ---

func (c *Client) PostTweet(text string, replyTo string, quoteTweetID string, pollOptions []string, pollDuration int) (map[string]any, error) {
	body := map[string]any{"text": text}
	if replyTo != "" {
		body["reply"] = map[string]any{"in_reply_to_tweet_id": replyTo}
	}
	if quoteTweetID != "" {
		body["quote_tweet_id"] = quoteTweetID
	}
	if len(pollOptions) > 0 {
		body["poll"] = map[string]any{"options": pollOptions, "duration_minutes": pollDuration}
	}
	return c.oauthRequest("POST", apiBase+"/tweets", body)
}

func (c *Client) DeleteTweet(tweetID string) (map[string]any, error) {
	return c.oauthRequest("DELETE", apiBase+"/tweets/"+tweetID, nil)
}

func (c *Client) GetTweet(tweetID string) (map[string]any, error) {
	params := url.Values{
		"tweet.fields": {"created_at,public_metrics,author_id,conversation_id,in_reply_to_user_id,referenced_tweets,attachments,entities,lang,note_tweet,article"},
		"expansions":   {"author_id,referenced_tweets.id,attachments.media_keys"},
		"user.fields":  {"name,username,verified,profile_image_url,public_metrics"},
		"media.fields": {"url,preview_image_url,type,width,height,alt_text"},
	}
	return c.bearerGet(apiBase + "/tweets/" + tweetID + "?" + params.Encode())
}

func (c *Client) SearchTweets(query string, maxResults int) (map[string]any, error) {
	if maxResults < 10 {
		maxResults = 10
	}
	if maxResults > 100 {
		maxResults = 100
	}
	params := url.Values{
		"query":        {query},
		"max_results":  {strconv.Itoa(maxResults)},
		"tweet.fields": {"created_at,public_metrics,author_id,conversation_id,entities,lang,note_tweet"},
		"expansions":   {"author_id,attachments.media_keys"},
		"user.fields":  {"name,username,verified,profile_image_url"},
		"media.fields": {"url,preview_image_url,type"},
	}
	return c.bearerGet(apiBase + "/tweets/search/recent?" + params.Encode())
}

func (c *Client) GetTweetMetrics(tweetID string) (map[string]any, error) {
	params := url.Values{
		"tweet.fields": {"public_metrics,non_public_metrics,organic_metrics"},
	}
	return c.oauthRequest("GET", apiBase+"/tweets/"+tweetID+"?"+params.Encode(), nil)
}

// --- users ---

func (c *Client) GetUser(username string) (map[string]any, error) {
	params := url.Values{
		"user.fields": {"created_at,description,public_metrics,verified,profile_image_url,url,location,pinned_tweet_id"},
	}
	return c.bearerGet(apiBase + "/users/by/username/" + username + "?" + params.Encode())
}

func (c *Client) GetTimeline(userID string, maxResults int, sinceID string) (map[string]any, error) {
	if maxResults < 5 {
		maxResults = 5
	}
	if maxResults > 100 {
		maxResults = 100
	}
	params := url.Values{
		"max_results":  {strconv.Itoa(maxResults)},
		"tweet.fields": {"created_at,public_metrics,author_id,conversation_id,entities,lang,note_tweet"},
		"expansions":   {"author_id,attachments.media_keys,referenced_tweets.id"},
		"user.fields":  {"name,username,verified"},
		"media.fields": {"url,preview_image_url,type"},
	}
	if sinceID != "" {
		params.Set("since_id", sinceID)
	}
	return c.bearerGet(apiBase + "/users/" + userID + "/tweets?" + params.Encode())
}

func (c *Client) GetFollowers(userID string, maxResults int) (map[string]any, error) {
	if maxResults < 1 {
		maxResults = 1
	}
	if maxResults > 1000 {
		maxResults = 1000
	}
	params := url.Values{
		"max_results": {strconv.Itoa(maxResults)},
		"user.fields": {"created_at,description,public_metrics,verified,profile_image_url"},
	}
	return c.bearerGet(apiBase + "/users/" + userID + "/followers?" + params.Encode())
}

func (c *Client) GetFollowing(userID string, maxResults int) (map[string]any, error) {
	if maxResults < 1 {
		maxResults = 1
	}
	if maxResults > 1000 {
		maxResults = 1000
	}
	params := url.Values{
		"max_results": {strconv.Itoa(maxResults)},
		"user.fields": {"created_at,description,public_metrics,verified,profile_image_url"},
	}
	return c.bearerGet(apiBase + "/users/" + userID + "/following?" + params.Encode())
}

func (c *Client) GetMentions(maxResults int) (map[string]any, error) {
	userID, err := c.GetAuthenticatedUserID()
	if err != nil {
		return nil, err
	}
	if maxResults < 5 {
		maxResults = 5
	}
	if maxResults > 100 {
		maxResults = 100
	}
	params := url.Values{
		"max_results":  {strconv.Itoa(maxResults)},
		"tweet.fields": {"created_at,public_metrics,author_id,conversation_id,entities,note_tweet"},
		"expansions":   {"author_id"},
		"user.fields":  {"name,username,verified"},
	}
	u := apiBase + "/users/" + userID + "/mentions?" + params.Encode()
	return c.oauthRequest("GET", u, nil)
}

// --- engagement ---

func (c *Client) LikeTweet(tweetID string) (map[string]any, error) {
	userID, err := c.GetAuthenticatedUserID()
	if err != nil {
		return nil, err
	}
	return c.oauthRequest("POST", apiBase+"/users/"+userID+"/likes", map[string]any{"tweet_id": tweetID})
}

func (c *Client) UnlikeTweet(tweetID string) (map[string]any, error) {
	userID, err := c.GetAuthenticatedUserID()
	if err != nil {
		return nil, err
	}
	return c.oauthRequest("DELETE", apiBase+"/users/"+userID+"/likes/"+tweetID, nil)
}

func (c *Client) Retweet(tweetID string) (map[string]any, error) {
	userID, err := c.GetAuthenticatedUserID()
	if err != nil {
		return nil, err
	}
	return c.oauthRequest("POST", apiBase+"/users/"+userID+"/retweets", map[string]any{"tweet_id": tweetID})
}

func (c *Client) Unretweet(tweetID string) (map[string]any, error) {
	userID, err := c.GetAuthenticatedUserID()
	if err != nil {
		return nil, err
	}
	return c.oauthRequest("DELETE", apiBase+"/users/"+userID+"/retweets/"+tweetID, nil)
}

// --- bookmarks ---

func (c *Client) GetBookmarks(maxResults int) (map[string]any, error) {
	userID, err := c.GetAuthenticatedUserID()
	if err != nil {
		return nil, err
	}
	if maxResults < 1 {
		maxResults = 1
	}
	if maxResults > 100 {
		maxResults = 100
	}
	params := url.Values{
		"max_results":  {strconv.Itoa(maxResults)},
		"tweet.fields": {"created_at,public_metrics,author_id,conversation_id,entities,lang,note_tweet"},
		"expansions":   {"author_id,attachments.media_keys"},
		"user.fields":  {"name,username,verified,profile_image_url"},
		"media.fields": {"url,preview_image_url,type"},
	}
	u := apiBase + "/users/" + userID + "/bookmarks?" + params.Encode()
	return c.oauthRequest("GET", u, nil)
}

func (c *Client) BookmarkTweet(tweetID string) (map[string]any, error) {
	userID, err := c.GetAuthenticatedUserID()
	if err != nil {
		return nil, err
	}
	return c.oauthRequest("POST", apiBase+"/users/"+userID+"/bookmarks", map[string]any{"tweet_id": tweetID})
}

func (c *Client) UnbookmarkTweet(tweetID string) (map[string]any, error) {
	userID, err := c.GetAuthenticatedUserID()
	if err != nil {
		return nil, err
	}
	return c.oauthRequest("DELETE", apiBase+"/users/"+userID+"/bookmarks/"+tweetID, nil)
}
