package vkHandler

type Products []Product
type Product struct {
	ID       int
	Title    string
	Price    int // в голосах
	PhotoURL string
}

// Каталог товаров.
// Ключ = то самое item, которое ты передаешь в VKWebAppShowOrderBox.
var items = map[string]*Product{
	"pack_1": {
		ID:       1,
		Title:    "1 кредит",
		Price:    10,
		PhotoURL: "https://graphai-pay.ai128.ru/credit_1.png",
	},
	"pack_5": {
		ID:       2,
		Title:    "5 кредитов",
		Price:    45,
		PhotoURL: "https://graphai-pay.ai128.ru/credit_5.png",
	},
	"pack_10": {
		ID:       3,
		Title:    "10 кредитов",
		Price:    80,
		PhotoURL: "https://graphai-pay.ai129.ru/credit_10.png",
	},
	"pack_25": {
		ID:       4,
		Title:    "25 кредитов",
		Price:    200,
		PhotoURL: "https://graphai-pay.ai128.ru/credit_25.png",
	},
}

func getProduct(name string) *Product {
	product, ok := items[name]
	if !ok {
		return nil
	}
	return product
}
