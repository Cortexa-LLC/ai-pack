package monitoring

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DailyUsage tracks token usage for a single day
type DailyUsage struct {
	Date              string                         `json:"date"` // YYYY-MM-DD
	TotalInputTokens  int64                          `json:"total_input_tokens"`
	TotalOutputTokens int64                          `json:"total_output_tokens"`
	TotalCost         float64                        `json:"total_cost"`    // Actual cost in USD
	TotalSavings      float64                        `json:"total_savings"` // Savings vs reference model in USD
	ProviderBreakdown map[string]*ProviderDailyUsage `json:"provider_breakdown"`
	LastUpdated       time.Time                      `json:"last_updated"`
}

// ProviderDailyUsage tracks daily usage per provider/model
type ProviderDailyUsage struct {
	Provider     string  `json:"provider"`
	Model        string  `json:"model"`
	Calls        int64   `json:"calls"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	Cost         float64 `json:"cost"`    // Calculated cost in USD
	Savings      float64 `json:"savings"` // Savings vs reference model in USD (for free/local providers)
}

// PersistentMetrics manages disk-based metrics storage
type PersistentMetrics struct {
	mu             sync.RWMutex
	dataDir        string
	currentDate    string
	currentDay     *DailyUsage
	costs          map[string][2]float64 // Cost per 1M tokens [input, output]
	freeProviders  map[string]bool       // Providers that are free (local/no-cost)
	referenceCosts [2]float64            // Reference model costs [input, output] per 1M tokens
}

// NewPersistentMetrics creates a new persistent metrics tracker.
// freeProviders lists provider names (e.g. "qwen") whose usage counts as savings.
// referenceModel is the model ID to compare against (e.g. "claude-sonnet-4-6");
// its cost is looked up from the model registry; falls back to claude-sonnet-4-6 pricing.
func NewPersistentMetrics(dataDir string, costs map[string][2]float64, freeProviders []string, referenceModel string) (*PersistentMetrics, error) {
	// Create data directory if it doesn't exist
	metricsDir := filepath.Join(dataDir, ".claude", "metrics", "daily")
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create metrics directory: %w", err)
	}

	// Build free providers set
	free := make(map[string]bool, len(freeProviders))
	for _, p := range freeProviders {
		free[p] = true
	}

	// Resolve reference model costs from the model registry.
	// Default to claude-sonnet-4-6 if not found.
	const defaultRefIn, defaultRefOut = 3.00, 15.00
	var refIn, refOut float64 = defaultRefIn, defaultRefOut
	if referenceModel != "" {
		for _, models := range ModelsByTier {
			for _, m := range models {
				if m.ID == referenceModel {
					refIn = m.CostPerMIn
					refOut = m.CostPerMOut
					break
				}
			}
		}
	}

	pm := &PersistentMetrics{
		dataDir:        metricsDir,
		costs:          costs,
		freeProviders:  free,
		referenceCosts: [2]float64{refIn, refOut},
	}

	// Load or create today's usage
	if err := pm.ensureCurrentDay(); err != nil {
		return nil, err
	}

	return pm, nil
}

// ensureCurrentDay ensures we have a DailyUsage for today
func (pm *PersistentMetrics) ensureCurrentDay() error {
	today := time.Now().Format("2006-01-02")

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// If we're already tracking today, nothing to do
	if pm.currentDate == today && pm.currentDay != nil {
		return nil
	}

	// Try to load today's data from disk
	dailyPath := filepath.Join(pm.dataDir, fmt.Sprintf("%s.json", today))
	var daily DailyUsage

	data, err := os.ReadFile(dailyPath)
	if err == nil {
		// File exists, load it
		if err := json.Unmarshal(data, &daily); err != nil {
			return fmt.Errorf("failed to parse daily metrics: %w", err)
		}
	} else if os.IsNotExist(err) {
		// File doesn't exist, create new
		daily = DailyUsage{
			Date:              today,
			ProviderBreakdown: make(map[string]*ProviderDailyUsage),
			LastUpdated:       time.Now(),
		}
	} else {
		return fmt.Errorf("failed to read daily metrics: %w", err)
	}

	pm.currentDate = today
	pm.currentDay = &daily

	return nil
}

// RecordUsage records token usage for a provider/model
func (pm *PersistentMetrics) RecordUsage(provider, model string, inputTokens, outputTokens int64) error {
	if err := pm.ensureCurrentDay(); err != nil {
		return err
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Update totals
	pm.currentDay.TotalInputTokens += inputTokens
	pm.currentDay.TotalOutputTokens += outputTokens

	// Update provider breakdown
	key := provider + ":" + model
	usage, exists := pm.currentDay.ProviderBreakdown[key]
	if !exists {
		usage = &ProviderDailyUsage{
			Provider: provider,
			Model:    model,
		}
		pm.currentDay.ProviderBreakdown[key] = usage
	}

	usage.Calls++
	usage.InputTokens += inputTokens
	usage.OutputTokens += outputTokens

	// Calculate actual cost (zero for free/local providers)
	if costs, ok := pm.costs[key]; ok {
		inputCost := float64(inputTokens) / 1_000_000 * costs[0]
		outputCost := float64(outputTokens) / 1_000_000 * costs[1]
		usage.Cost += inputCost + outputCost
		pm.currentDay.TotalCost += inputCost + outputCost
	}

	// For free/local providers, compute what this would have cost on the reference model
	if pm.freeProviders[provider] {
		savedIn := float64(inputTokens) / 1_000_000 * pm.referenceCosts[0]
		savedOut := float64(outputTokens) / 1_000_000 * pm.referenceCosts[1]
		usage.Savings += savedIn + savedOut
		pm.currentDay.TotalSavings += savedIn + savedOut
	}

	pm.currentDay.LastUpdated = time.Now()

	// Save to disk
	return pm.save()
}

// save writes current day's data to disk
func (pm *PersistentMetrics) save() error {
	if pm.currentDay == nil {
		return nil
	}

	dailyPath := filepath.Join(pm.dataDir, fmt.Sprintf("%s.json", pm.currentDate))

	data, err := json.MarshalIndent(pm.currentDay, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal daily metrics: %w", err)
	}

	if err := os.WriteFile(dailyPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write daily metrics: %w", err)
	}

	return nil
}

// GetToday returns today's usage
func (pm *PersistentMetrics) GetToday() (*DailyUsage, error) {
	if err := pm.ensureCurrentDay(); err != nil {
		return nil, err
	}

	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy
	copy := *pm.currentDay
	copy.ProviderBreakdown = make(map[string]*ProviderDailyUsage)
	for k, v := range pm.currentDay.ProviderBreakdown {
		usageCopy := *v
		copy.ProviderBreakdown[k] = &usageCopy
	}

	return &copy, nil
}

// GetDateRange returns usage for a date range
func (pm *PersistentMetrics) GetDateRange(startDate, endDate string) ([]*DailyUsage, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	var result []*DailyUsage

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dailyPath := filepath.Join(pm.dataDir, fmt.Sprintf("%s.json", dateStr))

		data, err := os.ReadFile(dailyPath)
		if os.IsNotExist(err) {
			// No data for this day, skip
			continue
		} else if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", dateStr, err)
		}

		var daily DailyUsage
		if err := json.Unmarshal(data, &daily); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", dateStr, err)
		}

		result = append(result, &daily)
	}

	return result, nil
}

// GetLast30Days returns usage for the last 30 days
func (pm *PersistentMetrics) GetLast30Days() ([]*DailyUsage, error) {
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	return pm.GetDateRange(startDate, endDate)
}
