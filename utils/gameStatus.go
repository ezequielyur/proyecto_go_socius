package utils

type GAME_STATUS string

const (
	Completado GAME_STATUS = "completado"
	Jugando    GAME_STATUS = "jugando"
	Pendiente  GAME_STATUS = "pendiente"
	Abandonado GAME_STATUS = "abandonado"
)

func IsValidStatus(s string) bool {
	switch GAME_STATUS(s) {
	case Completado, Jugando, Pendiente, Abandonado:
		return true
	}
	return false
}
