package feedback

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Status represents the current operational state of the scanner.
type Status string

const (
	// Idle indicates the scanner is not currently processing any files.
	Idle Status = "idle"
	// Scanning indicates a scan operation is in progress.
	Scanning Status = "scanning"
)

// ToastField defines the structure for toast notification messages sent to the frontend.
type ToastField struct {
	Type    string `json:"type"` // e.g., "info", "success", "error", "warning"
	Title   string `json:"title"`
	Message string `json:"message"`
}

// ScanProgressDTO defines the data structure for progress updates sent to the frontend.
type ScanProgressDTO struct {
	Current  int    `json:"current"`
	Total    int    `json:"total"`
	LastFile string `json:"lastFile"`
}

// Notifier defines the interface for sending notifications and updates to the user interface.
type Notifier interface {
	// SendToast sends a temporary popup notification.
	SendToast(ctx context.Context, msg ToastField)
	// SendScanProgress updates the progress bar with current scan statistics.
	SendScanProgress(ctx context.Context, current, total int, lastFile string)
	// SendScannerStatus updates the overall status of the scanner (e.g., Idle, Scanning).
	SendScannerStatus(ctx context.Context, status Status)
	// EmitAssetsChanged signals that the asset library has changed and views should refresh.
	EmitAssetsChanged(ctx context.Context)
}

// NewNotifier creates a new instance of the default Wails-based Notifier.
func NewNotifier() Notifier {
	return &WailsNotifier{}
}

// WailsNotifier is the production implementation of Notifier that uses Wails runtime events.
type WailsNotifier struct{}

// SendToast emits a "toast" event to the frontend.
func (n *WailsNotifier) SendToast(ctx context.Context, msg ToastField) {
	if ctx == nil {
		return
	}
	runtime.EventsEmit(ctx, "toast", msg)
}

// SendScannerStatus emits a "scan_status" event to the frontend.
func (n *WailsNotifier) SendScannerStatus(ctx context.Context, status Status) {
	if ctx == nil {
		return
	}
	runtime.EventsEmit(ctx, "scan_status", string(status))
}

// SendScanProgress emits a "scan_progress" event to the frontend.
func (n *WailsNotifier) SendScanProgress(ctx context.Context, current, total int, lastFile string) {
	if ctx == nil {
		return
	}
	payload := ScanProgressDTO{
		Current:  current,
		Total:    total,
		LastFile: lastFile,
	}
	runtime.EventsEmit(ctx, "scan_progress", payload)
}

// EmitAssetsChanged emits an "assets:changed" event to trigger a frontend refresh.
func (n *WailsNotifier) EmitAssetsChanged(ctx context.Context) {
	if ctx == nil {
		return
	}
	runtime.EventsEmit(ctx, "assets:changed", "refresh")
}
