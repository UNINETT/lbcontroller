// +build ignore
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/koki/json"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
)

//Service handled by the load balancers
type Service struct {
	Type     ServiceType              `json:"type,omitempty"`
	Metadata Metadata                 `json:"metadata,omitempty"`
	Config   Config                   `json:"config,omitempty"`
	Ingress  []v1.LoadBalancerIngress `json:"ingress,omitempty"` //TODO(gta) make our own type and remove dependancy from k8s?

}

//ListServices return a list of services
//configured on the loadbalancers.
//A token is needed to authenticate
func ListServices(url, token string) ([]Service, error) {
	url = svcURL(url)

	req, err := newRequest(http.MethodGet, url, token, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creatign http.Request")
	}
	res, err := http.DefaultClient.Do(req)
	//res, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to API endpoint: %s", url)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("error, returned status not 200 OK from API endpoint: %s", res.Status)
	}

	dec := json.NewDecoder(res.Body)
	svcs := []Service{}

	//read all the Messages and alter parse the cofigs
	for dec.More() {
		var s Service
		// decode an array value (Message)
		err := dec.Decode(&s)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding a Service object")
		}
		svcs = append(svcs, s)
	}
	res.Body.Close()

	return svcs, nil
}

//GetService get the configuration of the fronten specified by name, if the service
//is found GetService returnns a true boolean value as well
//A token is needed to authenticate
func GetService(name, url, token string) (Service, bool, error) {

	url = svcURL(url)
	ret := Service{}

	req, err := newRequest(http.MethodGet, url+"/"+name, token, nil)
	if err != nil {
		return ret, false, errors.Wrapf(err, "error creating http request")
	}
	res, err := http.DefaultClient.Do(req)
	//res, err := http.Get(url + "/" + name)
	if err != nil {
		return ret, false, errors.Wrapf(err, "error connecting to API endpoint: %s", url)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ret, false, errors.Wrapf(err, "error reading from API endpoint: %s", url)
	}

	var ingress []v1.LoadBalancerIngress
	//handle stautus 200 and 404
	switch res.StatusCode {
	case http.StatusNotFound:
		return ret, false, nil
	case http.StatusOK:
		location := res.Header.Get("Location")
		if location != "" {
			ingress, err = getIngress(location)
			if err != nil {
				return ret, false, errors.Wrapf(err, "error getting ingress form api: %s", location)
			}
		}

	default:
		return ret, false, errors.Errorf("error, returned status from API endpoint not supported: %s\n ", res.Status)
	}

	err = json.Unmarshal(body, &ret)
	if err != nil {
		return ret, false, errors.Wrap(err, "error decoding Service object")
	}
	ret.Ingress = ingress

	return ret, true, nil
}

//SyncService create or updates a new service
//A token is needed to authenticate
func SyncService(svc Service, url, token string) ([]v1.LoadBalancerIngress, error) {

	url = svcURL(url)
	data, err := json.Marshal(svc)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling Service")
	}
	buf := bytes.NewBuffer(data)

	req, err := newRequest(http.MethodPut, url+"/"+svc.Metadata.Name, token, buf)
	if err != nil {
		return nil, errors.Wrap(err, "error creatign http.Request")
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error sync-ing Service %s", svc.Metadata.Name)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint: %s", url)
	}

	switch res.StatusCode {
	case http.StatusCreated, http.StatusOK:
		//happy path
	default:
		return nil, errors.Errorf("API endpoint returned status %s, %s", res.Status, body)
	}

	location := res.Header.Get("Location")
	var ret []v1.LoadBalancerIngress

	if location != "" {
		ret, err = getIngress(location)
		if err != nil {
			return nil, errors.Wrapf(err, "error getting ingress form api: %s", location)
		}
		return ret, nil
	}

	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding Service object")
	}
	return ret, nil
}

//DeleteService deletes and exixting Service object, the new Service is retured.
//A token is needed to authenticate
func DeleteService(name, url, token string) error {
	url = svcURL(url)

	req, err := newRequest(http.MethodDelete, url+"/"+name, token, nil)
	if err != nil {
		return errors.Wrap(err, "error creatign http.Request")
	}
	req.Header.Set("Content-Type", jsonContent)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "error replacing Service %s", name)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return errors.Wrapf(err, "error reading from API endpoint: %s", url)
		}
		return errors.Errorf("API endpoint returned status %s, %s", res.Status, bytes.TrimSpace(body))
	}

	return nil
}

func svcURL(url string) string {
	return url + "/" + servicePath
}

//getIngress retrives the k8s loadBalancerIngress from the specified url
func getIngress(url string) ([]v1.LoadBalancerIngress, error) {

	//url = svcURL(url)
	var ret []v1.LoadBalancerIngress

	res, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to API endpoint to get ingress: %s", url)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from API endpoint to get ingress: %s", url)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("error, returned status from API endpoint not supported: %s\n ", res.Status)
	}

	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding LoadBalancerIngress object")
	}

	return ret, nil
}

func newRequest(method, endpoint, token string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	return req, nil
}
