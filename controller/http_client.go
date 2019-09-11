// Copyright © 2017 The virtual-kubelet authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	request = request.WithContext(ctx)

	response, err := p.client.Do(request)

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
