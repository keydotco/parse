package parse

import (
	"encoding/json"
	"errors"
	"net/url"
	"path"
	"reflect"
)

type Params map[string]interface{}

func CallFunction(client *ParseClient, name string, params Params, resp interface{}) error {
	return callFn(client, name, params, resp, nil)
}

type callFnT struct {
	name           string
	params         Params
	currentSession *sessionT
}

func (c *callFnT) method() string {
	return "POST"
}

func (c *callFnT) endpoint(client *ParseClient) (string, error) {
	u := url.URL{}
	if client.isHosted() {
		u.Path = path.Join(client.parseMountPoint, "functions", c.name)
	} else {
		u.Path = path.Join(client.version, "functions", c.name)
	}
	u.Scheme = client.parseScheme
	u.Host = client.parseHost

	return u.String(), nil
}

func (c *callFnT) body() (string, error) {
	b, err := json.Marshal(c.params)
	return string(b), err
}

func (c *callFnT) useMasterKey() bool {
	return false
}

func (c *callFnT) session() *sessionT {
	return c.currentSession
}

func (c *callFnT) contentType() string {
	return "application/json"
}

type fnRespT struct {
	Result interface{} `parse:"result"`
}

func callFn(client *ParseClient, name string, params Params, resp interface{}, currentSession *sessionT) error {
	rv := reflect.ValueOf(resp)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("resp must be a non-nil pointer")
	}

	if params == nil {
		params = Params{}
	}

	cr := &callFnT{
		name:           name,
		params:         params,
		currentSession: currentSession,
	}
	if b, err := client.doRequest(cr); err != nil {
		return err
	} else {
		r := fnRespT{}
		if err := json.Unmarshal(b, &r); err != nil {
			return err
		}
		return populateValue(resp, r.Result)
	}
}
