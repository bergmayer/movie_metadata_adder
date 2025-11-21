package tmdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	baseURL      = "https://api.themoviedb.org/3"
	imageBaseURL = "https://image.tmdb.org/t/p/original"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

type SearchResult struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
	Overview    string `json:"overview"`
	PosterPath  string `json:"poster_path"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

type MovieDetails struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
	Overview    string `json:"overview"`
	PosterPath  string `json:"poster_path"`
	Genres      []struct {
		Name string `json:"name"`
	} `json:"genres"`
	Credits struct {
		Cast []struct {
			Name string `json:"name"`
		} `json:"cast"`
		Crew []struct {
			Name string `json:"name"`
			Job  string `json:"job"`
		} `json:"crew"`
	} `json:"credits"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// SearchMovies searches for movies using the TMDB API
func (c *Client) SearchMovies(query string, year string) ([]SearchResult, error) {
	params := url.Values{}
	params.Add("api_key", c.apiKey)
	params.Add("query", query)
	if year != "" {
		params.Add("year", year)
	}

	searchURL := fmt.Sprintf("%s/search/movie?%s", baseURL, params.Encode())

	resp, err := c.httpClient.Get(searchURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	return searchResp.Results, nil
}

// FuzzySearch performs multiple searches with progressively simpler queries
func (c *Client) FuzzySearch(query string, year string) ([]SearchResult, error) {
	seenIDs := make(map[int]bool)
	var allResults []SearchResult

	// Try multiple search terms
	searchTerms := []string{
		query,
		strings.Join(strings.Fields(query)[:min(2, len(strings.Fields(query)))], " "),
		strings.Fields(query)[0],
	}

	for _, term := range searchTerms {
		if term == "" {
			continue
		}

		results, err := c.SearchMovies(term, year)
		if err != nil {
			return nil, err
		}

		for _, result := range results {
			if !seenIDs[result.ID] {
				seenIDs[result.ID] = true
				allResults = append(allResults, result)
			}
		}
	}

	return allResults, nil
}

// GetMovieDetails fetches full movie details including credits
func (c *Client) GetMovieDetails(movieID int) (*MovieDetails, error) {
	params := url.Values{}
	params.Add("api_key", c.apiKey)
	params.Add("append_to_response", "credits")

	detailsURL := fmt.Sprintf("%s/movie/%d?%s", baseURL, movieID, params.Encode())

	resp, err := c.httpClient.Get(detailsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	var details MovieDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}

	return &details, nil
}

// DownloadPoster downloads the movie poster and returns the image data
func (c *Client) DownloadPoster(posterPath string) ([]byte, error) {
	if posterPath == "" {
		return nil, fmt.Errorf("no poster path provided")
	}

	posterURL := imageBaseURL + posterPath

	resp, err := c.httpClient.Get(posterURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download poster: status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
