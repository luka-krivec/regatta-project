package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"regatta-project/pkg/db"

	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"

	"github.com/gorilla/mux"
)

// Define the base URL for the API
var baseURL = os.Getenv("BASE_URL")

// Types
type Regatta struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Location  string `json:"location"`
	Status    string `json:"status"`
}

type Team struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RegattaID string `json:"regattaId"`
}

type RaceResult struct {
	ID         string `json:"id"`
	RegattaID  string `json:"regattaId"`
	TeamID     string `json:"teamId"`
	RaceNumber int    `json:"raceNumber"`
	Position   int    `json:"position"`
	Points     int    `json:"points"`
}

type RaceScores struct {
	RaceNumber int          `json:"raceNumber"`
	Scores     []RaceResult `json:"scores"`
}

type TeamStanding struct {
	TeamID      string       `json:"teamId"`
	TeamName    string       `json:"name"`
	TotalPoints int          `json:"totalPoints"`
	Results     []RaceResult `json:"results"`
}

func main() {
	// Initialize database
	if err := db.InitDB(); err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	router := mux.NewRouter()

	// Enable CORS
	router.Use(corsMiddleware)

	// API routes
	router.HandleFunc("/api/regattas", createRegatta).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/regattas", getAllRegattas).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/regattas/{id}", getRegatta).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/regattas/{id}", updateRegatta).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/regattas/{id}", deleteRegatta).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/api/regattas/{regattaId}/teams", getRegattaTeams).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/regattas/{regattaId}/teams", addTeam).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/regattas/{regattaId}/results", addRaceResults).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/regattas/{regattaId}/standings", getRegattaStandings).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/regattas/{regattaId}/results", clearRegattaResults).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/api/dashboard/stats", getDashboardStats).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/regattas/{regattaId}/teams/{teamId}", deleteTeam).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/api/regattas/{regattaId}/teams/{teamId}", updateTeam).Methods("PUT", "OPTIONS")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Log the server starting address
	if port == "8081" { // Default port, log with baseURL and port
		log.Printf("API Server starting on %s:%s", baseURL, port)
	} else { // Custom port, log with baseURL only
		log.Printf("API Server starting on %s", baseURL)
	}

	// Combine baseURL and port for ListenAndServe
	address := baseURL
	if port == "8081" {
		address += ":" + port
	}

	if err := http.ListenAndServe(address, router); err != nil {
		log.Fatal(err)
	}
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from Heroku and localhost
		w.Header().Set("Access-Control-Allow-Origin", "https://regatta-project.onrender.com")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080") // Allow localhost for development
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Handler functions
func createRegatta(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request to create a regatta")

	var regatta Regatta
	if err := json.NewDecoder(r.Body).Decode(&regatta); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	regatta.ID = uuid.New().String()
	regatta.Status = "SCHEDULED"

	stmt, err := db.DB.Prepare("INSERT INTO regattas(id, name, start_date, end_date, location, status) VALUES($1, $2, $3, $4, $5, $6)")
	if err != nil {
		log.Printf("Error preparing SQL statement: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(regatta.ID, regatta.Name, regatta.StartDate, regatta.EndDate, regatta.Location, regatta.Status)
	if err != nil {
		log.Printf("Error executing SQL statement: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully created regatta: %+v", regatta)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(regatta)
}

func getRegattaStandings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	regattaId := vars["regattaId"]

	// Get all results for this regatta
	rows, err := db.DB.Query(`
		SELECT r.team_id, t.name, r.race_number, r.position, r.points 
		FROM race_results r 
		JOIN teams t ON r.team_id = t.id 
		WHERE r.regatta_id = $1`, regattaId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	standings := make(map[string]*TeamStanding)
	for rows.Next() {
		var result RaceResult
		var teamName string
		if err := rows.Scan(&result.TeamID, &teamName, &result.RaceNumber, &result.Position, &result.Points); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, exists := standings[result.TeamID]; !exists {
			standings[result.TeamID] = &TeamStanding{
				TeamID:   result.TeamID,
				TeamName: teamName,
			}
		}
		standings[result.TeamID].Results = append(standings[result.TeamID].Results, result)
		standings[result.TeamID].TotalPoints += result.Points
	}

	// Convert map to slice for output
	standingsList := make([]TeamStanding, 0, len(standings))
	for _, s := range standings {
		standingsList = append(standingsList, *s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(standingsList)
}

func getAllRegattas(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request to get all regattas")

	rows, err := db.DB.Query("SELECT id, name, start_date, end_date, location FROM regattas")
	if err != nil {
		log.Printf("Error fetching regattas: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var regattas []Regatta
	for rows.Next() {
		var regatta Regatta
		if err := rows.Scan(&regatta.ID, &regatta.Name, &regatta.StartDate, &regatta.EndDate, &regatta.Location); err != nil {
			log.Printf("Error scanning regatta: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		regattas = append(regattas, regatta)
	}

	log.Printf("Successfully retrieved %d regattas", len(regattas))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(regattas)
}

func getRegatta(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	log.Printf("Received request to get regatta with ID: %s", id)

	var regatta Regatta
	err := db.DB.QueryRow("SELECT id, name, start_date, end_date, location, status FROM regattas WHERE id = $1", id).
		Scan(&regatta.ID, &regatta.Name, &regatta.StartDate, &regatta.EndDate, &regatta.Location, &regatta.Status)

	if err != nil {
		log.Printf("Error fetching regatta: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("Successfully retrieved regatta: %+v", regatta)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(regatta)
}

func updateRegatta(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	log.Printf("Received request to update regatta with ID: %s", id)

	var regatta Regatta
	if err := json.NewDecoder(r.Body).Decode(&regatta); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.DB.Exec("UPDATE regattas SET name=$1, start_date=$2, end_date=$3, location=$4, status=$5 WHERE id=$6",
		regatta.Name, regatta.StartDate, regatta.EndDate, regatta.Location, regatta.Status, id)
	if err != nil {
		log.Printf("Error updating regatta: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully updated regatta with ID: %s", id)

	w.Header().Set("Content-Type", "application/json")
	regatta.ID = id
	json.NewEncoder(w).Encode(regatta)
}

func deleteRegatta(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	log.Printf("Received request to delete regatta with ID: %s", id)

	_, err := db.DB.Exec("DELETE FROM regattas WHERE id = $1", id)
	if err != nil {
		log.Printf("Error deleting regatta: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully deleted regatta with ID: %s", id)

	w.WriteHeader(http.StatusNoContent)
}

func getRegattaTeams(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	regattaId := vars["regattaId"]

	rows, err := db.DB.Query("SELECT id, name, regatta_id FROM teams WHERE regatta_id = $1", regattaId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var teams []Team
	for rows.Next() {
		var team Team
		if err := rows.Scan(&team.ID, &team.Name, &team.RegattaID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		teams = append(teams, team)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

func addTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	regattaId := vars["regattaId"]

	var team Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	team.ID = uuid.New().String()
	team.RegattaID = regattaId

	_, err := db.DB.Exec("INSERT INTO teams(id, name, regatta_id) VALUES($1, $2, $3)",
		team.ID, team.Name, team.RegattaID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

func addRaceResults(w http.ResponseWriter, r *http.Request) {
	log.Printf("addRaceResults handler called - Method: %s, URL: %s", r.Method, r.URL.Path)

	vars := mux.Vars(r)
	regattaId := vars["regattaId"]

	log.Printf("Path variable - RegattaID: %s", regattaId)

	var requestData struct {
		RaceNumber int          `json:"raceNumber"`
		Results    []RaceResult `json:"results"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received race number: %d", requestData.RaceNumber)
	log.Printf("Received results: %+v", requestData.Results)

	for _, result := range requestData.Results {
		result.ID = uuid.New().String()

		// Insert the race result into the database
		_, err := db.DB.Exec("INSERT INTO race_results (id, regatta_id, team_id, race_number, position, points) VALUES ($1, $2, $3, $4, $5, $6)",
			result.ID, result.RegattaID, result.TeamID, result.RaceNumber, result.Position, result.Points)

		if err != nil {
			log.Printf("Error adding race result to database: %v", err)
			// Handle unique constraint violation
			if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
				log.Printf("Unique constraint failed for ID: %s, updating existing record", result.ID)
				// Update logic here if necessary
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			log.Printf("Race result added successfully - TeamID: %s, RaceNumber: %d", result.TeamID, result.RaceNumber)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func clearRegattaResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	regattaId := vars["regattaId"]

	_, err := db.DB.Exec("DELETE FROM race_results WHERE regatta_id = $1", regattaId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getDashboardStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats := struct {
		ActiveRegattas int `json:"activeRegattas"`
		TotalTeams     int `json:"totalTeams"`
		RacesCompleted int `json:"racesCompleted"`
		UpcomingRaces  int `json:"upcomingRaces"`
	}{}

	// Get active regattas count
	err := db.DB.QueryRow("SELECT COUNT(*) FROM regattas WHERE status = 'active'").Scan(&stats.ActiveRegattas)
	if err != nil {
		log.Printf("Error getting active regattas: %v", err)
	}

	// Get total teams count
	err = db.DB.QueryRow("SELECT COUNT(*) FROM teams").Scan(&stats.TotalTeams)
	if err != nil {
		log.Printf("Error getting total teams: %v", err)
	}

	// Get completed races count
	err = db.DB.QueryRow("SELECT COUNT(*) FROM (SELECT DISTINCT regatta_id, race_number FROM race_results)").Scan(&stats.RacesCompleted)
	if err != nil {
		log.Printf("Error getting completed races: %v", err)
	}

	// Get upcoming races count
	err = db.DB.QueryRow("SELECT COUNT(*) FROM regattas WHERE status = 'SCHEDULED'").Scan(&stats.UpcomingRaces)
	if err != nil {
		log.Printf("Error getting upcoming races: %v", err)
	}

	json.NewEncoder(w).Encode(stats)
}

func deleteTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	regattaId := vars["regattaId"]
	teamId := vars["teamId"]

	if regattaId == "" || teamId == "" {
		http.Error(w, "regattaId and teamId are required", http.StatusBadRequest)
		return
	}

	// First verify the team belongs to the regatta
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM teams WHERE id = $1 AND regatta_id = $2", teamId, regattaId).Scan(&count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.Error(w, "Team not found or doesn't belong to this regatta", http.StatusNotFound)
		return
	}

	// Delete the team
	_, err = db.DB.Exec("DELETE FROM teams WHERE id = $1 AND regatta_id = $2", teamId, regattaId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func updateTeam(w http.ResponseWriter, r *http.Request) {
	// Request logging
	log.Printf("updateTeam handler called - Method: %s, URL: %s", r.Method, r.URL.Path)
	log.Printf("Request Headers: %+v", r.Header)

	vars := mux.Vars(r)
	regattaId := vars["regattaId"]
	teamId := vars["teamId"]

	log.Printf("Path variables - RegattaID: %s, TeamID: %s", regattaId, teamId)

	if regattaId == "" || teamId == "" {
		log.Printf("Error: Missing required parameters - RegattaID: %s, TeamID: %s", regattaId, teamId)
		http.Error(w, "regattaId and teamId are required", http.StatusBadRequest)
		return
	}

	// Read and log request body
	var team Team
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Received request body: %s", string(body))

	// Parse JSON body
	if err := json.Unmarshal(body, &team); err != nil {
		log.Printf("Error parsing JSON body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Parsed team data: %+v", team)

	// Verify team exists and belongs to regatta
	var count int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM teams WHERE id = $1 AND regatta_id = $2", teamId, regattaId).Scan(&count)
	if err != nil {
		log.Printf("Error checking team existence: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count == 0 {
		log.Printf("Team not found or doesn't belong to regatta - TeamID: %s, RegattaID: %s", teamId, regattaId)
		http.Error(w, "Team not found or doesn't belong to this regatta", http.StatusNotFound)
		return
	}
	log.Printf("Team verification successful - Found %d matching teams", count)

	// Update the team
	result, err := db.DB.Exec("UPDATE teams SET name = $1 WHERE id = $2 AND regatta_id = $3",
		team.Name, teamId, regattaId)
	if err != nil {
		log.Printf("Error executing update query: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
	} else {
		log.Printf("Update successful - Rows affected: %d", rowsAffected)
	}

	// Return updated team
	team.ID = teamId
	team.RegattaID = regattaId
	w.Header().Set("Content-Type", "application/json")

	responseJSON, err := json.Marshal(team)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Sending response: %s", string(responseJSON))

	json.NewEncoder(w).Encode(team)
}
