package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"encoding/csv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const initial_stocks = `Cyclops Industries,31966,1
Champion Intelligence,18803,1
Voyage Technologies,6163,0
Dwarf,52906,1
Phenomenon Enterprises,46994,1
White Wolf Sports,90973,1
Surge Aviation,12540,0
Turtle Co.,27772,0
Greatechnolgies,15117,1
Twisterecords,50189,1
Prodintelligence,84983,0
Solsticetems,30138,0
Freacrosystems,62412,0
Rootechnologies,23221,1
Luckytronics,80268,1
Aces,34272,1
Nymph cast,25578,1
Herb aid,70147,1
Mountain stones,96694,1
Vortex ex,96270,1
Ghost media,472,1
Riddle fly,1314,0
Globe mobile,1474,1
Tulip bit,1073,1
Sail air,811,1`

const startingCents = 1000000

var gamelist = []*gamestate{}

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
	//fmt.Printf("%s: %d old %d new\n", s.name, old, s.price)

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
			fmt.Println(game.GameID)
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

func main() {
	sqlURL, exists := os.LookupEnv("SQL_STRING")
	if !exists {
		fmt.Print("Environment Variable not set")
		return
	}
	db, err := sql.Open("mysql", sqlURL)
	if err != nil {
		fmt.Print(err)
		return
	}
	err = db.Ping()
	if err != nil {
		fmt.Print(err)
		return
	}

	corn := stock{
		"corn",
		5000,
		500,
		500,
		true,
	}
	egg := stock{
		"egg",
		25943,
		5000,
		4500,
		false,
	}
	p1 := dividend{
		&corn,
		"corn",
		67,
	}
	p2 := dividend{
		&egg,
		"egg",
		500,
	}
	port := []*dividend{&p1, &p2}
	nick := player{
		"nick",
		false,
		"dsad",
		port,
		100000,
	}
	plays := []*player{&nick}
	stocks := []*stock{&egg, &corn}
	game := gamestate{
		"23",
		plays,
		stocks,
	}
	gamelist = append(gamelist, &game)
	go waitAndUpdate()
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome")
	})
	r.HandleFunc("/register/{name}-{passHash}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		hash := vars["passHash"]
		register(db, name, hash)
		fmt.Fprintf(w, "Success")

	})
	r.HandleFunc("/login/{name}-{passHash}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		hash := vars["passHash"]
		// Check database if username/password hash exists. Send different error mesages for different cases.
		if authLogin(db, name, hash) {
			fmt.Fprintf(w, "Success")
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
	r.HandleFunc("/game/{gameID}/{buyOrSell}/{name}-{passHash}-{stockName}-{numShares}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		hash := vars["passHash"]
		gameID := vars["gameID"]
		// True for buy, false for sell
		buyOrSell := (vars["buyOrSell"] == "buy")
		currGame := getGameState(gameID, gamelist)
		if currGame == nil {
			http.Error(w, "No Such Game", http.StatusNotFound)
		} else if !authLogin(db, name, hash) {
			http.Error(w, "Forbidden", http.StatusForbidden)
		} else {
			//Check if User exists in game
			if buyOrSell {
				fmt.Fprintf(w, "Buy")
			} else {
				fmt.Fprintf(w, "Sell")
			}

		}
	})
	// Get status of all gamestate
	r.HandleFunc("/game/{gameID}/getGameStatus/{name}-{passHash}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		hash := vars["passHash"]
		gameID := vars["gameID"]
		currGame := getGameState(gameID, gamelist)
		if currGame == nil {
			http.Error(w, "No Such Game", http.StatusNotFound)
		} else if !authLogin(db, name, hash) {
			http.Error(w, "Forbidden", http.StatusForbidden)
		} else {
			//Check if User exists in game
			fmt.Fprint(w, "Status")

		}
	})
	// Creates game, adds to gamelist, adds to user's game list
	r.HandleFunc("/game/{gameID}/create/{name}-{passHash}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		hash := vars["passHash"]
		gameID := vars["gameID"]
		if !authLogin(db, name, hash) {
			http.Error(w, "Forbidden", http.StatusForbidden)
		} else {
			for _, g := range gamelist {
				if g.GameID == gameID {
					http.Error(w, "Game Already Exists with ID "+gameID, http.StatusConflict)
					return
				}
			}
			newList := initialzeGame(gamelist, &player{
				name,
				false,
				hash,
				[]*dividend{},
				startingCents,
			}, gameID)
			gamelist = newList
			fmt.Fprintf(w, "Game %s created", gameID)
		}
	})
	r.HandleFunc("/game/{gameID}/getStockStatus/{name}-{passHash}-{stockName}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		hash := vars["passHash"]
		gameID := vars["gameID"]
		stockName := vars["stockName"]

		currGame := getGameState(gameID, gamelist)
		if currGame == nil {
			http.Error(w, "No Such Game", http.StatusNotFound)
		} else if !authLogin(db, name, hash) {
			http.Error(w, "Forbidden", http.StatusForbidden)
		} else {
			//Check if User exists in game
			fmt.Fprintf(w, "Status of %s", stockName)

		}
	})

	http.Handle("/", r)
	http.ListenAndServe(":8090", nil)
}
