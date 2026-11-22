package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(NewServer(NewStore()))
	t.Cleanup(srv.Close)
	return srv
}

func post(t *testing.T, url, body string) *http.Response {
	t.Helper()
	resp, err := http.Post(url, "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func decode(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
}

const groceries = `{"description":"Groceries","amountCents":4550,"type":"expense","category":"Food","date":"2026-11-03"}`
const salary = `{"description":"Paycheck","amountCents":300000,"type":"income","category":"Salary","date":"2026-11-01"}`

func TestCreateAndGetTransaction(t *testing.T) {
	srv := newTestServer(t)

	resp := post(t, srv.URL+"/api/transactions", groceries)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	var created Transaction
	decode(t, resp, &created)
	if created.ID != 1 || created.AmountCents != 4550 {
		t.Errorf("created = %+v", created)
	}

	resp, err := http.Get(srv.URL + "/api/transactions/1")
	if err != nil {
		t.Fatal(err)
	}
	var got Transaction
	decode(t, resp, &got)
	if got.Description != "Groceries" {
		t.Errorf("got = %+v", got)
	}
}

func TestValidationRejectsBadInput(t *testing.T) {
	srv := newTestServer(t)

	cases := map[string]string{
		"missing description": `{"amountCents":100,"type":"expense","category":"Food","date":"2026-11-03"}`,
		"zero amount":         `{"description":"x","amountCents":0,"type":"expense","category":"Food","date":"2026-11-03"}`,
		"bad type":            `{"description":"x","amountCents":100,"type":"gift","category":"Food","date":"2026-11-03"}`,
		"bad date":            `{"description":"x","amountCents":100,"type":"expense","category":"Food","date":"03/11/2026"}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			resp := post(t, srv.URL+"/api/transactions", body)
			if resp.StatusCode != http.StatusUnprocessableEntity {
				t.Errorf("status = %d, want 422", resp.StatusCode)
			}
			resp.Body.Close()
		})
	}
}

func TestMalformedJSONReturns400(t *testing.T) {
	srv := newTestServer(t)

	resp := post(t, srv.URL+"/api/transactions", `{"description":`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestUnknownFieldReturns400(t *testing.T) {
	srv := newTestServer(t)

	resp := post(t, srv.URL+"/api/transactions", `{"description":"x","amountCents":1,"type":"expense","category":"a","date":"2026-11-03","extra":true}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestGetMissingReturns404(t *testing.T) {
	srv := newTestServer(t)

	resp, err := http.Get(srv.URL + "/api/transactions/99")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestDeleteTransaction(t *testing.T) {
	srv := newTestServer(t)
	post(t, srv.URL+"/api/transactions", groceries).Body.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/transactions/1", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204", resp.StatusCode)
	}
	resp.Body.Close()

	resp, _ = http.Get(srv.URL + "/api/transactions/1")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("after delete status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestListIsNewestFirst(t *testing.T) {
	srv := newTestServer(t)
	post(t, srv.URL+"/api/transactions", salary).Body.Close()    // 2026-11-01
	post(t, srv.URL+"/api/transactions", groceries).Body.Close() // 2026-11-03

	resp, _ := http.Get(srv.URL + "/api/transactions")
	var list []Transaction
	decode(t, resp, &list)
	if len(list) != 2 || list[0].Description != "Groceries" {
		t.Errorf("list order wrong: %+v", list)
	}
}

func TestSummary(t *testing.T) {
	srv := newTestServer(t)
	post(t, srv.URL+"/api/transactions", salary).Body.Close()
	post(t, srv.URL+"/api/transactions", groceries).Body.Close()

	resp, _ := http.Get(srv.URL + "/api/summary")
	var sum Summary
	decode(t, resp, &sum)

	if sum.IncomeCents != 300000 {
		t.Errorf("income = %d, want 300000", sum.IncomeCents)
	}
	if sum.ExpenseCents != 4550 {
		t.Errorf("expenses = %d, want 4550", sum.ExpenseCents)
	}
	if sum.BalanceCents != 295450 {
		t.Errorf("balance = %d, want 295450", sum.BalanceCents)
	}
	if sum.ExpenseByCategory["Food"] != 4550 {
		t.Errorf("Food = %d, want 4550", sum.ExpenseByCategory["Food"])
	}
}

func TestEmptySummaryHasEmptyCategoryMap(t *testing.T) {
	srv := newTestServer(t)

	resp, _ := http.Get(srv.URL + "/api/summary")
	var raw map[string]any
	decode(t, resp, &raw)
	if raw["expenseByCategory"] == nil {
		t.Error("expenseByCategory should be {} not null")
	}
}
