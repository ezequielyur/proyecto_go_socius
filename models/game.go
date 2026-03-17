package models

// Estructura para el juego
type Game struct {
	ID            int     `db:"id" json:"id"`
	RawgID        int     `db:"rawg_id" json:"rawg_id"`
	Title         string  `db:"title" json:"title"`
	Genre         string  `db:"genre" json:"genre"`
	Platform      string  `db:"platform" json:"platform"`
	CoverURL      string  `db:"cover_url" json:"cover_url"`
	PersonalNote  *string `db:"personal_note" json:"personal_note"`
	PersonalScore *int    `db:"personal_score" json:"personal_score"`
	Status        *string `db:"status" json:"status"`
	AddedAt       string  `db:"added_at" json:"added_at"`
}
