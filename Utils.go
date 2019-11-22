package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
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
	GameID    string
	Players   []*player
	Stocks    []*stock
	TicksLeft int
	Done      bool
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
	// First phase of stock adjustment, if huge change in shares bought, price changes. Change ratio in perMille
	// if s.NumShares != 0 {
	// 	changeRatio := int(((float64(s.NumShares) - float64(s.PreviousNumShares)) / float64(s.NumShares) * 1000))
	// 	s.changePriceByPermill(changeRatio)
	// }
	// Second phase, semi-random adjustment
	num := rand.Intn(1000)
	sign := 1
	if !s.Trend {
		sign = -1
	}
	// 90% chance we do nothing drastic, .1% change
	if num < 901 {
		changePerMill := sign * 1
		s.changePriceByPermill(changePerMill)
	} else if num > 900 && num < 951 {
		// 5% chance we  have a big change, 1% change
		changePerMill := sign * 10
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
func (game *gamestate) updateGamestateInDatabase(sqlURL string) {
	db, err := sql.Open("mysql", sqlURL)
	if err != nil {
		fmt.Println(err)
		db.Close()
		return
	}
	obj, err := json.Marshal(game)
	if err != nil {
		fmt.Println(err)
		db.Close()
		return
	}
	s := string(obj)
	rows, err := db.Query("INSERT INTO games (game_id, game_state) VALUES(?,?) ON DUPLICATE KEY UPDATE game_id=?, game_state=?;", game.GameID, s, game.GameID, s)
	rows.Close()
	db.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}
func waitAndUpdate(sqlURL string) {
	for {
		time.Sleep(1 * time.Minute)
		for i, game := range gamelist {
			if game.TicksLeft == 0 {
				game.Done = true
			} else {
				game.TicksLeft--
				game.updateStocks()
			}
			game.updateGamestateInDatabase(sqlURL)
			addHistory(game, i)
		}
	}

}
func addHistory(game *gamestate, index int) {
	if len(historyQueue) <= index {
		historyQueue = append(historyQueue, []gamestate{})
	}
	if len(historyQueue[index]) == 60 {
		historyQueue[index] = historyQueue[index][1:]
	}
	newStocks := []*stock{}
	for _, s := range game.Stocks {
		newStock := stock{
			s.Name,
			s.Price,
			s.NumShares,
			s.PreviousNumShares,
			s.Trend,
		}
		newStocks = append(newStocks, &newStock)
	}
	game.Stocks = newStocks
	historyQueue[index] = append(historyQueue[index], *game)
}

//Sets a player to deleted state
func (game *gamestate) removePlayer(oldPlayer player) {
	for _, p := range game.Players {
		if p.Name == oldPlayer.Name {
			p.Deleted = true
		}
	}
}
func authLogin(sqlURL string, name string, hash string) bool {
	//Database call here
	db, _ := sql.Open("mysql", sqlURL)
	rows, err := db.Query("SELECT name FROM users WHERE name=? AND hash=?;", name, hash)
	db.Close()
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
func register(sqlURL string, name string, hash string) bool {
	db, _ := sql.Open("mysql", sqlURL)
	var scanName string
	row := db.QueryRow("SELECT name FROM users WHERE name=? AND hash=?;", name, hash)
	err := row.Scan(&scanName)
	databaseCall := true
	if err == nil {
		db.Close()
		return false
	}
	rows, err := db.Query("INSERT INTO users (name,hash,game_list) VALUES(?,?,?);", name, hash, "")
	if err != nil {
		db.Close()
		return false
	}
	rows.Close()
	db.Close()
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
func (game *gamestate) checkPlayerExists(name string) bool {
	for _, p := range game.Players {
		fmt.Println(p.Name + " " + name)
		if p.Name == name && !p.Deleted {
			return true
		}
	}
	return false
}
func (game *gamestate) checkStockExists(name string) bool {
	for _, s := range game.Stocks {
		if s.Name == name {
			return true
		}
	}
	return false
}
func (game *gamestate) buyStock(userName string, stockName string, numShares int) bool {
	var play *player = nil
	var stock *stock = nil
	for _, p := range game.Players {
		if p.Name == userName && !p.Deleted {
			play = p
		}
	}
	for _, s := range game.Stocks {
		if s.Name == stockName {
			stock = s
		}
	}
	if numShares*stock.Price > play.TotalCash {
		// Player cannot afford to buy that stock
		return false
	}
	stock.NumShares += numShares
	owns, i := play.ownsStock(*stock)
	if owns {
		play.Portfolio[i].BoughtStock = stock
		play.Portfolio[i].NumShares += numShares
		play.TotalCash -= (numShares * stock.Price)
	} else {
		div := dividend{
			stock,
			stock.Name,
			numShares,
		}
		play.Portfolio = append(play.Portfolio, &div)
		play.TotalCash -= (numShares * stock.Price)
	}
	return true
}
func (game *gamestate) sellStock(userName string, stockName string, numShares int) bool {
	var play *player = nil
	var stock *stock = nil
	for _, p := range game.Players {
		if p.Name == userName && !p.Deleted {
			play = p
		}
	}
	for _, s := range game.Stocks {
		if s.Name == stockName {
			stock = s
		}
	}
	owns, i := play.ownsStock(*stock)
	if owns {
		if stock.NumShares < numShares || play.Portfolio[i].NumShares < numShares {
			// If there are not enough shares to sell
			return false
		}
		stock.NumShares -= numShares
		play.Portfolio[i].BoughtStock = stock
		play.Portfolio[i].NumShares -= numShares
		play.TotalCash += (numShares * stock.Price)
	} else {
		return false
	}
	return true
}

// Gets the correct gamestate pointer from a list of gamestates based on gameID
func getGameState(id string, gameList []*gamestate) (*gamestate, int) {
	for i, game := range gameList {
		if game.GameID == id {
			return game, i
		}
	}
	return nil, -1
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
		totalTicks,
		false,
	}
	games = append(games, &newGame)
	return games
}
