package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	// Parse flags
	url := flag.String("url", "", "url to get doc from")
	file := flag.String("file", "", "url to get doc from")
	query := flag.String("query", "", "doc html query")
	attr := flag.String("attr", "", "html element attribute")

	flag.Parse()
	if *url == "" && *file == "" {
		log.Fatal("url or file not provided")
	}
	if *query == "" {
		log.Fatal("query not provided")
	}

	// Create signal based context
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
			cancel()
		}
		signal.Stop(c)
	}()

	// Run bot
	if err := run(ctx, *file, *url, *query, *attr); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, file, url, query, attr string) error {
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: &transport{},
	}

	var reader io.Reader
	switch {
	case file != "":
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("couldn't open file: %w", err)
		}
		reader = f
	case url != "":
		r, err := client.Get(url)
		if err != nil {
			return fmt.Errorf("get request failed: %w", err)
		}
		if r.StatusCode != 200 {
			return fmt.Errorf("invalid status code: %s", r.Status)
		}
		reader = r.Body
		defer r.Body.Close()
	default:
		return fmt.Errorf("html source not provided")
	}

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return err
	}
	doc.Find(query).Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if attr != "" {
			if val, ok := s.Attr(attr); ok {
				text = val
			}
		}
		text = strings.TrimSpace(text)
		fmt.Println(text)
	})
	return nil
}

type transport struct{}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("cache-control", "max-age=0")
	r.Header.Set("rtt", "150")
	r.Header.Set("downlink", "10")
	r.Header.Set("ect", "4g")
	r.Header.Set("sec-ch-ua", `"Google Chrome";v="89", "Chromium";v="89", ";Not A Brand";v="99"`)
	r.Header.Set("sec-ch-ua-mobile", "?0")
	r.Header.Set("upgrade-insecure-requests", "1")
	r.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36")
	r.Header.Set("accept-language", "es-ES,es;q=0.9,en-US;q=0.8,en;q=0.7,eu;q=0.6,fr;q=0.5")
	return http.DefaultTransport.RoundTrip(r)
}
