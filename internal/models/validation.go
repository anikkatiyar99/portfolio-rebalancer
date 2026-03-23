package models

import (
	"fmt"
	"math"
	"strings"
)

const allocationTolerance = 1e-9

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (p Portfolio) Validate() error {
	if strings.TrimSpace(p.UserID) == "" {
		return &ValidationError{Field: "user_id", Message: "is required"}
	}
	if len(p.Allocation) == 0 {
		return &ValidationError{Field: "allocation", Message: "must not be empty"}
	}

	var total float64
	for asset, percent := range p.Allocation {
		if strings.TrimSpace(asset) == "" {
			return &ValidationError{Field: "allocation", Message: "asset name must not be empty"}
		}
		if percent < 0 || percent > 100 {
			return &ValidationError{Field: "allocation", Message: "percentages must be between 0 and 100"}
		}
		total += percent
	}

	if math.Abs(total-100.0) > allocationTolerance {
		return &ValidationError{Field: "allocation", Message: "total allocation must sum to 100"}
	}

	return nil
}

func (p UpdatedPortfolio) Validate() error {
	if strings.TrimSpace(p.UserID) == "" {
		return &ValidationError{Field: "user_id", Message: "is required"}
	}
	if len(p.NewAllocation) == 0 {
		return &ValidationError{Field: "new_allocation", Message: "must not be empty"}
	}

	var total float64
	for asset, percent := range p.NewAllocation {
		if strings.TrimSpace(asset) == "" {
			return &ValidationError{Field: "new_allocation", Message: "asset name must not be empty"}
		}
		if percent < 0 || percent > 100 {
			return &ValidationError{Field: "new_allocation", Message: "percentages must be between 0 and 100"}
		}
		total += percent
	}

	if math.Abs(total-100.0) > allocationTolerance {
		return &ValidationError{Field: "new_allocation", Message: "total allocation must sum to 100"}
	}

	return nil
}
