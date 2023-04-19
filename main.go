package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	binance_connector "github.com/binance/binance-connector-go"
	"github.com/gorilla/websocket"
)

const binanceBaseURL string = "https://testnet.binance.vision"
const apiKey string = ""
const secretKey string = ""

var pairs = []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
var binanceClient *binance_connector.Client

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func init() {
	// Initialize the Binance Client
	binanceClient = binance_connector.NewClient(apiKey, secretKey, binanceBaseURL)
}

func main() {
	http.HandleFunc("/get_pairs", GetPairs)
	http.HandleFunc("/ws", GetCurrentPrices)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error on running server: ", err)
	}
}

// GetPairs Controller, returns the list of pairs, no business logic here
func GetPairs(w http.ResponseWriter, r *http.Request) {
	WriteHttpJsonResponse(w, http.StatusOK, pairs)
}

func WriteHttpJsonResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// GetCurrentPrices Controller, returns the current prices of the pairs via WebSocket, returns current values for every 3 seconds 
func GetCurrentPrices(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
	}
	log.Println("Client Connected")

	for {
		currentPrices, err := CollectAllCurrentPrices()
		if err != nil {
			log.Println(err.Error())
			return
		}
		if err := conn.WriteJSON(currentPrices); err != nil {
			log.Println(err)
			return
		}
		time.Sleep(3 * time.Second)
	}
}

// Response Model for WebSocket Endpoint
type CurrentPrice struct {
	Pair  string    `json:"pair"`
	Price string    `json:"price"`
	Time  time.Time `json:"time"`
}

// CollectAllCurrentPrices Business Logic, returns the current prices of the pairs
func CollectAllCurrentPrices() (response []CurrentPrice, err error) {
	response = make([]CurrentPrice, 0)
	now := time.Now()

	for _, pair := range pairs {
		price, err := GetCurrentPriceFromBinance(pair)
		if err != nil {
			log.Println(err.Error())
			return response, err
		}

		var currentPrice CurrentPrice
		currentPrice.Pair = pair
		currentPrice.Price = price
		currentPrice.Time = now

		response = append(response, currentPrice)
	}

	return response, err
}

// GetCurrentPriceFromBinance Business Logic, returns the current price of a pair
func GetCurrentPriceFromBinance(pair string) (price string, err error) {
	res, err := binanceClient.
		NewAvgPriceService().
		Symbol(pair).
		Do(context.Background())
	if err != nil {
		log.Println(err.Error())
		return price, err
	}

	price = res.Price

	return price, err
}
