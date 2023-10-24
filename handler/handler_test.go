package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/julienschmidt/httprouter"
	"github.com/psmccarty/fetch-backend-apprentice-challenge/pkg"
)

func sampleReceipt() pkg.Receipt {
	receipt := pkg.Receipt{
		Retailer:     "Target",
		PurchaseDate: "2022-04-11",
		PurchaseTime: "13:01",
		Items:        []pkg.ReceiptItem{{ShortDescription: "Gatorade", Price: "22.90"}, {ShortDescription: "Food", Price: "10.00"}},
		Total:        "32.90",
		PointsEarned: 89,
	}
	return receipt
}

func dummyProcessReceiptRequest(sampleRequestBody pkg.Receipt) *http.Request {
	body, _ := jsoniter.Marshal(sampleRequestBody)
	r, _ := http.NewRequest("POST", "", bytes.NewBuffer(body))
	return r
}

func TestNewReceiptsHandler(t *testing.T) {
	t.Run("NewReceiptsHandler()", func(t *testing.T) {
		rHandler := NewReceiptsHandler()
		if rHandler.db == nil {
			t.Error("ReceiptsHandler was not initialized correctly")
		}
	})
}

func TestValidateReceipt(t *testing.T) {

	t.Run("Empty receipt", func(t *testing.T) {
		r := dummyProcessReceiptRequest(pkg.Receipt{})
		receipt, err := ValidateReceipt(r)
		if receipt != nil || err == nil {
			t.Error("Failed on empty request")
		}
	})

	t.Run("Invalid Retailer", func(t *testing.T) {
		sReceipt := sampleReceipt()
		sReceipt.Retailer = ""
		r := dummyProcessReceiptRequest(sReceipt)
		receipt, err := ValidateReceipt(r)
		if receipt != nil || err == nil {
			t.Error("Failed on Invalid Retailer")
		}
	})

	t.Run("Invalid Purchase Date", func(t *testing.T) {
		sReceipt := sampleReceipt()
		sReceipt.PurchaseDate = "March 23 2023"
		r := dummyProcessReceiptRequest(sReceipt)
		receipt, err := ValidateReceipt(r)
		if receipt != nil || err == nil {
			t.Error("Failed on Invalid Purchase Date")
		}
	})

	t.Run("Invalid Purchase Time", func(t *testing.T) {
		sReceipt := sampleReceipt()
		sReceipt.PurchaseTime = "9:30PM"
		r := dummyProcessReceiptRequest(sReceipt)
		receipt, err := ValidateReceipt(r)
		if receipt != nil || err == nil {
			t.Error("Failed on Invalid Purchase Date")
		}
	})

	t.Run("Invalid Short Description", func(t *testing.T) {
		sReceipt := sampleReceipt()
		sReceipt.Items[0].ShortDescription = "***"
		r := dummyProcessReceiptRequest(sReceipt)
		receipt, err := ValidateReceipt(r)
		if receipt != nil || err == nil {
			t.Error("Failed on Invalid Short Description")
		}
	})

	t.Run("Invalid Item Price", func(t *testing.T) {
		sReceipt := sampleReceipt()
		sReceipt.Items[0].Price = "$4"
		r := dummyProcessReceiptRequest(sReceipt)
		receipt, err := ValidateReceipt(r)
		if receipt != nil || err == nil {
			t.Error("Failed on Invalid Item Price")
		}
	})

	t.Run("Invalid Total", func(t *testing.T) {
		sReceipt := sampleReceipt()
		sReceipt.Total = "12.103"
		r := dummyProcessReceiptRequest(sReceipt)
		receipt, err := ValidateReceipt(r)
		if receipt != nil || err == nil {
			t.Error("Failed on Invalid Total")
		}
	})

	t.Run("Valid Receipt", func(t *testing.T) {
		sReceipt := sampleReceipt()
		r := dummyProcessReceiptRequest(sReceipt)
		receipt, err := ValidateReceipt(r)
		if receipt == nil || err != nil || !reflect.DeepEqual(sReceipt, *receipt) {
			t.Error("Failed on Valid Receipt")
		}
	})
}

func TestCalculatePointsFromReceipt(t *testing.T) {
	t.Run("Some points earned", func(t *testing.T) {
		sReceipt := sampleReceipt()
		points := CalculatePointsFromReceipt(&sReceipt)
		if points != 17 {
			t.Error("Total points value incorrect")
		}
	})

	t.Run("Points earned for every criteria", func(t *testing.T) {
		sReceipt := sampleReceipt()
		sReceipt.PurchaseDate = "2022-01-01"
		sReceipt.PurchaseTime = "15:00"
		sReceipt.Items[0].ShortDescription = "abc"
		sReceipt.Items[1].ShortDescription = "abcdef"
		sReceipt.Total = "1.00"
		points := CalculatePointsFromReceipt(&sReceipt)
		if points != 109 {
			t.Error("Total points value incorrect")
		}
	})
}

func TestGetPoints(t *testing.T) {

	t.Run("No params", func(t *testing.T) {
		rHandler := NewReceiptsHandler()
		r, _ := http.NewRequest("GET", "receipts//points", bytes.NewBuffer([]byte("")))
		w := httptest.NewRecorder()
		rHandler.GetPoints(w, r, nil)
		if w.Result().StatusCode != http.StatusBadRequest {
			t.Error("Wrong status code")
		}
	})

	t.Run("Receipt not found", func(t *testing.T) {
		rHandler := NewReceiptsHandler()
		r, _ := http.NewRequest("GET", "receipts/abc/points", bytes.NewBuffer([]byte("")))
		w := httptest.NewRecorder()
		ps := httprouter.Params{httprouter.Param{Key: "id", Value: "abc"}}
		rHandler.GetPoints(w, r, ps)
		if w.Result().StatusCode != http.StatusNotFound {
			t.Error("Wrong status code")
		}
	})

	t.Run("Receipt found", func(t *testing.T) {
		rHandler := NewReceiptsHandler()
		rHandler.db["abc"] = sampleReceipt()
		r, _ := http.NewRequest("GET", "receipts/abc/points", bytes.NewBuffer([]byte("")))
		w := httptest.NewRecorder()
		ps := httprouter.Params{httprouter.Param{Key: "id", Value: "abc"}}
		rHandler.GetPoints(w, r, ps)
		if w.Result().StatusCode != http.StatusOK {
			t.Error("Wrong status code")
		}
	})

}

func TestProcessReceipt(t *testing.T) {
	t.Run("Invalid Receipt", func(t *testing.T) {
		rHandler := NewReceiptsHandler()
		r, _ := http.NewRequest("POST", "receipts/process", bytes.NewBuffer([]byte("")))
		w := httptest.NewRecorder()
		rHandler.ProcessReceipt(w, r, nil)
		if w.Result().StatusCode != http.StatusBadRequest {
			t.Error("Wrong status code")
		}
		if len(rHandler.db) != 0 {
			t.Error("Invalid receipt was added")
		}
	})

	t.Run("Valid Receipt", func(t *testing.T) {
		rHandler := NewReceiptsHandler()
		body, _ := jsoniter.Marshal(sampleReceipt())
		r, _ := http.NewRequest("POST", "receipts/process", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		rHandler.ProcessReceipt(w, r, nil)
		if w.Result().StatusCode != http.StatusOK {
			t.Error("Wrong status code")
		}
		if len(rHandler.db) != 1 {
			t.Error("Valid receipt was not stored")
		}
	})
}
