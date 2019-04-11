package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/mock"
)

type MockedClient struct {
	mock.Mock
}

func init() {

	rocketChatMock := new(MockedClient)
	rocketChatClient = rocketChatMock

	rocketChatMock.On("GetChannelId", "prometheus-test-room").Return("test123")
	channel := &models.Channel{ID: "test123"}
	text := "**Firing: Oops, something happened!**\n**description**: \n**alertname**: something_happened\n**env**: prod\n**instance**: server01.int:9100\n**job**: node\n**service**: prometheus_bot\n**severity**: warning\n**supervisor**: runit\n"
	message := &models.Message{ID: "123", RoomID: channel.ID, Msg: text}
	rocketChatMock.On("SendMessage", channel, text).Return(message)

	*configFile = "config/rocketchat.yml"
	config = loadConfig(*configFile)
	user := &models.User{ID: "123", Name: "prometheus"}
	rocketChatMock.On("Login", config).Return(user)

}

func TestReadRequestBodyOk(t *testing.T) {

	// Load a simple example of a body coming from AlertManager
	data, err := ioutil.ReadFile("test_param.json")
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", bytes.NewReader(data))

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()

	dataReq, _ := readRequestBody(req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v, want %v", status, http.StatusOK)
	}

	// Check the returned data
	if reflect.DeepEqual(template.Data{}, dataReq) {
		t.Error("Struct shouldn't be empty")
	}
}

func TestReadRequestBodyError(t *testing.T) {
	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", nil)

	dataReq, err := readRequestBody(req)

	// Check the returned data
	if !reflect.DeepEqual(template.Data{}, dataReq) {
		t.Error("Struct should be empty")
	}

	// Check the response body
	expected := "EOF"
	if err.Error() != expected {
		t.Errorf("Unexpected body: got %v, want %v", err.Error(), expected)
	}
}

func TestWebhookHandler(t *testing.T) {

	// Load a simple example of a body coming from AlertManager
	data, err := ioutil.ReadFile("test_param.json")
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", bytes.NewReader(data))

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhook)

	// Test the handler with the request and record the result
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v, want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := `{"Status":200,"Message":"Success"}`
	if rr.Body.String() != expected {
		t.Errorf("Unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func TestWebhookHandlerError(t *testing.T) {
	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", nil)

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhook)

	// Test the handler with the request and record the result
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Wrong status code: got %v, want %v", status, http.StatusBadRequest)
	}

	// Check the response body
	expected := `{"Status":400,"Message":"EOF"}`
	if rr.Body.String() != expected {
		t.Errorf("Unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func (mock *MockedClient) GetChannelId(channelName string) (string, error) {
	args := mock.Called(channelName)
	return args.String(0), nil
}

func (mock *MockedClient) SendMessage(channel *models.Channel, text string) (*models.Message, error) {
	args := mock.Called(channel, text)
	return args.Get(0).(*models.Message), nil
}

func (mock *MockedClient) Login(credentials *models.UserCredentials) (*models.User, error) {
	args := mock.Called(&config.Credentials)
	return args.Get(0).(*models.User), nil
}
