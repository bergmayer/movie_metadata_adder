package metadata

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bergmayer/movie_metadata_adder/internal/tmdb"
)

type MovieMetadata struct {
	Title    string
	Year     string
	Director string
	Actors   string
	Genres   string
	Poster   []byte
}

// UpdateMovieFile updates a movie file with metadata and renames it
func UpdateMovieFile(filePath string, metadata MovieMetadata) (string, error) {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return "", fmt.Errorf("ffmpeg is not installed or not in PATH")
	}

	ext := filepath.Ext(filePath)
	dir := filepath.Dir(filePath)
	tempFile := filepath.Join(dir, "temp_metadata"+ext)

	// Build ffmpeg command
	args := []string{"-i", filePath}

	// If we have a poster, save it temporarily and add it
	var posterPath string
	if len(metadata.Poster) > 0 {
		posterPath = filepath.Join(dir, "temp_poster.jpg")
		if err := os.WriteFile(posterPath, metadata.Poster, 0644); err != nil {
			return "", err
		}
		defer os.Remove(posterPath)

		args = append(args, "-i", posterPath)
		// Map video, audio, and subtitle streams from input 0 (excludes existing artwork)
		// Then map the new poster from input 1
		args = append(args, "-map", "0:v", "-map", "0:a?", "-map", "0:s?", "-map", "1")
	} else {
		// No poster, but still exclude any existing attached pictures
		args = append(args, "-map", "0:v", "-map", "0:a?", "-map", "0:s?")
	}

	// Add metadata
	if metadata.Title != "" {
		args = append(args, "-metadata", fmt.Sprintf("title=%s", metadata.Title))
	}
	if metadata.Year != "" {
		args = append(args, "-metadata", fmt.Sprintf("year=%s", metadata.Year))
		args = append(args, "-metadata", fmt.Sprintf("date=%s", metadata.Year))
	}
	if metadata.Director != "" {
		args = append(args, "-metadata", fmt.Sprintf("director=%s", metadata.Director))
	}
	if metadata.Actors != "" {
		args = append(args, "-metadata", fmt.Sprintf("actors=%s", metadata.Actors))
	}
	if metadata.Genres != "" {
		args = append(args, "-metadata", fmt.Sprintf("genre=%s", metadata.Genres))
	}

	// Copy streams
	args = append(args, "-c", "copy")

	// If we have a poster, set it as attached pic
	if len(metadata.Poster) > 0 {
		args = append(args, "-c:v:1", "png", "-disposition:v:1", "attached_pic")
	}

	args = append(args, "-y", tempFile)

	// Run ffmpeg
	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed: %v\nOutput: %s", err, string(output))
	}

	// Generate new filename in Plex format: "Title (Year).ext"
	newName := fmt.Sprintf("%s (%s)%s", metadata.Title, metadata.Year, ext)
	newPath := filepath.Join(dir, newName)

	// Remove original file
	if err := os.Remove(filePath); err != nil {
		os.Remove(tempFile)
		return "", err
	}

	// Rename temp file to new name
	if err := os.Rename(tempFile, newPath); err != nil {
		return "", err
	}

	return newPath, nil
}

// ConvertMovieDetailsToMetadata converts TMDB movie details to MovieMetadata
func ConvertMovieDetailsToMetadata(details *tmdb.MovieDetails, posterData []byte) MovieMetadata {
	var year string
	if len(details.ReleaseDate) >= 4 {
		year = details.ReleaseDate[:4]
	}

	var director string
	for _, crew := range details.Credits.Crew {
		if crew.Job == "Director" {
			director = crew.Name
			break
		}
	}

	var actors []string
	for i, cast := range details.Credits.Cast {
		if i >= 5 {
			break
		}
		actors = append(actors, cast.Name)
	}

	var genres []string
	for _, genre := range details.Genres {
		genres = append(genres, genre.Name)
	}

	return MovieMetadata{
		Title:    details.Title,
		Year:     year,
		Director: director,
		Actors:   strings.Join(actors, ", "),
		Genres:   strings.Join(genres, ", "),
		Poster:   posterData,
	}
}
