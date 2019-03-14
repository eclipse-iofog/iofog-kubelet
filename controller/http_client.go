package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/cenkalti/backoff"
	"github.com/cpuguy83/strongerrors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type HttpClient struct {
	controllerEndpoint *url.URL
	controllerToken    string
	client             *http.Client
}

func NewHttpClient(controllerToken, controllerUrl string) (*HttpClient, error) {
	var client HttpClient

	client.client = &http.Client{}
	client.controllerToken = controllerToken

	epurl, err := url.Parse(controllerUrl)
	if err != nil {
		return nil, err
	}
	client.controllerEndpoint = epurl

	return &client, nil
}

func (p *HttpClient) DoGetRequest(urlPathStr string, v interface{}) error {
	response, err := p.DoGetRequestBytes(urlPathStr)
	if err != nil {
		return err
	}

	return json.Unmarshal(response, &v)
}

func (p *HttpClient) DoGetRequestBytes(urlPathStr string) ([]byte, error) {
	urlPath, err := url.Parse(urlPathStr)
	if err != nil {
		return nil, err
	}

	return p.DoRequest("GET", urlPath, nil, true)
}

func (p *HttpClient) DoRequest(method string, urlPath *url.URL, body []byte, readResponse bool) ([]byte, error) {
	requestURL := p.controllerEndpoint.ResolveReference(urlPath)

	// build the request
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	request, err := http.NewRequest(method, requestURL.String(), bodyReader)
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", p.controllerToken)

	// issue request
	retry := backoff.NewExponentialBackOff()
	retry.MaxElapsedTime = 5 * time.Minute

	var response *http.Response
	err = backoff.Retry(func() error {
		response, err = p.client.Do(request)
		return err
	}, retry)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		switch response.StatusCode {
		case http.StatusNotFound:
			return nil, strongerrors.NotFound(errors.New(response.Status))
		default:
			return nil, errors.New(response.Status)
		}
	}

	// read response body if asked to
	if readResponse {
		return ioutil.ReadAll(response.Body)
	}

	return nil, nil
}
