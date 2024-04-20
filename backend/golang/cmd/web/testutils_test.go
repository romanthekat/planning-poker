package main

//
//import (
//	"github.com/romanthekat/FaceRecognitionBackend/pkg/models/mock"
//	"github.com/romanthekat/FaceRecognitionBackend/pkg/services"
//	"io/ioutil"
//	"log"
//	"net/http"
//	"testing"
//)
//
//const mockMlEndpoint = "localhost:4242/ml"
//const mockValidAuthHeader = "test"
//
//func newTestApplication(t *testing.T) *Application {
//	PersonModel := &mock.PersonModel{}
//	return &Application{
//		errorLog:           log.New(ioutil.Discard, "", 0),
//		infoLog:            log.New(ioutil.Discard, "", 0),
//		persons:            PersonModel,
//		encodingComparator: services.NewEncodingComparator(PersonModel),
//		validAuthHeader:    mockValidAuthHeader,
//		mlEndpoint:         mockMlEndpoint,
//	}
//}
//
//func newGetRequest(t *testing.T, url, authHeader string) *http.Request {
//	r, err := http.NewRequest("GET", url, nil)
//	r.Header.Set("Authorization", authHeader)
//	if err != nil {
//		t.Fatal(err)
//	}
//	return r
//}
