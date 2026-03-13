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
	"sync/atomic"
	"time"
)

var haEntitiesWidgetTemplate = mustParseTemplate("ha-entities.html", "widget-base.html")

var (
	haEntitiesRegistryMu sync.RWMutex
	haEntitiesRegistry   = map[string]*haEntitiesWidget{}
)

var haEntitiesHTTPClient = &http.Client{Timeout: 10 * time.Second}
var haEntitiesIDCounter atomic.Int64

var haEntitiesToggleDomains = map[string]bool{
	"switch":        true,
	"light":         true,
	"input_boolean": true,
}

type haEntityConfig struct {
	EntityID   string          `yaml:"entity"`
	Title      string          `yaml:"title"`
	Icon       customIconField `yaml:"icon"`
	Domain     string          `yaml:"-"`
	Toggleable bool            `yaml:"-"`
	IsScript   bool            `yaml:"-"`
	IsSensor   bool            `yaml:"-"`
}

type haEntityState struct {
	EntityID string `json:"entity_id"`
	State    string `json:"state"`
	Unit     string `json:"unit,omitempty"`
}

type haEntitiesWidget struct {
	widgetBase `yaml:",inline"`
	cachedHTML template.HTML `yaml:"-"`
	WidgetID   string           `yaml:"id"`
	URL        string           `yaml:"url"`
	Token      string           `yaml:"token"`
	Columns    int              `yaml:"columns"`
	Entities   []haEntityConfig `yaml:"entities"`
	statesMu   sync.RWMutex     `yaml:"-"`
	states     map[string]haEntityState `yaml:"-"`
}

func (widget *haEntitiesWidget) initialize() error {
	widget.withTitle("Home Assistant").withCacheDuration(1 * time.Minute)

	if widget.URL == "" {
		return errors.New("ha-entities: url is required")
	}
	if widget.Token == "" {
		return errors.New("ha-entities: token is required")
	}
	if len(widget.Entities) == 0 {
		return errors.New("ha-entities: at least one entity is required")
	}
	if widget.Columns <= 0 {
		widget.Columns = 3
	}
	if widget.WidgetID == "" {
		n := haEntitiesIDCounter.Add(1)
		widget.WidgetID = fmt.Sprintf("ha-entities-%d", n)
	}

	for i := range widget.Entities {
		e := &widget.Entities[i]
		if idx := strings.IndexByte(e.EntityID, '.'); idx >= 0 {
			e.Domain = e.EntityID[:idx]
		}
		e.Toggleable = haEntitiesToggleDomains[e.Domain]
		e.IsScript = e.Domain == "script"
		e.IsSensor = e.Domain == "sensor"
		if e.Title == "" {
			e.Title = e.EntityID
		}
	}

	widget.states = make(map[string]haEntityState)

	haEntitiesRegistryMu.Lock()
	haEntitiesRegistry[widget.WidgetID] = widget
	haEntitiesRegistryMu.Unlock()

	widget.withError(nil)
	widget.cachedHTML = widget.renderTemplate(widget, haEntitiesWidgetTemplate)
	return nil
}

func (widget *haEntitiesWidget) update(ctx context.Context) {
	states, err := widget.fetchStates(ctx)
	if err != nil {
		widget.withError(err).scheduleEarlyUpdate()
		return
	}

	widget.statesMu.Lock()
	widget.states = states
	widget.statesMu.Unlock()

	widget.withError(nil).scheduleNextUpdate()
}

