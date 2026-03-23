package models

import "testing"

func TestPortfolioValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   Portfolio
		wantErr string
	}{
		{
			name: "valid portfolio",
			input: Portfolio{
				UserID: "user-1",
				Allocation: map[string]float64{
					"stocks": 60,
					"bonds":  30,
					"gold":   10,
				},
			},
		},
		{
			name: "missing user id",
			input: Portfolio{
				Allocation: map[string]float64{"stocks": 100},
			},
			wantErr: "user_id: is required",
		},
		{
			name: "empty allocation",
			input: Portfolio{
				UserID:     "user-1",
				Allocation: map[string]float64{},
			},
			wantErr: "allocation: must not be empty",
		},
		{
			name: "invalid total",
			input: Portfolio{
				UserID: "user-1",
				Allocation: map[string]float64{
					"stocks": 60,
					"bonds":  20,
				},
			},
			wantErr: "allocation: total allocation must sum to 100",
		},
		{
			name: "invalid percentage",
			input: Portfolio{
				UserID: "user-1",
				Allocation: map[string]float64{
					"stocks": 120,
				},
			},
			wantErr: "allocation: percentages must be between 0 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr == "" && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
			}
		})
	}
}

func TestUpdatedPortfolioValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   UpdatedPortfolio
		wantErr string
	}{
		{
			name: "valid updated portfolio",
			input: UpdatedPortfolio{
				UserID: "user-1",
				NewAllocation: map[string]float64{
					"stocks": 70,
					"bonds":  20,
					"gold":   10,
				},
			},
		},
		{
			name: "missing user id",
			input: UpdatedPortfolio{
				NewAllocation: map[string]float64{"stocks": 100},
			},
			wantErr: "user_id: is required",
		},
		{
			name: "missing allocation",
			input: UpdatedPortfolio{
				UserID:        "user-1",
				NewAllocation: map[string]float64{},
			},
			wantErr: "new_allocation: must not be empty",
		},
		{
			name: "blank asset name",
			input: UpdatedPortfolio{
				UserID: "user-1",
				NewAllocation: map[string]float64{
					"": 100,
				},
			},
			wantErr: "new_allocation: asset name must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr == "" && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
			}
		})
	}
}
