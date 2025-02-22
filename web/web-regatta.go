package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

//go:embed static/*
var content embed.FS

var baseAPIURL = os.Getenv("API_URL")
var baseWebURL = os.Getenv("BASE_URL")

type PageData struct {
	Title  string
	Active string
	Data   interface{}
}

type StandingsData struct {
	RegattaID   string `json:"regattaId"`
	TeamID      string `json:"teamId"`
	TeamName    string `json:"teamName"`
	TotalPoints int    `json:"totalPoints"`
	Position    int    `json:"position"`
}

type RaceResultData struct {
	RaceID     string    `json:"raceId"`
	TeamID     string    `json:"teamId"`
	TeamName   string    `json:"teamName"`
	Points     int       `json:"points"`
	FinishTime time.Time `json:"finishTime"`
}

type TeamData struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RegattaID string `json:"regattaId"`
}

type RegattaData struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	Status    string    `json:"status"`
}

type DashboardData struct {
	ActiveRegattas int `json:"activeRegattas"`
	TotalTeams     int `json:"totalTeams"`
	RacesCompleted int `json:"racesCompleted"`
	UpcomingRaces  int `json:"upcomingRaces"`
}

func renderTemplate(w http.ResponseWriter, tmpl string, data PageData) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting working directory: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	//cwd = filepath.Join(cwd, "web")

	// Define template files with absolute paths
	files := []string{
		filepath.Join(cwd, "templates", "layout.html"),
		filepath.Join(cwd, "templates", "nav.html"),
		filepath.Join(cwd, "templates", fmt.Sprintf("%s.html", tmpl)),
	}

	// Create a new template with functions
	t := template.New("layout").Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"formatDateTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"url_for": func(path string) string {
			return fmt.Sprintf("/%s", strings.TrimPrefix(path, "/"))
		},
	})

	// Parse all template files
	t, err = t.ParseFiles(files...)
	if err != nil {
		log.Printf("Error parsing template files: %v", err)
		log.Printf("Attempted to parse files: %v", files)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Execute the template
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Call the API to get dashboard stats
	resp, err := http.Get(fmt.Sprintf("%s/dashboard/stats", baseAPIURL))
	if err != nil {
		log.Printf("Error fetching dashboard stats from API: %v", err)
		renderTemplate(w, "dashboard", PageData{
			Title:  "Dashboard",
			Active: "dashboard",
			Data:   DashboardData{},
		})
		return
	}
	defer resp.Body.Close()

	// Add this debug line to see the actual response
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Raw API response: %s", string(body))

	// Create a new reader with the same data for the decoder
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	var stats DashboardData
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		log.Printf("Error decoding dashboard stats: %v", err)
		renderTemplate(w, "dashboard", PageData{
			Title:  "Dashboard",
			Active: "dashboard",
			Data:   DashboardData{},
		})
		return
	}

	renderTemplate(w, "dashboard", PageData{
		Title:  "Dashboard",
		Active: "dashboard",
		Data:   stats,
	})
}

func handleRegattas(w http.ResponseWriter, r *http.Request) {
	// Use the Heroku URL for API calls
	resp, err := http.Get(fmt.Sprintf("%s/regattas", baseAPIURL))
	if err != nil {
		log.Printf("Error fetching regattas: %v", err)
		renderTemplate(w, "regattas", PageData{
			Title:  "Manage Regattas",
			Active: "regattas",
			Data:   []RegattaData{},
		})
		return
	}
	defer resp.Body.Close()

	// Debug the raw response
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Raw regattas response: %s", string(body))

	// Create new reader for the decoder
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	var regattas []RegattaData
	if err := json.NewDecoder(resp.Body).Decode(&regattas); err != nil {
		log.Printf("Error decoding regattas: %v", err)
		renderTemplate(w, "regattas", PageData{
			Title:  "Manage Regattas",
			Active: "regattas",
			Data:   []RegattaData{},
		})
		return
	}

	renderTemplate(w, "regattas", PageData{
		Title:  "Manage Regattas",
		Active: "regattas",
		Data:   regattas,
	})
}

