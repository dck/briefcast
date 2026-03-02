package worker

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Poller struct {
	DB         *sql.DB
	HTTPClient *http.Client
}

func NewPoller(db *sql.DB) *Poller {
	return &Poller{
		DB: db,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RSS XML structs

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Image       rssImage `xml:"image"`
	ItunesImage struct {
		Href string `xml:"href,attr"`
	} `xml:"http://www.itunes.apple.com/dtds/podcast-1.0.dtd image"`
	Items []rssItem `xml:"item"`
}

type rssImage struct {
	URL string `xml:"url"`
}

type rssItem struct {
	Title         string       `xml:"title"`
	Description   string       `xml:"description"`
	GUID          string       `xml:"guid"`
	PubDate       string       `xml:"pubDate"`
	Enclosure     rssEnclosure `xml:"enclosure"`
	Content       string       `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
	ItunesSummary string       `xml:"http://www.itunes.apple.com/dtds/podcast-1.0.dtd summary"`
}

type rssEnclosure struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Length string `xml:"length,attr"`
}

// Poll checks all podcasts for new episodes.
// For each podcast:
// 1. Fetch RSS feed
// 2. Parse episodes from feed
// 3. For each episode not already in DB (by guid):
//   - If no audio enclosure: insert with status=skipped, skip_reason="No audio enclosure"
//   - If published_at <= podcast.created_at: skip (no backfill)
//   - Otherwise: insert with status=pending, current_step=download
//
// 4. Update podcast.last_checked_at
// Errors on individual feeds are logged and skipped (don't fail the whole poll).
func (p *Poller) Poll() error {
	rows, err := p.DB.Query(`SELECT id, rss_url, title, description, image_url, created_at FROM podcasts`)
	if err != nil {
		return fmt.Errorf("querying podcasts: %w", err)
	}
	defer rows.Close()

	type podcast struct {
		id          string
		rssURL      string
		title       string
		description string
		imageURL    sql.NullString
		createdAt   time.Time
	}

	var podcasts []podcast
	for rows.Next() {
		var pc podcast
		if err := rows.Scan(&pc.id, &pc.rssURL, &pc.title, &pc.description, &pc.imageURL, &pc.createdAt); err != nil {
			return fmt.Errorf("scanning podcast row: %w", err)
		}
		podcasts = append(podcasts, pc)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating podcast rows: %w", err)
	}

	for _, pc := range podcasts {
		if err := p.pollFeed(pc.id, pc.rssURL, pc.title, pc.description, pc.imageURL.String, pc.createdAt); err != nil {
			log.Printf("error polling feed for podcast %s (%s): %v", pc.id, pc.rssURL, err)
		}
	}

	return nil
}

func (p *Poller) pollFeed(podcastID, rssURL, currentTitle, currentDesc, currentImageURL string, createdAt time.Time) error {
	feed, err := p.fetchFeed(rssURL)
	if err != nil {
		return fmt.Errorf("fetching feed: %w", err)
	}

	// Update podcast metadata if changed
	if err := p.updatePodcastMetadata(podcastID, currentTitle, currentDesc, currentImageURL, feed); err != nil {
		log.Printf("error updating metadata for podcast %s: %v", podcastID, err)
	}

	for _, item := range feed.Channel.Items {
		if err := p.processItem(podcastID, item, createdAt); err != nil {
			log.Printf("error processing item %q for podcast %s: %v", item.GUID, podcastID, err)
		}
	}

	// Update last_checked_at
	if _, err := p.DB.Exec(`UPDATE podcasts SET last_checked_at = $1 WHERE id = $2`, time.Now().UTC(), podcastID); err != nil {
		return fmt.Errorf("updating last_checked_at: %w", err)
	}

	return nil
}

func (p *Poller) fetchFeed(rssURL string) (*rssFeed, error) {
	req, err := http.NewRequest("GET", rssURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "Briefcast/1.0")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var feed rssFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("parsing RSS XML: %w", err)
	}

	return &feed, nil
}

func (p *Poller) updatePodcastMetadata(podcastID, currentTitle, currentDesc, currentImageURL string, feed *rssFeed) error {
	newTitle := strings.TrimSpace(feed.Channel.Title)
	newDesc := strings.TrimSpace(feed.Channel.Description)
	newImage := feedImageURL(feed)

	if newTitle == currentTitle && newDesc == currentDesc && newImage == currentImageURL {
		return nil
	}

	_, err := p.DB.Exec(
		`UPDATE podcasts SET title = $1, description = $2, image_url = $3 WHERE id = $4`,
		newTitle, newDesc, newImage, podcastID,
	)
	if err != nil {
		return fmt.Errorf("updating podcast metadata: %w", err)
	}
	log.Printf("updated metadata for podcast %s", podcastID)
	return nil
}

func feedImageURL(feed *rssFeed) string {
	if feed.Channel.ItunesImage.Href != "" {
		return feed.Channel.ItunesImage.Href
	}
	return feed.Channel.Image.URL
}

func (p *Poller) processItem(podcastID string, item rssItem, podcastCreatedAt time.Time) error {
	guid := item.GUID
	if guid == "" {
		guid = item.Enclosure.URL
	}
	if guid == "" {
		return fmt.Errorf("item has no GUID and no enclosure URL, skipping")
	}

	// Check if episode already exists
	var exists bool
	err := p.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM episodes WHERE podcast_id = $1 AND guid = $2)`,
		podcastID, guid,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("checking episode existence: %w", err)
	}
	if exists {
		return nil
	}

	audioURL := item.Enclosure.URL
	showNotes := itemShowNotes(item)
	publishedAt := parseRSSDate(item.PubDate)

	// No audio enclosure → insert as skipped
	if audioURL == "" {
		log.Printf("new episode (skipped, no audio): podcast=%s guid=%s title=%q", podcastID, guid, item.Title)
		_, err := p.DB.Exec(
			`INSERT INTO episodes (podcast_id, guid, title, description, audio_url, show_notes, status, skip_reason, published_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			podcastID, guid, item.Title, item.Description, "", showNotes, "skipped", "No audio enclosure", nullTime(publishedAt),
		)
		return err
	}

	// Skip episodes published before or at the podcast's created_at (no backfill)
	if publishedAt != nil && !publishedAt.After(podcastCreatedAt) {
		return nil
	}

	log.Printf("new episode: podcast=%s guid=%s title=%q", podcastID, guid, item.Title)
	_, err = p.DB.Exec(
		`INSERT INTO episodes (podcast_id, guid, title, description, audio_url, show_notes, status, current_step, published_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		podcastID, guid, item.Title, item.Description, audioURL, showNotes, "pending", "download", nullTime(publishedAt),
	)
	return err
}

func itemShowNotes(item rssItem) string {
	if item.Content != "" {
		return item.Content
	}
	if item.Description != "" {
		return item.Description
	}
	return item.ItunesSummary
}

// parseRSSDate tries multiple common RSS date formats.
func parseRSSDate(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	formats := []string{
		time.RFC1123Z,                    // Mon, 02 Jan 2006 15:04:05 -0700
		time.RFC1123,                     // Mon, 02 Jan 2006 15:04:05 MST
		time.RFC822Z,                     // 02 Jan 06 15:04 -0700
		time.RFC822,                      // 02 Jan 06 15:04 MST
		"Mon, 2 Jan 2006 15:04:05 -0700", // single-digit day
		"Mon, 2 Jan 2006 15:04:05 MST",
		"2006-01-02T15:04:05Z", // ISO 8601
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
		"02 Jan 2006 15:04:05 -0700", // without weekday
		"02 Jan 2006 15:04:05 MST",
	}

	for _, f := range formats {
		t, err := time.Parse(f, s)
		if err == nil {
			return &t
		}
	}

	log.Printf("unable to parse date: %q", s)
	return nil
}

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
