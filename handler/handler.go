package handler

import (
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/psmccarty/fetch-backend-apprentice-challenge/pkg"
)

const (
	DateLayout = time.DateOnly
	TimeLayout = "15:04"
	TwoPM      = "14:00"
	FourPM     = "16:00"
)

type ReceiptsHandler struct {
	db          map[string]pkg.Receipt
	pointsCache map[string]int
	dbLock      sync.Mutex
}

func NewReceiptsHandler() *ReceiptsHandler {
	rHandler := ReceiptsHandler{db: make(map[string]pkg.Receipt), pointsCache: make(map[string]int)}
	return &rHandler
}

func (rHandler *ReceiptsHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/receipts/process", rHandler.ProcessReceipt)
	router.GET("/receipts/:id/points", rHandler.GetPoints)
}

func (rHandler *ReceiptsHandler) GetPoints(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if id == "" {
		log.Println("GetPoints::No id")
		ServerResponse(w, http.StatusBadRequest, []byte(""))
		return
	}

	rHandler.dbLock.Lock()
	defer rHandler.dbLock.Unlock()

	// check if result is cached
	if cachedPoints, ok := rHandler.pointsCache[id]; ok {
		resp := pkg.GetPointsResponse{Points: cachedPoints}
		payload, err := jsoniter.Marshal(resp)
		if err != nil {
			log.Println(err)
			ServerResponse(w, http.StatusInternalServerError, []byte(""))
			return
		}
		ServerResponse(w, http.StatusOK, payload)
		return
	}

	// check if receipt is stored in db
	receipt, ok := rHandler.db[id]
	if !ok {
		log.Println("GetPoints::Receipt not found")
		ServerResponse(w, http.StatusNotFound, []byte(""))
		return
	}

	// compute points
	points := CalculatePointsFromReceipt(&receipt)
	resp := pkg.GetPointsResponse{Points: points}
	payload, err := jsoniter.Marshal(resp)
	if err != nil {
		log.Println(err)
		ServerResponse(w, http.StatusInternalServerError, []byte(""))
		return
	}

	// store in cache
	rHandler.pointsCache[id] = points

	// respond with points
	ServerResponse(w, http.StatusOK, payload)
}

func (rHandler *ReceiptsHandler) ProcessReceipt(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// verify receipt
	receipt, err := ValidateReceipt(r)
	if err != nil {
		log.Println(err)
		ServerResponse(w, http.StatusBadRequest, []byte(""))
		return
	}

	// generate id for receipt
	id := uuid.NewString()
	resp := pkg.ProcessReceiptResponse{Id: id}
	payload, err := jsoniter.Marshal(resp)
	if err != nil {
		log.Println(err)
		ServerResponse(w, http.StatusInternalServerError, []byte(""))
		return
	}

	// store receipt in dv
	rHandler.dbLock.Lock()
	rHandler.db[id] = *receipt
	rHandler.dbLock.Unlock()

	// respond with id
	ServerResponse(w, http.StatusOK, payload)
}

func ValidateReceipt(r *http.Request) (*pkg.Receipt, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "ValidateReceipt::ReadAll")
	}

	var receipt pkg.Receipt
	err = jsoniter.Unmarshal(body, &receipt)
	if err != nil {
		return nil, errors.Wrap(err, "ValidateReceipt::Unmarshal")
	}

	// check retailer
	validRetailer := regexp.MustCompile(`^\S+$`)
	if !validRetailer.MatchString(receipt.Retailer) {
		return nil, errors.New("Invalid retailer")
	}

	// check purchase date
	validDate := regexp.MustCompile(`^[1,2]\d{3}-[0,1]\d-[0,1,2,3]\d$`)
	if !validDate.MatchString(receipt.PurchaseDate) {
		return nil, errors.New("Invalid date")
	}

	// check purchase time
	validTime := regexp.MustCompile(`^(0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$`) // found from stack overflow
	if !validTime.MatchString(receipt.PurchaseTime) {
		return nil, errors.New("Invalid time")
	}

	// check total
	validMoney := regexp.MustCompile(`^\d+.\d{2}$`)
	if !validMoney.MatchString(receipt.Total) {
		return nil, errors.New("Invalid total")
	}

	// check items
	validDescription := regexp.MustCompile(`^[\w\s\-]+$`)
	for _, item := range receipt.Items {
		if !validDescription.MatchString(item.ShortDescription) || !validMoney.MatchString(item.Price) {
			return nil, errors.New("invalid item")
		}
	}

	return &receipt, nil
}

func CalculatePointsFromReceipt(receipt *pkg.Receipt) int {
	points := 0

	// One point for every alphanumeric character in the retailer name.
	for _, x := range receipt.Retailer {
		if (x >= 'a' && x <= 'z') || (x >= 'A' && x <= 'Z') || (x >= '0' && x <= '9') {
			points++
		}
	}

	receiptTotalInCents, _ := strconv.ParseInt(strings.ReplaceAll(receipt.Total, ".", ""), 10, 32)
	// 50 points if the total is a round dollar amount with no cents.
	if receiptTotalInCents%100 == 0 {
		points += 50
	}

	// 25 points if the total is a multiple of 0.25.
	if receiptTotalInCents%25 == 0 {
		points += 25
	}

	// 5 points for every two items on the receipt.
	points += (len(receipt.Items) / 2) * 5

	// If the trimmed length of the item description is a multiple of 3, multiply the price by 0.2 and round up to the nearest integer. The result is the number of points earned.
	for _, item := range receipt.Items {
		if len(strings.TrimSpace(item.ShortDescription))%3 == 0 {
			itemPriceInCents, _ := strconv.ParseFloat(strings.ReplaceAll(item.Price, ".", ""), 10)
			points += int(math.Ceil(itemPriceInCents * 0.2 / 100.0))
		}
	}

	// 6 points if the day in the purchase date is odd.
	formattedDate, _ := time.Parse(DateLayout, receipt.PurchaseDate)
	if formattedDate.Month()%2 == 1 {
		points += 6
	}

	//10 points if the time of purchase is after 2:00pm and before 4:00pm.
	formattedTime, _ := time.Parse(TimeLayout, receipt.PurchaseTime)
	formattedTwoPM, _ := time.Parse(TimeLayout, TwoPM)
	formattedFourPM, _ := time.Parse(TimeLayout, FourPM)
	if formattedTime.After(formattedTwoPM) && formattedTime.Before(formattedFourPM) {
		points += 10
	}

	return points
}

func ServerResponse(w http.ResponseWriter, statusCode int, payload []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(payload)
}
