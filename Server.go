package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	sqlURL, exists := os.LookupEnv("MONGO_URL")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(sqlURL))
	if err != nil {
		log.Fatal(err)
	}
	clientOptions := options.Client().ApplyURI(sqlURL)
	if !exists {
		fmt.Print("Environment Variable not set")
		return
	}
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("cpnnect")
	go waitAndUpdate(sqlURL)
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome")
	})
	r.HandleFunc("/register/{name}-{passHash}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		hash := vars["passHash"]
		if register(sqlURL, name, hash) {
			fmt.Fprintf(w, "Success")
		} else {
			http.Error(w, "User Already exists", http.StatusConflict)
		}

	})
	r.HandleFunc("/login/{name}-{passHash}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		hash := vars["passHash"]
		// Check database if username/password hash exists. Send different error mesages for different cases.
		if authLogin(sqlURL, name, hash) {
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
		stockName := vars["stockName"]
		numShares, _ := strconv.Atoi(vars["numShares"])
		currGame := getGameState(gameID, gamelist)
		if currGame == nil {
			http.Error(w, "No Such Game", http.StatusNotFound)
		} else if !authLogin(sqlURL, name, hash) {
			http.Error(w, "Forbidden, Login Failed", http.StatusForbidden)
		} else {
			//Check if User exists in game
			if currGame.checkPlayerExists(name) {
				if buyOrSell {
					if currGame.buyStock(name, stockName, numShares) {
						fmt.Fprintf(w, "Bought %d shares of stock %s", numShares, stockName)
					} else {
						http.Error(w, "Forbidden, not enough money to buy stock", http.StatusForbidden)
					}
				} else {
					if currGame.sellStock(name, stockName, numShares) {
						fmt.Fprintf(w, "Sold %d shares of stock %s", numShares, stockName)
					} else {
						http.Error(w, "Forbidden, not enough shares to sell", http.StatusForbidden)
					}
				}
			} else {
				http.Error(w, "Forbidden, user does not exist in game", http.StatusForbidden)
			}
		}
	})
	r.HandleFunc("/allGames/{name}-{passHash}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		hash := vars["passHash"]
		list := []*gamestate{}
		// Check database if username/password hash exists. Send different error mesages for different cases.
		if authLogin(sqlURL, name, hash) {
			for _, g := range gamelist {
				fmt.Println(g.GameID)
				if g.checkPlayerExists(name) {
					list = append(list, g)
				}
			}
			j, _ := json.Marshal(list)
			w.Write(j)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
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
		} else if !authLogin(sqlURL, name, hash) {
			http.Error(w, "Forbidden", http.StatusForbidden)
		} else {
			//Check if User exists in game
			if currGame.checkPlayerExists(name) {
				j, _ := json.Marshal(currGame)
				w.Write(j)
			} else {
				http.Error(w, "Forbidden, user does not exist in game", http.StatusForbidden)
			}

		}
	})

	// Creates game, adds to gamelist, adds to user's game list
	r.HandleFunc("/game/{gameID}/create/{name}-{passHash}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		hash := vars["passHash"]
		gameID := vars["gameID"]
		if !authLogin(sqlURL, name, hash) {
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
	// Creates game, adds to gamelist, adds to user's game list
	r.HandleFunc("/game/{gameID}/join/{name}-{passHash}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		hash := vars["passHash"]
		gameID := vars["gameID"]
		if !authLogin(sqlURL, name, hash) {
			http.Error(w, "Forbidden", http.StatusForbidden)
		} else {
			for _, g := range gamelist {
				if g.GameID == gameID {
					if g.checkPlayerExists(name) {
						http.Error(w, "Player Lready Exists", http.StatusConflict)
						return
					}
					g.addPlayer(&player{
						name,
						false,
						hash,
						[]*dividend{},
						startingCents,
					})
					j, _ := json.Marshal(g)
					w.Write(j)
				}
			}
			http.Error(w, "Game does not exist", http.StatusNotFound)
		}
	})

	http.Handle("/", r)
	http.ListenAndServe(":8090", nil)
}
