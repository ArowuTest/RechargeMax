package services

import (
	"errors"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/infrastructure/persistence"

	"github.com/google/uuid"
)

type DrawTypeService struct {
	repo *persistence.DrawTypeRepositoryGORM
}

func NewDrawTypeService(repo *persistence.DrawTypeRepositoryGORM) *DrawTypeService {
	return &DrawTypeService{repo: repo}
}

func (s *DrawTypeService) CreateDrawType(name, description string) (*entities.DrawType, error) {
	if name == "" {
		return nil, errors.New("draw type name cannot be empty")
	}

	existing, _ := s.repo.FindByName(name)
	if existing != nil {
		return nil, errors.New("draw type with this name already exists")
	}

	var desc *string
	if description != "" {
		desc = &description
	}

	drawType := &entities.DrawType{
		Name:        name,
		Description: desc,
	}

	if err := s.repo.Create(drawType); err != nil {
		return nil, err
	}

	return drawType, nil
}

func (s *DrawTypeService) GetDrawType(id uuid.UUID) (*entities.DrawType, error) {
	return s.repo.FindByID(id)
}

func (s *DrawTypeService) GetAllDrawTypes() ([]entities.DrawType, error) {
	return s.repo.FindAll()
}

func (s *DrawTypeService) UpdateDrawType(id uuid.UUID, name, description string) (*entities.DrawType, error) {
	drawType, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("draw type not found")
	}

	if name != "" {
		existing, _ := s.repo.FindByName(name)
		if existing != nil && existing.ID != id {
			return nil, errors.New("draw type with this name already exists")
		}
		drawType.Name = name
	}

	if description != "" {
		drawType.Description = &description
	}

	if err := s.repo.Update(drawType); err != nil {
		return nil, err
	}

	return drawType, nil
}

func (s *DrawTypeService) DeleteDrawType(id uuid.UUID) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("draw type not found")
	}

	return s.repo.Delete(id)
}

func (s *DrawTypeService) GetDrawTypeByName(name string) (*entities.DrawType, error) {
	return s.repo.FindByName(name)
}
