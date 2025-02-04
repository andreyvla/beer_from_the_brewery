package utils

import (
	"beer_from_the_brewery/models"
	"fmt"
	"strings"
)

func FormatBeerInfo(beer models.Beer, detailed bool) string {
	beerInfo := fmt.Sprintf("*%s - %s*\nЦена: %.2f\nВ наличии: %d", beer.Name, beer.Type, beer.Price, beer.Quantity)
	if detailed {
		beerInfo += fmt.Sprintf("\n%s", beer.Description)
	}
	return beerInfo
}

func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
