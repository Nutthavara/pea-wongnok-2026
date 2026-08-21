package recipe

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *repository {
	return &repository{
		db: db,
	}
}

func (repo *repository) Create(ctx context.Context, recipe Recipe) (*Recipe, error) {
	if err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).Create(&recipe).Error; err != nil {
			return err
		}

		for index := range recipe.Ingredients {
			recipe.Ingredients[index].RecipeID = recipe.ID
		}
		if len(recipe.Ingredients) > 0 {
			if err := tx.Omit(clause.Associations).Create(&recipe.Ingredients).Error; err != nil {
				return err
			}
		}

		for index := range recipe.Instructions {
			recipe.Instructions[index].RecipeID = recipe.ID
		}
		if len(recipe.Instructions) > 0 {
			if err := tx.Omit(clause.Associations).Create(&recipe.Instructions).Error; err != nil {
				return err
			}
		}

		return nil

	}); err != nil {
		return nil, err

	}

	return &recipe, nil
}
