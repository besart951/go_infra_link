package logger

import (
	"os"
	"log/slog"
)

// Setup initialisiert den Logger.
// env: "prod" für JSON output, "dev" für lesbaren Text
func Setup(env string) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug, // Oder aus Env-Variable laden
        // AddSource: true, // Zeigt Dateiname und Zeilennummer (teuer in Prod!)
	}

	if env == "prod" {
		// JSON für Log-Aggregatoren (Datadog, ELK, CloudWatch)
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Text für lokale Entwicklung (bunt und lesbar wenn unterstützt)
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
    
    // Optional: Setze ihn auch als globalen Default, falls mal eine 3rd Party Lib loggt
    slog.SetDefault(logger)

	return logger
}