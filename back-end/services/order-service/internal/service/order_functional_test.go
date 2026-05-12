//go:build functional

package service_test

import (
	"context"
	"os"
	"testing"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/services/order-service/internal/model"
	"github.com/nusaroute/services/order-service/internal/repository"
	"github.com/nusaroute/services/order-service/internal/service"
)

func TestOrderFunctional(t *testing.T) {
	// Setup real database connection
	host := os.Getenv("DB_HOST")
	if host == "" { host = "localhost" }
	port := os.Getenv("DB_PORT")
	if port == "" { port = "5432" }
	user := os.Getenv("DB_USER")
	if user == "" { user = "postgres" }
	pass := os.Getenv("DB_PASSWORD")
	if pass == "" { pass = "postgres" }
	dbname := os.Getenv("DB_NAME")
	if dbname == "" { dbname = "nusaroute_order" }

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: host, Port: port, User: user, Password: pass, DBName: dbname,
	})
	if err != nil {
		t.Skipf("Skipping functional test: failed to connect to DB: %v", err)
		return
	}
	defer db.Close()

	repo := repository.NewOrderRepository(db)
	svc := service.NewOrderService(repo, nil) // nil producer for test

	ctx := context.Background()
	testUserID := "func-user-123"

	// Cleanup
	_, _ = db.Exec("DELETE FROM orders WHERE user_id = $1", testUserID)

	var createdOrder *model.Order
	t.Run("CreateOrder", func(t *testing.T) {
		order, err := svc.CreateOrder(ctx, testUserID, model.CreateOrderRequest{
			SenderName: "Budi", SenderPhone: "08123456789",
			SenderAddress: "Bandung", ReceiverName: "Siti",
			ReceiverAddress: "Jakarta", ServiceType: "REG",
			WeightKg: 1.5, ShippingCost: 15000, TotalCost: 15000,
		})
		if err != nil { t.Fatalf("failed to create order: %v", err) }
		if order.AWB == "" { t.Error("AWB missing") }
		createdOrder = order
	})

	t.Run("GetOrder", func(t *testing.T) {
		order, err := svc.GetOrder(ctx, createdOrder.ID)
		if err != nil { t.Fatalf("failed to get order: %v", err) }
		if order.ID != createdOrder.ID { t.Errorf("ID mismatch: %s != %s", order.ID, createdOrder.ID) }
	})

	t.Run("CancelOrder", func(t *testing.T) {
		err := svc.CancelOrder(ctx, createdOrder.ID, testUserID)
		if err != nil { t.Fatalf("failed to cancel order: %v", err) }
		
		order, _ := svc.GetOrder(ctx, createdOrder.ID)
		if order.Status != "CANCELLED" { t.Errorf("expected CANCELLED, got %s", order.Status) }
	})

	// Final cleanup
	_, _ = db.Exec("DELETE FROM orders WHERE user_id = $1", testUserID)
}
