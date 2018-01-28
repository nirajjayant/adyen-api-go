// Package adyen is Adyen API Library for GO
package adyen

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/go-querystring/query"
)

// DefaultCurrency default currency for transactions
const DefaultCurrency = "EUR"

// Version of a current Adyen API
const (
	APIVersion = "v25"
)

// Enpoint service to use
const (
	PaymentService   = "Payment"
	RecurringService = "Recurring"
)

// New - creates Adyen instance
//
// Description:
//
//     - env - Environment for next API calls
//     - username - API username for authentication
//     - password - API password for authentication
//
// You can create new API user there: https://ca-test.adyen.com/ca/ca/config/users.shtml
func New(env environment, username, password string) *Adyen {
	return &Adyen{
		Credentials: newCredentials(env, username, password),
		Currency:    DefaultCurrency,
	}
}

// NewWithHMAC - create new Adyen instance with HPP credentials
//
// Use this constructor when you need to use Adyen HPP API.
//
// Description:
//
//     - env - Environment for next API calls
//     - username - API username for authentication
//     - password - API password for authentication
//     - hmac - is generated when new Skin is created in Adyen Customer Area
//
// New skin can be created there https://ca-test.adyen.com/ca/ca/skin/skins.shtml
func NewWithHMAC(env environment, username, password, hmac string) *Adyen {
	return &Adyen{
		Credentials: newCredentialsWithHMAC(env, username, password, hmac),
		Currency:    DefaultCurrency,
	}
}

// Adyen - base structure with configuration options
type Adyen struct {
	Credentials apiCredentials
	Currency    string
	Logger      *log.Logger
}

// ClientURL - returns URl, that need to loaded in UI, to encrypt Credit Card information
//
//           - clientID - Used to load external JS files from Adyen, to encrypt client requests
func (a *Adyen) ClientURL(clientID string) string {
	return a.Credentials.Env.ClientURL(clientID)
}

// adyenURL returns Adyen backend URL
func (a *Adyen) adyenURL(service string, requestType string) string {
	return a.Credentials.Env.BaseURL(service, APIVersion) + "/" + requestType + "/"
}

// createHPPUrl returns Adyen HPP url
func (a *Adyen) createHPPUrl(requestType string) string {
	return a.Credentials.Env.HppURL(requestType)
}

// AttachLogger attach logger to API instance
//
// Example:
//
//  instance := adyen.New(....)
//  Logger = log.New(os.Stdout, "Adyen Playground: ", log.Ldate|log.Ltime|log.Lshortfile)
//  instance.AttachLogger(Logger)
func (a *Adyen) AttachLogger(Logger *log.Logger) {
	a.Logger = Logger
}

// SetCurrency set default currency for transactions
func (a *Adyen) SetCurrency(currency string) {
	a.Currency = currency
}

// execute request on Adyen side, transforms "requestEntity" into JSON representation
func (a *Adyen) execute(service string, method string, requestEntity interface{}) (*Response, error) {
	body, err := json.Marshal(requestEntity)
	if err != nil {
		return nil, err
	}

	url := a.adyenURL(service, method)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if a.Logger != nil {
		a.Logger.Printf("[Request]: %s %s\n%s", method, url, body)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(a.Credentials.Username, a.Credentials.Password)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer func() {
		err = resp.Body.Close()
	}()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)

	if err != nil {
		return nil, err
	}

	if a.Logger != nil {
		a.Logger.Printf("[Response]: %s %s\n%s", method, url, buf.String())
	}

	providerResponse := &Response{
		Response: resp,
		Body:     buf.Bytes(),
	}

	err = providerResponse.handleHTTPError()

	if err != nil {
		return nil, err
	}

	return providerResponse, nil
}

// executeHpp - execute request without authorization to Adyen Hosted Payment API
func (a *Adyen) executeHpp(method string, requestEntity interface{}) (*Response, error) {
	url := a.createHPPUrl(method)

	v, _ := query.Values(requestEntity)
	url = url + "?" + v.Encode()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if a.Logger != nil {
		a.Logger.Printf("[Request]: %s %s", method, url)
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer func() {
		err = resp.Body.Close()
	}()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)

	if err != nil {
		return nil, err
	}

	if a.Logger != nil {
		a.Logger.Printf("[Response]: %s %s\n%s", method, url, buf.String())
	}

	providerResponse := &Response{
		Response: resp,
		Body:     buf.Bytes(),
	}

	return providerResponse, nil
}

// Payment - returns PaymentGateway
func (a *Adyen) Payment() *PaymentGateway {
	return &PaymentGateway{a}
}

// Modification - returns ModificationGateway
func (a *Adyen) Modification() *ModificationGateway {
	return &ModificationGateway{a}
}

// Recurring - returns RecurringGateway
func (a *Adyen) Recurring() *RecurringGateway {
	return &RecurringGateway{a}
}
