package textbelt

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// MessageStatus type can be used to check message state
type MessageStatus string

const (
	key    = "textbelt"
	apiURL = "https://textbelt.com"

	StatusDelivered MessageStatus = "DELIVERED"
	StatusSent      MessageStatus = "SENT"
	StatusSending   MessageStatus = "SENDING"
	StatusFailed    MessageStatus = "FAILED"
	StatusUnknown   MessageStatus = "UNKNOWN"
)

// New creates the new Textbelt object executing passed options
func New(options ...func(*Textbelt)) *Textbelt {
	t := &Textbelt{
		key:     key,
		url:     apiURL,
		timeout: 5 * time.Second,
	}

	for _, opt := range options {
		opt(t)
	}

	return t
}

// Textbelt struct is the main struct using which you will interact with textbelt API
type Textbelt struct {
	key     string
	url     string
	timeout time.Duration
}

type response struct {
	Success        bool   `json:"success"`
	Status         string `json:"status"`
	ID             string `json:"textId"`
	Error          string `json:"error"`
	QuotaRemaining int    `json:"quotaRemaining"`
	OTP            string `json:"otp"`
	ValidOTP       bool   `json:"isValidOtp"`
}

// Quota returns the number of remaining amount of messages that can be sent
func (t *Textbelt) Quota() (int, error) {
	c := &http.Client{
		Timeout: t.timeout,
	}

	u := fmt.Sprintf("%s/quota/%s", t.url, t.key)
	resp, err := c.Get(u)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return -1, err
	}
	return r.QuotaRemaining, nil
}

// Status returns the message status for specific message ID
func (t *Textbelt) Status(id string) (MessageStatus, error) {
	c := &http.Client{
		Timeout: t.timeout,
	}

	u := fmt.Sprintf("%s/status/%s", t.url, id)
	resp, err := c.Get(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	return MessageStatus(r.Status), nil
}

// Send will send the message and will return the ID of the message
func (t *Textbelt) Send(phone, content string) (string, error) {
	values := url.Values{
		"phone":   {phone},
		"message": {content},
		"key":     {t.key},
	}

	c := &http.Client{
		Timeout: t.timeout,
	}

	u := t.url + "/text"

	resp, err := c.PostForm(u, values)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}

	if !r.Success {
		return "", errors.New(r.Error)
	}

	return r.ID, nil
}

// CustomOTP enables you to customize your OTP messages
type CustomOTP struct {
	Phone    string // Phone number of the receiver
	UserID   string // UserID - arbitrary ID for the generated OTP
	Message  string // Custom message, $OTP inside will hold the actual content
	Lifetime int    // How long the OTP should last
	Length   int    // Number of digits inside the OTP
}

// GenerateCustomOTP enables you to customize your OTP message by providing CustomOTP pointer
func (t *Textbelt) GenerateCustomOTP(otp *CustomOTP) (string, error) {
	values := url.Values{
		"phone":  {otp.Phone},
		"userid": {otp.UserID},
		"key":    {t.key},
	}

	if otp.Message != "" {
		values.Add("message", otp.Message)
	}

	if otp.Lifetime > 0 {
		values.Add("lifetime", strconv.Itoa(otp.Lifetime))
	}

	if otp.Length > 0 {
		values.Add("length", strconv.Itoa(otp.Length))
	}

	return t.sendOTP(values)
}

// GenerateOTP will generate the OTP and send the message to the user and will return the generated OTP
func (t *Textbelt) GenerateOTP(phone, userid string) (string, error) {
	values := url.Values{
		"phone":  {phone},
		"userid": {userid},
		"key":    {t.key},
	}

	return t.sendOTP(values)
}

// VerifyOTP checks whether the specified otp and userid are valid
func (t *Textbelt) VerifyOTP(otp, userid string) (bool, error) {
	u := fmt.Sprintf("%s/otp/verify", t.url)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return false, err
	}

	q := req.URL.Query()
	q.Add("otp", otp)
	q.Add("userid", userid)
	q.Add("key", t.key)

	req.URL.RawQuery = q.Encode()

	c := &http.Client{
		Timeout: t.timeout,
	}

	resp, err := c.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return false, err
	}

	if !r.Success {
		return false, errors.New(r.Error)
	}

	return r.ValidOTP, err
}

// WithURL enables you to pass custom textbelt API endpoint
func WithURL(url string) func(*Textbelt) {
	return func(t *Textbelt) {
		t.url = url
	}
}

// WithURL enables you to pass your own API key, otherwise free "textbelt" key will be used
func WithKey(key string) func(*Textbelt) {
	return func(t *Textbelt) {
		t.key = key
	}
}

// WithTimeout enables you to set timeout for requests, otherwise 5 seconds will be used
func WithTimeout(timeout time.Duration) func(*Textbelt) {
	return func(t *Textbelt) {
		t.timeout = timeout
	}
}

func (t *Textbelt) sendOTP(values url.Values) (string, error) {
	c := &http.Client{
		Timeout: t.timeout,
	}

	u := fmt.Sprintf("%s/otp/generate", t.url)
	resp, err := c.PostForm(u, values)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}

	if !r.Success {
		return "", errors.New(r.Error)
	}

	return r.OTP, nil
}
