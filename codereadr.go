package codereadr

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const (
	apiURL = "https://api.codereadr.com/api/"
)

// Section consts
const (
	SectionDatabases = "databases"
	SectionScans     = "scans"
	SectionServices  = "services"
	SectionUsers     = "users"
)

// Action consts
const (
	ActionAddQuestion       = "addquestion"
	ActionAddUserPermission = "adduserpermission"
	ActionAddValue          = "addvalue"
	ActionCreate            = "create"
	ActionDelete            = "delete"
	ActionRemoveQuestion    = "removequestion"
	ActionRetrieve          = "retrieve"
	ActionShowValues        = "showvalues"
	ActionUpdate            = "update"
	ActionUpload            = "upload"
)

// ServiceValidationMethod consts
const (
	ServiceValidationMethodRecord           = "record"
	ServiceValidationMethodOnDeviceRecord   = "ondevicerecord"
	ServiceValidationMethodDatabase         = "database"
	ServiceValidationMethodOnDeviceDatabase = "ondevicedatabase"
	ServiceValidationMethodPostback         = "postback"
)

// New func
func New(apiKey string) *Client {
	return &Client{
		APIKey: apiKey,
	}
}

// Client type
type Client struct {
	APIKey string
}

// Do func
func (c *Client) Do(o interface{}, section, action string, args ...map[string]interface{}) error {
	reqBody := &bytes.Buffer{}
	reqBodyWriter := multipart.NewWriter(reqBody)
	reqBodyWriter.WriteField("api_key", c.APIKey)
	reqBodyWriter.WriteField("section", section)
	reqBodyWriter.WriteField("action", action)
	if len(args) > 0 && args[0] != nil {
		for n, v := range args[0] {
			name, value := fmt.Sprintf("%v", n), fmt.Sprintf("%v", v)
			if strings.HasPrefix(name, "@") {
				name = name[1:]
				f, _ := reqBodyWriter.CreateFormFile(name, name)
				fmt.Fprint(f, value)
				continue
			}
			reqBodyWriter.WriteField(name, value)
		}
	}
	reqBodyWriter.Close()

	req, _ := http.NewRequest(http.MethodPost, apiURL, reqBody)
	req.Header.Add("Content-Type", reqBodyWriter.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var resBody Response
	if err := xml.Unmarshal(b, &resBody); err != nil {
		return err
	}

	if resBody.Status != 1 {
		return fmt.Errorf("%s [Error #%d]",
			resBody.Error.Message, resBody.Error.Code)
	}

	if o == nil {
		return nil
	}
	return xml.Unmarshal(b, o)
}

// Response type
type Response struct {
	Status int `xml:"status"`
	Error  *struct {
		Code    int    `xml:"code,attr"`
		Message string `xml:",chardata"`
	} `xml:"error"`
}

// ResponseCreate type
type ResponseCreate struct {
	Response
	ID int `xml:"id"`
}

// ResponseUsersRetrieve type
type ResponseUsersRetrieve struct {
	Response
	Count int `xml:"count"`
	Users []struct {
		ID       int    `xml:"id,attr"`
		Username string `xml:"username"`
	} `xml:"user"`
}

// ResponseDatabasesRetrieve type
type ResponseDatabasesRetrieve struct {
	Response
	Count     int `xml:"count"`
	Databases []struct {
		ID int `xml:"id,attr"`
	} `xml:"database"`
}

// ResponseDatabasesShowValues type
type ResponseDatabasesShowValues struct {
	Response
	Count  int `xml:"count"`
	Values []struct {
		Response string `xml:"response"`
	} `xml:"value"`
}

// ResponseServicesRetrieve type
type ResponseServicesRetrieve struct {
	Response
	Count    int `xml:"count"`
	Services []struct {
		ID          int    `xml:"id,attr"`
		Name        string `xml:"name"`
		Description string `xml:"descriptionFromDb"`
	} `xml:"service"`
}

// ResponseScansRetrieve type
type ResponseScansRetrieve struct {
	Response
	Count int `xml:"count"`
	Scans []struct {
		Service struct {
			ID   int    `xml:"id,attr"`
			Name string `xml:",chardata"`
		} `xml:"service"`
		ID        int      `xml:"id,attr"`
		TID       string   `xml:"tid"`
		Result    string   `xml:"result"`
		Timestamp DateTime `xml:"timestamp"`
		Answer    string   `xml:"answer"`
	} `xml:"scan"`
}

// DateTime type
type DateTime struct {
	time.Time
}

// UnmarshalXML func
func (t *DateTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	d.DecodeElement(&v, &start)

	parsed, err := time.Parse("2006-01-02 15:04:05", v)
	if err != nil {
		return err
	}

	*t = DateTime{parsed}
	return nil
}
