package persistence

import (
	"rechargemax/internal/domain/entities"
	"gorm.io/gorm"
)

type PrizeTemplateRepositoryGORM struct {
	db *gorm.DB
}

func NewPrizeTemplateRepositoryGORM(db *gorm.DB) *PrizeTemplateRepositoryGORM {
	return &PrizeTemplateRepositoryGORM{db: db}
}

func (r *PrizeTemplateRepositoryGORM) Create(template *entities.PrizeTemplate) error {
	return r.db.Create(template).Error
}

func (r *PrizeTemplateRepositoryGORM) FindByID(id uint) (*entities.PrizeTemplate, error) {
	var template entities.PrizeTemplate
	err := r.db.Preload("PrizeCategories").First(&template, id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *PrizeTemplateRepositoryGORM) FindAll() ([]entities.PrizeTemplate, error) {
	var templates []entities.PrizeTemplate
	err := r.db.Preload("PrizeCategories").Find(&templates).Error
	return templates, err
}

func (r *PrizeTemplateRepositoryGORM) FindByDrawTypeID(drawTypeID uint) ([]entities.PrizeTemplate, error) {
	var templates []entities.PrizeTemplate
	err := r.db.Where("draw_type_id = ?", drawTypeID).Preload("PrizeCategories").Find(&templates).Error
	return templates, err
}

func (r *PrizeTemplateRepositoryGORM) Update(template *entities.PrizeTemplate) error {
	return r.db.Save(template).Error
}

func (r *PrizeTemplateRepositoryGORM) Delete(id uint) error {
	// Delete associated prize categories first
	if err := r.db.Where("prize_template_id = ?", id).Delete(&entities.PrizeCategory{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&entities.PrizeTemplate{}, id).Error
}

func (r *PrizeTemplateRepositoryGORM) FindByName(name string) (*entities.PrizeTemplate, error) {
	var template entities.PrizeTemplate
	err := r.db.Where("name = ?", name).Preload("PrizeCategories").First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *PrizeTemplateRepositoryGORM) SetAsDefault(id uint, drawTypeID uint) error {
	// First, unset all defaults for this draw type
	if err := r.db.Model(&entities.PrizeTemplate{}).
		Where("draw_type_id = ?", drawTypeID).
		Update("is_default", false).Error; err != nil {
		return err
	}
	
	// Then set the specified template as default
	return r.db.Model(&entities.PrizeTemplate{}).
		Where("id = ?", id).
		Update("is_default", true).Error
}
