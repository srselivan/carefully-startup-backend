package main

import (
	"investment-game-backend/internal/app"
	"investment-game-backend/internal/config"
)

func main() {
	app.Run(config.New())
}
