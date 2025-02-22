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

	"github.com/gin-gonic/gin"
)

//go:embed static/*
var content embed.FS

var baseAPIURL = os.Getenv("API_URL")
var baseWebURL = os.Getenv("BASE_URL")

type PageData struct {
	Title   string
	Active  string
	Data    interface{}
	API_URL string
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

func handleDashboard(c *gin.Context) {
	// Call the API to get dashboard stats
	resp, err := http.Get(fmt.Sprintf("%s/dashboard/stats", baseAPIURL))
	if err != nil {
		log.Printf("Error fetching dashboard stats from API: %v", err)
		c.HTML(http.StatusOK, "dashboard.html", PageData{
			Title:   "Dashboard",
			Active:  "dashboard",
			Data:    DashboardData{},
			API_URL: baseAPIURL,
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
		c.HTML(http.StatusOK, "dashboard.html", PageData{
			Title:   "Dashboard",
			Active:  "dashboard",
			Data:    DashboardData{},
			API_URL: baseAPIURL,
		})
		return
	}

	c.HTML(http.StatusOK, "dashboard.html", PageData{
		Title:   "Dashboard",
		Active:  "dashboard",
		Data:    stats,
		API_URL: baseAPIURL,
	})
}

func handleRegattas(c *gin.Context) {
	// Use the Heroku URL for API calls
	resp, err := http.Get(fmt.Sprintf("%s/regattas", baseAPIURL))
	if err != nil {
		log.Printf("Error fetching regattas: %v", err)
		c.HTML(http.StatusOK, "regattas.html", PageData{
			Title:   "Manage Regattas",
			Active:  "regattas",
			Data:    []RegattaData{},
			API_URL: baseAPIURL,
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
		c.HTML(http.StatusOK, "regattas.html", PageData{
			Title:   "Manage Regattas",
			Active:  "regattas",
			Data:    []RegattaData{},
			API_URL: baseAPIURL,
		})
		return
	}

	c.HTML(http.StatusOK, "regattas.html", PageData{
		Title:   "Manage Regattas",
		Active:  "regattas",
		Data:    regattas,
		API_URL: baseAPIURL,
	})
}

func handleTeams(c *gin.Context) {
	// Get regattaId from query parameter
	regattaId := c.Query("regattaId")
	if regattaId == "" {
		log.Printf("No regattaId provided")
		c.HTML(http.StatusOK, "teams.html", PageData{
			Title:   "Manage Teams",
			Active:  "teams",
			Data:    []TeamData{},
			API_URL: baseAPIURL,
		})
		return
	}

	// Call the correct API endpoint with regattaId
	resp, err := http.Get(fmt.Sprintf("%s/regattas/%s/teams", baseAPIURL, regattaId))
	if err != nil {
		log.Printf("Error fetching teams from API: %v", err)
		c.HTML(http.StatusOK, "teams.html", PageData{
			Title:   "Manage Teams",
			Active:  "teams",
			Data:    []TeamData{},
			API_URL: baseAPIURL,
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
		c.HTML(http.StatusOK, "teams.html", PageData{
			Title:   "Manage Teams",
			Active:  "teams",
			Data:    []TeamData{},
			API_URL: baseAPIURL,
		})
		return
	}

	c.HTML(http.StatusOK, "teams.html", PageData{
		Title:   "Manage Teams",
		Active:  "teams",
		Data:    teams,
		API_URL: baseAPIURL,
	})
}

func handleResults(c *gin.Context) {
	// Call the API to get results
	resp, err := http.Get(fmt.Sprintf("%s/results", baseAPIURL))
	if err != nil {
		log.Printf("Error fetching results from API: %v", err)
		c.HTML(http.StatusOK, "results.html", PageData{
			Title:   "Race Results",
			Active:  "results",
			Data:    []RaceResultData{},
			API_URL: baseAPIURL,
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
		c.HTML(http.StatusOK, "results.html", PageData{
			Title:   "Race Results",
			Active:  "results",
			Data:    []RaceResultData{},
			API_URL: baseAPIURL,
		})
		return
	}

	c.HTML(http.StatusOK, "results.html", PageData{
		Title:   "Race Results",
		Active:  "results",
		Data:    results,
		API_URL: baseAPIURL,
	})
}

func handleStandings(c *gin.Context) {
	// Call the API to get standings
	resp, err := http.Get(fmt.Sprintf("%s/standings", baseAPIURL))
	if err != nil {
		log.Printf("Error fetching standings from API: %v", err)
		c.HTML(http.StatusOK, "standings.html", PageData{
			Title:   "Current Standings",
			Active:  "standings",
			Data:    []StandingsData{},
			API_URL: baseAPIURL,
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
		c.HTML(http.StatusOK, "standings.html", PageData{
			Title:   "Current Standings",
			Active:  "standings",
			Data:    []StandingsData{},
			API_URL: baseAPIURL,
		})
		return
	}

	c.HTML(http.StatusOK, "standings.html", PageData{
		Title:   "Current Standings",
		Active:  "standings",
		Data:    standings,
		API_URL: baseAPIURL,
	})
}

func handleDashboardStats(c *gin.Context) {
	resp, err := http.Get(fmt.Sprintf("%s/dashboard/stats", baseAPIURL))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching dashboard stats"})
		return
	}
	defer resp.Body.Close()

	// Copy the API response to our response
	c.Header("Content-Type", "application/json")
	io.Copy(c.Writer, resp.Body)
}

func main() {
	router := gin.Default()

	// Serve static files
	router.StaticFS("/static", http.FS(content))

	// API routes
	router.GET("/api/dashboard/stats", handleDashboardStats)

	// Page routes
	router.GET("/", handleDashboard)
	router.GET("/dashboard", handleDashboard)
	router.GET("/regattas", handleRegattas)
	router.GET("/teams", handleTeams)
	router.GET("/results", handleResults)
	router.GET("/standings", handleStandings)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Web Server starting on %s:%s", baseWebURL, port)
	if err := router.Run(fmt.Sprintf("%s:%s", baseWebURL, port)); err != nil {
		log.Fatal(err)
	}
}
