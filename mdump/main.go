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
	"text/template"

	"github.com/mmcdole/gofeed"
	"github.com/sourcegraph/conc/pool"
)

const rssLink = "https://rss.listen.style/p/magicalfm/rss"

type Item struct {
	Title       string
	Description string
	Link        string
	Transcript  string
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

	buf := bytes.NewBuffer(nil)
	if err := tmpl.Execute(buf, items); err != nil {
		return err
	}

	if err := os.WriteFile("dist/output.md", buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}
