package main

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/hypebeast/go-osc/osc"
)

// Resolume connects a column's clips as a side effect of triggering the column,
// so clip connections seen shortly after a column trigger are treated as the same event.
const columnClipDedupeWindow = 2 * time.Second

func isBlackoutName(name string) bool {
	return strings.Contains(strings.ToLower(name), "blackout")
}

type clipRef struct {
	columnName string
	known      bool
	connected  bool
}

// BlackoutMonitor sends a QLab cue whenever a blackout column, or a clip inside one, is triggered.
type BlackoutMonitor struct {
	mu                sync.Mutex
	client            *osc.Client // nil when QLab is not configured
	cue               string
	clips             map[int64]*clipRef // keyed by the clip's "connected" parameter ID
	lastColumnTrigger map[string]time.Time
}

func NewBlackoutMonitor(qlab *QLab) *BlackoutMonitor {
	m := &BlackoutMonitor{
		clips:             map[int64]*clipRef{},
		lastColumnTrigger: map[string]time.Time{},
	}
	if qlab != nil {
		m.client = osc.NewClient(qlab.IP, int(qlab.OSCPort))
		m.cue = qlab.BlackoutCue
	}
	return m
}

// SetComposition rebuilds the tracked clips from a freshly received composition
// and returns the parameter IDs that should be subscribed to.
func (m *BlackoutMonitor) SetComposition(c *CompositionConfig) []int64 {
	if m.client == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// Resolume re-sends the whole composition on every change, so keep the
	// known connect state for clips that are still present.
	previous := m.clips
	m.clips = map[int64]*clipRef{}
	var ids []int64
	for colIndex, col := range c.Columns {
		if !isBlackoutName(col.Name.Value) {
			continue
		}
		// Clips live on the layers: layer.Clips[colIndex] is this column's clip on that layer.
		for _, layer := range c.Layers {
			if colIndex >= len(layer.Clips) {
				continue
			}
			clipID := layer.Clips[colIndex].Connected.ID
			if clipID == 0 {
				continue
			}
			ref := &clipRef{columnName: col.Name.Value}
			if old, ok := previous[clipID]; ok {
				ref.known, ref.connected = old.known, old.connected
			}
			m.clips[clipID] = ref
			ids = append(ids, clipID)
		}
	}
	return ids
}

// ColumnTriggered is called every time a column is triggered over OSC.
func (m *BlackoutMonitor) ColumnTriggered(columnName string) {
	if m.client == nil || !isBlackoutName(columnName) {
		return
	}
	m.mu.Lock()
	m.lastColumnTrigger[columnName] = time.Now()
	m.mu.Unlock()

	m.fire("column '" + columnName + "'")
}

// ClipUpdate is called when Resolume pushes a new value for a subscribed parameter.
// A cue fires on each transition of a clip from disconnected to connected.
func (m *BlackoutMonitor) ClipUpdate(id int64, value interface{}) {
	if m.client == nil {
		return
	}
	on := isConnected(value)

	m.mu.Lock()
	ref, ok := m.clips[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	wasKnown, wasOn := ref.known, ref.connected
	ref.known, ref.connected = true, on

	last, seen := m.lastColumnTrigger[ref.columnName]
	fromColumnTrigger := seen && time.Since(last) < columnClipDedupeWindow
	m.mu.Unlock()

	// The first value seen is just the current state, not a trigger.
	if !wasKnown || wasOn || !on || fromColumnTrigger {
		return
	}
	m.fire("clip in '" + ref.columnName + "'")
}

func (m *BlackoutMonitor) fire(reason string) {
	if m.cue == "" {
		log.Printf("Blackout triggered by %s, but qlab.blackout_cue is not set", reason)
		return
	}
	address := "/cue/" + m.cue + "/go"
	if err := m.client.Send(osc.NewMessage(address)); err != nil {
		log.Printf("Failed to send blackout cue to QLab: %v", err)
		return
	}
	log.Printf("Blackout triggered by %s: sent %s to QLab", reason, address)
}

func isConnected(v interface{}) bool {
	switch x := v.(type) {
	case bool:
		return x
	case float64:
		return x != 0
	case string:
		// ParamState values, e.g. "Connected" or "Connected & previewing".
		return strings.HasPrefix(x, "Connected")
	}
	return false
}
