package faker

import (
	"fmt"
	"time"

	"github.com/MikhaylovMaks/wb_techl0/internal/models"
	"github.com/brianvoe/gofakeit/v7"
)

// GenerateFakeOrder — функция для генерации случайного заказа.
func GenerateFakeOrder() *models.Order {
	gofakeit.Seed(time.Now().UnixNano())
	numItems := gofakeit.Number(1, 5)
	items := make([]models.Items, numItems)
	for i := 0; i < numItems; i++ {
		items[i] = models.Items{
			ChrtID:      gofakeit.Number(1000, 9999),
			TrackNumber: gofakeit.UUID(),
			Price:       gofakeit.Number(100, 10000),
			RID:         gofakeit.UUID(),
			Name:        gofakeit.ProductName(),
			Sale:        gofakeit.Number(0, 50),
			Size:        gofakeit.RandomString([]string{"S", "M", "L", "XL"}),
			TotalPrice:  gofakeit.Number(100, 10000),
			NmID:        gofakeit.Number(100000, 999999),
			Brand:       gofakeit.Company(),
			Status:      gofakeit.Number(0, 1),
		}
	}

	return &models.Order{
		OrderUID:    gofakeit.UUID(),
		TrackNumber: gofakeit.UUID(),
		Entry:       gofakeit.Word(),
		Delivery: models.Delivery{
			Name:    gofakeit.Name(),
			Phone:   fmt.Sprintf("+1%010d", gofakeit.Number(0, 9999999999)),
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: gofakeit.Address().Address,
			Region:  gofakeit.State(),
			Email:   gofakeit.Email(),
		},
		Payment: models.Payment{
			Transaction:  gofakeit.UUID(),
			RequestID:    gofakeit.UUID(),
			Currency:     "USD",
			Provider:     gofakeit.Company(),
			Amount:       gofakeit.Number(100, 10000),
			PaymentDT:    time.Now().Unix(),
			Bank:         gofakeit.Company(),
			DeliveryCost: gofakeit.Number(100, 500),
			GoodsTotal:   gofakeit.Number(500, 15000),
			CustomFee:    gofakeit.Number(0, 100),
		},
		Items:             items,
		Locale:            gofakeit.Language(),
		InternalSignature: gofakeit.UUID(),
		CustomerID:        gofakeit.UUID(),
		DeliveryService:   gofakeit.Company(),
		ShardKey:          gofakeit.LetterN(10),
		SmID:              gofakeit.Number(1, 100),
		DateCreated:       time.Now(),
		OofShard:          gofakeit.LetterN(10),
	}
}
