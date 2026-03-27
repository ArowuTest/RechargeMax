package persistence

import (
	"strings"

	"rechargemax/internal/domain/entities"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PrizeTemplateRepositoryGORM struct {
	db *gorm.DB
}

func NewPrizeTemplateRepositoryGORM(db *gorm.DB) *PrizeTemplateRepositoryGORM {
	return &PrizeTemplateRepositoryGORM{db: db}
}

func (r *PrizeTemplateRepositoryGORM) Create(template *entities.PrizeTemplate) error {
	err := r.db.Create(template).Error
	if err != nil && strings.Contains(err.Error(), "is_default") {
		// Column may not exist yet (pending migration); retry without it
		return r.db.Omit("is_default").Create(template).Error
	}
	return err
}

func (r *PrizeTemplateRepositoryGORM) Update(template *entities.PrizeTemplate) error {
	err := r.db.Save(template).Error
	if err != nil && strings.Contains(err.Error(), "is_default") {
		return r.db.Omit("is_default").Save(template).Error
	}
	return err
}

func (r *PrizeTemplateRepositoryGORM) FindByID(id uuid.UUID) (*entities.PrizeTemplate, error) {
	var template entities.PrizeTemplate
	err := r.db.Preload("PrizeCategories").First(&template, "id = ?", id).Error
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

func (r *PrizeTemplateRepositoryGORM) FindByDrawTypeID(drawTypeID uuid.UUID) ([]entities.PrizeTemplate, error) {
	var templates []entities.PrizeTemplate
	err := r.db.Where("draw_type_id = ?", drawTypeID).Preload("PrizeCategories").Find(&templates).Error
	return templates, err
}

func (r *PrizeTemplateRepositoryGORM) Delete(id uuid.UUID) error {
	if err := r.db.Where("template_id = ?", id).Delete(&entities.PrizeCategory{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&entities.PrizeTemplate{}, "id = ?", id).Error
}

func (r *PrizeTemplateRepositoryGORM) FindByName(name string) (*entities.PrizeTemplate, error) {
	var template entities.PrizeTemplate
	err := r.db.Where("name = ?", name).Preload("PrizeCategories").First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *PrizeTemplateRepositoryGORM) SetAsDefault(id uuid.UUID, drawTypeID uuid.UUID) error {
	if err := r.db.Model(&entities.PrizeTemplate{}).
		Where("draw_type_id = ?", drawTypeID).
		Update("is_default", false).Error; err != nil {
		return err
	}
	return r.db.Model(&entities.PrizeTemplate{}).
		Where("id = ?", id).
		Update("is_default", true).Error
}
