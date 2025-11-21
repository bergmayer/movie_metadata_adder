# Movie Metadata Adder - Technical Documentation

This document provides detailed technical information about the Movie Metadata Adder application.

## Table of Contents

- [Architecture](#architecture)
- [File and Directory Locations](#file-and-directory-locations)
- [API Usage](#api-usage)
- [Metadata Tags](#metadata-tags)
- [FFmpeg Commands](#ffmpeg-commands)
- [Error Handling](#error-handling)
- [Building and Distribution](#building-and-distribution)

## Architecture

### Overview

The application follows a modular architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────┐
│           main.go (GUI Layer)               │
│  - Fyne GUI components                      │
│  - User interaction handling                │
│  - Application lifecycle                    │
└─────────────────────────────────────────────┘
           │          │          │
           ▼          ▼          ▼
┌──────────────┐ ┌──────────┐ ┌──────────────┐
│    Config    │ │   TMDB   │ │   Parser     │
│   Package    │ │  Client  │ │   Package    │
└──────────────┘ └──────────┘ └──────────────┘
           │          │          │
           └──────────┴──────────┘
                      │
                      ▼
            ┌──────────────────┐
            │    Metadata      │
            │     Writer       │
            │  (FFmpeg wrap)   │
            └──────────────────┘
```

### Package Descriptions

#### `internal/config`

**Purpose**: Cross-platform configuration management

**Key Functions**:
- `GetConfigDir()` - Returns platform-specific config directory
- `GetConfigPath()` - Returns full path to config.json
- `Load()` - Reads configuration from disk
- `Save()` - Writes configuration to disk with proper permissions

**Platform Behavior**:
- macOS: Uses `~/Library/Application Support/`
- Linux (including WSL): Respects `XDG_CONFIG_HOME` environment variable
- Creates directories with 0700 permissions (Unix)
- Creates config files with 0600 permissions (Unix)

#### `internal/tmdb`

**Purpose**: TMDB API client for movie data retrieval

**Key Functions**:
- `NewClient(apiKey)` - Creates a new TMDB client instance
- `SearchMovies(query, year)` - Searches for movies by title and optional year
- `FuzzySearch(query, year)` - Performs multiple searches with varying query lengths
- `GetMovieDetails(movieID)` - Fetches complete movie details including credits
- `DownloadPoster(posterPath)` - Downloads movie poster image

**API Endpoints Used**:
- `GET /3/search/movie` - Movie search
- `GET /3/movie/{id}` - Movie details with credits

#### `internal/parser`

**Purpose**: Filename parsing to extract movie information

**Key Functions**:
- `ParseFilename(filePath)` - Extracts title and year from filename

**Parsing Logic**:
1. Extract base filename (remove path and extension)
2. Replace `.` and `_` with spaces
3. Extract year using regex: `\(?([12][0-9]{3})\)?`
4. Remove quality indicators (720p, BluRay, x264, etc.)
5. Clean up whitespace

**Recognized Quality Indicators**:
- Resolutions: 720p, 1080p, 2160p, 4k
- Sources: BRRip, BDRip, BluRay, WEBRip, HDTV, WEB-DL
- Codecs: x264, x265, 10bit, HEVC
- Audio: AAC, AC3, DTS
- Release groups: GalaxyRG, RARBG, YTS, YIFY
- File sizes: Patterns like "2GB", "700MB"

#### `internal/metadata`

**Purpose**: Metadata embedding using FFmpeg

**Key Functions**:
- `UpdateMovieFile(filePath, metadata)` - Main function to update file
- `ConvertMovieDetailsToMetadata(details, posterData)` - Converts TMDB data to metadata struct

**Process**:
1. Validates FFmpeg installation
2. Creates temporary files for processing
3. Builds FFmpeg command with appropriate flags
4. Executes FFmpeg
5. Replaces original file
6. Cleans up temporary files

## File and Directory Locations

### Configuration Storage

#### macOS
```
~/Library/Application Support/MovieMetadataAdder/
└── config.json
```

Environment variables used: None (uses standard macOS location)

#### Linux
```
$XDG_CONFIG_HOME/MovieMetadataAdder/config.json
```
or if `XDG_CONFIG_HOME` is not set:
```
~/.config/MovieMetadataAdder/config.json
```

Environment variables used:
- `XDG_CONFIG_HOME` - Optional, follows XDG Base Directory Specification

**WSL Note**: Uses the same Linux configuration paths within your WSL home directory.

### Temporary Files

Temporary files are created in the **same directory** as the movie file being processed:

#### temp_metadata{extension}
- **Purpose**: Temporary output file during FFmpeg processing
- **Lifetime**: Created during processing, deleted after successful rename or on error
- **Size**: Same as original file (stream copy, no re-encoding)
- **Permissions**: Inherits from parent directory

#### temp_poster.jpg
- **Purpose**: Temporary storage for downloaded poster image
- **Lifetime**: Created before FFmpeg runs, deleted immediately after FFmpeg completes
- **Size**: Varies, typically 500KB-2MB for high-resolution posters
- **Permissions**: 0644 (read/write for owner, read for others)

### Cache

**The application does not use any cache.**

All data is fetched from TMDB for each operation. This ensures:
- Always up-to-date movie information
- No stale data
- No disk space used for caching
- No cache invalidation logic needed

### Logs

**The application does not create log files.**

Errors and status updates are displayed in the GUI. FFmpeg output is captured but not logged to disk.

## API Usage

### TMDB API

#### Authentication
- Method: API key in URL query parameter
- Parameter: `api_key={key}`
- All requests use HTTPS

#### Rate Limits
TMDB enforces rate limits:
- **40 requests per 10 seconds** per IP address
- **Unlimited** requests per day for standard accounts

The application makes:
- 1-3 requests during fuzzy search (deduplicated)
- 1 request to fetch movie details
- 1 request to download poster

**Total per movie**: ~3-5 requests

#### Endpoints

##### Search Movies
```
GET https://api.themoviedb.org/3/search/movie
Parameters:
  - api_key (required)
  - query (required) - Search term
  - year (optional) - Filter by release year

Response: JSON with array of movie results
```

##### Get Movie Details
```
GET https://api.themoviedb.org/3/movie/{movie_id}
Parameters:
  - api_key (required)
  - append_to_response=credits - Include cast/crew data

Response: JSON with complete movie details
```

##### Download Poster
```
GET https://image.tmdb.org/t/p/original{poster_path}
Parameters: None (public CDN)

Response: Binary image data (JPEG)
```

### Network Requirements

- **Internet connection required**: Application cannot function offline
- **Firewall rules**: Must allow HTTPS (443) to:
  - `api.themoviedb.org`
  - `image.tmdb.org`
- **Proxy support**: Inherits from system HTTP proxy settings (Go standard library)

## Metadata Tags

### Standard Tags Used

The application embeds the following metadata tags:

| Tag Name | Description | Example |
|----------|-------------|---------|
| `title` | Movie title | "The Matrix" |
| `year` | Release year (4-digit) | "1999" |
| `date` | Release year (alternate format) | "1999" |
| `director` | Director name | "Lana Wachowski" |
| `actors` | Comma-separated cast list (top 5) | "Keanu Reeves, Laurence Fishburne, ..." |
| `genre` | Comma-separated genre list | "Action, Science Fiction" |

### Tag Support by Format

| Format | title | year | director | actors | genre | poster |
|--------|-------|------|----------|--------|-------|--------|
| MP4 (.mp4) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Matroska (.mkv) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| AVI (.avi) | ✓ | ~ | ~ | ~ | ~ | ✗ |
| QuickTime (.mov) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| M4V (.m4v) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

Legend: ✓ = Full support, ~ = Limited support, ✗ = Not supported

**Note**: MP4, MKV, MOV, and M4V provide the best metadata support.

### Poster Embedding

Posters are embedded as:
- **Stream type**: Attached picture / Cover art
- **Codec**: PNG (converted from JPEG during embedding)
- **Disposition**: `attached_pic`
- **Resolution**: Original TMDB resolution (typically 2000x3000px)

## FFmpeg Commands

### Without Poster

```bash
ffmpeg -i input.mkv \
  -metadata title="Movie Title" \
  -metadata year="2023" \
  -metadata date="2023" \
  -metadata director="Director Name" \
  -metadata actors="Actor 1, Actor 2" \
  -metadata genre="Action, Drama" \
  -c copy \
  -y output.mkv
```

### With Poster

```bash
ffmpeg -i input.mkv -i poster.jpg \
  -map 0 -map 1 \
  -metadata title="Movie Title" \
  -metadata year="2023" \
  -metadata date="2023" \
  -metadata director="Director Name" \
  -metadata actors="Actor 1, Actor 2" \
  -metadata genre="Action, Drama" \
  -c copy \
  -c:v:1 png \
  -disposition:v:1 attached_pic \
  -y output.mkv
```

### Command Explanation

- `-i input.mkv` - Input movie file
- `-i poster.jpg` - Input poster image (if present)
- `-map 0` - Map all streams from first input
- `-map 1` - Map poster from second input
- `-metadata key=value` - Set metadata tags
- `-c copy` - Copy all streams without re-encoding (fast)
- `-c:v:1 png` - Convert poster to PNG
- `-disposition:v:1 attached_pic` - Mark poster as attached picture
- `-y` - Overwrite output file without asking

## Error Handling

### Validation Errors

**Invalid file type**:
- Checked before processing begins
- Error displayed in status label
- File is not processed

**Missing FFmpeg**:
- Checked before running FFmpeg command
- Error: "ffmpeg is not installed or not in PATH"
- Processing halts

### API Errors

**HTTP Status Codes**:
- `401 Unauthorized` - Invalid API key
- `404 Not Found` - Movie ID doesn't exist
- `429 Too Many Requests` - Rate limit exceeded (rare)
- Other non-200 codes - Generic error message

**Network Errors**:
- Connection timeout
- DNS resolution failure
- No internet connection
- All displayed as error messages in GUI

### Processing Errors

**FFmpeg Failures**:
- Captures stderr output
- Displays error with FFmpeg output in status label
- Temporary files are cleaned up
- Original file remains unchanged

**File System Errors**:
- Permission denied (can't read/write)
- Disk full
- File already in use
- All displayed as error messages

### Error Recovery

The application provides automatic cleanup:
1. If FFmpeg fails, `temp_metadata` file is deleted
2. If poster download fails, metadata is still embedded (without poster)
3. If file rename fails, temporary file is deleted
4. Original file is only deleted after successful processing

## Building and Distribution

### Build Commands

#### Standard Build
```bash
go build -o movie_metadata_adder
```

#### Optimized Build (smaller binary)
```bash
go build -ldflags="-s -w" -o movie_metadata_adder
```

#### Cross-Platform Builds

**macOS (from any OS)**:
```bash
GOOS=darwin GOARCH=amd64 go build -o movie_metadata_adder-macos-amd64
GOOS=darwin GOARCH=arm64 go build -o movie_metadata_adder-macos-arm64
```

**Linux (from any OS)**:
```bash
GOOS=linux GOARCH=amd64 go build -o movie_metadata_adder-linux-amd64
GOOS=linux GOARCH=arm64 go build -o movie_metadata_adder-linux-arm64
```

**Note**: The Linux build works on WSL2 as well.

### Dependencies

Runtime dependencies (required by end users):
- FFmpeg (external binary)

Build dependencies (managed by Go modules):
- `fyne.io/fyne/v2` - GUI framework
- All other dependencies are transitive (automatically managed)

### Binary Size

Typical binary sizes:
- macOS: ~30MB
- Linux: ~25MB

The binaries are relatively large because:
- Fyne embeds GUI rendering code
- Go produces statically-linked binaries
- All dependencies are compiled in

To reduce size:
- Use `-ldflags="-s -w"` to strip debug info (~20% reduction)
- Use `upx` compression tool (~50% reduction, but slower startup)

### Distribution

**Recommended distribution format**:
- Single executable binary
- README.md file
- UNLICENSE file

**No installer needed** - application is self-contained.

**User must install separately**:
- FFmpeg (not bundled due to licensing and size)

### Testing

Manual testing checklist:
1. First run - API key prompt appears
2. Drag and drop file - filename appears in drop area
3. Search results appear automatically
4. Edit search field - results update after clicking Search
5. Select movie - Process button enables
6. Process movie - file is updated and renamed
7. Drag another file - interface resets correctly

## Performance Considerations

### Processing Time

Factors affecting processing speed:
- **Network speed**: TMDB API requests and poster download
- **Disk I/O**: Reading/writing large video files
- **File size**: Larger files take longer (linear relationship)

Typical processing time for a 2GB movie file:
- Fast connection, SSD: 5-10 seconds
- Slow connection, HDD: 30-60 seconds

### Memory Usage

The application has minimal memory footprint:
- GUI: ~20MB
- JSON parsing: <1MB
- Poster download: ~2MB (transient)
- FFmpeg: <10MB

**Total**: ~30-40MB typical usage

**Note**: FFmpeg processes files in streaming mode, so even very large files don't cause memory issues.

### Disk Space

No additional disk space required beyond:
- Original movie file
- Temporary space for processing (equals file size, freed after completion)

## Security Considerations

### Input Validation

- File paths: Validated by OS
- API responses: Parsed with Go's JSON decoder (safe)
- Metadata strings: Escaped for FFmpeg command line

### Command Injection Prevention

The application uses Go's `exec.Command()` which properly handles arguments, preventing shell injection.

Metadata values are passed as separate arguments, not through shell interpolation.

### API Key Storage

- Stored in plain text JSON (standard for API keys)
- File permissions set to 0600 on Unix (owner read/write only)
- Never transmitted except to TMDB over HTTPS

### HTTPS Certificate Validation

Go's HTTP client automatically validates SSL/TLS certificates using the system certificate store.

## Future Enhancement Possibilities

Potential features for future versions:
- Batch processing (multiple files at once)
- TV show support
- Custom metadata field mapping
- Preview before processing
- Undo/restore original file
- Alternative metadata providers (IMDb, etc.)
- Subtitle download and embedding

This is a public domain project - anyone is welcome to fork and enhance it!
