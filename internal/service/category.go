package service

import (
	"context"
	"io"

	"github.com/eduardo-js/go-fc-grpc/internal/database"
	"github.com/eduardo-js/go-fc-grpc/internal/pb"
)

type CategoryService struct {
	pb.UnimplementedCategoryServiceServer
	CategoryDB database.Category
}

func NewCategoryService(categoryDB database.Category) *CategoryService {
	return &CategoryService{
		CategoryDB: categoryDB,
	}
}

func (c *CategoryService) CreateCategory(ctx context.Context, in *pb.CreateCategoryRequest) (*pb.Category, error) {
	category, err := c.CategoryDB.Create(in.Name, in.Description)

	if err != nil {
		return nil, err
	}

	return &pb.Category{
		Id:          category.ID,
		Name:        category.Name,
		Description: category.Description,
	}, nil
}

func (c *CategoryService) ListCategories(ctx context.Context, in *pb.Empty) (*pb.CategoryList, error) {
	categories, err := c.CategoryDB.FindAll()
	if err != nil {
		return nil, err
	}

	var categoriesResponse []*pb.Category
	for _, c := range categories {
		categoriesResponse = append(categoriesResponse, &pb.Category{
			Id:          c.ID,
			Name:        c.Name,
			Description: c.Description,
		})
	}
	return &pb.CategoryList{Categories: categoriesResponse}, nil
}

func (c *CategoryService) GetCategoryById(ctx context.Context, in *pb.Id) (*pb.Category, error) {
	category, err := c.CategoryDB.FindByCourseID(in.Id)
	if err != nil {
		return nil, err
	}

	return &pb.Category{
		Id:          category.ID,
		Name:        category.Name,
		Description: category.Description,
	}, nil
}

func (c *CategoryService) CreateCategoryStream(stream pb.CategoryService_CreateCategoryStreamServer) error {
	categories := &pb.CategoryList{}

	for {
		category, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(categories)
		}
		if err != nil {
			return err
		}
		c, err := c.CategoryDB.Create(category.Name, category.Description)
		if err != nil {
			return err
		}
		categories.Categories = append(categories.Categories, &pb.Category{
			Id:          c.ID,
			Description: c.Description,
			Name:        c.Name,
		})
	}
}

func (c *CategoryService) CreateCategoryStreamBidirectional(stream pb.CategoryService_CreateCategoryStreamBidirectionalServer) error {

	for {
		category, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		c, err := c.CategoryDB.Create(category.Name, category.Description)
		if err != nil {
			return err
		}
		err = stream.Send(&pb.Category{
			Id:          c.ID,
			Description: c.Description,
			Name:        c.Name,
		})
		if err != nil {
			return err
		}
	}
}
