package postgres

import (
	"context"
	"fmt"

	"github.com/MikhaylovMaks/wb_techl0/internal/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

type OrderRepository interface {
	SaveOrder(ctx context.Context, order *models.Order) error
	GetOrderByUID(ctx context.Context, uid string) (*models.Order, error)
}

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
		order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region,
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
		return fmt.Errorf("Insert order failed %w", err)
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

func (r *Repository) GetOrderByUID(ctx context.Context, uid string) (*models.Order, error) {
	// 1. Order + delivery_id + payment_id
	var order models.Order
	var deliveryID, paymentID int

	err := r.db.QueryRow(ctx, `SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
			   o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created,
			   o.oof_shard, o.delivery_id, o.payment_id
				FROM orders o
				WHERE o.order_uid = $1
				`, uid).Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SmID,
		&order.DateCreated, &order.OofShard, &deliveryID, &paymentID)
	if err != nil {
		return nil, fmt.Errorf("get order failed: %w", err)
	}

	// 2. Delivery
	err = r.db.QueryRow(ctx, `SELECT name, phone, zip, city, address, region, email
	FROM delivery WHERE id = $1`,
		deliveryID).Scan(&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City, &order.Delivery.Address,
		&order.Delivery.Region, &order.Delivery.Email)
	if err != nil {
		return nil, fmt.Errorf("get delivery failed: %w", err)
	}

	// 3. Payment
	err = r.db.QueryRow(ctx, `SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
	FROM payment WHERE id = $1`,
		paymentID).Scan(&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDT, &order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee)
	if err != nil {
		return nil, fmt.Errorf("get payment failed: %w", err)
	}

	// 4. Items  - order_items
	rows, err := r.db.Query(ctx, `
	SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
	FROM items
	WHERE order_uid = $1`,
		order.OrderUID)
	if err != nil {
		return nil, fmt.Errorf("get items failed: %w", err)
	}
	defer rows.Close()
	order.Items = []models.Items{}
	for rows.Next() {
		var item models.Items
		if err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		); err != nil {
			return nil, fmt.Errorf("scan item failed: %w", err)
		}
		order.Items = append(order.Items, item)
	}
	return &order, nil
}
