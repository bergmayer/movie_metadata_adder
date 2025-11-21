# Movie Metadata Adder

A cross-platform GUI application for adding metadata to movie files using The Movie Database (TMDB) API. Built with Go and Fyne.

## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Supported File Formats](#supported-file-formats)
- [Configuration Storage](#configuration-storage)
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
- **Secure API Key Storage** - API keys stored following platform conventions (see [Configuration Storage](#configuration-storage))
- **Cross-Platform** - Works on macOS, Linux, and Windows (Linux: Tested on KDE Plasma; drag-and-drop not supported in all environments)

## Prerequisites

**To use:**
- FFmpeg (must be in system PATH)
- TMDB API key (free from themoviedb.org)

**To build:**
- Go 1.21+
- C compiler (for Fyne GUI framework)

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

## Configuration Storage

Configuration files (including your TMDB API key) are stored in:

- **macOS**: `~/Library/Application Support/MovieMetadataAdder/`
- **Linux**: `$XDG_CONFIG_HOME/MovieMetadataAdder/` (or `~/.config/MovieMetadataAdder/`)
- **Windows**: `%APPDATA%\MovieMetadataAdder\`

During processing, temporary files are created in the same directory as the movie file and automatically deleted after completion.

## License

This project is released into the **public domain** under the Unlicense. See the [UNLICENSE](UNLICENSE) file for details.

You are free to use, modify, and distribute this software for any purpose, commercial or non-commercial, without any restrictions.

## Acknowledgments

- Movie data provided by [The Movie Database (TMDB)](https://www.themoviedb.org/)
- This product uses the TMDB API but is not endorsed or certified by TMDB
- GUI framework: [Fyne](https://fyne.io/)
- Media processing: [FFmpeg](https://ffmpeg.org/)
