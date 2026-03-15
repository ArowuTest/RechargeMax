package persistence

import (
	"rechargemax/internal/domain/entities"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PrizeCategoryRepositoryGORM struct {
	db *gorm.DB
}

func NewPrizeCategoryRepositoryGORM(db *gorm.DB) *PrizeCategoryRepositoryGORM {
	return &PrizeCategoryRepositoryGORM{db: db}
}

func (r *PrizeCategoryRepositoryGORM) Create(category *entities.PrizeCategory) error {
	return r.db.Create(category).Error
}

func (r *PrizeCategoryRepositoryGORM) FindByID(id uuid.UUID) (*entities.PrizeCategory, error) {
	var category entities.PrizeCategory
	err := r.db.First(&category, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *PrizeCategoryRepositoryGORM) FindAll() ([]entities.PrizeCategory, error) {
	var categories []entities.PrizeCategory
	err := r.db.Order("display_order ASC").Find(&categories).Error
	return categories, err
}

// FindByTemplateID uses the DB column name "template_id" (not prize_template_id)
func (r *PrizeCategoryRepositoryGORM) FindByTemplateID(templateID uuid.UUID) ([]entities.PrizeCategory, error) {
	var categories []entities.PrizeCategory
	err := r.db.Where("template_id = ?", templateID).
		Order("display_order ASC").
		Find(&categories).Error
	return categories, err
}

func (r *PrizeCategoryRepositoryGORM) Update(category *entities.PrizeCategory) error {
	return r.db.Save(category).Error
}

func (r *PrizeCategoryRepositoryGORM) Delete(id uuid.UUID) error {
	return r.db.Delete(&entities.PrizeCategory{}, "id = ?", id).Error
}

func (r *PrizeCategoryRepositoryGORM) DeleteByTemplateID(templateID uuid.UUID) error {
	return r.db.Where("template_id = ?", templateID).Delete(&entities.PrizeCategory{}).Error
}

func (r *PrizeCategoryRepositoryGORM) BulkCreate(categories []entities.PrizeCategory) error {
	return r.db.Create(&categories).Error
}

func (r *PrizeCategoryRepositoryGORM) UpdateDisplayOrder(categoryID uuid.UUID, order int) error {
	return r.db.Model(&entities.PrizeCategory{}).
		Where("id = ?", categoryID).
		Update("display_order", order).Error
}
