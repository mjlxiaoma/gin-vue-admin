package service

import (
	"context"
	"errors"

	"gva/internal/model"
	"gva/internal/pkg/apperr"
	"gva/internal/pkg/pagination"
	"gva/internal/repository"

	"gorm.io/gorm"
)

// ProductService 商品业务服务。
type ProductService struct {
	repo repository.ProductRepository
}

// NewProductService 构造商品服务。
func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// List 商品分页列表。
func (s *ProductService) List(ctx context.Context, q pagination.Query) (pagination.Result[model.Product], error) {
	q.Normalize()
	return s.repo.List(ctx, q)
}

// Get 商品详情；记录不存在或已软删时返回 404。
func (s *ProductService) Get(ctx context.Context, id uint) (*model.Product, error) {
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("商品不存在")
		}
		return nil, err
	}
	return e, nil
}

// Create 创建商品；创建人和时间由 GORM 回调自动写入。
func (s *ProductService) Create(ctx context.Context, e *model.Product) error {
	return s.repo.Create(ctx, e)
}

// Update 全量更新商品。
func (s *ProductService) Update(ctx context.Context, e *model.Product) error {
	return s.repo.Update(ctx, e)
}

// Delete 软删除商品，先确认记录存在。
func (s *ProductService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.NotFound("商品不存在")
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

// BatchDelete 批量软删除商品。
func (s *ProductService) BatchDelete(ctx context.Context, ids []uint) error {
	return s.repo.BatchDelete(ctx, ids)
}
