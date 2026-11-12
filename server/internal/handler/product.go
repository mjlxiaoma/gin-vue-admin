package handler

import (
	"strconv"

	"gva/internal/model"
	"gva/internal/pkg/apperr"
	"gva/internal/pkg/pagination"
	"gva/internal/pkg/response"
	"gva/internal/service"

	"github.com/gin-gonic/gin"
)

// productUpsertReq 创建和更新商品共用的请求体。
type productUpsertReq struct {
	Name     string  `json:"name" binding:"required,max=100"`
	Category string  `json:"category" binding:"required,max=50"`
	Price    float64 `json:"price" binding:"gte=0"`
	Stock    int     `json:"stock" binding:"gte=0"`
	Status   string  `json:"status" binding:"required,oneof=on off"`
}

// productBatchDeleteReq 批量删除请求体。
type productBatchDeleteReq = batchDeleteReq

// ProductHandler 商品 HTTP 处理器。
type ProductHandler struct {
	svc *service.ProductService
}

// NewProductHandler 构造商品处理器。
func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// List GET /api/product
// @Summary      商品分页列表
// @Tags         product
// @Produce      json
// @Security     BearerAuth
// @Param        page    query int    false "页码" default(1)
// @Param        size    query int    false "每页" default(10)
// @Param        keyword query string false "商品名称关键词"
// @Param        status  query string false "on/off"
// @Success      200  {object} response.ApiResult
// @Failure      422  {object} response.ProblemDetail
// @Router       /product [get]
func (h *ProductHandler) List(c *gin.Context) {
	var q pagination.Query
	if err := c.ShouldBindQuery(&q); err != nil {
		apperr.Write(c, apperr.Validation(err.Error(), nil))
		return
	}
	res, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		apperr.Write(c, err)
		return
	}
	response.Success(c, res)
}

// Get GET /api/product/:id
// @Summary      商品详情
// @Tags         product
// @Produce      json
// @Security     BearerAuth
// @Param        id   path int true "商品 ID"
// @Success      200  {object} response.ApiResult
// @Failure      404  {object} response.ProblemDetail
// @Failure      422  {object} response.ProblemDetail
// @Router       /product/{id} [get]
func (h *ProductHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		apperr.Write(c, apperr.Validation("无效的商品 id", nil))
		return
	}
	e, err := h.svc.Get(c.Request.Context(), uint(id))
	if err != nil {
		apperr.Write(c, err)
		return
	}
	response.Success(c, e)
}

// Create POST /api/product
// @Summary      创建商品
// @Tags         product
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body productUpsertReq true "商品信息"
// @Success      200  {object} response.ApiResult
// @Failure      422  {object} response.ProblemDetail
// @Router       /product [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var req productUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Write(c, apperr.Validation(err.Error(), nil))
		return
	}
	e := &model.Product{
		Name: req.Name, Category: req.Category, Price: req.Price,
		Stock: req.Stock, Status: req.Status,
	}
	if err := h.svc.Create(c.Request.Context(), e); err != nil {
		apperr.Write(c, err)
		return
	}
	response.Success(c, e, "创建成功")
}

// Update PUT /api/product/:id
// @Summary      更新商品
// @Tags         product
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "商品 ID"
// @Param        request body productUpsertReq true "商品信息"
// @Success      200  {object} response.ApiResult
// @Failure      404  {object} response.ProblemDetail
// @Failure      422  {object} response.ProblemDetail
// @Router       /product/{id} [put]
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		apperr.Write(c, apperr.Validation("无效的商品 id", nil))
		return
	}
	var req productUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Write(c, apperr.Validation(err.Error(), nil))
		return
	}
	e, err := h.svc.Get(c.Request.Context(), uint(id))
	if err != nil {
		apperr.Write(c, err)
		return
	}
	e.Name = req.Name
	e.Category = req.Category
	e.Price = req.Price
	e.Stock = req.Stock
	e.Status = req.Status
	if err := h.svc.Update(c.Request.Context(), e); err != nil {
		apperr.Write(c, err)
		return
	}
	response.Success(c, e, "更新成功")
}

// Delete DELETE /api/product/:id
// @Summary      删除商品
// @Tags         product
// @Produce      json
// @Security     BearerAuth
// @Param        id   path int true "商品 ID"
// @Success      200  {object} response.ApiResult
// @Failure      404  {object} response.ProblemDetail
// @Failure      422  {object} response.ProblemDetail
// @Router       /product/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		apperr.Write(c, apperr.Validation("无效的商品 id", nil))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		apperr.Write(c, err)
		return
	}
	response.Success(c, true, "删除成功")
}

// BatchDelete DELETE /api/product
// @Summary      批量删除商品
// @Tags         product
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body productBatchDeleteReq true "商品 ID 列表"
// @Success      200  {object} response.ApiResult
// @Failure      422  {object} response.ProblemDetail
// @Router       /product [delete]
func (h *ProductHandler) BatchDelete(c *gin.Context) {
	var req productBatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Write(c, apperr.Validation(err.Error(), nil))
		return
	}
	ids := make([]uint, 0, len(req.IDs))
	for _, s := range req.IDs {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			apperr.Write(c, apperr.Validation("无效的商品 id: "+s, nil))
			return
		}
		ids = append(ids, uint(id))
	}
	if err := h.svc.BatchDelete(c.Request.Context(), ids); err != nil {
		apperr.Write(c, err)
		return
	}
	response.Success(c, true, "删除成功")
}
