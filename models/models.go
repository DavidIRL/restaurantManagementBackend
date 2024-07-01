package models

import (
	"go.mongodb.org/mongo-drive/bson/primitive"
	"time"
)

type Food struct {
	ID         primitive.ObjectID `bson:"_id"`
	Name       *string            `json:"name" validate:"required,min=2,max=100"`
	Price      *float64           `json:"price" validate:"required"`
	Food_image *string            `json:"food_image" validate:"required"`
	Created_at time.time          `json:"created_at"`
	Updated_at time.time          `json:"updated_at"`
	Food_id    string             `json:"food_id"`
	Menu_id    *string            `json:"menu_id" validate:"required"`
}
