package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	dbUser := "nusaroute"
	dbPass := "nusaroute_secret"
	dbHost := "localhost"
	dbPort := "5433"

	// Connect to Order DB
	orderDB, err := sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=nusaroute_order sslmode=disable", dbHost, dbPort, dbUser, dbPass))
	if err != nil {
		log.Fatalf("Failed to connect to Order DB: %v", err)
	}
	defer orderDB.Close()

	// Connect to Courier DB
	courierDB, err := sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=nusaroute_courier sslmode=disable", dbHost, dbPort, dbUser, dbPass))
	if err != nil {
		log.Fatalf("Failed to connect to Courier DB: %v", err)
	}
	defer courierDB.Close()

	// Connect to Hub DB
	hubDB, err := sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=nusaroute_hub sslmode=disable", dbHost, dbPort, dbUser, dbPass))
	if err != nil {
		log.Fatalf("Failed to connect to Hub DB: %v", err)
	}
	defer hubDB.Close()

	log.Println("Seeding Hub Data...")
	seedHubs(hubDB)

	log.Println("Seeding Courier Data...")
	seedCouriers(courierDB)

	log.Println("Seeding Order Data...")
	seedOrders(orderDB)

	log.Println("✅ Real data seeded successfully!")
}

func seedHubs(db *sqlx.DB) {
	cities := []string{"Jakarta", "Bandung", "Surabaya", "Medan", "Makassar", "Bali", "Semarang", "Yogyakarta"}
	provinces := []string{"DKI Jakarta", "Jawa Barat", "Jawa Timur", "Sumatera Utara", "Sulawesi Selatan", "Bali", "Jawa Tengah", "DI Yogyakarta"}
	codes := []string{"JKT-02", "BDG-02", "SBY-02", "MDN-02", "MKS-02", "DPS-02", "SMG-02", "YK-02"}
	for i, city := range cities {
		_, err := db.Exec(`
			INSERT INTO hubs (id, name, code, city, province, lat, lng, type, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO NOTHING
		`, fmt.Sprintf("hub-%d", i), "Hub "+city, codes[i], city, provinces[i], -6.2+float64(i)*0.1, 106.8+float64(i)*0.1, "SORTATION", true)
		if err != nil {
			log.Printf("Error seeding hub: %v", err)
		}
	}
}

func seedCouriers(db *sqlx.DB) {
	for i := 0; i < 50; i++ {
		_, err := db.Exec(`
			INSERT INTO couriers (id, user_id, full_name, phone, vehicle_type, vehicle_plate, max_capacity_kg, current_lat, current_lng, is_online, is_available, rating, total_deliveries, is_active, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			ON CONFLICT (id) DO NOTHING
		`, uuid.New().String(), uuid.New().String(), fmt.Sprintf("Kurir %d", i), fmt.Sprintf("0812%08d", i), "MOTORCYCLE", fmt.Sprintf("B %d ABC", 1000+i), 50.0, -6.2, 106.8, true, true, 4.8, 120, true, time.Now())
		if err != nil {
			log.Printf("Error seeding courier: %v", err)
		}
	}
}

func seedOrders(db *sqlx.DB) {
	now := time.Now()
	statuses := []string{"DELIVERED", "IN_TRANSIT", "PENDING_PAYMENT", "OUT_FOR_DELIVERY"}
	
	for i := 0; i < 300; i++ {
		// Random past date within 7 days
		daysAgo := rand.Intn(7)
		createdAt := now.AddDate(0, 0, -daysAgo).Add(time.Duration(-rand.Intn(24)) * time.Hour)
		
		status := statuses[rand.Intn(len(statuses))]
		if daysAgo > 2 {
			status = "DELIVERED" // Older orders are likely delivered
		}

		cost := 15000 + rand.Intn(50000)
		weight := 1.0 + rand.Float64()*5.0
		_, err := db.Exec(`
			INSERT INTO orders (
				id, awb, user_id, status, service_type, 
				sender_name, sender_phone, sender_address, 
				receiver_name, receiver_phone, receiver_address, 
				shipping_cost, total_cost, weight_kg,
				created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			ON CONFLICT (id) DO NOTHING
		`, 
			uuid.New().String(), 
			fmt.Sprintf("NR%08d", rand.Intn(100000000)), 
			uuid.New().String(), 
			status, 
			"REGULAR", 
			"Sender "+fmt.Sprint(i), 
			"08123456789",
			"Jl. Sender Address "+fmt.Sprint(i),
			"Receiver "+fmt.Sprint(i), 
			"08987654321",
			"Jl. Receiver Address "+fmt.Sprint(i),
			cost, 
			cost,
			weight,
			createdAt, 
			createdAt,
		)
		if err != nil {
			log.Printf("Error seeding order: %v", err)
		}
	}
}
