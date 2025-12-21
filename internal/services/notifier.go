package services

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Status string

const (
	Idle     Status = "idle"
	Scanning Status = "scanning"
)

// ToastField definiuje strukturę wiadomości
type ToastField struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

// ScanProgressDTO - to jest to, co Frontend dostanie jako "data"
type ScanProgressDTO struct {
	Current  int    `json:"current"`
	Total    int    `json:"total"`
	LastFile string `json:"lastFile"`
}

// Notifier to nasz kontrakt. Mówi: "Umiem wysłać toasta".
type Notifier interface {
	SendToast(ctx context.Context, msg ToastField)
	SendScanProgress(ctx context.Context, current, total int, lastFile string)
	SendScannerStatus(ctx context.Context, status Status)
}

// NewNotifier tworzy nową instancję Notifier.
func NewNotifier() Notifier {
	return &WailsNotifier{}
}

// WailsNotifier to implementacja produkcyjna.
type WailsNotifier struct{}

// SendToast implementuje interfejs Notifier używając Wailsa.
func (n *WailsNotifier) SendToast(ctx context.Context, msg ToastField) {
	runtime.EventsEmit(ctx, "toast", msg)
}

func (n *WailsNotifier) SendScannerStatus(ctx context.Context, status Status) {
	runtime.EventsEmit(ctx, "scan_status", string(status))
}

func (n *WailsNotifier) SendScanProgress(ctx context.Context, current, total int, lastFile string) {
	payload := ScanProgressDTO{
		Current:  current,
		Total:    total,
		LastFile: lastFile,
	}
	runtime.EventsEmit(ctx, "scan_progress", payload)
}
