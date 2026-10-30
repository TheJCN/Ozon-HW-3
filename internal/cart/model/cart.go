package model

type Item struct {
	SKU   int64  `json:"sku"`
	Name  string `json:"name"`
	Count uint32 `json:"count"`
	Price uint32 `json:"price"`
}

type AddRequest struct {
	Count uint32 `json:"count"`
}

type Cart struct {
	Items      []Item `json:"items"`
	TotalPrice uint32 `json:"total_price"`
}
