package ravendb

type GeekPerson struct {
	Name                    string  `json:"Name"`
	FavoritePrimes          []int   `json:"FavoritePrimes"`
	FavoriteVeryLargePrimes []int64 `json:"FavoriteVeryLargePrimes"`
}

func NewGeekPerson() *GeekPerson {
	return &GeekPerson{}
}

func (p *GeekPerson) getName() string {
	return p.Name
}

func (p *GeekPerson) setName(name string) {
	p.Name = name
}

func (p *GeekPerson) getFavoritePrimes() []int {
	return p.FavoritePrimes
}

func (p *GeekPerson) setFavoritePrimes(favoritePrimes []int) {
	p.FavoritePrimes = favoritePrimes
}

func (p *GeekPerson) getFavoriteVeryLargePrimes() []int64 {
	return p.FavoriteVeryLargePrimes
}

func (p *GeekPerson) setFavoriteVeryLargePrimes(favoriteVeryLargePrimes []int64) {
	p.FavoriteVeryLargePrimes = favoriteVeryLargePrimes
}
