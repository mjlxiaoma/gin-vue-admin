package model

// Product 商品实体。嵌入 Model 获得主键、时间戳、审计字段与软删除能力。
type Product struct {
	Model
	Name     string  `gorm:"size:100;not null" json:"name"`
	Category string  `gorm:"size:50;not null" json:"category"`
	Price    float64 `gorm:"type:decimal(10,2);not null;default:0" json:"price"`
	Stock    int     `gorm:"not null;default:0" json:"stock"`
	Status   string  `gorm:"size:16;not null;default:on;index" json:"status"` // on=上架，off=下架
}

// TableName 指定商品表名。
func (Product) TableName() string { return "products" }
