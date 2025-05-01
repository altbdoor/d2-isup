package main

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"
)

func GetRss(rssUrl string) (string, error) {
	var err error

	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		fmt.Sprintf("Chrome/135.0.%d.%d ", rand.IntN(9999), rand.IntN(99)) +
		"Safari/537.36"

	slog.Debug(userAgent)

	req, err := http.NewRequest("GET", rssUrl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http error %d", resp.StatusCode)
	}

	var rss Rss
	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(&rss)

	if err != nil {
		return "", fmt.Errorf("failed to parse xml: %w", err)
	}

	var filteredItems []RssItem
	slog.Info(fmt.Sprintf("found %d rss entries", len(rss.Channel.Items)))

	for _, item := range rss.Channel.Items {
		cleanedDesc := strings.ToLower(item.Description)
		cleanedDesc = strings.ReplaceAll(cleanedDesc, "bung.ie/destiny2help", "HELP_LINK")

		if strings.Contains(cleanedDesc, "destiny 2") || strings.Contains(cleanedDesc, "destiny2") {
			filteredItems = append(filteredItems, item)
		}
	}

	slog.Info(fmt.Sprintf("found %d filtered rss entries", len(filteredItems)))
	xmlData, err := xml.MarshalIndent(filteredItems, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal xml: %w", err)
	}

	content := string(xmlData)
	slog.Debug(content)
	return content, nil
}
