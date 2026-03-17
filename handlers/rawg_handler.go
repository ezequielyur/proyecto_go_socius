package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler para buscar juegos en RAWG
func searchGames(c *gin.Context, rawgKey string) {
	query := c.Query("search")
	resp, err := http.Get(fmt.Sprintf("https://api.rawg.io/api/games?key=%s&search=%s", rawgKey, query))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	c.JSON(http.StatusOK, body)
}

// Handler para obtener detalles de un juego
func getGameDetails(c *gin.Context) {
	rawgID := c.Param("rawg_id")
	resp, err := http.Get(fmt.Sprintf("https://api.rawg.io/api/games/%s?key=YOUR_API_KEY", rawgID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	c.JSON(http.StatusOK, body)
}
