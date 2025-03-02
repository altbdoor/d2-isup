# d2-isup

Google Gemini powered, Destiny 2 down time visualizer

## What does it do?

1. ~~Get the HTML for [Bungie server and updates status webpage](https://help.bungie.net/hc/en-us/articles/360049199271-Destiny-Server-and-Update-Status)~~
   1. This page has become too unreliable to web scrape with all the protection it has
1. Get the XML RSS feed from [Unofficial BungieHelp](https://mastodon.social/@bungiehelp)
1. Pass the content to Google Gemini, and ask Google Gemini to return a JSON data on maintenance and down time
   1. See the prompt in `scripts/main.go`!
1. Serve the website on GitHub Pages with:
   1. <https://github.com/andybrewer/mvp>
   1. <https://www.chartjs.org/>
   1. <https://alpinejs.dev/>
1. Website renders the down time based on web visitor's own time zone!
1. Repeat every 8 hours

## Development

### Gemini API key

You will need to get a Google Gemini API key, see <https://ai.google.dev/gemini-api/docs/api-key>.

Provide the key as `GOOGLE_API_KEY` environment variable. Example:

```sh
GOOGLE_API_KEY=SeCrEt_KeY123
```

### Final steps

1. Install Golang 1.23.x
1. `git clone`
1. `cd scripts`
1. `go mod download`
1. `go run main.go`
1. `cd ../_site`
1. `python -m http.server 8000` or any static server of your choice
