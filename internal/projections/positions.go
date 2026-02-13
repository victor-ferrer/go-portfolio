package projections

import (
	"context"
	"math"
	"time"

	"go-portfolio/internal/domain"
)

// Position represents the current position in an instrument.
type Position struct {
	Instrument    string    `json:"instrument"`
	Quantity      float64   `json:"quantity"`
	AverageCost   float64   `json:"average_cost"`
	CurrentValue  float64   `json:"current_value"`
	TotalCost     float64   `json:"total_cost"`
	FirstBuyDate  time.Time `json:"first_buy_date"`
	LastTradeDate time.Time `json:"last_trade_date"`
}

// ProjectOpenPositions aggregates events by instrument and calculates open positions.
func ProjectOpenPositions(ctx context.Context, events []domain.Event) (map[string]Position, error) {
	positions := make(map[string]Position)

	for _, event := range events {
		if event.Type != "TransactionImported" {
			continue
		}

		instrument := event.Payload.Instrument
		if instrument == "" {
			continue
		}

		position, exists := positions[instrument]
		if !exists {
			position = Position{
				Instrument:   instrument,
				FirstBuyDate: event.Payload.CreatedAt,
			}
		}

		// Update position based on transaction category
		if event.Payload.Category == "Trade" {
			// Extract quantity from Amount (for now, using Amount as quantity)
			// In a real implementation, this would come from a dedicated Quantity field
			quantity := event.Payload.Amount

			// Update total cost and quantity
			position.TotalCost += event.Payload.Amount * event.Payload.Amount // Amount as both price and qty for now
			position.Quantity += quantity

			// Recalculate average cost
			if position.Quantity > 0 {
				position.AverageCost = position.TotalCost / position.Quantity
			}
		}
		// Corporate actions typically don't change cost basis in the same way

		position.LastTradeDate = event.Payload.CreatedAt
		positions[instrument] = position
	}

	return positions, nil
}

// ProjectAnnualizedReturn calculates annualized return since first investment.
func ProjectAnnualizedReturn(ctx context.Context, events []domain.Event, position Position) float64 {
	if len(events) == 0 || position.Quantity == 0 {
		return 0
	}

	// Find first and last trade dates
	var firstDate, lastDate time.Time
	totalInvested := 0.0

	for _, event := range events {
		if event.AggregateID != position.Instrument || event.Payload.Category != "Trade" {
			continue
		}

		if firstDate.IsZero() {
			firstDate = event.Payload.CreatedAt
		}
		lastDate = event.Payload.CreatedAt

		// Accumulate invested amount (cost basis)
		totalInvested += event.Payload.Amount
	}

	if firstDate.IsZero() || totalInvested == 0 {
		return 0
	}

	// Calculate time held in years
	daysHeld := lastDate.Sub(firstDate).Hours() / 24
	yearsHeld := daysHeld / 365.25
	if yearsHeld == 0 {
		yearsHeld = 1 // Avoid division by zero
	}

	// Calculate current value (simplified: current price = average cost for now)
	currentValue := position.Quantity * position.AverageCost

	// Calculate simple return
	totalReturn := (currentValue - totalInvested) / totalInvested
	if math.IsNaN(totalReturn) || math.IsInf(totalReturn, 0) {
		return 0
	}

	// Annualize the return
	annualizedReturn := math.Pow(1+totalReturn, 1/yearsHeld) - 1

	if math.IsNaN(annualizedReturn) || math.IsInf(annualizedReturn, 0) {
		return 0
	}

	return annualizedReturn
}

// CalculatePortfolioMetrics aggregates metrics across all positions.
type PortfolioMetrics struct {
	TotalValue         float64
	TotalCost          float64
	UnrealizedGain     float64
	UnrealizedGainPct  float64
	AnnualizedReturn   float64
	PositionCount      int
}

// ProjectPortfolioMetrics calculates aggregated portfolio metrics.
func ProjectPortfolioMetrics(ctx context.Context, events []domain.Event, positions map[string]Position) (PortfolioMetrics, error) {
	metrics := PortfolioMetrics{}

	for _, position := range positions {
		metrics.TotalValue += position.CurrentValue
		metrics.TotalCost += position.TotalCost
		metrics.PositionCount++
	}

	if metrics.TotalCost > 0 {
		metrics.UnrealizedGain = metrics.TotalValue - metrics.TotalCost
		metrics.UnrealizedGainPct = (metrics.UnrealizedGain / metrics.TotalCost) * 100
	}

	// Calculate portfolio-wide annualized return (weighted average)
	if metrics.PositionCount > 0 {
		totalAnnualized := 0.0
		for _, position := range positions {
			// Get events for this position
			var positionEvents []domain.Event
			for _, event := range events {
				if event.AggregateID == position.Instrument {
					positionEvents = append(positionEvents, event)
				}
			}
			weight := position.TotalCost / metrics.TotalCost
			annualized := ProjectAnnualizedReturn(ctx, positionEvents, position)
			totalAnnualized += annualized * weight
		}
		metrics.AnnualizedReturn = totalAnnualized
	}

	return metrics, nil
}
