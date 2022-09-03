package main

import (
	"context"
	"os"

	app "github.com/youla-dev/schema/internal/app"

	"github.com/joho/godotenv"
)

var Version = "0.0"

func main() {
	_ = godotenv.Load(os.Getenv("SCHEMA_CONFIG"))

	app := app.NewApp(Version)
	app.Run(context.Background())
}
