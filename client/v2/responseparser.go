package client

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gogo/protobuf/proto"
	data "github.com/influxdb/client/v2/internal"
)

//ResponseParser ...
type ResponseParser interface {
	// WriteResponse writes a response.
	ParseResponse(resp *http.Response) (interface{}, error)

	//io.ReadCloser
}

type responseParser struct {
	formatter interface {
		ParseResponse(resp *http.Response) (interface{}, error)
	}
	*http.Response
}

// NewResponseParser creates a new ResponseWriter based on the Accept header
// in the request that wraps the ResponseWriter.
func NewResponseParser(resp *http.Response) ResponseParser {
	rw := &responseParser{Response: resp}
	switch resp.Header.Get("Content-Type") {

	case "application/x-protobuf", "application/octet-stream":
		rw.formatter = &protoParser{Response: resp}
	case "application/json":
		fallthrough
	default:
		rw.formatter = &jsonParser{Response: resp}
	}
	return rw
}

func (w *responseParser) ParseResponse(resp *http.Response) (interface{}, error) {
	return w.formatter.ParseResponse(resp)
}

type protoParser struct {
	*http.Response
	respData data.Results
}

func (w *protoParser) ParseResponse(resp *http.Response) (interface{}, error) {

	buf := bytes.NewBuffer(make([]byte, 0))
	_, readErr := buf.ReadFrom(resp.Body)
	failOnError(readErr)
	body := buf.Bytes()
	err := proto.Unmarshal(body, &w.respData)
	failOnError(err)
	// log.Println("proto: ", w.respData)
	return w.respData, err

}

func failOnError(err error) {
	if err != nil {
		log.Println("FAIL: ", err)
	}
}

type jsonParser struct {
	*http.Response
	resp Response
	//respData map[string]interface{} //.InfluxResponse
}

func (w *jsonParser) ParseResponse(resp *http.Response) (interface{}, error) {

	buf := bytes.NewBuffer(make([]byte, 0))
	_, readErr := buf.ReadFrom(resp.Body)
	failOnError(readErr)
	body := buf.Bytes()
	err := json.Unmarshal(body, &w.resp)
	failOnError(err)
	// log.Println("json: ", w.respData)
	return w.resp, err
}
