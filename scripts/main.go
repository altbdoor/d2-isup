package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const PAGE_URL = "https://bsky.app/profile/bungiehelp.bungie.net/rss"

const systemInstruction = `You are an expert data parser specializing in extracting information from XML RSS feeds, specifically the Destiny 2 Twitter feed.

Your _sole_ task is to identify and parse entries that _explicitly_ announce _upcoming_, _scheduled_ Destiny 2 maintenance that impacts gameplay or access. You will ignore _all_ other entries.

**Here's how you should operate:**

1.  **Input:** You will receive XML data representing an RSS feed from the Destiny 2 Twitter account.

2.  **Strict Filtering (Critical):** You _must_ implement very strict filtering criteria. An entry _must_ meet _all_ of the following conditions to be considered relevant:

    - **Keywords (Required):** The "<description>" _must_ contain _at least one_ of the following phrases, used in the context of announcing _future_ maintenance:
      - "Upcoming maintenance"
      - "Scheduled maintenance"
      - "Maintenance scheduled"
      - "Downtime scheduled"
    - **Time Indicator (Required):** The "<description>" _must_ contain a clear indicator of _future_ time, such as:
      - "on [Date/Time]"
      - "starting [Date/Time]"
      - "at [Time]" (in conjunction with a date elsewhere in the title/description)
      - "tomorrow"
      - "next [Day of the Week]"
    - **Negative Constraints (Critical):** The following conditions _must not_ be present for an entry to be considered relevant:
      - Mentions of "known issues," "investigating," "issue," "bug," "hotfix," "patch," or "update" (unless _explicitly_ tied to a _scheduled_ maintenance period announced for the _future_).
      - Any language indicating past or ongoing maintenance.
      - Any language that pertains to events that have already occurred.
    - **If _any_ of the negative constraints are met, _immediately_ reject the entry.**

3.  **Extraction:** If, and _only if_, an entry passes _all_ the filtering criteria above, extract the following information:

    - "maintenance_time_start": when the maintenance window starts
    - "maintenance_time_end": when the maintenance window ends
    - "server_down_start": when the servers are down, or when players can no longer play
    - "server_down_end": when the servers are up, or when players can start playing
    - "description": a short and simple description of what the downtime is about

4.  **Parsing and Structuring:** Analyze the "<description>" to determine:

    - **Date and time**: Please pay attention to the provided timezone or time offset in the description. If any of the date time values cannot be determined, the value will be set to a default of "1970-01-01T00:00:00Z".

5.  **JSON Conversion:** Convert the extracted and parsed information into a JSON object with the following structure:

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

6.  **Output:** Return a single, valid JSON string containing an array of upcoming maintenance announcements. If no _upcoming_ maintenance announcements are found that meet the strict filtering criteria, return an empty JSON array: "[]".
`

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
		Timeout: time.Second * 10,
	}

	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		fmt.Sprintf("Chrome/130.0.0.%d Safari/537.%d", rand.Intn(9999), rand.Intn(99))

	req, _ := http.NewRequest("GET", PAGE_URL, nil)
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("(!) failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("(!) http error: %d", resp.StatusCode)
	}

	log.Println("(i) parsing document")
	var rss Rss
	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(&rss)

	if err != nil {
		log.Fatalf("(!) failed to parse xml: %v", err)
	}

	xmlData, err := xml.MarshalIndent(rss.Channel.Items, "", "  ")
	if err != nil {
		log.Fatalf("(!) failed to marshal xml: %v", err)
	}

	content := string(xmlData)

	// ========================================
	// send the xml to gemini
	// ========================================

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatalln("(!) failed to get GOOGLE_API_KEY")
	}

	aiClient := openai.NewClient(
		option.WithBaseURL("https://generativelanguage.googleapis.com/v1beta/openai/"),
		option.WithAPIKey(apiKey),
	)

	ctx := context.Background()
	var genResp *openai.ChatCompletion
	attempts := 0
	maxAttempts := 5

	for attempts < maxAttempts {
		genResp, err = aiClient.Chat.Completions.New(
			ctx,
			openai.ChatCompletionNewParams{
				Model:       openai.String("gemini-2.0-flash"),
				Temperature: openai.Float(0.6),
				MaxTokens:   openai.Int(8000),
				N:           openai.Int(1),
				Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
					openai.SystemMessage(
						strings.ReplaceAll(systemInstruction, "'''", "```"),
					),
					openai.UserMessage(content),
				}),
				ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
					openai.ResponseFormatJSONObjectParam{
						Type: openai.F(openai.ResponseFormatJSONObjectTypeJSONObject),
					},
				),
			},
		)

		if err == nil {
			break
		}

		attempts++
		log.Printf("(!) attempt %d, failed to generate content in gemini: %v\n", attempts, err)
		log.Println("(!) retrying...")
		time.Sleep(2 * time.Second)
	}

	if attempts == maxAttempts {
		log.Fatalln("(!) max retries reached")
	}

	aiResponse := genResp.Choices[0].Message.Content

	maintenanceData := []MaintenanceData{}
	if err := json.Unmarshal([]byte(aiResponse), &maintenanceData); err != nil {
		log.Fatalf("(!) failed to json parse: %v\n(!) data is: %v", err, aiResponse)
	}

	outputPath := filepath.Join(baseDir, "./_site/data.json")
	outputBytes, _ := json.MarshalIndent(maintenanceData, "", "  ")
	os.WriteFile(outputPath, outputBytes, 0644)
	log.Println("(i) done")
}
