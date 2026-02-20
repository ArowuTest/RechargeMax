package persistence

import (
	"rechargemax/internal/domain/entities"
	"gorm.io/gorm"
)

type DrawTypeRepositoryGORM struct {
	db *gorm.DB
}

func NewDrawTypeRepositoryGORM(db *gorm.DB) *DrawTypeRepositoryGORM {
	return &DrawTypeRepositoryGORM{db: db}
}

func (r *DrawTypeRepositoryGORM) Create(drawType *entities.DrawType) error {
	return r.db.Create(drawType).Error
}

func (r *DrawTypeRepositoryGORM) FindByID(id uint) (*entities.DrawType, error) {
	var drawType entities.DrawType
	err := r.db.Preload("PrizeTemplates").First(&drawType, id).Error
	if err != nil {
		return nil, err
	}
	return &drawType, nil
}

func (r *DrawTypeRepositoryGORM) FindAll() ([]entities.DrawType, error) {
	var drawTypes []entities.DrawType
	err := r.db.Preload("PrizeTemplates").Find(&drawTypes).Error
	return drawTypes, err
}

func (r *DrawTypeRepositoryGORM) Update(drawType *entities.DrawType) error {
	return r.db.Save(drawType).Error
}

func (r *DrawTypeRepositoryGORM) Delete(id uint) error {
	return r.db.Delete(&entities.DrawType{}, id).Error
}

func (r *DrawTypeRepositoryGORM) FindByName(name string) (*entities.DrawType, error) {
	var drawType entities.DrawType
	err := r.db.Where("name = ?", name).Preload("PrizeTemplates").First(&drawType).Error
	if err != nil {
		return nil, err
	}
	return &drawType, nil
}
