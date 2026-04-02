package vkHandler

type vkError struct {
	Error struct {
		Code     int    `json:"error_code"`
		Msg      string `json:"error_msg"`
		Critical bool   `json:"critical"`
	} `json:"error"`
}

type vkSuccessGetItem struct {
	Response struct {
		ItemID   int    `json:"item_id"`
		Title    string `json:"title"`
		PhotoURL string `json:"photo_url,omitempty"`
		Price    int    `json:"price"`
	} `json:"response"`
}

type vkSuccessChargeable struct {
	Response struct {
		OrderID    int64 `json:"order_id"`
		AppOrderID int64 `json:"app_order_id"`
	} `json:"response"`
}
