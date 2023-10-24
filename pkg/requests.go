package pkg

type Receipt struct {
	Retailer     string        `json:"retailer"`
	PurchaseDate string        `json:"purchaseDate"`
	PurchaseTime string        `json:"purchaseTime"`
	Items        []ReceiptItem `json:"items"`
	Total        string        `json:"total"`
	PointsEarned int           `json:"pointsEarned,omitempty"`
}

type ReceiptItem struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}
