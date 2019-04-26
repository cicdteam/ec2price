package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/pricing/pricingiface"
	"encoding/json"
	"reflect"
	"strings"
)

// JSONBytesEqual compares the JSON in two byte slices.
func JSONBytesEqual(a, b []byte) bool {
    var j, j2 interface{}
    json.Unmarshal(a, &j)
    json.Unmarshal(b, &j2)
    return reflect.DeepEqual(j2, j)
}

type mockPricingClient struct {
    pricingiface.PricingAPI
}

var mockPriceData = map[string]map[string]string{}

func (m *mockPricingClient) GetProductsPages(*pricing.GetProductsInput, func(*pricing.GetProductsOutput, bool) bool) error {

	if mockPriceData["eu-mock-1"] == nil {
		mockPriceData["eu-mock-1"] = map[string]string{}
	}
	mockPriceData["eu-mock-1"]["t1.mock"] = "1.1111"
	mockPriceData["eu-mock-1"]["t2.mock"] = "1.2222"
	mockPriceData["eu-mock-1"]["t3.mock"] = "1.3333"

	if mockPriceData["eu-mock-2"] == nil {
		mockPriceData["eu-mock-2"] = map[string]string{}
	}
	mockPriceData["eu-mock-2"]["t1.mock"] = "2.1111"
	mockPriceData["eu-mock-2"]["t2.mock"] = "2.2222"
	mockPriceData["eu-mock-2"]["t3.mock"] = "2.3333"

	if mockPriceData["eu-mock-3"] == nil {
		mockPriceData["eu-mock-3"] = map[string]string{}
	}
	mockPriceData["eu-mock-3"]["t1.mock"] = "3.1111"
	mockPriceData["eu-mock-3"]["t2.mock"] = "3.2222"
	mockPriceData["eu-mock-3"]["t3.mock"] = "3.3333"

	return nil
}

func TestUsagePageHandler(t *testing.T) {

	req := httptest.NewRequest("GET", "/", nil)
	rr  := httptest.NewRecorder()
	NewRouter().ServeHTTP(rr, req)

	statusExpected := http.StatusOK
	if status := rr.Code; status != statusExpected {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, statusExpected)
	}

	ctypeExpected := "text/html; charset=UTF-8"
	if ctype := rr.Header().Get("Content-Type"); ctype != ctypeExpected {
		t.Errorf("content type header does not match: got %v want %v",
			ctype, ctypeExpected)
	}
}

func TestWrongRegion(t *testing.T) {

	req := httptest.NewRequest("GET", "/eu-wrong-1", nil)
	rr  := httptest.NewRecorder()
	NewRouter().ServeHTTP(rr, req)

	statusExpected := http.StatusNotFound
	if status := rr.Code; status != statusExpected {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, statusExpected)
	}

	ctypeExpected := "text/html; charset=UTF-8"
	if ctype := rr.Header().Get("Content-Type"); ctype != ctypeExpected {
		t.Errorf("content type header does not match: got %v want %v",
			ctype, ctypeExpected)
	}
}

func TestAnswerRegion(t *testing.T) {

	mockSvc := &mockPricingClient{}
	err := getPrices(mockSvc)
	if err != nil {
		t.Errorf(err.Error())
	}

	priceData = mockPriceData

	req := httptest.NewRequest("GET", "/eu-mock-2", nil)
	rr  := httptest.NewRecorder()
	NewRouter().ServeHTTP(rr, req)

	statusExpected := http.StatusOK
	if status := rr.Code; status != statusExpected {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, statusExpected)
	}

	ctypeExpected := "application/json; charset=UTF-8"
	if ctype := rr.Header().Get("Content-Type"); ctype != ctypeExpected {
		t.Errorf("content type header does not match: got %v want %v",
			ctype, ctypeExpected)
	}

	bodyExpected := `
	{
	    "t1.mock": "2.1111",
	    "t2.mock": "2.2222",
	    "t3.mock": "2.3333"
	}
`
	body := rr.Body.String()
	if JSONBytesEqual([]byte(body), []byte(bodyExpected)) != true {
		t.Errorf("handler returned unexpected body:\ngot\n%v\nwant\n%v",
		body, bodyExpected)
	}
}

func TestAnswerInstance(t *testing.T) {

	mockSvc := &mockPricingClient{}
	err := getPrices(mockSvc)
	if err != nil {
		t.Errorf(err.Error())
	}

	priceData = mockPriceData

	req := httptest.NewRequest("GET", "/eu-mock-2/t3.mock", nil)
	rr  := httptest.NewRecorder()
	NewRouter().ServeHTTP(rr, req)

	statusExpected := http.StatusOK
	if status := rr.Code; status != statusExpected {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, statusExpected)
	}

	ctypeExpected := "text/html; charset=UTF-8"
	if ctype := rr.Header().Get("Content-Type"); ctype != ctypeExpected {
		t.Errorf("content type header does not match: got %v want %v",
			ctype, ctypeExpected)
	}

	bodyExpected := "2.3333"
	body := strings.TrimSpace(rr.Body.String())
	if strings.Compare(body, bodyExpected) !=0 {
		t.Errorf("handler returned unexpected body:\ngot\n'%v'\nwant\n'%v'",
		body, bodyExpected)
	}
}

func TestAnswerAll(t *testing.T) {

	mockSvc := &mockPricingClient{}
	err := getPrices(mockSvc)
	if err != nil {
		t.Errorf(err.Error())
	}

	priceData = mockPriceData

	req := httptest.NewRequest("GET", "/all", nil)
	rr  := httptest.NewRecorder()
	NewRouter().ServeHTTP(rr, req)

	statusExpected := http.StatusOK
	if status := rr.Code; status != statusExpected {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, statusExpected)
	}

	ctypeExpected := "application/json; charset=UTF-8"
	if ctype := rr.Header().Get("Content-Type"); ctype != ctypeExpected {
		t.Errorf("content type header does not match: got %v want %v",
			ctype, ctypeExpected)
	}

	bodyExpected := `
	{
	    "eu-mock-1": {
	        "t1.mock": "1.1111",
	        "t2.mock": "1.2222",
	        "t3.mock": "1.3333"
	    },
	    "eu-mock-2": {
	        "t1.mock": "2.1111",
	        "t2.mock": "2.2222",
	        "t3.mock": "2.3333"
	    },
	    "eu-mock-3": {
	        "t1.mock": "3.1111",
	        "t2.mock": "3.2222",
	        "t3.mock": "3.3333"
	    }
	}
`
	body := rr.Body.String()
	if JSONBytesEqual([]byte(body), []byte(bodyExpected)) != true {
		t.Errorf("handler returned unexpected body:\ngot\n%v\nwant\n%v",
		body, bodyExpected)
	}
}
