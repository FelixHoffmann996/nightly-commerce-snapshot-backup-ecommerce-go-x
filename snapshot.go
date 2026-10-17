package snapshot

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

type Checkout struct {
	ID          string `json:"id"`
	CustomerID  string `json:"customer_id"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
}

type Fulfillment struct {
	CheckoutID string `json:"checkout_id"`
	Status     string `json:"status"`
	CarrierRef string `json:"carrier_ref"`
}

type Receipt struct {
	CheckoutID string `json:"checkout_id"`
	ReceiptID  string `json:"receipt_id"`
	IssuedAt   string `json:"issued_at"`
}

type OrderUpdate struct {
	CheckoutID string `json:"checkout_id"`
	Status     string `json:"status"`
	ChangedAt  string `json:"changed_at"`
}

type CommerceData struct {
	Checkouts    []Checkout    `json:"checkouts"`
	Fulfillments []Fulfillment `json:"fulfillments"`
	Receipts     []Receipt     `json:"receipts"`
	OrderUpdates []OrderUpdate `json:"order_updates"`
}

type Snapshot struct {
	BusinessDate string          `json:"business_date"`
	Orders       []OrderSnapshot `json:"orders"`
}

type OrderSnapshot struct {
	Checkout    Checkout      `json:"checkout"`
	Fulfillment Fulfillment   `json:"fulfillment"`
	Receipt     Receipt       `json:"receipt"`
	Updates     []OrderUpdate `json:"updates"`
}

// Build includes only orders with both fulfillment and receipt evidence.
func Build(day time.Time, input CommerceData) (Snapshot, error) {
	fulfillments := make(map[string]Fulfillment, len(input.Fulfillments))
	for _, item := range input.Fulfillments {
		fulfillments[item.CheckoutID] = item
	}
	receipts := make(map[string]Receipt, len(input.Receipts))
	for _, item := range input.Receipts {
		receipts[item.CheckoutID] = item
	}
	updates := make(map[string][]OrderUpdate)
	for _, item := range input.OrderUpdates {
		updates[item.CheckoutID] = append(updates[item.CheckoutID], item)
	}

	orders := make([]OrderSnapshot, 0, len(input.Checkouts))
	for _, checkout := range input.Checkouts {
		fulfillment, hasFulfillment := fulfillments[checkout.ID]
		receipt, hasReceipt := receipts[checkout.ID]
		if !hasFulfillment || !hasReceipt {
			continue
		}
		orderUpdates := updates[checkout.ID]
		sort.Slice(orderUpdates, func(i, j int) bool {
			return orderUpdates[i].ChangedAt < orderUpdates[j].ChangedAt
		})
		orders = append(orders, OrderSnapshot{checkout, fulfillment, receipt, orderUpdates})
	}
	sort.Slice(orders, func(i, j int) bool { return orders[i].Checkout.ID < orders[j].Checkout.ID })
	return Snapshot{BusinessDate: day.UTC().Format("2006-01-02"), Orders: orders}, nil
}

func Encode(value Snapshot) ([]byte, error) {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode snapshot: %w", err)
	}
	return append(body, '\n'), nil
}
