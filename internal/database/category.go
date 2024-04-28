package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	db          *gorm.DB
	ID          string
	Name        string
	Description string
	gorm.Model
}

func NewCategory(db *gorm.DB) *Category {
	return &Category{db: db}
}

func (c *Category) Create(name string, description string) (*Category, error) {
	category := &Category{Name: name, Description: description, ID: uuid.New().String()}

	err := c.db.Create(category).Error
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (c *Category) FindAll() ([]Category, error) {
	categories := []Category{}
	err := c.db.Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (c *Category) FindByCourseID(courseID string) (*Category, error) {
	var category Category

	err := c.db.Find(&category, "id = ?", courseID).Error

	if err != nil {
		return nil, err
	}
	return &category, nil
}
