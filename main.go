package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/bergmayer/movie_metadata_adder/internal/config"
	"github.com/bergmayer/movie_metadata_adder/internal/metadata"
	"github.com/bergmayer/movie_metadata_adder/internal/parser"
	"github.com/bergmayer/movie_metadata_adder/internal/tmdb"
)

var (
	cfg           *config.Config
	tmdbClient    *tmdb.Client
	currentFile   string
	extractedYear string
	dropLabel     *widget.Label
	searchEntry   *widget.Entry
	resultsList   *widget.List
	searchButton  *widget.Button
	processBtn    *widget.Button
	statusLabel   *widget.Label
	searchResults []tmdb.SearchResult
	selectedIdx   int = -1
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Movie Metadata Adder")
	myWindow.Resize(fyne.NewSize(800, 600))

	// Load config
	var err error
	cfg, err = config.Load()
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load config: %v", err), myWindow)
	}

	// If no API key, prompt for it
	if cfg.TMDBAPIKey == "" {
		showAPIKeyDialog(myWindow, myApp)
	} else {
		tmdbClient = tmdb.NewClient(cfg.TMDBAPIKey)
	}

	// Create UI components
	statusLabel = widget.NewLabel("Drag and drop a movie file to begin")
	statusLabel.Wrapping = fyne.TextWrapWord

	searchEntry = widget.NewEntry()
	searchEntry.SetPlaceHolder("Enter movie title to search...")
	searchEntry.Disable()

	searchButton = widget.NewButton("Search TMDB", func() {
		performSearch(false) // Manual search - don't use extracted year
	})
	searchButton.Disable()

	resultsList = widget.NewList(
		func() int {
			return len(searchResults)
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			result := searchResults[id]
			year := ""
			if len(result.ReleaseDate) >= 4 {
				year = result.ReleaseDate[:4]
			}

			vbox := item.(*fyne.Container)
			vbox.Objects[0].(*widget.Label).SetText(fmt.Sprintf("%s (%s)", result.Title, year))

			overview := result.Overview
			if len(overview) > 100 {
				overview = overview[:100] + "..."
			}
			vbox.Objects[1].(*widget.Label).SetText(overview)
		},
	)

	resultsList.OnSelected = func(id widget.ListItemID) {
		selectedIdx = id
		processBtn.Enable()
	}

	processBtn = widget.NewButton("Process movie with selected result", func() {
		processMovie()
	})
	processBtn.Disable()

	configBtn := widget.NewButton("Set API Key", func() {
		showAPIKeyDialog(myWindow, myApp)
	})

	// Drag and drop target
	dropLabel = widget.NewLabel("Drop movie file here")
	dropLabel.Alignment = fyne.TextAlignCenter

	dropContainer := container.NewStack(
		widget.NewCard("", "", dropLabel),
	)

	// Set up drag and drop
	myWindow.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		if len(uris) > 0 {
			handleFileDrop(uris[0])
		}
	})

	// Layout
	searchBox := container.NewBorder(nil, nil, nil, searchButton, searchEntry)

	content := container.NewBorder(
		container.NewVBox(
			dropContainer,
			statusLabel,
			searchBox,
			widget.NewSeparator(),
		),
		container.NewVBox(
			processBtn,
			configBtn,
		),
		nil,
		nil,
		resultsList,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func handleFileDrop(uri fyne.URI) {
	filePath := uri.Path()

	// Validate file type
	if !isValidVideoFile(filePath) {
		statusLabel.SetText("Error: Only MPEG video files are supported")
		return
	}

	currentFile = filePath

	// Update drop label to show filename
	dropLabel.SetText(filepath.Base(filePath))

	statusLabel.SetText(fmt.Sprintf("File loaded: %s", filepath.Base(filePath)))

	// Parse filename
	movieInfo := parser.ParseFilename(filePath)
	extractedYear = movieInfo.Year
	searchEntry.SetText(movieInfo.Title)
	searchEntry.Enable()
	searchButton.Enable()

	// Auto-search with extracted year
	performSearch(true)
}

func isValidVideoFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	validExtensions := []string{".mp4", ".mkv", ".avi", ".mov", ".m4v", ".mpg", ".mpeg", ".wmv", ".flv"}

	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

func performSearch(useYear bool) {
	if tmdbClient == nil {
		statusLabel.SetText("Error: Please set your TMDB API key first")
		return
	}

	query := searchEntry.Text
	if query == "" {
		statusLabel.SetText("Error: Please enter a search query")
		return
	}

	statusLabel.SetText("Searching TMDB...")

	// Use extracted year only for auto-search, not manual search
	year := ""
	if useYear {
		year = extractedYear
	}

	results, err := tmdbClient.FuzzySearch(query, year)
	if err != nil {
		statusLabel.SetText(fmt.Sprintf("Error: %v", err))
		return
	}

	if len(results) == 0 {
		statusLabel.SetText("No results found")
		searchResults = nil
		resultsList.Refresh()
		return
	}

	searchResults = results
	selectedIdx = -1
	resultsList.UnselectAll()
	resultsList.Refresh()
	processBtn.Disable()

	statusLabel.SetText(fmt.Sprintf("Found %d results", len(results)))
}

func processMovie() {
	if selectedIdx < 0 || selectedIdx >= len(searchResults) {
		statusLabel.SetText("Error: Please select a movie from the results")
		return
	}

	if currentFile == "" {
		statusLabel.SetText("Error: No file loaded")
		return
	}

	selectedMovie := searchResults[selectedIdx]
	statusLabel.SetText(fmt.Sprintf("Fetching details for: %s", selectedMovie.Title))

	// Fetch full movie details
	details, err := tmdbClient.GetMovieDetails(selectedMovie.ID)
	if err != nil {
		statusLabel.SetText(fmt.Sprintf("Error fetching details: %v", err))
		return
	}

	// Download poster
	var posterData []byte
	if details.PosterPath != "" {
		statusLabel.SetText("Downloading poster...")
		posterData, err = tmdbClient.DownloadPoster(details.PosterPath)
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Warning: Could not download poster: %v", err))
		}
	}

	// Convert to metadata
	movieMetadata := metadata.ConvertMovieDetailsToMetadata(details, posterData)

	// Update file
	statusLabel.SetText("Updating file metadata...")
	newPath, err := metadata.UpdateMovieFile(currentFile, movieMetadata)
	if err != nil {
		statusLabel.SetText(fmt.Sprintf("Error updating file: %v", err))
		return
	}

	statusLabel.SetText(fmt.Sprintf("Success! File saved as: %s", filepath.Base(newPath)))

	// Reset UI
	currentFile = ""
	extractedYear = ""
	dropLabel.SetText("Drop movie file here")
	searchEntry.SetText("")
	searchEntry.Disable()
	searchButton.Disable()
	searchResults = nil
	resultsList.Refresh()
	processBtn.Disable()
}

func showAPIKeyDialog(win fyne.Window, app fyne.App) {
	apiKeyEntry := widget.NewEntry()
	apiKeyEntry.SetPlaceHolder("Enter your TMDB API key")
	if cfg.TMDBAPIKey != "" {
		apiKeyEntry.SetText(cfg.TMDBAPIKey)
	}

	formDialog := dialog.NewForm(
		"TMDB API Key",
		"Save",
		"Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("API Key", apiKeyEntry),
		},
		func(confirmed bool) {
			if confirmed {
				cfg.TMDBAPIKey = apiKeyEntry.Text
				if err := config.Save(cfg); err != nil {
					dialog.ShowError(fmt.Errorf("failed to save config: %v", err), win)
					return
				}
				tmdbClient = tmdb.NewClient(cfg.TMDBAPIKey)
				statusLabel.SetText("API key saved successfully")
			} else if cfg.TMDBAPIKey == "" {
				dialog.ShowInformation("API Key Required",
					"You need to set a TMDB API key to use this application.\n\n"+
					"Get one at: https://www.themoviedb.org/settings/api", win)
			}
		},
		win,
	)

	formDialog.Resize(fyne.NewSize(400, 150))
	formDialog.Show()
}
