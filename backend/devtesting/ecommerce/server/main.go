package main

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const appPort = 8080

func main() {
	db := openDB("ecommerce.db")
	defer db.Close()
	seed(db)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Content-Type"},
	}))

	r.GET("/healthz", healthz)
	r.GET("/api/products", listProducts(db))
	r.GET("/api/products/:id", getProduct(db))
	r.POST("/api/cart", addToCart(db))
	r.GET("/api/cart", getCart(db))
	r.POST("/api/promo", applyPromoHandler(db))
	r.POST("/api/checkout", checkout(db))

	printBanner()
	log.Fatal(r.Run(fmt.Sprintf(":%d", appPort)))
}

func printBanner() {
	fmt.Println()
	fmt.Println("===========================================")
	fmt.Printf("  Ecommerce Demo — listening on :%d\n", appPort)
	fmt.Println("-------------------------------------------")
	fmt.Println("  GET    /healthz")
	fmt.Println("  GET    /api/products")
	fmt.Println("  GET    /api/products/:id   (id=42 panics)")
	fmt.Println("  POST   /api/cart")
	fmt.Println("  GET    /api/cart")
	fmt.Println("  POST   /api/promo")
	fmt.Println("  POST   /api/checkout       (~4s)")
	fmt.Println("-------------------------------------------")
	fmt.Println("  Frontend dev: http://localhost:5173")
	fmt.Println("===========================================")
	fmt.Println()
}
