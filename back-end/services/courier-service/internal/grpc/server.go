package grpc_server

import (
	"context"
	"fmt"
	"math/rand"
	"net"

	"github.com/nusaroute/services/courier-service/pkg/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type CourierGRPCServer struct {
	pb.UnimplementedCourierServiceServer
}

func NewCourierGRPCServer() *CourierGRPCServer {
	return &CourierGRPCServer{}
}

func (s *CourierGRPCServer) GetCourierLocation(ctx context.Context, req *pb.CourierLocationRequest) (*pb.CourierLocationResponse, error) {
	// Mock location data based on courier ID
	lat := -6.200000 + (rand.Float64() * 0.05)
	lng := 106.816666 + (rand.Float64() * 0.05)
	
	return &pb.CourierLocationResponse{
		Latitude:  lat,
		Longitude: lng,
		Status:    "ON_DELIVERY",
	}, nil
}

func (s *CourierGRPCServer) GetAvailableCouriers(ctx context.Context, req *pb.AvailableCouriersRequest) (*pb.AvailableCouriersResponse, error) {
	// Mock available couriers based on requested radius
	couriers := []*pb.CourierInfo{
		{
			Id:         "C001",
			FullName:   "Budi Santoso",
			Phone:      "08123456789",
			CurrentLat: req.Latitude + 0.01,
			CurrentLng: req.Longitude + 0.01,
		},
		{
			Id:         "C002",
			FullName:   "Siti Aminah",
			Phone:      "08198765432",
			CurrentLat: req.Latitude - 0.01,
			CurrentLng: req.Longitude - 0.01,
		},
	}
	
	return &pb.AvailableCouriersResponse{
		Data: couriers,
	}, nil
}

func StartGRPCServer(port string) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCourierServiceServer(grpcServer, NewCourierGRPCServer())
	
	// Enable reflection for tools like grpcurl
	reflection.Register(grpcServer)

	return grpcServer.Serve(lis)
}
