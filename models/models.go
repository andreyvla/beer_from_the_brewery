package models

type Beer struct {
	ID          int
	Name        string
	Description string
	Price       float64
	Quantity    int
	ImageURL    string
	Type        string
}

type CartItem struct {
	BeerID   int
	Quantity int
}
