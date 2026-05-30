package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"qvarkk/kvault/internal/domain"

	"github.com/PuerkitoBio/goquery"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

type UrlMetadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	SiteName    string `json:"site_name"`
}

type UrlTaskItemRepo interface {
	GetActiveByIDForUpdate(context.Context, *sqlx.Tx, string) (*domain.Item, error)
	UpdateUrlContentTx(context.Context, *sqlx.Tx, *domain.Item) error
	SetUrlStatusTx(ctx context.Context, tx *sqlx.Tx, itemID, status string) error
}

type UrlTaskService struct {
	itemRepo   UrlTaskItemRepo
	transactor Transactor
	cache      CacheStore
	httpClient *http.Client
}

func NewUrlTaskService(itemRepo UrlTaskItemRepo, transactor Transactor, cache CacheStore) *UrlTaskService {
	return &UrlTaskService{
		itemRepo:   itemRepo,
		transactor: transactor,
		cache:      cache,
		httpClient: newGuardedHTTPClient(10 * time.Second),
	}
}

func (s *UrlTaskService) FetchAndExtract(ctx context.Context, userID, itemID string) error {
	// Phase 1: authorize and read the source URL. The lock is released before
	// the (slow, external) HTTP fetch so we never hold a row lock across the network.
	var sourceURL string
	err := s.transactor.WithTx(ctx, func(tx *sqlx.Tx) error {
		item, err := s.itemRepo.GetActiveByIDForUpdate(ctx, tx, itemID)
		if err != nil {
			return NewServiceError(ErrItemNotFound, "not found", err)
		}
		if item.UserID != userID {
			return NewServiceError(ErrItemNotFound, "forbidden", nil)
		}
		if !item.SourceURL.Valid || item.SourceURL.String == "" {
			return NewServiceError(ErrInternal, "item has no source_url", nil)
		}
		sourceURL = item.SourceURL.String
		return nil
	})
	if err != nil {
		return err
	}

	meta, extractedText, fetchErr := s.fetchURL(sourceURL)
	if fetchErr != nil {
		s.markUrlStatus(ctx, itemID, domain.UrlStatusError)
		invalidateSingleCache(ctx, s.cache, itemKey(itemID))
		invalidateListCache(ctx, s.cache, itemListVersionKey(userID))
		return fetchErr
	}

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	// Phase 2: re-lock the row and write the extracted content (sets status=ready).
	err = s.transactor.WithTx(ctx, func(tx *sqlx.Tx) error {
		item, err := s.itemRepo.GetActiveByIDForUpdate(ctx, tx, itemID)
		if err != nil {
			return NewServiceError(ErrItemNotFound, "not found", err)
		}
		if item.UserID != userID {
			return NewServiceError(ErrItemNotFound, "forbidden", nil)
		}
		item.UrlMetadata = NewNullString(string(metaJSON))
		item.ExtractedContent = NewNullString(extractedText)
		return s.itemRepo.UpdateUrlContentTx(ctx, tx, item)
	})
	if err != nil {
		return err
	}

	invalidateSingleCache(ctx, s.cache, itemKey(itemID))
	invalidateListCache(ctx, s.cache, itemListVersionKey(userID))

	return nil
}

// markUrlStatus best-effort updates url_status; failures are logged, not fatal.
func (s *UrlTaskService) markUrlStatus(ctx context.Context, itemID string, status domain.UrlStatus) {
	err := s.transactor.WithTx(ctx, func(tx *sqlx.Tx) error {
		return s.itemRepo.SetUrlStatusTx(ctx, tx, itemID, string(status))
	})
	if err != nil {
		zap.L().Warn("failed to set url_status",
			zap.String("item_id", itemID),
			zap.String("status", string(status)),
			zap.Error(err),
		)
	}
}

func (s *UrlTaskService) fetchURL(rawURL string) (UrlMetadata, string, error) {
	if _, err := validatePublicURL(rawURL); err != nil {
		return UrlMetadata{}, "", err
	}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return UrlMetadata{}, "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; kvault/1.0)")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return UrlMetadata{}, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return UrlMetadata{}, "", fmt.Errorf("fetch returned status %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "" &&
		!strings.Contains(ct, "text/html") && !strings.Contains(ct, "application/xhtml") {
		return UrlMetadata{}, "", fmt.Errorf("unsupported content-type %q", ct)
	}

	// Cap how much we read so a huge/streaming response can't exhaust memory.
	limited := io.LimitReader(resp.Body, maxFetchBodyBytes)
	doc, err := goquery.NewDocumentFromReader(limited)
	if err != nil {
		return UrlMetadata{}, "", err
	}

	meta := UrlMetadata{}

	// og:title > <title>
	if og := doc.Find(`meta[property="og:title"]`).AttrOr("content", ""); og != "" {
		meta.Title = og
	} else {
		meta.Title = strings.TrimSpace(doc.Find("title").Text())
	}

	// og:description > meta[name=description]
	if og := doc.Find(`meta[property="og:description"]`).AttrOr("content", ""); og != "" {
		meta.Description = og
	} else {
		meta.Description = doc.Find(`meta[name="description"]`).AttrOr("content", "")
	}

	meta.Image = doc.Find(`meta[property="og:image"]`).AttrOr("content", "")
	meta.SiteName = doc.Find(`meta[property="og:site_name"]`).AttrOr("content", "")

	// Extract body text: <article> > <main> > <body>
	var textNode *goquery.Selection
	if a := doc.Find("article"); a.Length() > 0 {
		textNode = a.First()
	} else if m := doc.Find("main"); m.Length() > 0 {
		textNode = m.First()
	} else {
		textNode = doc.Find("body")
	}

	textNode.Find("script, style, noscript").Remove()

	// goquery's .Text() concatenates adjacent elements with no separator
	// (e.g. "<p>foo</p><p>bar</p>" -> "foobar"), which fuses distinct words and
	// hurts search. Collect text node-by-node with spaces between them instead.
	var rawText string
	if len(textNode.Nodes) > 0 {
		rawText = nodeText(textNode.Nodes[0])
	}
	words := strings.Fields(stripNullBytes(rawText))
	extractedText := strings.Join(words, " ")

	return meta, extractedText, nil
}

// nodeText gathers all descendant text, appending a space after each text node
// so adjacent elements' words stay separated. strings.Fields later collapses
// the redundant whitespace.
func nodeText(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			sb.WriteString(node.Data)
			sb.WriteByte(' ')
			return
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return sb.String()
}
