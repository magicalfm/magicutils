# mdump

mdump is a CLI tool that downloads and processes podcast transcripts from RSS feeds, specifically designed for MagicalFM podcast episodes.

## Features

- Downloads podcast episode transcripts from RSS feeds
- Processes VTT format transcripts into markdown files
- Splits output into multiple files (default: 20 episodes per file)
- Concurrent transcript fetching for improved performance
- Sorts episodes by publication date

## Requirements

- Go 1.23 or later

## Configuration

You can modify the following constants in `main.go`:

- `rssLink`: The RSS feed URL to fetch episodes from
- `episodesPerFile`: Number of episodes per output file (default: 20)
