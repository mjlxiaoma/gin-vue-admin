package repository

import (
	"context"

	"gva/internal/model"
	"gva/internal/pkg/pagination"

	"gorm.io/gorm"
)

// ProductRepository 商品数据访问接口。
type ProductRepository interface {
	List(ctx context.Context, q pagination.Query) (pagination.Result[model.Product], error)
	FindByID(ctx context.Context, id uint) (*model.Product, error)
	Create(ctx context.Context, e *model.Product) error
	Update(ctx context.Context, e *model.Product) error
	Delete(ctx context.Context, id uint) error
	BatchDelete(ctx context.Context, ids []uint) error
}

// productRepository 复用通用 CRUD，仅补充商品列表过滤条件。
type productRepository struct {
	*GenericRepository[model.Product]
}

// NewProductRepository 构造商品仓储。
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{GenericRepository: NewGenericRepository[model.Product](db)}
}

// List 商品分页列表，支持名称模糊搜索和状态精确过滤。
func (r *productRepository) List(ctx context.Context, q pagination.Query) (pagination.Result[model.Product], error) {
	return r.GenericRepository.List(ctx, q, func(db *gorm.DB) *gorm.DB {
		if q.Keyword != "" {
			db = db.Where("name LIKE ?", "%"+q.Keyword+"%")
		}
		if q.Status != "" {
			db = db.Where("status = ?", q.Status)
		}
		return db
	})
}
