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
		Price:    1,
		PhotoURL: "https://example.com/static/starter.png",
	},
	"pack_5": {
		ID:       2,
		Title:    "5 кредитов",
		Price:    5,
		PhotoURL: "https://example.com/static/pack_5.png",
	},
	"pack_10": {
		ID:       3,
		Title:    "10 кредитов",
		Price:    10,
		PhotoURL: "https://example.com/static/pack_10.png",
	},
}

func getProduct(name string) *Product {
	product, ok := items[name]
	if !ok {
		return nil
	}
	return product
}
