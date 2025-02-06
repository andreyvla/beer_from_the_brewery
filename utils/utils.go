package utils

import (
	"beer_from_the_brewery/models"
	"fmt"
	"strings"
)

// FormatBeerInfo форматирует информацию о пиве для отправки пользователю.
//
// beer - структура с информацией о пиве.
// detailed - флаг, указывающий, нужно ли выводить подробное описание.
func FormatBeerInfo(beer models.Beer, detailed bool) string {
	beerInfo := fmt.Sprintf("*%s - %s*\nЦена: %.2f\nВ наличии: %d", beer.Name, beer.Type, beer.Price, beer.Quantity)
	if detailed {
		beerInfo += fmt.Sprintf("\n%s", beer.Description) // Добавляем описание, если нужно
	}
	return beerInfo
}

// ContainsIgnoreCase проверяет, содержит ли строка s подстроку substr без учета регистра.
func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
