package snapshot

import (
	"testing"
	"time"
)

func TestBuildRequiresSettlementEvidence(t *testing.T) {
	tests := []struct {
		name       string
		input      CommerceData
		wantOrders []string
	}{
		{
			name: "retain reconciled order and sort its updates",
			input: CommerceData{
				Checkouts:    []Checkout{{ID: "co-2"}, {ID: "co-1"}},
				Fulfillments: []Fulfillment{{CheckoutID: "co-1"}, {CheckoutID: "co-2"}},
				Receipts:     []Receipt{{CheckoutID: "co-1"}, {CheckoutID: "co-2"}},
				OrderUpdates: []OrderUpdate{
					{CheckoutID: "co-1", ChangedAt: "2026-09-06T12:00:00Z"},
					{CheckoutID: "co-1", ChangedAt: "2026-09-06T08:00:00Z"},
				},
			},
			wantOrders: []string{"co-1", "co-2"},
		},
		{
			name: "exclude order without receipt evidence",
			input: CommerceData{
				Checkouts:    []Checkout{{ID: "co-pending"}},
				Fulfillments: []Fulfillment{{CheckoutID: "co-pending"}},
			},
			wantOrders: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Build(time.Date(2026, 9, 6, 23, 0, 0, 0, time.UTC), tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Orders) != len(tt.wantOrders) {
				t.Fatalf("got %d orders, want %d", len(got.Orders), len(tt.wantOrders))
			}
			for i, id := range tt.wantOrders {
				if got.Orders[i].Checkout.ID != id {
					t.Fatalf("order %d = %q, want %q", i, got.Orders[i].Checkout.ID, id)
				}
			}
			if len(got.Orders) > 0 && len(got.Orders[0].Updates) == 2 && got.Orders[0].Updates[0].ChangedAt != "2026-09-06T08:00:00Z" {
				t.Fatal("updates are not chronological")
			}
		})
	}
}
