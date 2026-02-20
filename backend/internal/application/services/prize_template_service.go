package services

import (
	"errors"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/infrastructure/persistence"
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
	drawTypeID uint,
	isDefault bool,
	categories []entities.PrizeCategory,
) (*entities.PrizeTemplate, error) {
	if name == "" {
		return nil, errors.New("template name cannot be empty")
	}

	if len(categories) == 0 {
		return nil, errors.New("template must have at least one prize category")
	}

	// Check if template with same name already exists
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

	// Create associated prize categories
	for i := range categories {
		categories[i].PrizeTemplateID = template.ID
		if categories[i].DisplayOrder == 0 {
			categories[i].DisplayOrder = i + 1
		}
	}

	if err := s.categoryRepo.BulkCreate(categories); err != nil {
		// Rollback template creation if categories fail
		s.templateRepo.Delete(template.ID)
		return nil, err
	}

	// If this is set as default, unset other defaults
	if isDefault {
		s.templateRepo.SetAsDefault(template.ID, drawTypeID)
	}

	// Reload with categories
	return s.templateRepo.FindByID(template.ID)
}

func (s *PrizeTemplateService) GetTemplate(id uint) (*entities.PrizeTemplate, error) {
	return s.templateRepo.FindByID(id)
}

func (s *PrizeTemplateService) GetAllTemplates() ([]entities.PrizeTemplate, error) {
	return s.templateRepo.FindAll()
}

func (s *PrizeTemplateService) GetTemplatesByDrawType(drawTypeID uint) ([]entities.PrizeTemplate, error) {
	return s.templateRepo.FindByDrawTypeID(drawTypeID)
}

func (s *PrizeTemplateService) UpdateTemplate(
	id uint,
	name, description string,
	isDefault *bool,
) (*entities.PrizeTemplate, error) {
	template, err := s.templateRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("template not found")
	}

	if name != "" {
		// Check if another template with same name exists
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
			s.templateRepo.SetAsDefault(id, template.DrawTypeID)
		}
	}

	if err := s.templateRepo.Update(template); err != nil {
		return nil, err
	}

	return s.templateRepo.FindByID(id)
}

func (s *PrizeTemplateService) DeleteTemplate(id uint) error {
	// Check if template exists
	_, err := s.templateRepo.FindByID(id)
	if err != nil {
		return errors.New("template not found")
	}

	return s.templateRepo.Delete(id)
}

func (s *PrizeTemplateService) AddCategoryToTemplate(
	templateID uint,
	categoryName string,
	prizeAmount float64,
	winnerCount, runnerUpCount int,
) (*entities.PrizeCategory, error) {
	template, err := s.templateRepo.FindByID(templateID)
	if err != nil {
		return nil, errors.New("template not found")
	}

	// Get current categories to determine display order
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
	id uint,
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

func (s *PrizeTemplateService) DeleteCategory(id uint) error {
	return s.categoryRepo.Delete(id)
}

func (s *PrizeTemplateService) GetCategoriesByTemplate(templateID uint) ([]entities.PrizeCategory, error) {
	return s.categoryRepo.FindByTemplateID(templateID)
}

func (s *PrizeTemplateService) CalculateTotalPrizePool(templateID uint) (float64, error) {
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
