package main

import (
	_ "embed"
	"flag"
	"log"
	"log/slog"
	"os"
)

const PAGE_URL = "https://bsky.app/profile/did:plc:pekfvt52gjy5qunf3jcdvze4/rss"

//go:embed prompt.md
var systemInstruction string

func main() {
	logLevel := slog.LevelInfo
	isDebugVal, isDebugPresent := os.LookupEnv("DEBUG")

	if isDebugPresent && isDebugVal != "" {
		logLevel = slog.LevelDebug
	}

	slog.SetLogLoggerLevel(logLevel)

	getRssFlag := flag.Bool("rss", false, "Load and parse RSS data into XML")
	parseGeminiFlag := flag.Bool("gemini", false, "Send RSS data to Gemini and output JSON data")
	outFileFlag := flag.String("outFile", "", "Path to a file to save JSON data")

	flag.Parse()

	if !*getRssFlag {
		flag.PrintDefaults()
		return
	}

	slog.Info("fetching rss data")
	rssContent, err := GetRss(PAGE_URL)
	if err != nil {
		log.Fatal(err)
	}

	if !*parseGeminiFlag {
		return
	}

	slog.Info("sending rss data to gemini")
	bytesContent, err := ParseGemini(systemInstruction, rssContent)
	if err != nil {
		log.Fatal(err)
	}

	slog.Debug(string(bytesContent))
	if *outFileFlag == "" {
		slog.Info("no output set")
		return
	}

	slog.Info("writing to output file")
	err = os.WriteFile(*outFileFlag, bytesContent, 0644)
	if err != nil {
		log.Fatalf("failed to write file %s: %v", *outFileFlag, err)
	}

	slog.Info("done")
}
