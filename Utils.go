package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type dividend struct {
	BoughtStock *stock
	Name        string
	NumShares   int
}

type player struct {
	Name         string
	Deleted      bool
	passwordHash string
	Portfolio    []*dividend
	TotalCash    int
}
type gamestate struct {
	GameID  string
	Players []*player
	Stocks  []*stock
}

//Prices in integer cents to avoid floating point comparisons etc.
type stock struct {
	Name              string
	Price             int
	NumShares         int
	PreviousNumShares int
	Trend             bool
}

// Returns an int representing the total value of a player's portfoilio
func (p *player) getAccountBalance() int {
	var bal = 0
	for _, d := range p.Portfolio {
		bal = bal + (d.NumShares * (d.BoughtStock.Price))
	}
	return bal
}
func (p *player) ownsStock(s stock) (bool, int) {
	for i, d := range p.Portfolio {
		if d.BoughtStock.equals(s) {
			return true, i
		}
	}
	return false, -1
}

//Returns info about a certain stock and a player. If the player owns that stock, returns how many shares and its price
func (p *player) getStockInfo(s stock) (int, int) {
	owns, index := p.ownsStock(s)
	if owns {
		return p.Portfolio[index].NumShares, p.Portfolio[index].BoughtStock.Price
	}
	return -1, -1
}

// Directly changes the price of a stock
func (s *stock) changePrice(price int) {
	s.Price = price
}

// Change the price by change/1000
func (s *stock) changePriceByPermill(change int) {
	changeInPrice := int(math.Round(float64(s.Price) * (float64(change) / 1000.0)))
	newPrice := s.Price + changeInPrice
	s.changePrice(newPrice)
}

//Directly changes the name of a stock
func (s *stock) changeName(name string) {
	s.Name = name
}

//Directly changes the number of shares of a stock
func (s *stock) changeShares(shares int) {
	s.NumShares = shares
}

// TODO Adjust numbers
func (s *stock) statisticalUpdate() {
	//old := s.price
	// First phase of stock adjustment, if huge change in shares bought, price changes. Change ratio in perMille
	changeRatio := int(((float64(s.NumShares) - float64(s.PreviousNumShares)) / float64(s.NumShares)) * 1000)
	s.changePriceByPermill(changeRatio)
	// Second phase, semi-random adjustment
	num := rand.Intn(1000)
	sign := 1
	if !s.Trend {
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
		s.Trend = !s.Trend
	}

}
func (s *stock) equals(other stock) bool {
	return (s.Name == other.Name)
}

//Adds a player to the game
func (game *gamestate) addPlayer(newPlayer *player) {
	game.Players = append(game.Players, newPlayer)
}
func (game *gamestate) updateStocks() {
	for _, s := range game.Stocks {
		s.statisticalUpdate()
	}
}
func waitAndUpdate() {
	//TODO every iteration updates db
	for {
		time.Sleep(3 * time.Second)
		for _, game := range gamelist {
			game.updateStocks()
		}
	}

}

//Sets a player to deleted state
func (game *gamestate) removePlayer(oldPlayer player) {
	for _, p := range game.Players {
		if p.Name == oldPlayer.Name {
			p.Deleted = true
		}
	}
}
func authLogin(db *sql.DB, name string, hash string) bool {
	//Database call here
	rows, err := db.Query("SELECT name FROM users WHERE name=? AND hash=?", name, hash)
	databaseCall := false
	if err != nil {
		fmt.Print(err)
		return false
	}
	if rows.Next() {
		databaseCall = true
	}
	rows.Close()
	return databaseCall
}
func register(db *sql.DB, name string, hash string) bool {
	rows, err := db.Query("SELECT name FROM users WHERE name=? AND hash=?", name, hash)
	databaseCall := false
	if err != nil {
		fmt.Print(err)
		return false
	}
	if rows.Next() {
		return false
	}
	rows.Close()
	rows, err = db.Query("INSERT INTO users (name,hash,game_list) VALUES(?,?,?)", name, hash, "")
	if err != nil {
		return false
	}
	rows.Close()
	return databaseCall
}

//Returns a tuple of active and deleted players
func (game *gamestate) getNum() (int, int) {
	var i = 0
	var j = 0
	for _, p := range game.Players {
		if !p.Deleted {
			i = i + 1
		} else {
			j = j + 1
		}
	}
	return i, j
}

// Gets the correct gamestate pointer from a list of gamestates based on gameID
func getGameState(id string, gameList []*gamestate) *gamestate {
	for _, game := range gameList {
		if game.GameID == id {
			return game
		}
	}
	return nil
}
func getInitialStocks() []*stock {
	stocks := []*stock{}

	r := csv.NewReader(strings.NewReader(initial_stocks))
	records, err := r.ReadAll()
	if err != nil {
		log.Fatalln("Error reading from csv file", err)
	}
	for index := 0; index < len(records); index++ {
		name := records[index][0]
		price, _ := strconv.Atoi(records[index][1])
		outlook, _ := strconv.Atoi(records[index][2])
		b := false
		if outlook > 1 {
			b = true
		}
		s := stock{
			name,
			price,
			0,
			0,
			b,
		}
		stocks = append(stocks, &s)
	}
	return stocks
}
func initialzeGame(games []*gamestate, user *player, gameID string) []*gamestate {
	stocks := getInitialStocks()
	players := []*player{user}
	newGame := gamestate{
		gameID,
		players,
		stocks,
	}
	games = append(games, &newGame)
	return games
}
