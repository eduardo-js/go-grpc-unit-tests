package main

import (
	"net"

	"github.com/eduardo-js/go-fc-grpc/internal/database"
	"github.com/eduardo-js/go-fc-grpc/internal/pb"
	"github.com/eduardo-js/go-fc-grpc/internal/service"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("./db.sqlite"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&database.Category{})
	categoryDB := database.NewCategory(db)

	categoryService := service.NewCategoryService(*categoryDB)

	grpcServer := grpc.NewServer()

	pb.RegisterCategoryServiceServer(grpcServer, categoryService)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		panic(err)
	}
	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}
