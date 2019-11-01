package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

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
