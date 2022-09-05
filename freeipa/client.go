/*
This package is UNMAINTANED!!!
See https://github.com/tehwalris/go-freeipa for details

Package freeipa provides a client for the FreeIPA API.

It provides access to almost all methods available through the API.
Every API method has generated go structs for request parameters and output.

This code is generated from a schema which was queried from a FreeIPA
server using its "schema" method. This client performs basic response validation.
Since the FreeIPA server does not always conform to its own schema, it can
happen that this libary fails to unmarshal a response from FreeIPA.
If you run into that, please open an issue for this client library.
With that said, this is still the most extensive golang FreeIPA client
and it's probably easier to fix those issues here than to write
a new client from scratch.

Since FreeIPA cares about the presence or abscence of fields in requests,
all optional fields are defined as pointers. There are utility functions like
freeipa.String to make filling these less painful.

The client uses FreeIPA's JSON-RPC interface with username/password authentication.
There is no support for connecting to FreeIPA with Kerberos authentication.
There is currently no support for batched requests.

See https://github.com/tehwalris/go-freeipa/blob/master/developing.md
for information on how this library is generated.
*/
package freeipa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	k5client "github.com/jcmturner/gokrb5/v8/client"
	k5config "github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/keytab"
	"github.com/jcmturner/gokrb5/v8/spnego"
	"github.com/pkg/errors"
)

// Client holds a connection to a FreeIPA server.
type Client struct {
	host     string
	hc       *http.Client
	user     string
	pw       string
	k5client *k5client.Client
}

// Error is an error returned by the FreeIPA server in a JSON response.
type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Name    string `json:"name"`
}

func (t *Error) Error() string {
	return fmt.Sprintf("%v (%v): %v", t.Name, t.Code, t.Message)
}

// Connect connects to the FreeIPA server and performs an initial login.
func Connect(host string, tspt *http.Transport, user, pw string) (*Client, error) {
	jar, e := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: nil, // this should be fine, since we only use one server
	})
	if e != nil {
		return nil, e
	}
	c := &Client{
		host: host,
		hc: &http.Client{
			Transport: tspt,
			Jar:       jar,
		},
		user: user,
		pw:   pw,
	}
	if e := c.login(); e != nil {
		return nil, fmt.Errorf("initial login falied: %v", e)
	}
	return c, nil
}

func ConnectWithKerberos(host string, tspt *http.Transport, k5ConnectOpts *KerberosConnectOptions) (*Client, error) {
	jar, e := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: nil, // this should be fine, since we only use one server
	})
	if e != nil {
		return nil, e
	}

	krb5Config, err := k5config.NewFromReader(k5ConnectOpts.Krb5ConfigReader)
	if err != nil {
		return nil, errors.WithMessage(err, "reading kerberos configuration")
	}

	ktBytes, err := ioutil.ReadAll(k5ConnectOpts.KeytabReader)
	if err != nil {
		return nil, errors.WithMessage(err, "reading keytab")
	}

	kt := keytab.New()
	if err := kt.Unmarshal(ktBytes); err != nil {
		return nil, errors.WithMessage(err, "parsing keytab")
	}

	k5client := k5client.NewWithKeytab(k5ConnectOpts.Username, k5ConnectOpts.Realm, kt, krb5Config)

	c := &Client{
		host: host,
		hc: &http.Client{
			Transport: tspt,
			Jar:       jar,
		},
		user:     k5ConnectOpts.Username,
		k5client: k5client,
	}
	if e := c.login(); e != nil {
		return nil, fmt.Errorf("initial login falied: %v", e)
	}
	return c, nil
}

func (c *Client) exec(req *request) (io.ReadCloser, error) {
	res, e := c.sendRequest(req)
	if e != nil {
		return nil, e
	}

	if res.StatusCode == http.StatusUnauthorized {
		res.Body.Close()
		if e := c.login(); e != nil {
			return nil, fmt.Errorf("renewed login failed: %v", e)
		}
		res, e = c.sendRequest(req)
		if e != nil {
			return nil, e
		}
	}

	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, fmt.Errorf("unexpected http status code: %v", res.StatusCode)
	}
	return res.Body, nil
}

func (c *Client) login() error {
	if c.k5client != nil {
		return c.loginWithKerberos()
	}

	data := url.Values{
		"user":     []string{c.user},
		"password": []string{c.pw},
	}
	res, e := c.hc.PostForm(fmt.Sprintf("https://%v/ipa/session/login_password", c.host), data)
	if e != nil {
		return e
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http status code: %v", res.StatusCode)
	}
	return nil
}

func (c *Client) loginWithKerberos() error {

	k5LoginEndpoint := fmt.Sprintf("https://%s/ipa/session/login_kerberos", c.host)
	spnegoCl := spnego.NewClient(c.k5client, c.hc, "")

	req, err := http.NewRequest("POST", k5LoginEndpoint, nil)
	if err != nil {
		return errors.WithMessage(err, "building login HTTP request")
	}

	req.Header.Add("Referer", fmt.Sprintf("https://%s/ipa", c.host))

	res, err := spnegoCl.Do(req)
	if err != nil {
		return errors.Wrap(err, "logging in using Kerberos")
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http status code: %v", res.StatusCode)
	}

	return nil
}

func (c *Client) sendRequest(req *request) (*http.Response, error) {
	reqB, e := json.Marshal(req)
	if e != nil {
		return nil, e
	}
	reqH, e := http.NewRequest("POST", fmt.Sprintf("https://%v/ipa/session/json", c.host), bytes.NewBuffer(reqB))
	if e != nil {
		return nil, e
	}
	reqH.Header.Set("Content-Type", "application/json")
	reqH.Header.Set("Accept", "application/json")
	reqH.Header.Set("Referer", fmt.Sprintf("https://%v/ipa/ui", c.host))
	return c.hc.Do(reqH)
}
