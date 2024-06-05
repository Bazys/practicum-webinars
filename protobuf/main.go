package main

import (
	"context"
	"log"
	"net"
	"net/http"

	proto "webinars/protobuf/service"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type server struct {
	proto.UnimplementedUserServiceServer
}

func (s *server) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	// Логика создания пользователя
	return &proto.CreateUserResponse{Message: "User created successfully"}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_validator.UnaryServerInterceptor(),    // Валидация данных
			grpc_auth.UnaryServerInterceptor(authFunc), // Авторизация
		),
	)
	proto.RegisterUserServiceServer(s, &server{})

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Настройка grpc-gateway
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = proto.RegisterUserServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts)
	if err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	gwServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Printf("Serving gRPC-Gateway on http://localhost:8080")
	log.Fatal(gwServer.ListenAndServe())
}

func authFunc(ctx context.Context) (context.Context, error) {
	// Простая проверка авторизации. В реальном приложении здесь будет логика проверки токена.
	md, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}
	if md != "valid-token" {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}
	return ctx, nil
}
