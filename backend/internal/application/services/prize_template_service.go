package services

import (
	"errors"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/infrastructure/persistence"

	"github.com/google/uuid"
)

type PrizeTemplateService struct {
	templateRepo *persistence.PrizeTemplateRepositoryGORM
	categoryRepo *persistence.PrizeCategoryRepositoryGORM
}

func NewPrizeTemplateService(
	templateRepo *persistence.PrizeTemplateRepositoryGORM,
	categoryRepo *persistence.PrizeCategoryRepositoryGORM,
) *PrizeTemplateService {
	return &PrizeTemplateService{
		templateRepo: templateRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *PrizeTemplateService) CreateTemplate(
	name, description string,
	drawTypeID uuid.UUID,
	isDefault bool,
	categories []entities.PrizeCategory,
) (*entities.PrizeTemplate, error) {
	if name == "" {
		return nil, errors.New("template name cannot be empty")
	}
	if len(categories) == 0 {
		return nil, errors.New("template must have at least one prize category")
	}

	existing, _ := s.templateRepo.FindByName(name)
	if existing != nil {
		return nil, errors.New("template with this name already exists")
	}

	var desc *string
	if description != "" {
		desc = &description
	}

	template := &entities.PrizeTemplate{
		Name:        name,
		Description: desc,
		DrawTypeID:  drawTypeID,
		IsDefault:   isDefault,
	}

	if err := s.templateRepo.Create(template); err != nil {
		return nil, err
	}

	for i := range categories {
		categories[i].PrizeTemplateID = template.ID
		if categories[i].DisplayOrder == 0 {
			categories[i].DisplayOrder = i + 1
		}
	}

	if err := s.categoryRepo.BulkCreate(categories); err != nil {
		_ = s.templateRepo.Delete(template.ID)
		return nil, err
	}

	if isDefault {
		_ = s.templateRepo.SetAsDefault(template.ID, drawTypeID)
	}

	return s.templateRepo.FindByID(template.ID)
}

func (s *PrizeTemplateService) GetTemplate(id uuid.UUID) (*entities.PrizeTemplate, error) {
	return s.templateRepo.FindByID(id)
}

func (s *PrizeTemplateService) GetAllTemplates() ([]entities.PrizeTemplate, error) {
	return s.templateRepo.FindAll()
}

func (s *PrizeTemplateService) GetTemplatesByDrawType(drawTypeID uuid.UUID) ([]entities.PrizeTemplate, error) {
	return s.templateRepo.FindByDrawTypeID(drawTypeID)
}

func (s *PrizeTemplateService) UpdateTemplate(
	id uuid.UUID,
	name, description string,
	isDefault *bool,
) (*entities.PrizeTemplate, error) {
	template, err := s.templateRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("template not found")
	}

	if name != "" {
		existing, _ := s.templateRepo.FindByName(name)
		if existing != nil && existing.ID != id {
			return nil, errors.New("template with this name already exists")
		}
		template.Name = name
	}

	if description != "" {
		template.Description = &description
	}

	if isDefault != nil {
		template.IsDefault = *isDefault
		if *isDefault {
			_ = s.templateRepo.SetAsDefault(id, template.DrawTypeID)
		}
	}

	if err := s.templateRepo.Update(template); err != nil {
		return nil, err
	}

	return s.templateRepo.FindByID(id)
}

func (s *PrizeTemplateService) DeleteTemplate(id uuid.UUID) error {
	_, err := s.templateRepo.FindByID(id)
	if err != nil {
		return errors.New("template not found")
	}
	return s.templateRepo.Delete(id)
}

func (s *PrizeTemplateService) AddCategoryToTemplate(
	templateID uuid.UUID,
	categoryName string,
	prizeAmount float64,
	winnerCount, runnerUpCount int,
) (*entities.PrizeCategory, error) {
	template, err := s.templateRepo.FindByID(templateID)
	if err != nil {
		return nil, errors.New("template not found")
	}

	categories, _ := s.categoryRepo.FindByTemplateID(templateID)
	displayOrder := len(categories) + 1

	category := &entities.PrizeCategory{
		PrizeTemplateID: template.ID,
		CategoryName:    categoryName,
		PrizeAmount:     prizeAmount,
		WinnerCount:     winnerCount,
		RunnerUpCount:   runnerUpCount,
		DisplayOrder:    displayOrder,
	}

	if err := s.categoryRepo.Create(category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *PrizeTemplateService) UpdateCategory(
	id uuid.UUID,
	categoryName *string,
	prizeAmount *float64,
	winnerCount, runnerUpCount *int,
) (*entities.PrizeCategory, error) {
	category, err := s.categoryRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("category not found")
	}

	if categoryName != nil {
		category.CategoryName = *categoryName
	}
	if prizeAmount != nil {
		category.PrizeAmount = *prizeAmount
	}
	if winnerCount != nil {
		category.WinnerCount = *winnerCount
	}
	if runnerUpCount != nil {
		category.RunnerUpCount = *runnerUpCount
	}

	if err := s.categoryRepo.Update(category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *PrizeTemplateService) DeleteCategory(id uuid.UUID) error {
	return s.categoryRepo.Delete(id)
}

func (s *PrizeTemplateService) GetCategoriesByTemplate(templateID uuid.UUID) ([]entities.PrizeCategory, error) {
	return s.categoryRepo.FindByTemplateID(templateID)
}

func (s *PrizeTemplateService) CalculateTotalPrizePool(templateID uuid.UUID) (float64, error) {
	categories, err := s.categoryRepo.FindByTemplateID(templateID)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, cat := range categories {
		total += cat.PrizeAmount * float64(cat.WinnerCount)
	}
	return total, nil
}
