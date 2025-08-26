CREATE TABLE delivery (
   id SERIAL PRIMARY KEY,
   name VARCHAR(255) NOT NULL,
   phone VARCHAR(50) NOT NULL,
   zip VARCHAR(20) NOT NULL,
   city VARCHAR(255) NOT NULL,
   address VARCHAR(255) NOT NULL,
   region VARCHAR(255) NOT NULL,
   email VARCHAR(255) NOT NULL
);

CREATE TABLE payment (
   id SERIAL PRIMARY KEY,
   transaction VARCHAR(255) NOT NULL,
   request_id VARCHAR(255),
   currency VARCHAR(10) NOT NULL,
   provider VARCHAR(255) NOT NULL,
   amount INT NOT NULL,
   payment_dt INT NOT NULL,
   bank VARCHAR(255) NOT NULL,
   delivery_cost INT NOT NULL,
   goods_total INT NOT NULL,
   custom_fee INT NOT NULL
);

CREATE TABLE items (
   id SERIAL PRIMARY KEY,
   chrt_id INT NOT NULL,
   track_number VARCHAR(255) NOT NULL,
   price INT NOT NULL,
   rid VARCHAR(255) NOT NULL,
   name VARCHAR(255) NOT NULL,
   sale INT NOT NULL,
   size VARCHAR(50) NOT NULL,
   total_price INT NOT NULL,
   nm_id INT NOT NULL,
   brand VARCHAR(255) NOT NULL,
   status INT NOT NULL
);

CREATE TABLE orders (
   order_uid VARCHAR(255) PRIMARY KEY,
   track_number VARCHAR(255) NOT NULL,
   entry VARCHAR(255) NOT NULL,
   locale VARCHAR(255) NOT NULL,
   internal_signature VARCHAR(255),
   customer_id VARCHAR(255) NOT NULL,
   delivery_service VARCHAR(255) NOT NULL,
   shardkey VARCHAR(255) NOT NULL,
   sm_id INT NOT NULL,
   date_created TIMESTAMP NOT NULL,
   oof_shard VARCHAR(255) NOT NULL,
   delivery_id INT NOT NULL,
   payment_id INT NOT NULL,
   items_id INT NOT NULL,
   CONSTRAINT orders_delivery_id_fkey FOREIGN KEY (delivery_id) REFERENCES delivery(id) ON DELETE CASCADE,
   CONSTRAINT orders_payment_id_fkey FOREIGN KEY (payment_id) REFERENCES payment(id) ON DELETE CASCADE,
   CONSTRAINT orders_items_id_fkey FOREIGN KEY (items_id) REFERENCES items(id) ON DELETE CASCADE
);
