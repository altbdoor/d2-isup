package main

import (
	"encoding/xml"
	"time"
)

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Items []RssItem `xml:"item"`
	} `xml:"channel"`
}

type RssItem struct {
	XMLName     xml.Name `xml:"item"`
	Description string   `xml:"description"`
	// Guid        string   `xml:"guid"`
	// Link        string   `xml:"link"`
	// PubDate     string   `xml:"pubDate"`
	// Category    string   `xml:"category"`
}

type MaintenanceData struct {
	MaintenanceTimeStart time.Time `json:"maintenance_time_start"`
	MaintenanceTimeEnd   time.Time `json:"maintenance_time_end"`
	ServerDownStart      time.Time `json:"server_down_start"`
	ServerDownEnd        time.Time `json:"server_down_end"`
	Description          string    `json:"description"`
}
