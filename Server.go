package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	//"math/rand"
)

type stock struct {
	name      string
	price     float64
	numShares int
}

// Directly changes the price of a stock
func (s *stock) changePrice(price float64) {
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

func main() {
	jame := stock{
		"corn",
		float64(3.5),
		3,
	}
	fmt.Print(jame.price)
	jame.changePrice(float64(99.20))
	fmt.Print(jame.price)
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to my website!")
	})
	http.Handle("/", r)
	http.ListenAndServe(":8090", nil)
}
