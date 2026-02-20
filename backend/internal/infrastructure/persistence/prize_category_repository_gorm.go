package persistence

import (
	"rechargemax/internal/domain/entities"
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

func (r *PrizeCategoryRepositoryGORM) FindByID(id uint) (*entities.PrizeCategory, error) {
	var category entities.PrizeCategory
	err := r.db.First(&category, id).Error
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

func (r *PrizeCategoryRepositoryGORM) FindByTemplateID(templateID uint) ([]entities.PrizeCategory, error) {
	var categories []entities.PrizeCategory
	err := r.db.Where("prize_template_id = ?", templateID).
		Order("display_order ASC").
		Find(&categories).Error
	return categories, err
}

func (r *PrizeCategoryRepositoryGORM) Update(category *entities.PrizeCategory) error {
	return r.db.Save(category).Error
}

func (r *PrizeCategoryRepositoryGORM) Delete(id uint) error {
	return r.db.Delete(&entities.PrizeCategory{}, id).Error
}

func (r *PrizeCategoryRepositoryGORM) DeleteByTemplateID(templateID uint) error {
	return r.db.Where("prize_template_id = ?", templateID).Delete(&entities.PrizeCategory{}).Error
}

func (r *PrizeCategoryRepositoryGORM) BulkCreate(categories []entities.PrizeCategory) error {
	return r.db.Create(&categories).Error
}

func (r *PrizeCategoryRepositoryGORM) UpdateDisplayOrder(categoryID uint, order int) error {
	return r.db.Model(&entities.PrizeCategory{}).
		Where("id = ?", categoryID).
		Update("display_order", order).Error
}
