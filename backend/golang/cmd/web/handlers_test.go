package main

import (
	"encoding/json"
	"github.com/romanthekat/FaceRecognitionBackend/pkg/models"
	"github.com/romanthekat/FaceRecognitionBackend/pkg/models/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPerson(t *testing.T) {
	//given
	app := newTestApplication(t)

	ts := httptest.NewServer(app.routes())
	defer ts.Close()

	r := newGetRequest(t, ts.URL+"/person/get?id=1", app.validAuthHeader)

	//when
	rs, err := ts.Client().Do(r)
	if err != nil {
		t.Fatal(err)
	}

	//then
	if rs.StatusCode != http.StatusOK {
		t.Fatalf("want %d; got %d", http.StatusOK, rs.StatusCode)
	}

	var person *models.Person

	err = json.NewDecoder(rs.Body).Decode(&person)
	if err != nil {
		t.Fatal(err)
	}

	if person.ID != mock.Person.ID ||
		person.FirstName != mock.Person.FirstName ||
		person.LastName != mock.Person.LastName ||
		person.Email != mock.Person.Email {
		t.Errorf("want body contains json with mock person, got %q", person)
	}

}
func TestGetAllPerson(t *testing.T) {
	//given
	app := newTestApplication(t)

	ts := httptest.NewServer(app.routes())
	defer ts.Close()

	r := newGetRequest(t, ts.URL+"/person/all", app.validAuthHeader)

	//when
	rs, err := ts.Client().Do(r)
	if err != nil {
		t.Fatal(err)
	}

	//then
	if rs.StatusCode != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, rs.StatusCode)
	}

	var persons []*models.Person

	err = json.NewDecoder(rs.Body).Decode(&persons)
	if err != nil {
		t.Fatal(err)
	}

	if len(persons) == 0 {
		t.Fatalf("persons must be provided, found none")
	}

	for _, person := range persons {
		if person.ID != mock.Person.ID ||
			person.FirstName != mock.Person.FirstName ||
			person.LastName != mock.Person.LastName ||
			person.Email != mock.Person.Email {
			t.Errorf("want body contains json with mock person, got %q", person)
		}
	}
}

func TestGetEncodingStringByMlResponse(t *testing.T) {
	response := []byte("[[1 2 3]]")

	encodingString := getEncodingStringByMlResponse(response)

	const correctEncodingString = "1 2 3"
	if encodingString != correctEncodingString {
		t.Errorf("want '%q', got %q", correctEncodingString, encodingString)
	}
}
