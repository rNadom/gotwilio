package gotwilio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const IPMsgURL = "https://ip-messaging.twilio.com/v1"
const IPMsgChannelTypePublic = "public"
const IPMsgChannelTypePrivate = "private"

type IPMessageChannel struct {
	Sid          string `json:"sid"`
	AccountSid   string `json:"account_sid"`
	ServiceSid   string `json:"service_sid"`
	FriendlyName string `json:"friendly_name"`
	Attributes   string `json:"attributes"`
	Type         string `json:"type"`
	DateCreated  string `json:"date_created"`
	DateUpdated  string `json:"date_created"`
	CreatedBy    string `json:"created_by"`
	Url          string `json:"url"`
}

type ListIPMsgChanResp struct {
	Meta     IPMessageMeta      `json:"meta"`
	Channels []IPMessageChannel `json:"channels"`
}

type ListIPMsgResp struct {
	Meta     IPMessageMeta `json:"meta"`
	Messages []IPMessage   `json:"messages"`
}

type IPMessageMeta struct {
	Page            int64  `json:"page"`
	PageSize        int64  `json:"page_size"`
	FirstPageUrl    string `json:"first_page_url"`
	PreviousPageUrl string `json:"previous_page_url"`
	Url             string `json:"url"`
	NextPageUrl     string `json:"next_page_url"`
	Key             string `json:"key"`
}

type IPMessage struct {
	Sid         string `json:"sid"`
	AccountSid  string `json:"account_sid"`
	ServiceSid  string `json:"service_sid"`
	To          string `json:"to"`
	DateCreated string `json:"date_created"`
	DateUpdated string `json:"date_updated"`
	WasEdited   bool   `json:"was_edited"`
	From        string `json:"from"`
	Body        string `json:"body"`
	Url         string `json:"url"`
}

func (twilio *Twilio) ListIPMessageChannels(serviceSid string) ([]IPMessageChannel, error) {
	twilioUrl := fmt.Sprintf(
		"%s/Services/%s/Channels",
		IPMsgURL,
		serviceSid,
	)
	resp := ListIPMsgChanResp{}

	_, err := twilio.sendIPMsgRequest(nil, twilioUrl, "GET", &resp)
	return resp.Channels, err
}

// https://www.twilio.com/docs/api/ip-messaging/rest/channels#action-create
func (twilio *Twilio) CreateIPMessageChannel(serviceSid, channelType, friendlyName, uniqueName string) (string, error) {
	twilioUrl := fmt.Sprintf(
		"%s/Services/%s/Channels",
		IPMsgURL,
		serviceSid,
	)
	resp := IPMessageChannel{}
	formValues := url.Values{}
	if friendlyName != "" {
		formValues.Set("FriendlyName", friendlyName)
	}
	if uniqueName != "" {
		formValues.Set("UniqueName", uniqueName)
	}
	formValues.Set("Type", channelType)
	_, err := twilio.sendIPMsgRequest(formValues, twilioUrl, "POST", &resp)
	return resp.Sid, err
}

func (twilio *Twilio) DeleteIPMessageChannel(serviceSid, channelSid string) error {
	twilioUrl := fmt.Sprintf(
		"%s/Services/%s/Channels/%s",
		IPMsgURL,
		serviceSid,
		channelSid,
	)

	res, err := twilio.delete(twilioUrl)
	if err != nil {
		return err
	}
	if res.StatusCode != 204 {
		return errors.New("failed deleting twilio channel")
	}
	return nil
}

// https://www.twilio.com/docs/api/ip-messaging/rest/messages#action-create
func (twilio *Twilio) SendIPMessage(serviceSid, channelSid, memberSid, body, from string) error {
	twilioUrl := fmt.Sprintf(
		"%s/Services/%s/Channels/%s/Messages",
		IPMsgURL,
		serviceSid,
		channelSid,
	)
	resp := IPMessage{}
	formValues := url.Values{}
	formValues.Set("Body", body)
	formValues.Set("From", from)
	_, err := twilio.sendIPMsgRequest(formValues, twilioUrl, "POST", &resp)

	return err
}

// https://www.twilio.com/docs/api/ip-messaging/rest/messages#action-list
func (twilio *Twilio) ListIPMessages(serviceSid, channelSid string) ([]IPMessage, error) {
	twilioUrl := fmt.Sprintf(
		"%s/Services/%s/Channels/%s/Messages",
		IPMsgURL,
		serviceSid,
		channelSid,
	)
	resp := ListIPMsgResp{}
	_, err := twilio.sendIPMsgRequest(nil, twilioUrl, "GET", &resp)

	return resp.Messages, err
}

func (twilio *Twilio) sendIPMsgRequest(formValues url.Values, twilioUrl, method string, respStruct interface{}) (exception *Exception, err error) {
	var res *http.Response

	if method == "POST" {
		res, err = twilio.post(formValues, twilioUrl)
	} else if method == "GET" {
		res, err = twilio.get(twilioUrl)
	}

	if err != nil {
		return exception, err
	}
	defer res.Body.Close()

	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return exception, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		exception = new(Exception)
		err = json.Unmarshal(responseBody, exception)

		// We aren't checking the error because we don't actually care.
		// It's going to be passed to the client either way.
		return exception, err
	}

	err = json.Unmarshal(responseBody, respStruct)
	return exception, err
}