func (widget *haEntitiesWidget) fetchStates(ctx context.Context) (map[string]haEntityState, error) {
	req, err := http.NewRequestWithContext(
		ctx, "GET",
		strings.TrimRight(widget.URL, "/")+"/api/states",
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+widget.Token)

	resp, err := haEntitiesHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("home assistant returned status %d", resp.StatusCode)
	}

	var all []struct {
		EntityID   string `json:"entity_id"`
		State      string `json:"state"`
		Attributes struct {
			Unit string `json:"unit_of_measurement"`
		} `json:"attributes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&all); err != nil {
		return nil, fmt.Errorf("decoding HA response: %w", err)
	}

	needed := make(map[string]bool, len(widget.Entities))
	for _, e := range widget.Entities {
		if !e.IsScript {
			needed[e.EntityID] = true
		}
	}

	states := make(map[string]haEntityState, len(widget.Entities))
	for _, item := range all {
		if needed[item.EntityID] {
			states[item.EntityID] = haEntityState{
				EntityID: item.EntityID,
				State:    item.State,
				Unit:     item.Attributes.Unit,
			}
		}
	}
	return states, nil
}

func (widget *haEntitiesWidget) callHAService(ctx context.Context, domain, service, entityID string) error {
	reqBody, _ := json.Marshal(map[string]string{"entity_id": entityID})

	req, err := http.NewRequestWithContext(
		ctx, "POST",
		strings.TrimRight(widget.URL, "/")+"/api/services/"+domain+"/"+service,
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+widget.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := haEntitiesHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("home assistant returned status %d", resp.StatusCode)
	}
	return nil
}

func (widget *haEntitiesWidget) refreshAndReturnStates(ctx context.Context, w http.ResponseWriter) {
	states, err := widget.fetchStates(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	widget.statesMu.Lock()
	widget.states = states
	widget.statesMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(states)
}

func (widget *haEntitiesWidget) getHandlerFunc() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"GET /{id}":         haEntitiesHandleGet,
		"POST /{id}/toggle": haEntitiesHandleToggle,
		"POST /{id}/run":    haEntitiesHandleRun,
	}
}

func (widget *haEntitiesWidget) Render() template.HTML {
	return widget.cachedHTML
}

func lookupHaEntities(w http.ResponseWriter, r *http.Request) *haEntitiesWidget {
	id := r.PathValue("id")
	haEntitiesRegistryMu.RLock()
	widget, ok := haEntitiesRegistry[id]
	haEntitiesRegistryMu.RUnlock()
	if !ok {
		http.Error(w, "ha-entities: widget not found", http.StatusNotFound)
		return nil
	}
	return widget
}

func haEntitiesHandleGet(w http.ResponseWriter, r *http.Request) {
	widget := lookupHaEntities(w, r)
	if widget == nil {
		return
	}

	widget.refreshAndReturnStates(r.Context(), w)
}

func haEntitiesHandleToggle(w http.ResponseWriter, r *http.Request) {
	widget := lookupHaEntities(w, r)
	if widget == nil {
		return
	}

	var body struct {
		EntityID string `json:"entity_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.EntityID == "" {
		http.Error(w, "entity_id is required", http.StatusBadRequest)
		return
	}

	// Verify entity is in this widget and is toggleable
	var domain string
	for _, e := range widget.Entities {
		if e.EntityID == body.EntityID {
			if !e.Toggleable {
				http.Error(w, "entity is not toggleable", http.StatusBadRequest)
				return
			}
			domain = e.Domain
			break
		}
	}
	if domain == "" {
		http.Error(w, "entity not found in widget", http.StatusNotFound)
		return
	}

	// Determine turn_on or turn_off from current cached state
	widget.statesMu.RLock()
	current, ok := widget.states[body.EntityID]
	widget.statesMu.RUnlock()

	service := "turn_on"
	if ok && current.State == "on" {
		service = "turn_off"
	}

	if err := widget.callHAService(r.Context(), domain, service, body.EntityID); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	widget.refreshAndReturnStates(r.Context(), w)
}

func haEntitiesHandleRun(w http.ResponseWriter, r *http.Request) {
	widget := lookupHaEntities(w, r)
	if widget == nil {
		return
	}

	var body struct {
		EntityID string `json:"entity_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.EntityID == "" {
		http.Error(w, "entity_id is required", http.StatusBadRequest)
		return
	}

	found := false
	for _, e := range widget.Entities {
		if e.EntityID == body.EntityID && e.IsScript {
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "script not found in widget", http.StatusNotFound)
		return
	}

	if err := widget.callHAService(r.Context(), "script", "turn_on", body.EntityID); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
}
