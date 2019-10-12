package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	//"math/rand"
)

type player struct {
	name string
}
type gamestate struct {
	gameID  string
	players []player
	stocks  []stock
}

//Prices in integer cents to avoid floating point comparisons etc.
type stock struct {
	name      string
	price     int
	numShares int
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
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, fmt.Sprint(jame.price))
	})
	http.Handle("/", r)
	// Go keyword launches the function in another thread
	http.ListenAndServe(":8090", nil)
}

// Need multithreading to change state asynchronoously
