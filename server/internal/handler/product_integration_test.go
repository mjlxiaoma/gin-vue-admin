package handler

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func productBody(name string) map[string]any {
	return map[string]any{
		"name": name, "category": "数码", "price": 199.9,
		"stock": 10, "status": "on",
	}
}

func productID(t *testing.T, body map[string]any, token string, r http.Handler) string {
	t.Helper()
	w := authedReq(t, r, http.MethodPost, "/api/product", token, body)
	require.Equal(t, 200, w.Code, w.Body.String())
	data, ok := decodeResult(t, w).Data.(map[string]any)
	require.True(t, ok)
	return strconv.Itoa(int(data["id"].(float64)))
}

func TestProductHandler_CRUD(t *testing.T) {
	r, token := newAppServer(t)
	body := productBody("测试商品")
	id := productID(t, body, token, r)

	w := authedReq(t, r, http.MethodGet, "/api/product/"+id, token, nil)
	require.Equal(t, 200, w.Code, w.Body.String())
	data, _ := decodeResult(t, w).Data.(map[string]any)
	assert.Equal(t, "测试商品", data["name"])

	body["name"] = "测试商品改名"
	body["price"] = 299.9
	body["stock"] = 8
	body["status"] = "off"
	w = authedReq(t, r, http.MethodPut, "/api/product/"+id, token, body)
	require.Equal(t, 200, w.Code, w.Body.String())

	w = authedReq(t, r, http.MethodGet, "/api/product/"+id, token, nil)
	require.Equal(t, 200, w.Code)
	data, _ = decodeResult(t, w).Data.(map[string]any)
	assert.Equal(t, "测试商品改名", data["name"])
	assert.Equal(t, 299.9, data["price"])
	assert.Equal(t, "off", data["status"])

	w = authedReq(t, r, http.MethodDelete, "/api/product/"+id, token, nil)
	require.Equal(t, 200, w.Code, w.Body.String())
	w = authedReq(t, r, http.MethodGet, "/api/product/"+id, token, nil)
	assert.Equal(t, 404, w.Code)
}

func TestProductHandler_ListFilters(t *testing.T) {
	r, token := newAppServer(t)
	onBody := productBody("上架键盘")
	productID(t, onBody, token, r)
	offBody := productBody("下架鼠标")
	offBody["status"] = "off"
	productID(t, offBody, token, r)

	w := authedReq(t, r, http.MethodGet, "/api/product?page=1&size=10&keyword=键盘&status=on", token, nil)
	require.Equal(t, 200, w.Code, w.Body.String())
	data, _ := decodeResult(t, w).Data.(map[string]any)
	assert.Equal(t, float64(1), data["total"])
	records, _ := data["records"].([]any)
	require.Len(t, records, 1)
}

func TestProductHandler_BatchDelete(t *testing.T) {
	r, token := newAppServer(t)
	ids := []string{
		productID(t, productBody("批量商品1"), token, r),
		productID(t, productBody("批量商品2"), token, r),
	}
	w := authedReq(t, r, http.MethodDelete, "/api/product", token, batchDeleteReq{IDs: ids})
	require.Equal(t, 200, w.Code, w.Body.String())

	w = authedReq(t, r, http.MethodGet, "/api/product?page=1&size=10", token, nil)
	require.Equal(t, 200, w.Code)
	data, _ := decodeResult(t, w).Data.(map[string]any)
	assert.Equal(t, float64(0), data["total"])
}

func TestProductHandler_Validation(t *testing.T) {
	r, token := newAppServer(t)
	tests := []struct {
		name string
		body map[string]any
	}{
		{name: "缺少名称", body: map[string]any{"category": "数码", "status": "on"}},
		{name: "负价格", body: map[string]any{"name": "商品", "category": "数码", "price": -1, "status": "on"}},
		{name: "负库存", body: map[string]any{"name": "商品", "category": "数码", "stock": -1, "status": "on"}},
		{name: "非法状态", body: map[string]any{"name": "商品", "category": "数码", "status": "unknown"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := authedReq(t, r, http.MethodPost, "/api/product", token, tt.body)
			assert.Equal(t, 422, w.Code, w.Body.String())
		})
	}
}

func TestProductHandler_NoToken(t *testing.T) {
	r, _ := newAppServer(t)
	w := authedReq(t, r, http.MethodGet, "/api/product", "", nil)
	assert.Equal(t, 401, w.Code)
}

func TestProductHandler_NormalUserForbidden(t *testing.T) {
	r, _ := newAppServer(t)
	w := doJSON(t, r, http.MethodPost, "/api/auth/sessions", LoginRequest{Username: "user", Password: "123456"})
	require.Equal(t, 200, w.Code, w.Body.String())
	data, _ := decodeResult(t, w).Data.(map[string]any)
	token := data["accessToken"].(string)

	w = authedReq(t, r, http.MethodGet, "/api/product", token, nil)
	assert.Equal(t, 403, w.Code)
}
