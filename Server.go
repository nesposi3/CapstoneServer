package main

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type dividend struct {
	boughtStock stock
	name        string
	numShares   int
}

type player struct {
	name         string
	deleted      bool
	passwordHash int
	portfolio    []dividend
}
type gamestate struct {
	gameID     string
	players    []player
	stocks     []stock
	tickLength int
}

//Prices in integer cents to avoid floating point comparisons etc.
type stock struct {
	name              string
	price             int
	numShares         int
	previousNumShares int
	trend             bool
}

// Returns an int representing the total value of a player's portfoilio
func (p *player) getAccountBalance() int {
	var bal = 0
	for _, d := range p.portfolio {
		bal = bal + (d.numShares * (d.boughtStock.price))
	}
	return bal
}
func (p *player) ownsStock(s stock) (bool, int) {
	for i, d := range p.portfolio {
		if d.boughtStock.equals(s) {
			return true, i
		}
	}
	return false, -1
}

//Returns info about a certain stock and a player. If the player owns that stock, returns how many shares and its price
func (p *player) getStockInfo(s stock) (int, int) {
	owns, index := p.ownsStock(s)
	if owns {
		return p.portfolio[index].numShares, p.portfolio[index].boughtStock.price
	}
	return -1, -1
}

// Directly changes the price of a stock
func (s *stock) changePrice(price int) {
	s.price = price
}

// Change the price by change/1000
func (s *stock) changePriceByPermill(change int) {
	newPrice := int(math.Round(float64(s.price) * (float64(change) / 1000.0)))
	s.changePrice(newPrice)
}

//Directly changes the name of a stock
func (s *stock) changeName(name string) {
	s.name = name
}

//Directly changes the number of shares of a stock
func (s *stock) changeShares(shares int) {
	s.numShares = shares
}

// TODO Adjust numbers
func (s *stock) statisticalUpdate() {

	// First phase of stock adjustment, if huge change in shares bought, price changes. Change ratio in perMille
	changeRatio := int(((float64(s.numShares) - float64(s.previousNumShares)) / float64(s.numShares)) * 1000)
	s.changePriceByPermill(changeRatio)
	// Second phase, semi-random adjustment
	num := rand.Intn(1000)
	sign := 1
	if !s.trend {
		sign = -1
	}
	// 90% chance we do nothing drastic, 1% change
	if num < 901 {
		changePerMill := sign * 10
		s.changePriceByPermill(changePerMill)
	} else if num > 900 && num < 951 {
		// 5% chance we  have a big change, 10% change
		changePerMill := sign * 100
		s.changePriceByPermill(changePerMill)
	} else if num > 950 && num < 1001 {
		// Change the trend of the stock
		s.trend = !s.trend
	}
}
func (s *stock) equals(other stock) bool {
	return (s.name == other.name)
}

//Adds a player to the game
func (game *gamestate) addPlayer(newPlayer player) {
	game.players = append(game.players, newPlayer)
}
func (game *gamestate) updateStocks() {
	for _, s := range game.stocks {
		s.statisticalUpdate()
	}
}
func waitAndUpdate(gameList []gamestate) {
	time.Sleep(time.Minute)
	for _, game := range gameList {
		game.updateStocks()
	}
}

//Sets a player to deleted state
func (game *gamestate) removePlayer(oldPlayer player) {
	for _, p := range game.players {
		if p.name == oldPlayer.name {
			p.deleted = true
		}
	}
}

//Returns a tuple of active and deleted players
func (game *gamestate) getNum() (int, int) {
	var i = 0
	var j = 0
	for _, p := range game.players {
		if !p.deleted {
			i = i + 1
		} else {
			j = j + 1
		}
	}
	return i, j
}
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome")
	})
	r.HandleFunc("/register/{name}-{passHash}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fmt.Printf(vars["name"])
	})
	http.Handle("/", r)
	// Go keyword launches the function in another thread
	http.ListenAndServe(":8090", nil)
}

// Need multithreading to change state asynchronoously
