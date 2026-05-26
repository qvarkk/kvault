package services

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"qvarkk/kvault/internal/domain"

	"github.com/PuerkitoBio/goquery"
	"github.com/jmoiron/sqlx"
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
}

type UrlTaskService struct {
	itemRepo   UrlTaskItemRepo
	transactor Transactor
	httpClient *http.Client
}

func NewUrlTaskService(itemRepo UrlTaskItemRepo, transactor Transactor) *UrlTaskService {
	return &UrlTaskService{
		itemRepo:   itemRepo,
		transactor: transactor,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *UrlTaskService) FetchAndExtract(ctx context.Context, userID, itemID string) error {
	return s.transactor.WithTx(ctx, func(tx *sqlx.Tx) error {
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

		meta, extractedText, fetchErr := s.fetchURL(item.SourceURL.String)
		if fetchErr != nil {
			return fetchErr
		}

		metaJSON, err := json.Marshal(meta)
		if err != nil {
			return err
		}

		item.UrlMetadata = NewNullString(string(metaJSON))
		item.ExtractedContent = NewNullString(extractedText)

		return s.itemRepo.UpdateUrlContentTx(ctx, tx, item)
	})
}

func (s *UrlTaskService) fetchURL(rawURL string) (UrlMetadata, string, error) {
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

	doc, err := goquery.NewDocumentFromReader(resp.Body)
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
	rawText := textNode.Text()
	words := strings.Fields(rawText)
	extractedText := strings.Join(words, " ")

	return meta, extractedText, nil
}
