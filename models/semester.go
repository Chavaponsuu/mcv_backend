
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)
// Semester represents an academic semester

type Semester struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Year      int                `bson:"year" json:"year"`
	Term      int                `bson:"term" json:"term"` // 1, 2, 3 (summer)
	
	StartDate time.Time `bson:"start_date" json:"start_date"`
	EndDate   time.Time `bson:"end_date" json:"end_date"`
	
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
