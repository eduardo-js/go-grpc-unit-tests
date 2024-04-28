package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"testing"

	"github.com/eduardo-js/go-fc-grpc/internal/database"
	"github.com/eduardo-js/go-fc-grpc/internal/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "github.com/mattn/go-sqlite3"
)

func setupDB() (*database.Category, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&database.Category{})
	categoryDB := *database.NewCategory(db)
	return &categoryDB, nil
}

func makeTestServer(ctx context.Context, db database.Category) (pb.CategoryServiceClient, func() error, error) {
	buffer := 1024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	pb.RegisterCategoryServiceServer(baseServer, NewCategoryService(db))
	go func() {
		err := baseServer.Serve(lis)
		if err != nil {
			log.Fatalf("Error baseServer.Serve %v", err)
		}
	}()

	conn, err := grpc.DialContext(
		ctx,
		"",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, err
	}

	closer := func() error {
		err := lis.Close()
		if err != nil {
			log.Fatalf("Error lis.Close %v", err)
		}
		baseServer.GracefulStop()
		return nil
	}

	client := pb.NewCategoryServiceClient(conn)
	return client, closer, nil
}

func TestCategoryService_CreateCategory(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)
	ctx := context.Background()
	client, closer, err := makeTestServer(ctx, *db)
	assert.Nil(t, err)
	defer closer()

	c := &pb.CreateCategoryRequest{Name: "test name", Description: "test description"}
	out, err := client.CreateCategory(ctx, c)

	assert.Nil(t, err)
	assert.NotNil(t, out.Id)
	assert.Equal(t, out.Description, c.Description)
	assert.Equal(t, out.Name, c.Name)
}

func TestCategoryService_CreateCategoryStream(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)
	ctx := context.Background()
	client, closer, err := makeTestServer(ctx, *db)

	defer closer()
	assert.Nil(t, err)
	stream, err := client.CreateCategoryStream(ctx)
	assert.Nil(t, err)

	for i := 0; i < 3; i++ {
		c := &pb.CreateCategoryRequest{Name: fmt.Sprintf("test name %d", i), Description: fmt.Sprintf("test description %d", i)}
		err := stream.SendMsg(c)
		assert.Nil(t, err)
	}
	out, err := stream.CloseAndRecv()

	assert.Nil(t, err)
	assert.Len(t, out.Categories, 3)
	assert.NotNil(t, out.Categories[0].Id)
	assert.NotNil(t, out.Categories[1].Id)
	assert.NotNil(t, out.Categories[2].Id)
}

func TestCategoryService_CreateCategoryBidirectionalStream(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)
	ctx := context.Background()
	client, closer, err := makeTestServer(ctx, *db)
	defer closer()
	assert.Nil(t, err)
	conn, err := client.CreateCategoryStreamBidirectional(ctx)
	assert.Nil(t, err)

	for i := 0; i < 3; i++ {
		c := &pb.CreateCategoryRequest{Name: fmt.Sprintf("test name %d", i), Description: fmt.Sprintf("test description %d", i)}
		err := conn.Send(c)
		assert.Nil(t, err)
	}

	err = conn.CloseSend()
	assert.Nil(t, err)

	var out []*pb.Category
	for {
		o, err := conn.Recv()
		if err == io.EOF {
			break
		}
		out = append(out, o)
	}

	assert.Nil(t, err)
	assert.Len(t, out, 3)
	assert.NotNil(t, out[0].Id)
	assert.NotNil(t, out[1].Id)
	assert.NotNil(t, out[2].Id)
}

func TestCategoryService_GetCategoryById(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)
	ctx := context.Background()
	client, closer, err := makeTestServer(ctx, *db)
	assert.Nil(t, err)

	defer closer()

	c, err := db.Create("test name", "desc")
	assert.Nil(t, err)

	out, err := client.GetCategoryById(ctx, &pb.Id{Id: c.ID})
	assert.Nil(t, err)
	assert.NotNil(t, out.Id)
	assert.Equal(t, out.Description, c.Description)
	assert.Equal(t, out.Name, c.Name)
}

func TestCategoryService_ListCategories(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)
	ctx := context.Background()
	client, closer, err := makeTestServer(ctx, *db)
	defer closer()
	assert.Nil(t, err)

	_, err = db.Create("nam", "desc")
	assert.Nil(t, err)
	_, err = db.Create("foo", "bar")
	assert.Nil(t, err)

	out, err := client.ListCategories(ctx, &pb.Empty{})
	assert.Nil(t, err)
	assert.Len(t, out.Categories, 2)
}
