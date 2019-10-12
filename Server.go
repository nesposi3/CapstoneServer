package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	//"math/rand"
)

type dividend struct {
	boughtStock stock
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
	name      string
	price     int
	numShares int
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

//Directly changes the name of a stock
func (s *stock) changeName(name string) {
	s.name = name
}

//Directly changes the number of shares of a stock
func (s *stock) changeShares(shares int) {
	s.numShares = shares
}
func (s *stock) waitAndChange() {
	for {
		time.Sleep(2 * time.Second)
		s.price += 3
	}
}
func (s *stock) equals(other stock) bool {
	return (s.name == other.name)
}

//Adds a player to the game
func (game *gamestate) addPlayer(newPlayer player) {
	game.players = append(game.players, newPlayer)
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
	http.Handle("/", r)
	// Go keyword launches the function in another thread
	http.ListenAndServe(":8090", nil)
}

// Need multithreading to change state asynchronoously
