package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"golang.org/x/net/html"
)

const PAGE_URL = "https://help.bungie.net/hc/en-us/articles/360049199271-Destiny-Server-and-Update-Status"

const systemInstruction = `You will be receiving a HTML page content, which describes zero or more maintenance windows for the game Destiny 2.

Read and understand the HTML page content. Please pay attention to the provided timezone or time offset in the HTML table. Based on the content, please format a JSON data which describes the following:

- "maintenance_time_start": when the maintenance window starts
- "maintenance_time_end": when the maintenance window ends
- "server_down_start": when the servers are down, or when players can no longer play
- "server_down_end": when the servers are up, or when players can start playing
- "description": a simple description of what the downtime is about

If any of the date time values cannot be determined, the value will be set to a default of "1970-01-01T00:00:00Z".

Example output:

'''json
[
  {
    "maintenance_time_start": "2024-10-01T08:00:00-5:00",
    "maintenance_time_end": "2024-10-01T12:30:00-5:00",
    "server_down_start": "2024-10-01T09:00:00-5:00",
    "server_down_end": "2024-10-01T11:30:00-5:00",
    "description": "Downtime for server maintenance"
  },
  {
    "maintenance_time_start": "2024-10-13T13:00:00-8:00",
    "maintenance_time_end": "2024-10-13T18:00:00-8:00",
    "server_down_start": "1970-01-01T00:00:00Z",
    "server_down_end": "1970-01-01T00:00:00Z",
    "description": "Game will be brought down for maintenance"
  },
]
'''
`

type MaintenanceData struct {
	MaintenanceTimeStart time.Time `json:"maintenance_time_start"`
	MaintenanceTimeEnd   time.Time `json:"maintenance_time_end"`
	ServerDownStart      time.Time `json:"server_down_start"`
	ServerDownEnd        time.Time `json:"server_down_end"`
	Description          string    `json:"description"`
}

func removeAllAttributes(node *html.Node) {
	if node.Type == html.ElementNode {
		node.Attr = nil
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		removeAllAttributes(c)
	}
}

func main() {
	var err error

	// ========================================
	// get base dir
	// ========================================

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("unable to retrieve script path")
		os.Exit(1)
	}

	baseDir := filepath.Join(filepath.Dir(filename), "../")

	// ========================================
	// read and parse the html page
	// ========================================

	httpClient := &http.Client{
		Timeout: time.Minute,
		Transport: &http.Transport{
			// https://github.com/sweetbbak/go-cloudflare-bypass/blob/main/reqwest/reqwest.go
			// need to somehow set a tls config?
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
			},
		},
	}

	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		fmt.Sprintf("Chrome/130.0.0.%d Safari/537.%d", rand.Intn(9999), rand.Intn(99))

	finalPageUrl := os.Getenv("OVERRIDE_URL")
	if finalPageUrl == "" {
		finalPageUrl = PAGE_URL
	} else {
		log.Print("(i) overriding url")
	}

	req, _ := http.NewRequest("GET", finalPageUrl, nil)
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("(!) failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("(!) http error: %d", resp.StatusCode)
	}

	log.Print("(i) parsing document")
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()

	if err != nil {
		log.Fatalf("(!) failed to parse html: %v", err)
	}

	article := doc.Find("[itemprop='articleBody']").First()
	if article.Length() == 0 {
		log.Fatal("(!) failed to find article element")
	}

	log.Print("(i) stripping attributes")
	article.Each(func(i int, s *goquery.Selection) {
		for _, node := range s.Nodes {
			removeAllAttributes(node)
		}
	})

	content, _ := article.Html()
	content = strings.TrimSpace(content)

	// ========================================
	// send the html to gemini
	// ========================================

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("(!) failed to get GOOGLE_API_KEY")
	}

	aiClient := openai.NewClient(
		option.WithBaseURL("https://generativelanguage.googleapis.com/v1beta/"),
		option.WithAPIKey(apiKey),
	)

	ctx := context.Background()
	genResp, err := aiClient.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model:       openai.F("gemini-1.5-flash-002"),
			Temperature: openai.F(0.6),
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(
					strings.ReplaceAll(systemInstruction, "'''", "```"),
				),
				openai.UserMessage(content),
			}),

			// https://ai.google.dev/gemini-api/docs/openai
			// ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			// 	openai.ResponseFormatJSONObjectParam{},
			// ),
			// MaxTokens:   openai.F(int64(8000)),
		},
	)

	if err != nil {
		log.Fatalf("(!) failed to generate content in gemini: %v", err)
	}

	aiResponse := genResp.Choices[0].Message.Content

	// TODO: temp fix because gemini compat does not have response_format yet
	aiResponse = strings.TrimSpace(aiResponse)
	if strings.HasPrefix(aiResponse, "```") {
		aiResponse = strings.TrimPrefix(aiResponse, "```json")
		aiResponse = strings.TrimSuffix(aiResponse, "```")
	}

	maintenanceData := []MaintenanceData{}
	if err := json.Unmarshal([]byte(aiResponse), &maintenanceData); err != nil {
		log.Fatalf("(!) failed to json parse: %v", err)
	}

	outputPath := filepath.Join(baseDir, "./_site/data.json")
	outputBytes, _ := json.MarshalIndent(maintenanceData, "", "  ")
	os.WriteFile(outputPath, outputBytes, 0644)
	log.Print("(i) done")
}
