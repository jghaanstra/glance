package glance

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"time"
)

var haTodoWidgetTemplate = mustParseTemplate("ha-todo.html", "widget-base.html")

var (
	haTodoRegistryMu sync.RWMutex
	haTodoRegistry   = map[string]*haTodoWidget{}
)

var haTodoHTTPClient = &http.Client{Timeout: 10 * time.Second}

type haTodoItem struct {
	UID     string `json:"uid"`
	Summary string `json:"summary"`
	Status  string `json:"status"`
	Due     string `json:"due,omitempty"`
}

type haTodoWidget struct {
	widgetBase       `yaml:",inline"`
	cachedHTML       template.HTML `yaml:"-"`
	WidgetID         string        `yaml:"id"`
	URL              string        `yaml:"url"`
	Token            string        `yaml:"token"`
	Entity           string        `yaml:"entity"`
	ShowNeedsAction  *bool         `yaml:"show-needs-action"`
	ShowCompleted    *bool         `yaml:"show-completed"`
	EmptyText        string        `yaml:"empty-text"`
	AddPlaceholder   string        `yaml:"add-placeholder"`
	itemsMu          sync.RWMutex  `yaml:"-"`
	items            []haTodoItem  `yaml:"-"`
}

func (widget *haTodoWidget) initialize() error {
	widget.withTitle("To-Do").withCacheDuration(5 * time.Minute)

	if widget.URL == "" {
		return errors.New("ha-todo: url is required")
	}
	if widget.Token == "" {
		return errors.New("ha-todo: token is required")
	}
	if widget.Entity == "" {
		return errors.New("ha-todo: entity is required")
	}
	if widget.WidgetID == "" {
		widget.WidgetID = strings.NewReplacer(".", "-", "_", "-", " ", "-").Replace(widget.Entity)
	}

	if widget.EmptyText == "" {
		widget.EmptyText = "No open tasks!"
	}
	if widget.AddPlaceholder == "" {
		widget.AddPlaceholder = "Add a task\u2026"
	}

	haTodoRegistryMu.Lock()
	haTodoRegistry[widget.WidgetID] = widget
	haTodoRegistryMu.Unlock()

	widget.items = []haTodoItem{}
	widget.withError(nil)
	widget.cachedHTML = widget.renderTemplate(widget, haTodoWidgetTemplate)
	return nil
}

func (widget *haTodoWidget) update(ctx context.Context) {
	items, err := widget.fetchFromHA(ctx)
	if err != nil {
		widget.withError(err).scheduleEarlyUpdate()
		return
	}

	widget.itemsMu.Lock()
	widget.items = items
	widget.itemsMu.Unlock()

	widget.withError(nil).scheduleNextUpdate()
}

func (widget *haTodoWidget) fetchFromHA(ctx context.Context) ([]haTodoItem, error) {
	showNeedsAction := widget.ShowNeedsAction == nil || *widget.ShowNeedsAction
	showCompleted := widget.ShowCompleted != nil && *widget.ShowCompleted

	statusFilter := make([]string, 0, 2)
	if showNeedsAction {
		statusFilter = append(statusFilter, "needs_action")
	}
	if showCompleted {
		statusFilter = append(statusFilter, "completed")
	}
	if len(statusFilter) == 0 {
		return []haTodoItem{}, nil
	}

	reqBody, _ := json.Marshal(map[string]any{
		"entity_id": widget.Entity,
		"status":    statusFilter,
	})

	req, err := http.NewRequestWithContext(
		ctx, "POST",
		strings.TrimRight(widget.URL, "/")+"/api/services/todo/get_items?return_response",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+widget.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := haTodoHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("home assistant returned status %d", resp.StatusCode)
	}

	var result struct {
		ServiceResponse map[string]struct {
			Items []haTodoItem `json:"items"`
		} `json:"service_response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding HA response: %w", err)
	}

	for _, v := range result.ServiceResponse {
		if v.Items == nil {
			return []haTodoItem{}, nil
		}
		return v.Items, nil
	}
	return []haTodoItem{}, nil
}

func (widget *haTodoWidget) callHA(ctx context.Context, action string, body map[string]any) error {
	body["entity_id"] = widget.Entity
	reqBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(
		ctx, "POST",
		strings.TrimRight(widget.URL, "/")+"/api/services/todo/"+action,
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+widget.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := haTodoHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("home assistant returned status %d", resp.StatusCode)
	}
	return nil
}

// refreshAndRespond fetches fresh items from HA, updates the cache, and writes them as JSON.
func (widget *haTodoWidget) refreshAndRespond(ctx context.Context, w http.ResponseWriter) {
	items, err := widget.fetchFromHA(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	if items == nil {
		items = []haTodoItem{}
	}

	widget.itemsMu.Lock()
	widget.items = items
	widget.itemsMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (widget *haTodoWidget) getHandlerFunc() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"GET /{id}":          haTodoHandleGet,
		"POST /{id}/items":   haTodoHandleAdd,
		"PATCH /{id}/items":  haTodoHandleUpdate,
		"DELETE /{id}/items": haTodoHandleDelete,
	}
}

func (widget *haTodoWidget) Render() template.HTML {
	return widget.cachedHTML
}

func lookupHaTodo(w http.ResponseWriter, r *http.Request) *haTodoWidget {
	id := r.PathValue("id")
	haTodoRegistryMu.RLock()
	widget, ok := haTodoRegistry[id]
	haTodoRegistryMu.RUnlock()
	if !ok {
		http.Error(w, "ha-todo: widget not found", http.StatusNotFound)
		return nil
	}
	return widget
}

func haTodoHandleGet(w http.ResponseWriter, r *http.Request) {
	widget := lookupHaTodo(w, r)
	if widget == nil {
		return
	}

	widget.itemsMu.RLock()
	items := widget.items
	widget.itemsMu.RUnlock()

	if items == nil {
		items = []haTodoItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func haTodoHandleAdd(w http.ResponseWriter, r *http.Request) {
	widget := lookupHaTodo(w, r)
	if widget == nil {
		return
	}

	var body struct {
		Summary string `json:"summary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Summary == "" {
		http.Error(w, "summary is required", http.StatusBadRequest)
		return
	}

	if err := widget.callHA(r.Context(), "add_item", map[string]any{"item": body.Summary}); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	widget.refreshAndRespond(r.Context(), w)
}

func haTodoHandleUpdate(w http.ResponseWriter, r *http.Request) {
	widget := lookupHaTodo(w, r)
	if widget == nil {
		return
	}

	var body struct {
		UID    string `json:"uid"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.UID == "" {
		http.Error(w, "uid is required", http.StatusBadRequest)
		return
	}
	if body.Status != "completed" && body.Status != "needs_action" {
		http.Error(w, "status must be 'completed' or 'needs_action'", http.StatusBadRequest)
		return
	}

	if err := widget.callHA(r.Context(), "update_item", map[string]any{
		"item":   body.UID,
		"status": body.Status,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	widget.refreshAndRespond(r.Context(), w)
}

func haTodoHandleDelete(w http.ResponseWriter, r *http.Request) {
	widget := lookupHaTodo(w, r)
	if widget == nil {
		return
	}

	var body struct {
		UID string `json:"uid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.UID == "" {
		http.Error(w, "uid is required", http.StatusBadRequest)
		return
	}

	if err := widget.callHA(r.Context(), "remove_item", map[string]any{"item": body.UID}); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	widget.refreshAndRespond(r.Context(), w)
}
