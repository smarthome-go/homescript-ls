package main

import (
	"time"

	"github.com/smarthome-go/homescript/v2/homescript"
)

type dummyExecutor struct{}

func (self dummyExecutor) IsAnalyzer() bool {
	return true
}

func (self dummyExecutor) ResolveModule(_ string) (string, bool, bool, error) {
	return "", true, false, nil
}

func (self dummyExecutor) Sleep(sleepTime float64) {
}

func (self dummyExecutor) Print(args ...string) error {
	return nil
}
func (self dummyExecutor) Println(args ...string) error {
	return nil
}

func (self dummyExecutor) GetSwitch(name string) (homescript.SwitchResponse, error) {
	return homescript.SwitchResponse{}, nil
}

func (self dummyExecutor) Switch(name string, power bool) error {
	return nil
}

func (self dummyExecutor) Ping(ip string, timeout float64) (bool, error) {
	return false, nil
}

func (self dummyExecutor) Notify(title string, description string, level homescript.NotificationLevel) error {
	return nil
}

func (self dummyExecutor) Remind(title string, description string, urgency homescript.ReminderUrgency, dueDate time.Time) (uint, error) {
	return 0, nil
}

func (self dummyExecutor) Log(title string, description string, level homescript.LogLevel) error {
	return nil
}

func (self dummyExecutor) Exec(id string, args map[string]string) (homescript.ExecResponse, error) {
	return homescript.ExecResponse{
		RuntimeSecs: 0.2,
		ReturnValue: homescript.ValueNull{},
	}, nil
}

func (self dummyExecutor) Get(url string) (homescript.HttpResponse, error) {
	return homescript.HttpResponse{
		Status:     "OK",
		StatusCode: 200,
		Body:       "{\"foo\": \"bar\"}",
	}, nil
}

func (self dummyExecutor) Http(url string, method string, body string, headers map[string]string) (homescript.HttpResponse, error) {
	return homescript.HttpResponse{
		Status:     "Internal Server Error",
		StatusCode: 500,
		Body:       "{\"error\": \"the server is currently running on JavaScript\"}",
	}, nil
}

func (self dummyExecutor) GetUser() string {
	return "john_doe"
}

func (self dummyExecutor) GetWeather() (homescript.Weather, error) {
	return homescript.Weather{
		WeatherTitle:       "Rain",
		WeatherDescription: "light rain",
		Temperature:        17.0,
		FeelsLike:          16.0,
		Humidity:           87,
	}, nil
}

func (self dummyExecutor) GetStorage(_ string) (*string, error) {
	return nil, nil
}

func (self dummyExecutor) SetStorage(_ string, _ string) error {
	return nil
}
