package service

import (
	"encoding/json"
	"os"

	"ap_final/internal/domain"
)

func startAuditWorker(ch <-chan domain.AuditEvent) {
	f, err := os.OpenFile("audit.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// если файл не открылся, просто молча не логируем (milestone, не NASA)
		return
	}

	enc := json.NewEncoder(f)
	for ev := range ch {
		_ = enc.Encode(ev) // одна JSON-строка на событие
	}
}
