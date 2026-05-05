package main

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

const sessionID = "demo-session"

type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	PriceCents  int    `json:"priceCents"`
	Description string `json:"description"`
}

type CartItem struct {
	Product Product `json:"product"`
	Qty     int     `json:"qty"`
}

func openDB(path string) *sql.DB {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			price_cents INTEGER NOT NULL,
			description TEXT NOT NULL DEFAULT ''
		);
		CREATE TABLE IF NOT EXISTS cart_items (
			session_id TEXT NOT NULL,
			product_id INTEGER NOT NULL,
			qty INTEGER NOT NULL,
			PRIMARY KEY (session_id, product_id)
		);
	`); err != nil {
		log.Fatalf("create tables: %v", err)
	}
	return db
}

func seed(db *sql.DB) {
	products := []Product{
		{1, "Linen Crewneck Tee", 2900, "Soft midweight tee in oat heather."},
		{2, "Selvedge Denim Jeans", 12800, "14oz Japanese selvedge, slim straight."},
		{3, "Merino Wool Beanie", 3400, "Itch-free fine merino, ribbed cuff."},
		{4, "Leather Card Holder", 4500, "Vegetable-tanned full-grain leather."},
		{5, "Cotton Field Jacket", 18900, "Waxed cotton canvas with corduroy collar."},
		{6, "Linen Drawstring Pants", 8400, "Relaxed crop, side pockets, garment-dyed."},
		{7, "Suede Chukka Boots", 19500, "Crepe sole, unlined, hand-stitched welt."},
		{8, "Flannel Overshirt", 9800, "Brushed flannel, two chest pockets."},
		{9, "Wool Socks (3-pack)", 3600, "Cushioned merino, reinforced heel."},
		{42, "Rare Gold Bomber Jacket", 79900, "Limited drop. Oversized fit. Probably broken."},
	}
	for _, p := range products {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO products (id, name, price_cents, description) VALUES (?, ?, ?, ?)`,
			p.ID, p.Name, p.PriceCents, p.Description,
		); err != nil {
			log.Fatalf("seed product %d: %v", p.ID, err)
		}
	}
}
