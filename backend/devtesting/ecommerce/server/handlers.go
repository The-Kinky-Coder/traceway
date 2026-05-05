package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func listProducts(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`SELECT id, name, price_cents, description FROM products ORDER BY id`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var out []Product
		for rows.Next() {
			var p Product
			if err := rows.Scan(&p.ID, &p.Name, &p.PriceCents, &p.Description); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			out = append(out, p)
		}
		c.JSON(http.StatusOK, out)
	}
}

func getProduct(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Bug 1 — nil map assignment panics. gin.Recovery() turns it into a 500.
		if id == "42" {
			var m map[string]int
			m["count"]++
		}

		idInt, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var p Product
		err = db.QueryRow(
			`SELECT id, name, price_cents, description FROM products WHERE id = ?`,
			idInt,
		).Scan(&p.ID, &p.Name, &p.PriceCents, &p.Description)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, p)
	}
}

type addToCartReq struct {
	ProductID int `json:"productId" binding:"required"`
	Qty       int `json:"qty" binding:"required,min=1"`
}

func addToCart(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body addToCartReq
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err := db.Exec(`
			INSERT INTO cart_items (session_id, product_id, qty) VALUES (?, ?, ?)
			ON CONFLICT(session_id, product_id) DO UPDATE SET qty = qty + excluded.qty
		`, sessionID, body.ProductID, body.Qty)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func getCart(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT p.id, p.name, p.price_cents, p.description, ci.qty
			FROM cart_items ci JOIN products p ON p.id = ci.product_id
			WHERE ci.session_id = ?
			ORDER BY p.id
		`, sessionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		items := []CartItem{}
		for rows.Next() {
			var ci CartItem
			if err := rows.Scan(&ci.Product.ID, &ci.Product.Name, &ci.Product.PriceCents, &ci.Product.Description, &ci.Qty); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			items = append(items, ci)
		}
		c.JSON(http.StatusOK, items)
	}
}

type promoReq struct {
	Code string `json:"code"`
}

func applyPromoHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body promoReq
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if body.Code == "" {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "promo code required"})
			return
		}
		// All promo codes are invalid in the demo — keeps the bug surface small.
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "promo code not recognized"})
	}
}

// Bug 2 — slow checkout. Each step is its own helper so spans wrap cleanly
// when Traceway is added during the live demo.
func checkout(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		started := time.Now()
		log.Printf("INFO  checkout started session=%s", sessionID)

		validateCart(ctx)
		log.Printf("INFO  cart validated")

		invStart := time.Now()
		lookupInventory(ctx)
		invMs := time.Since(invStart).Milliseconds()
		log.Printf("WARN  slow inventory lookup detected (%dms > 500ms threshold)", invMs)

		chargePayment(ctx)
		log.Printf("INFO  payment charged")

		sendReceiptEmail(ctx)
		log.Printf("INFO  checkout completed in %dms", time.Since(started).Milliseconds())

		// Clear the cart so the demo is repeatable.
		if _, err := db.Exec(`DELETE FROM cart_items WHERE session_id = ?`, sessionID); err != nil {
			log.Printf("ERROR clear cart: %v", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"orderId":    "ord_demo_001",
			"durationMs": time.Since(started).Milliseconds(),
		})
	}
}

func validateCart(_ context.Context) {
	time.Sleep(10 * time.Millisecond)
}

func lookupInventory(_ context.Context) {
	// Simulates a missing index / N+1 query — the slow span Traceway will reveal.
	time.Sleep(3500 * time.Millisecond)
}

func chargePayment(_ context.Context) {
	time.Sleep(300 * time.Millisecond)
}

func sendReceiptEmail(_ context.Context) {
	time.Sleep(200 * time.Millisecond)
}
