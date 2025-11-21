# Movie Metadata Adder

A cross-platform GUI application for adding metadata to movie files using The Movie Database (TMDB) API. Built with Go and Fyne.

## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Building from Source](#building-from-source)
- [Usage](#usage)
- [Supported File Formats](#supported-file-formats)
- [Configuration and Data Storage](#configuration-and-data-storage)
- [How It Works](#how-it-works)
- [Troubleshooting](#troubleshooting)
- [Privacy and Security](#privacy-and-security)
- [License](#license)

## Features

- **Drag and Drop Interface** - Simple, intuitive GUI for movie file processing
- **Automatic Filename Parsing** - Intelligently extracts movie title and year from filenames
- **Fuzzy Search** - Multiple search strategies to find the right movie even with partial or messy filenames
- **Interactive Search Results** - Browse and select from search results with movie descriptions
- **Editable Search** - Manually edit search terms if automatic parsing doesn't match
- **Complete Metadata Embedding**:
  - Title
  - Year
  - Director
  - Actors (top 5 cast members)
  - Genres
  - Movie poster (embedded as cover art)
- **Automatic File Renaming** - Renames files to Plex-compatible format: `Title (Year).ext`
- **In-Place Processing** - Original file is replaced with the metadata-enriched version
- **Secure API Key Storage** - API keys stored following platform conventions (see [Configuration](#configuration-and-data-storage))
- **Cross-Platform** - Works on macOS, Linux, and Windows

## Prerequisites

### Required

1. **FFmpeg** - Must be installed and available in your system PATH
   - **macOS**: `brew install ffmpeg`
   - **Linux (Debian/Ubuntu)**: `sudo apt install ffmpeg`
   - **Linux (RHEL/Fedora)**: `sudo dnf install ffmpeg`
   - **Windows**: `choco install ffmpeg` (via Chocolatey) or download from [ffmpeg.org](https://ffmpeg.org/download.html)

2. **TMDB API Key** - Free API key from The Movie Database
   - Sign up at [themoviedb.org](https://www.themoviedb.org/signup)
   - Get your API key at [themoviedb.org/settings/api](https://www.themoviedb.org/settings/api)

### Optional (for building from source)

3. **Go 1.21 or later** - Only needed if building from source
   - Download from [go.dev](https://go.dev/dl/)

## Installation

### Pre-built Binaries

Download the pre-built binary for your platform from the releases page.

### Building from Source

**macOS/Linux:**
```bash
# Clone or download the source code
cd movie_metadata_adder

# Download dependencies
go mod download

# Build the application
go build -o movie_metadata_adder
```

**Windows:**
```powershell
# Clone or download the source code
cd movie_metadata_adder

# Download dependencies
go mod download

# Build the application (hides console window)
$env:CC = "gcc"
go build -ldflags "-H windowsgui" -o movie_metadata_adder.exe
```

## Usage

### First-Time Setup

1. Launch the application:
   ```bash
   ./movie_metadata_adder
   ```

2. On first run, you'll be prompted to enter your TMDB API key
   - The key will be securely stored in your system's standard configuration directory
   - You only need to do this once

### Processing a Movie File

1. **Drag and drop** a movie file onto the application window
   - The filename will appear in the drop area
   - The search field will be auto-populated with the parsed movie title
   - An automatic search will be performed

2. **Review search results**
   - Results are displayed with title, year, and description
   - If the automatic search doesn't find the right movie, edit the search field and click "Search TMDB"

3. **Select the correct movie** from the list
   - Click on the movie that matches your file

4. **Click "Process Selected Movie"**
   - The application will:
     - Download full movie details from TMDB
     - Download the movie poster
     - Embed all metadata into the video file
     - Embed the poster as cover art
     - Rename the file to "Title (Year).ext"

5. **Done!**
   - The original file has been replaced with the metadata-enriched version
   - The file has been renamed to Plex-compatible format
   - You can drag another movie file to process more files

### Changing Your API Key

Click the "Set API Key" button at any time to update your TMDB API key.

## Supported File Formats

The application validates file types and only accepts common video container formats:

- **MP4** (.mp4) - MPEG-4 Part 14
- **Matroska** (.mkv) - Open source container format
- **AVI** (.avi) - Audio Video Interleave
- **QuickTime** (.mov) - Apple QuickTime Movie
- **M4V** (.m4v) - iTunes video format
- **MPEG** (.mpg, .mpeg) - MPEG-1/MPEG-2 video
- **Windows Media Video** (.wmv)
- **Flash Video** (.flv)

The application uses FFmpeg for metadata embedding, which supports reading metadata from and writing metadata to all of these formats.

## Configuration and Data Storage

The application follows platform-specific standards for storing configuration files and temporary data.

### Configuration Directory

Configuration files (including your TMDB API key) are stored in:

- **macOS**: `~/Library/Application Support/MovieMetadataAdder/`
- **Linux**: `$XDG_CONFIG_HOME/MovieMetadataAdder/` (or `~/.config/MovieMetadataAdder/` if XDG_CONFIG_HOME is not set)
- **Windows**: `%APPDATA%\MovieMetadataAdder\`

### Configuration File

The configuration is stored as JSON at:
```
<config_dir>/config.json
```

Example content:
```json
{
  "tmdb_api_key": "your_api_key_here"
}
```

The file has restricted permissions (0600 on Unix-like systems) to prevent unauthorized access.

### Temporary Files

During processing, the application creates temporary files in the same directory as the movie file being processed:

- **temp_metadata{.ext}** - Temporary file created by FFmpeg during metadata embedding
- **temp_poster.jpg** - Temporary poster image file

These temporary files are automatically deleted after processing completes, whether successful or not.

### Cache

The application does not maintain any cache. All data is fetched fresh from TMDB for each search.

### Network Usage

The application makes HTTPS requests to:
- `https://api.themoviedb.org/3/*` - TMDB API endpoints
- `https://image.tmdb.org/t/p/original/*` - TMDB image CDN for poster downloads

No other network connections are made. No telemetry or usage data is sent anywhere.

## How It Works

### 1. Filename Parsing

When you drop a file, the application:
- Removes the file extension
- Replaces dots and underscores with spaces
- Extracts the year (4-digit number, possibly in parentheses)
- Removes common quality indicators (720p, 1080p, BluRay, WEBRip, x264, etc.)
- Cleans up extra spaces

Example transformations:
```
The.Matrix.1999.1080p.BluRay.x264.mkv → "The Matrix" (1999)
Inception (2010) BDRip.mp4 → "Inception" (2010)
pulp_fiction_1994_720p.avi → "pulp fiction" (1994)
```

### 2. Fuzzy Search

The application performs multiple searches to maximize chances of finding the correct movie:
1. First, searches with the full parsed title
2. Then, searches with just the first two words
3. Finally, searches with just the first word

All results are combined and deduplicated by movie ID.

If a year was extracted from the filename, it's included in the search to narrow results.

### 3. Metadata Retrieval

When you select a movie, the application:
1. Fetches complete movie details using the TMDB API
2. Extracts:
   - Title and release year
   - Director (from crew data)
   - Top 5 cast members
   - All genres
3. Downloads the movie poster (highest resolution available)

### 4. File Processing

The application uses FFmpeg to:
1. Copy all existing streams from the original file (no re-encoding)
2. Add the movie poster as an attached picture stream
3. Embed metadata tags:
   - `title` - Movie title
   - `year` and `date` - Release year
   - `director` - Director name
   - `actors` - Comma-separated cast list
   - `genre` - Comma-separated genre list
4. Write the result to a temporary file
5. Delete the original file
6. Rename the temporary file to the Plex format: `Title (Year).ext`

**Important**: The original file is permanently replaced. Make sure you have backups if needed.

## Troubleshooting

### "ffmpeg is not installed or not in PATH"

FFmpeg must be installed and accessible from the command line.

Test by running:
```bash
ffmpeg -version
```

If this doesn't work, FFmpeg is not properly installed or not in your PATH.

### "TMDB API returned status 401"

Your API key is invalid or has been revoked. Click "Set API Key" to enter a new key.

### "TMDB API returned status 404"

The movie ID doesn't exist in TMDB. This shouldn't happen during normal use.

### "No results found"

The search query didn't match any movies in TMDB. Try:
1. Manually editing the search field
2. Simplifying the search (e.g., just the main title, no year)
3. Checking the spelling

### "Error: Only MPEG video files are supported"

The file you dropped has an unsupported extension. See [Supported File Formats](#supported-file-formats).

### Poster not embedded

Some older video formats don't support attached pictures. The metadata will still be embedded, but the poster won't be included.

### File processing is slow

FFmpeg is copying all streams without re-encoding, which is fast. However:
- Large files take longer to process
- Slow disk I/O can slow things down
- Downloading high-resolution posters takes time on slow connections

## Privacy and Security

### Data Collection

This application does **NOT**:
- Send any telemetry or analytics
- Track usage
- Store or transmit your API key anywhere except local configuration
- Store or cache movie files or posters beyond temporary processing

### API Key Security

- Your TMDB API key is stored locally in a JSON file
- On Unix-like systems (macOS, Linux), the file has 0600 permissions (read/write for owner only)
- The key is never transmitted except to TMDB's API over HTTPS

### File Safety

- The application modifies files in place
- Original files are deleted after successful processing
- **Always maintain backups of important files before processing**

### Network Security

- All API communication uses HTTPS
- No other network connections are made
- Certificate validation is performed by the Go HTTP client

## Project Structure

```
movie_metadata_adder/
├── main.go                          # Main GUI application
├── internal/
│   ├── config/
│   │   └── config.go                # Cross-platform config management
│   ├── tmdb/
│   │   └── client.go                # TMDB API client
│   ├── parser/
│   │   └── filename.go              # Filename parser
│   └── metadata/
│       └── writer.go                # Metadata writer (FFmpeg wrapper)
├── go.mod                           # Go module definition
├── go.sum                           # Go module checksums
├── README.md                        # This file
├── DOCUMENTATION.md                 # Additional documentation
└── UNLICENSE                        # Public domain dedication
```

## License

This project is released into the **public domain** under the Unlicense. See the [UNLICENSE](UNLICENSE) file for details.

You are free to use, modify, and distribute this software for any purpose, commercial or non-commercial, without any restrictions.

## Acknowledgments

- Movie data provided by [The Movie Database (TMDB)](https://www.themoviedb.org/)
- This product uses the TMDB API but is not endorsed or certified by TMDB
- GUI framework: [Fyne](https://fyne.io/)
- Media processing: [FFmpeg](https://ffmpeg.org/)

## Support

This is a public domain project with no official support. However:
- Check the [Troubleshooting](#troubleshooting) section first
- Review the [DOCUMENTATION.md](DOCUMENTATION.md) for additional details
- FFmpeg documentation: https://ffmpeg.org/documentation.html
- Fyne documentation: https://developer.fyne.io/
- TMDB API documentation: https://developers.themoviedb.org/3
