package vkHandler

type Products []Product
type Product struct {
	ID       int64
	Title    string
	Name     string
	Credits  int64
	Price    int // в голосах
	PhotoURL string
}

// Каталог товаров.
// Ключ = то самое item, которое ты передаешь в VKWebAppShowOrderBox.
var items = map[string]*Product{
	"pack_1": {
		ID:       1,
		Title:    "1 кредит",
		Name:     "pack_1",
		Credits:  1,
		Price:    10,
		PhotoURL: "https://graphai-pay.ai128.ru/credit_1.png",
	},
	"pack_5": {
		ID:       2,
		Title:    "5 кредитов",
		Name:     "pack_5",
		Credits:  5,
		Price:    45,
		PhotoURL: "https://graphai-pay.ai128.ru/credit_5.png",
	},
	"pack_10": {
		ID:       3,
		Title:    "10 кредитов",
		Name:     "pack_10",
		Credits:  10,
		Price:    80,
		PhotoURL: "https://graphai-pay.ai129.ru/credit_10.png",
	},
	"pack_25": {
		ID:       4,
		Title:    "25 кредитов",
		Name:     "pack_25",
		Credits:  25,
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

func getProductById(id int64) *Product {
	for _, product := range items {
		if product.ID == id {
			return product
		}
	}
	return nil
}
