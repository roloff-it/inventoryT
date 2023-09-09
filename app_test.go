package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

// tests

var a App

func TestMain(m *testing.M) {
	err := a.Initialize(DBUser, DBPassword, "test")
	if err != nil {
		log.Fatal("Error occured while initialising the database")
	}
	createTable()
	m.Run()
}

func createTable() {
	// `` for multiline string
	createTableQuery := `CREATE TABLE IF NOT EXISTS products ( 
		id int NOT NULL AUTO_INCREMENT,
		name varchar(255) NOT NULL,
		quantity int,
		price float(10,7),
		PRIMARY KEY (id)
	);`

	_, err := a.DB.Exec(createTableQuery)
	if err != nil {
		log.Fatal("Can't create DB table", err)
	}
}

func clearTable() {
	// Delete all records
	_, err := a.DB.Exec("DELETE FROM products")
	if err != nil {
		log.Fatal("Can't delete all records from table", err)
	}
	_, err = a.DB.Exec("ALTER TABLE products AUTO_INCREMENT=1")
	if err != nil {
		log.Fatal("AUTO_INCREMENT cant set again to 1", err)
	}
}

func addProduct(name string, quantity int, price float64) {
	query := fmt.Sprintf(`INSERT INTO products(name, quantity, price) 
						   VALUES ('%v', %v, %v)`, name, quantity, price)
	_, err := a.DB.Exec(query)
	if err != nil {
		log.Fatal("Can't add product to table", err)
	}
}

func TestGetProduct(t *testing.T) {
	clearTable()
	addProduct("Iluminati", 17, 23.50)
	request, _ := http.NewRequest("GET", "/product/1", nil)
	respone := sendRequest(request)
	checkStatusCode(t, http.StatusOK, respone.Code)
}

func checkStatusCode(t *testing.T, expectedStatusCode int, actualStatusCode int) {
	if expectedStatusCode != actualStatusCode {
		t.Errorf("Expected status: %v, Received: %v", expectedStatusCode, actualStatusCode)
	}
}

func sendRequest(request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	a.Router.ServeHTTP(recorder, request)
	return recorder
}

func TestCreateProduct(t *testing.T) {
	clearTable()
	var product = []byte(`{"name":"jam", "quantity":5, "price":400.89}`)
	req, err := http.NewRequest("POST", "/product", bytes.NewBuffer(product))
	if err != nil {
		log.Fatal("Cant create Product with POST", err)
	}
	req.Header.Set("Content-Type", "application/json")
	response := sendRequest(req)
	checkStatusCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	// log.Printf("Type %T", m["name"])
	if m["name"] != "jam" {
		t.Errorf("CreateProduct - Expected name: %v, got: %v", "jam", m["name"])
	}
	// log.Printf("Type %T", m["quantity"])
	if m["quantity"] != float64(5) {
		t.Errorf("CreateProduct - Expected quantity: %v, got: %v", 5, m["quantity"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProduct("cia", 100, 20)

	req, err := http.NewRequest("GET", "/product/1", nil)
	if err != nil {
		log.Fatal("TestDeleteProduct: Error by request to create a product", err)
	}
	response := sendRequest(req)
	checkStatusCode(t, http.StatusOK, response.Code)

	req, err = http.NewRequest("DELETE", "/product/1", nil)
	if err != nil {
		log.Fatal("TestDeleteProduct: Error by request to delete the product", err)
	}
	response = sendRequest(req)
	checkStatusCode(t, http.StatusOK, response.Code)
}

func TestUpdateProduct(t *testing.T) {
	clearTable()
	addProduct("discord", 150, 900)

	req, err := http.NewRequest("GET", "/product/1", nil)
	if err != nil {
		log.Fatal("TestDeleteProduct: Error by request to create a product", err)
	}
	response := sendRequest(req)
	checkStatusCode(t, http.StatusOK, response.Code)

	var oldValue map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &oldValue)

	var product = []byte(`{"name":"discord", "quantity":500, "price":900}`)
	req, err = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(product))
	if err != nil {
		log.Fatal("Cant update Product with PUT", err)
	}
	req.Header.Set("Content-Type", "application/json")
	response = sendRequest(req)

	var newValue map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &newValue)

	if oldValue["id"] != newValue["id"] {
		t.Errorf("Expected id: %v, Got: %v", newValue["id"], oldValue["id"])
	}

	if oldValue["name"] != newValue["name"] {
		t.Errorf("Expected name: %v, Got: %v", newValue["name"], oldValue["name"])
	}

	if oldValue["quantity"] == newValue["quantity"] {
		t.Errorf("Expected not equal quantity: %v, Got: %v", 500, newValue["quantity"])
	}

	if oldValue["price"] != newValue["price"] {
		t.Errorf("Expected quantity: %v, Got: %v", oldValue["price"], newValue["price"])
	}
}
