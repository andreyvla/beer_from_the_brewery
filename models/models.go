package models

// Beer представляет информацию о пиве.
type Beer struct {
	ID          int     `json:"id"`          // Уникальный идентификатор пива.
	Name        string  `json:"name"`        // Название пива.
	Description string  `json:"description"` // Описание пива.
	Price       float64 `json:"price"`       // Цена пива.
	Quantity    int     `json:"quantity"`    // Количество пива в наличии.
	ImageURL    string  `json:"image_url"`   // URL изображения пива.
	Type        string  `json:"type"`        // Тип пива (например, "Лагер", "Стаут" и т.д.).
}

// CartItem представляет элемент в корзине пользователя.
type CartItem struct {
	BeerID   int `json:"beer_id"`  // ID пива в корзине.
	Quantity int `json:"quantity"` // Количество пива в корзине.
}
