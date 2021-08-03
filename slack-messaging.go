package slack_messaging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type SlackChan struct {
	token          string
	comChan        chan string
	slackChannelId string
	deliveryChan   chan string
	errorChan      chan string
}

type SlackMsg struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type SlackResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

func NewSlackChan(token string, comChan chan string, slackChannelId string, deliveryChan chan string, errorChan chan string) *SlackChan {
	s := &SlackChan{
		token:          token,
		comChan:        comChan,
		slackChannelId: slackChannelId,
		deliveryChan:   deliveryChan,
		errorChan:      errorChan,
	}

	go s.slackSend()

	return s
}

func (s *SlackChan) SendMsg(msg string) {
	fmt.Println(s.comChan)
	s.comChan <- msg
}

func (s *SlackChan) slackSend() {
	fmt.Println(s.comChan)
	for msg := range s.comChan {
		fmt.Println("Sending Message", msg)
		s_msg := &SlackMsg{
			Channel: s.slackChannelId,
			Text:    msg,
		}
		bmsg, _ := json.Marshal(s_msg)
		fmt.Println(string(bmsg))
		req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(bmsg))
		if err != nil {
			panic(err)
		}
		req.Header.Set("Authorization", "Bearer "+s.token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			if s.errorChan != nil {
				s.deliveryChan <- "Error in sending Slack Message: " + err.Error()
			}
		}

		body, _ := ioutil.ReadAll(resp.Body)

		sr := &SlackResponse{}
		errM := json.Unmarshal(body, sr)
		if errM != nil {
			fmt.Println(errM)

		}
		fmt.Println(sr)
		if s.deliveryChan != nil && sr.Ok {
			s.deliveryChan <- "Slack Messaging Delivered: " + msg
		}

		if s.errorChan != nil && !sr.Ok {
			s.deliveryChan <- "Error in Slack Messaging: " + sr.Error
		}
	}
}
