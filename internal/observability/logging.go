package observability

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
)

func ConfigureLogging(level string) {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	LogEvent("logging.configured", map[string]any{
		"level": strings.ToUpper(strings.TrimSpace(level)),
	})
}

func LogEvent(event string, fields map[string]any) {
	payload := map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"event":     event,
	}
	for key, value := range fields {
		payload[key] = value
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf(`{"timestamp":"%s","event":"logging.error","detail":"marshal failed","original_event":%q}`, time.Now().UTC().Format(time.RFC3339Nano), event)
		return
	}
	log.Print(string(data))
}
