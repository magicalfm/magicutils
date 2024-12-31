package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"text/template"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/sourcegraph/conc/pool"
)

const (
	rssLink         = "https://rss.listen.style/p/magicalfm/rss"
	episodesPerFile = 20 // 1ファイルあたりのエピソード数
)

type Item struct {
	Title       string
	Description string
	Link        string
	Transcript  string
	Published   *time.Time
}

func getRssFeed(url string) (*gofeed.Feed, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}

	return feed, nil
}

func fetchTranscript(ctx context.Context, itemLink string) (string, error) {
	transcriptLink, err := url.JoinPath(itemLink, "transcript.vtt")
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", transcriptLink, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func constructItems(ctx context.Context, feed *gofeed.Feed) ([]Item, error) {
	p := pool.New().WithErrors().WithContext(ctx).WithMaxGoroutines(10)
	items := make([]Item, len(feed.Items))
	for i, item := range feed.Items {
		i, item := i, item

		p.Go(func(ctx context.Context) error {
			fmt.Println("fetching transcript for ", item.Title)
			transcript, err := fetchTranscript(ctx, item.Link)
			if err != nil {
				return err
			}

			items[i] = Item{
				Title:       item.Title,
				Description: item.Content,
				Link:        item.Link,
				Transcript:  transcript,
				Published:   item.PublishedParsed,
			}

			return nil
		})
	}
	if err := p.Wait(); err != nil {
		return nil, err
	}
	return items, nil
}

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func realMain() error {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	feed, err := getRssFeed(rssLink)
	if err != nil {
		return err
	}

	tmpl, err := template.ParseFiles("markdown.tmpl")
	if err != nil {
		return err
	}

	items, err := constructItems(ctx, feed)
	if err != nil {
		return err
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Published.Before(*items[j].Published)
	})

	dir := "output"
	if err := os.RemoveAll("output"); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	for i := 0; i < len(items); i += episodesPerFile {
		end := i + episodesPerFile
		if end > len(items) {
			end = len(items)
		}

		chunk := items[i:end]
		buf := bytes.NewBuffer(nil)
		if err := tmpl.Execute(buf, chunk); err != nil {
			return err
		}

		filename := fmt.Sprintf("%s/transcript_%d.md", dir, i/episodesPerFile+1)
		if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
			return err
		}
	}

	return nil
}
