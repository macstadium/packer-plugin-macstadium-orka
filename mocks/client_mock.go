package mocks

import (
	"bytes"
	"net/http"
	"io/ioutil"
	"time"
)

	// ############################
	// # 	    ENDPOINTS         #
	// ############################
/***************************************************
*	/token  			POST	   *
*	/token  			DELETE     *
*	/resources/image/copy  		POST       *
*	/resources/image/commit  	POST	   *
*	/resources/image/save  		POST       *
*	/resources/image/delete  	DELETE     *
*	/resources/vm/create  		POST       *
*	/resources/vm/deploy  		POST       *
*	/resources/vm/purge  		DELETE     *
****************************************************/


type Client struct {
	ErrorType string
	Timeout time.Duration
}

// Do is the moc client's 'Do' func
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	Request := req.URL.Path + ":" + req.Method
	jsonError := `{"message":"", "errors": [{"message": "Error"}]}`
	switch Request { 
	case "/token:POST":
		if c.ErrorType == "Login" {
			return Response(req, jsonError, 500, "500 Internal Server Error")
		}
		json := `{"token":"sometoken","message":"success", "errors": []}`
		return Response(req, json, 200, "200 OK")
	case "/token:DELETE":
		if c.ErrorType == "Logout" {
			return Response(req, jsonError, 500, "500 Internal Server Error")
		}
		json := `{"tokensRevoked":"1","message":"", "errors": []}`
		return Response(req, json, 200, "200 OK")
	case "/resources/image/copy:POST":
		if c.ErrorType == "ImageCopy" {
			return Response(req, jsonError, 500, "500 Internal Server Error")
		}
		json := `{"message":"Successfully copied", "errors": []}`
		return Response(req, json, 200, "200 OK")
	case "/resources/image/delete:DELETE":
		if c.ErrorType == "ImageDelete" {
			return Response(req, jsonError, 500, "500 Internal Server Error")
		}
		json := `{"message":"Successfully deleted", "errors": []}`
		return Response(req, json, 200, "200 OK")
	case "/resources/image/commit:POST":
		if c.ErrorType == "ImageCommit" {
			return Response(req, jsonError, 500, "500 Internal Server Error")
		}
		json := `{"message":"committed", "errors": []}`
		return Response(req, json, 200, "200 OK")
	case "/resources/image/save:POST":
		if c.ErrorType == "ImageSave" {
			return Response(req, jsonError, 500, "500 Internal Server Error")
		}
		json := `{"message":"saved", "errors": []}`
		return Response(req, json, 200, "200 OK")
	case "/resources/vm/create:POST":
		if c.ErrorType == "VMCreate" {
			return Response(req, jsonError, 500, "500 Internal Server Error")
		}
		json := `{"message":"Successfully Created", "errors": []}`
		return Response(req, json, 201, "201 Created")
	case "/resources/vm/deploy:POST":
		if c.ErrorType == "VMDeploy" {
			return Response(req, jsonError, 500, "500 Internal Server Error")
		}
		json := `{
			"vm_id": "05ca969973999",
			"ip": "10.221.188.11",
			"ssh_port": "8823"
		}`
		return Response(req, json, 200, "200 OK")
	case "/resources/vm/purge:DELETE":
		if c.ErrorType == "VMPurge" {
			return Response(req, jsonError, 500, "500 Internal Server Error")
		}
		json := `{"message":"Successfully purged VM", "errors": []}`
		return Response(req, json, 200, "200 OK")
	}
	return nil, nil
}

func Response(req *http.Request, json string, StatusCode int, Status string) (*http.Response, error) {
	// create a new reader with that JSON
	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	return &http.Response{
		StatusCode: StatusCode,
		Status: Status,
		Body:       r,
	}, nil
}