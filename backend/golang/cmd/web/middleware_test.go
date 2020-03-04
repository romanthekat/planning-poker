package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var next = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
})

func TestAuthValidHeader(t *testing.T) {
	//given
	rr := httptest.NewRecorder()
	r := newGetRequest(t, "/", mockValidAuthHeader)

	app := newTestApplication(t)

	//when
	app.authorization(next).ServeHTTP(rr, r)

	//then
	rs := rr.Result()
	checkStatusCodeOk(t, rs)

	defer rs.Body.Close()
	body := readBody(t, rs)
	if string(body) != "OK" {
		t.Errorf("want body to equal %q", "OK")
	}
}

func TestAuthNotValidHeader(t *testing.T) {
	//given
	rr := httptest.NewRecorder()
	r := newGetRequest(t, "/", "wrong auth header")

	app := newTestApplication(t)

	//when
	app.authorization(next).ServeHTTP(rr, r)

	//then
	rs := rr.Result()
	checkStatusCodeOk(t, rs)

	defer rs.Body.Close()
	body := readBody(t, rs)
	if string(body) == "OK" {
		t.Errorf("want body to not equal %q", "OK")
	}
}

func readBody(t *testing.T, rs *http.Response) []byte {
	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func checkStatusCodeOk(t *testing.T, rs *http.Response) {
	if rs.StatusCode != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, rs.StatusCode)
	}
}
