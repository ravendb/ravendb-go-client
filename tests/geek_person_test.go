package tests

type GeekPerson struct {
	Name                    string  `json:"Name"`
	FavoritePrimes          []int   `json:"FavoritePrimes"`
	FavoriteVeryLargePrimes []int64 `json:"FavoriteVeryLargePrimes"`
}
