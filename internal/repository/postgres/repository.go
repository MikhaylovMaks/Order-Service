package postgres

import (
	"context"
	"fmt"

	"github.com/MikhaylovMaks/wb_techl0/internal/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveOrder(ctx context.Context, order *models.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Delivery
	var deliveryID int
	err = tx.QueryRow(ctx, `INSERT INTO delivery (name, phone, zip, city, address, region, email)
	VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		order.Delivery.Name, order.Delivery.Phone, order.Delivery, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region,
		order.Delivery.Email).Scan(&deliveryID)
	if err != nil {
		return fmt.Errorf("Insert delivery failed: %w", err)
	}

	// 2. Payment
	var paymentID int
	err = tx.QueryRow(ctx, `INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee).Scan(&paymentID)
	if err != nil {
		return fmt.Errorf("Insert payment failed: %w", err)
	}

	// 3. Order
	_, err = tx.Exec(ctx, `INSERT INTO orders (order_uid, track_number, entry, locale,
	internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created,
	oof_shard, delivery_id, payment_id)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.ShardKey, order.SmID, order.DateCreated, order.OofShard,
		deliveryID, paymentID)
	if err != nil {
		return fmt.Errorf("Insert order failed %w, err")
	}

	// 4. Items
	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("Insert item failed: %w", err)
		}
	}

	// Commit
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit failed %w", err)
	}
	return nil
}
