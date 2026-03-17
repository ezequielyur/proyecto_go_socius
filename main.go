package main

import (
	"encoding/json"
	"fmt"
	"game-library-api/config"
	"game-library-api/models"
	"game-library-api/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

type ByStatus struct {
	Completado int `json:"completado"`
	Jugando    int `json:"jugando"`
	Pendiente  int `json:"pendiente"`
	Abandonado int `json:"abandonado"`
}

type StatsResponse struct {
	Total        int      `json:"total"`
	ByStatus     ByStatus `json:"by_status"`
	AverageScore float64  `json:"average_score"`
}

func main() {
	cfg := config.LoadConfig()

	var err error
	db, err = sqlx.Connect("mysql", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Error al conectar DB:", err)
	}
	defer db.Close()

	log.Println("DB conectada correctamente")

	// RAWG API ORIGINAL
	http.HandleFunc("/api/games", searchGames)
	http.HandleFunc("/api/games/", getGameDetails)

	// Library
	http.HandleFunc("/api/library", libraryHandler)
	http.HandleFunc("/api/library/", libraryWithIDHandler)
	http.HandleFunc("/api/library/stats", getLibraryStats)

	log.Println("Servidor corriendo en :8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}

// Helpers

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func parseIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

// RAWG API ORIGINAL

func searchGames(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("search")
	cfg := config.LoadConfig()

	url := fmt.Sprintf("https://api.rawg.io/api/games?key=%s&search=%s", cfg.RawgKey, query)

	resp, err := http.Get(url)

	if resp.StatusCode != http.StatusOK {
		log.Printf("RAWG devolvió status %d", resp.StatusCode)
		writeJSON(w, 502, map[string]string{"error": "Error en servicio externo (RAWG)"})
		return
	}

	if err != nil {
		log.Printf("Error desconocido al intentar consultar RAWG: %v", err)
		writeJSON(w, 502, map[string]string{"error": "Error consultando RAWG, intente nuevamente"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error desconocido al intentar decodificar JSON de RAWG: %v", err)
		writeJSON(w, 502, map[string]string{"error": "Error en la respuesta de RAWG, intente nuevament"})
		return
	}

	writeJSON(w, 200, result)
}

func getGameDetails(w http.ResponseWriter, r *http.Request) {
	id := parseIDFromPath(r.URL.Path)
	cfg := config.LoadConfig()

	url := fmt.Sprintf("https://api.rawg.io/api/games/%s?key=%s", id, cfg.RawgKey)

	resp, err := http.Get(url)

	if resp.StatusCode != http.StatusOK {
		log.Printf("RAWG devolvió status %d", resp.StatusCode)
		writeJSON(w, 502, map[string]string{"error": "Error en servicio externo (RAWG)"})
		return
	}

	if err != nil {
		log.Printf("Error desconocido al intentar consultar RAWG: %v", err)
		writeJSON(w, 502, map[string]string{"error": "Error consultando RAWG, intente nuevamente"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error desconocido al intentar decodificar JSON de RAWG: %v", err)
		writeJSON(w, 502, map[string]string{"error": "Error en la respuesta de RAWG, intente nuevamente"})
		return
	}

	writeJSON(w, 200, result)
}

// LIBRARY ROUTERS

func libraryHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listGames(w, r)
	case http.MethodPost:
		addGame(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func libraryWithIDHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		updateGame(w, r)
	case http.MethodDelete:
		deleteGame(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// HANDLERS

func listGames(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	var games []models.Game
	query := "SELECT * FROM game_library"

	var err error
	if status != "" {
		query += " WHERE status = ?"
		err = db.Select(&games, query, status)
	} else {
		err = db.Select(&games, query)
	}

	if err != nil {
		log.Printf("Error desconocido al intentar consultar los recursos: %v", err)
		writeJSON(w, 500, map[string]string{"error": "Error al consultar los recursos, por favor intente nuevamente."})
		return
	}

	writeJSON(w, 200, games)
}

func addGame(w http.ResponseWriter, r *http.Request) {
	var game models.Game

	if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
		writeJSON(w, 400, map[string]string{"error": "El JSON propocionado es inválido"})
		return
	}

	if game.RawgID == 0 || game.Title == "" {
		writeJSON(w, 400, map[string]string{"error": "Campos faltantes en el JSON"})
		return
	}

	result, err := db.Exec(
		"INSERT INTO game_library (rawg_id, title, genre, platform, cover_url) VALUES (?, ?, ?, ?, ?)",
		game.RawgID, game.Title, game.Genre, game.Platform, game.CoverURL,
	)

	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == utils.RAWG_ID_DUPLICATED {
				log.Printf("El recurso con RAWG ID %v ya existe en la BD", game.RawgID)
				writeJSON(w, 409, map[string]string{"error": "Recurso duplicado"})
				return
			} else {
				log.Printf("Error desconocido al intentar agregar recurso, RAWG ID %v, Error: %v\n", game.RawgID, err)
			}
		}
		writeJSON(w, 500, map[string]string{"error": "Error desconocido, vuelva a intentar nuevamente"})
		return
	}

	id, _ := result.LastInsertId()
	game.ID = int(id)
	game.AddedAt = time.Now().Format("2006-01-02")

	writeJSON(w, 201, game)
}

func updateGame(w http.ResponseWriter, r *http.Request) {
	id := parseIDFromPath(r.URL.Path)

	var updates models.Game
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSON(w, 400, map[string]string{"error": "JSON inválido"})
		return
	}

	query := "UPDATE game_library SET "
	args := []interface{}{}
	set := []string{}

	if updates.PersonalNote != nil {
		set = append(set, "personal_note = ?")
		args = append(args, updates.PersonalNote)
	}

	if updates.PersonalScore != nil {
		if *updates.PersonalScore < 0 || *updates.PersonalScore > 10 {
			log.Println("La puntuacion no se encuentra en el rango [0,10]")
			writeJSON(w, 400, map[string]string{"error": "La puntuacion debe pertenecer al rango [0,10]"})
			return
		}
		set = append(set, "personal_score = ?")
		args = append(args, updates.PersonalScore)
	}

	if updates.Status != nil {
		if !utils.IsValidStatus(*updates.Status) {
			writeJSON(w, 400, map[string]string{"error": "Status inválido"})
			return
		}
		set = append(set, "status = ?")
		args = append(args, updates.Status)
	}

	if len(set) == 0 {
		writeJSON(w, 400, map[string]string{"error": "No se proporciono ningun campo para actualizar"})
		return
	}

	query += strings.Join(set, ", ") + " WHERE rawg_id = ?"
	args = append(args, id)

	result, err := db.Exec(query, args...)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			log.Printf("MySQL error (%d): %s\n", mysqlErr.Number, mysqlErr.Message)
		} else {
			log.Printf("Error desconocido al intentar actualizar recurso, RAWG ID %v, Error: %v\n", id, err)
		}
		writeJSON(w, 500, map[string]string{"error": "Error desconocido, vuelva a intentar"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		writeJSON(w, 404, map[string]string{"error": "Recurso no encontrado"})
		log.Println("No se encontro el recurso a actualizar en la BD. RAWG ID: ", id)
		return
	}

	writeJSON(w, 200, map[string]string{"message": "Juego actualizado correctamente"})
}

func deleteGame(w http.ResponseWriter, r *http.Request) {
	rawg_id := parseIDFromPath(r.URL.Path)

	result, err := db.Exec("DELETE FROM game_library WHERE rawg_id = ?", rawg_id)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			log.Printf("MySQL error (%d): %s\n", mysqlErr.Number, mysqlErr.Message)
		} else {
			log.Printf("Error desconocido, RAWG ID %v, Error: %v\n", rawg_id, err)
		}

		writeJSON(w, 500, map[string]string{"error": "Error desconocido, vuelva a intentar"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		writeJSON(w, 404, map[string]string{"error": "Recurso no encontrado"})
		log.Println("No se encontro el recurso a eliminar en la BD. RAWG ID: ", rawg_id)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getLibraryStats(w http.ResponseWriter, r *http.Request) {
	var totalGames, completados, jugando, pendientes, abandonados int
	var averageScore, score float64

	db.Get(&totalGames, "SELECT COUNT(*) FROM game_library")
	db.Get(&completados, "SELECT COUNT(*) FROM game_library WHERE status = 'completado'")
	db.Get(&jugando, "SELECT COUNT(*) FROM game_library WHERE status = 'jugando'")
	db.Get(&pendientes, "SELECT COUNT(*) FROM game_library WHERE status = 'pendiente'")
	db.Get(&abandonados, "SELECT COUNT(*) FROM game_library WHERE status = 'abandonado'")
	db.Get(&score, "SELECT COALESCE(SUM(personal_score), 0) FROM game_library")

	if totalGames > 0 {
		averageScore = float64(score) / float64(totalGames)
	}

	response := StatsResponse{
		Total: totalGames,
		ByStatus: ByStatus{
			Completado: completados,
			Jugando:    jugando,
			Pendiente:  pendientes,
			Abandonado: abandonados,
		},
		AverageScore: averageScore,
	}

	writeJSON(w, 200, response)
}
