package repository

import (
	"context"
	"testing"

	"gva/internal/model"
	"gva/internal/pkg/pagination"
	"gva/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newProductRepo(t *testing.T) ProductRepository {
	t.Helper()
	return NewProductRepository(testutil.NewTestDB(t))
}

func TestProductRepository_CRUDAndList(t *testing.T) {
	repo := newProductRepo(t)
	ctx := context.Background()

	e := &model.Product{Name: "机械键盘", Category: "数码", Price: 299.90, Stock: 20, Status: "on"}
	require.NoError(t, repo.Create(ctx, e))
	require.NotZero(t, e.ID)

	got, err := repo.FindByID(ctx, e.ID)
	require.NoError(t, err)
	assert.Equal(t, "机械键盘", got.Name)
	assert.Equal(t, 299.90, got.Price)

	got.Stock = 18
	got.Status = "off"
	require.NoError(t, repo.Update(ctx, got))
	updated, err := repo.FindByID(ctx, e.ID)
	require.NoError(t, err)
	assert.Equal(t, 18, updated.Stock)
	assert.Equal(t, "off", updated.Status)

	require.NoError(t, repo.Create(ctx, &model.Product{Name: "马克杯", Category: "家居", Price: 39.9, Stock: 50, Status: "on"}))

	all, err := repo.List(ctx, pagination.Query{Page: 1, Size: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(2), all.Total)

	byName, err := repo.List(ctx, pagination.Query{Keyword: "键盘", Page: 1, Size: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), byName.Total)
	assert.Equal(t, "机械键盘", byName.Records[0].Name)

	byStatus, err := repo.List(ctx, pagination.Query{Status: "off", Page: 1, Size: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), byStatus.Total)
	assert.Equal(t, e.ID, byStatus.Records[0].ID)

	require.NoError(t, repo.Delete(ctx, e.ID))
	_, err = repo.FindByID(ctx, e.ID)
	assert.Error(t, err)
}

func TestProductRepository_BatchDelete(t *testing.T) {
	repo := newProductRepo(t)
	ctx := context.Background()
	ids := make([]uint, 0, 3)
	for _, name := range []string{"商品A", "商品B", "商品C"} {
		e := &model.Product{Name: name, Category: "测试", Status: "on"}
		require.NoError(t, repo.Create(ctx, e))
		ids = append(ids, e.ID)
	}

	require.NoError(t, repo.BatchDelete(ctx, ids))
	res, err := repo.List(ctx, pagination.Query{Page: 1, Size: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(0), res.Total)
}