func handleTeams(w http.ResponseWriter, r *http.Request) {
	// Get regattaId from query parameter
	regattaId := r.URL.Query().Get("regattaId")
	if regattaId == "" {
		log.Printf("No regattaId provided")
		renderTemplate(w, "teams", PageData{
			Title:  "Manage Teams",
			Active: "teams",
			Data:   []TeamData{},
		})
		return
	}

	// Call the correct API endpoint with regattaId
	resp, err := http.Get(fmt.Sprintf("%s/regattas/%s/teams", baseAPIURL, regattaId))
	if err != nil {
		log.Printf("Error fetching teams from API: %v", err)
		renderTemplate(w, "teams", PageData{
			Title:  "Manage Teams",
			Active: "teams",
			Data:   []TeamData{},
		})
		return
	}
	defer resp.Body.Close()

	// Debug logging
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Raw teams response: %s", string(body))
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	var teams []TeamData
	if err := json.NewDecoder(resp.Body).Decode(&teams); err != nil {
		log.Printf("Error decoding teams: %v", err)
		renderTemplate(w, "teams", PageData{
			Title:  "Manage Teams",
			Active: "teams",
			Data:   []TeamData{},
		})
		return
	}

	renderTemplate(w, "teams", PageData{
		Title:  "Manage Teams",
		Active: "teams",
		Data:   teams,
	})
}

func handleResults(w http.ResponseWriter, r *http.Request) {
	// Call the API to get results
	resp, err := http.Get(fmt.Sprintf("%s/results", baseAPIURL))
	if err != nil {
		log.Printf("Error fetching results from API: %v", err)
		renderTemplate(w, "results", PageData{
			Title:  "Race Results",
			Active: "results",
			Data:   []RaceResultData{},
		})
		return
	}
	defer resp.Body.Close()

	// Add debug logging
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Raw results response: %s", string(body))
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	var results []RaceResultData
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		log.Printf("Error decoding results: %v", err)
		renderTemplate(w, "results", PageData{
			Title:  "Race Results",
			Active: "results",
			Data:   []RaceResultData{},
		})
		return
	}

	renderTemplate(w, "results", PageData{
		Title:  "Race Results",
		Active: "results",
		Data:   results,
	})
}

func handleStandings(w http.ResponseWriter, r *http.Request) {
	// Call the API to get standings
	resp, err := http.Get(fmt.Sprintf("%s/standings", baseAPIURL))
	if err != nil {
		log.Printf("Error fetching standings from API: %v", err)
		renderTemplate(w, "standings", PageData{
			Title:  "Current Standings",
			Active: "standings",
			Data:   []StandingsData{},
		})
		return
	}
	defer resp.Body.Close()

	// Debug logging
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Raw standings response: %s", string(body))
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	var standings []StandingsData
	if err := json.NewDecoder(resp.Body).Decode(&standings); err != nil {
		log.Printf("Error decoding standings: %v", err)
		renderTemplate(w, "standings", PageData{
			Title:  "Current Standings",
			Active: "standings",
			Data:   []StandingsData{},
		})
		return
	}

	renderTemplate(w, "standings", PageData{
		Title:  "Current Standings",
		Active: "standings",
		Data:   standings,
	})
}

func handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(fmt.Sprintf("%s/dashboard/stats", baseAPIURL))

	if err != nil {
		http.Error(w, "Error fetching dashboard stats", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy the API response to our response
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

func main() {
	router := mux.NewRouter()

	// Create a custom file server with proper MIME types
	router.PathPrefix("/static/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set correct MIME types based on file extension
		switch {
		case strings.HasSuffix(r.URL.Path, ".js"):
			w.Header().Set("Content-Type", "application/javascript")
		case strings.HasSuffix(r.URL.Path, ".css"):
			w.Header().Set("Content-Type", "text/css")
		}
		// Serve the file from the embedded filesystem
		http.FileServer(http.FS(content)).ServeHTTP(w, r)
	})

	// API routes
	router.HandleFunc("/api/dashboard/stats", handleDashboardStats)

	// Page routes
	router.HandleFunc("/", handleDashboard)
	router.HandleFunc("/dashboard", handleDashboard)
	router.HandleFunc("/regattas", handleRegattas)
	router.HandleFunc("/teams", handleTeams)
	router.HandleFunc("/results", handleResults)
	router.HandleFunc("/standings", handleStandings)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Web Server starting on %s:%s", baseWebURL, port)

	// Combine baseWebURL and port for ListenAndServe
	address := baseWebURL + ":" + port

	log.Fatal(http.ListenAndServe(address, router))
}
