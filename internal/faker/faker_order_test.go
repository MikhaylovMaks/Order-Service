package faker_test

import (
	"testing"

	"github.com/MikhaylovMaks/wb_techl0/internal/faker"
)

func TestGenerateFakeOrder(t *testing.T) {
	order := faker.GenerateFakeOrder()

	if order == nil {
		t.Fatal("expected non-nil order")
	}
	if order.OrderUID == "" {
		t.Error("expected non-empty OrderUID")
	}
	if len(order.Items) == 0 {
		t.Error("expected at least one item")
	}
}
