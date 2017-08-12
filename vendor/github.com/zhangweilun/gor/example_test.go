package gor

import (
	"log"
	"testing"
	"net/http"
	"net/url"
	"io"
	"encoding/xml"
)

/**
* 
* @author willian
* @created 2017-01-24 19:18
* @email 18702515157@163.com  
**/

func Test_get(t *testing.T) {
	// This is a very basic GET request
	resp, err := Get("http://httpbin.org/get", nil)

	if err != nil {
		log.Println(err)
	}

	if resp.Ok != true {
		log.Println("Request did not return OK")
	}

	log.Println(resp.String())
}

func Test_customHTTPClient(t *testing.T) {
	resp, err := Get("http://httpbin.org/get", &Request_options{Http_client: http.DefaultClient})

	if err != nil {
		log.Println(err)
	}

	if resp.Ok != true {
		log.Println("Request did not return OK")
	}

	log.Println(resp.String())
}

func Test_proxy(t *testing.T) {
	proxyURL, err := url.Parse("http://119.182.101.167:9999") // Proxy URL
	if err != nil {
		log.Panicln(err)
	}

	resp, err := Get("http://ip.chinaz.com/getip.aspx",
		&Request_options{Proxies: map[string]*url.URL{proxyURL.Scheme: proxyURL}})

	if err != nil {
		log.Println(err)
	}

	if resp.Ok != true {
		log.Println("Request did not return OK")
	}

	log.Println(resp)
}
func Example_cookies() {
	resp, err := Get("http://httpbin.org/cookies",
		&Request_options{
			Cookies: []*http.Cookie{
				{
					Name:     "TestCookie",
					Value:    "Random Value",
					HttpOnly: true,
					Secure:   false,
				}, {
					Name:     "AnotherCookie",
					Value:    "Some Value",
					HttpOnly: true,
					Secure:   false,
				},
			},
		})

	if err != nil {
		log.Println("Unable to make request", err)
	}

	if resp.Ok != true {
		log.Println("Request did not return OK")
	}

	log.Println(resp.String())
}

//func Example_session() {
//	session := grequests.NewSession(nil)
//
//	resp, err := session.Get("http://httpbin.org/cookies/set", &grequests.RequestOptions{Params: map[string]string{"one": "two"}})
//
//	if err != nil {
//		log.Fatal("Cannot set cookie: ", err)
//	}
//
//	if resp.Ok != true {
//		log.Println("Request did not return OK")
//	}
//
//	log.Println(resp.String())
//
//}

func Example_parse_XML() {
	type GetXMLSample struct {
		XMLName xml.Name `xml:"slideshow"`
		Title   string   `xml:"title,attr"`
		Date    string   `xml:"date,attr"`
		Author  string   `xml:"author,attr"`
		Slide   []struct {
			Type  string `xml:"type,attr"`
			Title string `xml:"title"`
		} `xml:"slide"`
	}

	resp, err := Get("http://httpbin.org/xml", nil)

	if err != nil {
		log.Println("Unable to make request", err)
	}

	if resp.Ok != true {
		log.Println("Request did not return OK")
	}

	userXML := &GetXMLSample{}

	// func xmlASCIIDecoder(charset string, input io.Reader) (io.Reader, error) {
	// 	return input, nil
	// }

	// If the server returns XML encoded in another charset (not UTF-8) – you
	// must provide an encoder function that looks like the one I wrote above.

	// If you an consuming UTF-8 just pass `nil` into the second arg
	if err := resp.XML(userXML, xmlASCIIDecoder); err != nil {
		log.Println("Unable to consume the response as XML: ", err)
	}

	if userXML.Title != "Sample Slide Show" {
		log.Printf("Invalid XML serialization %#v", userXML)
	}
}

func Example_customUserAgent() {
	ro := &Request_options{UserAgent: "LeviBot 0.1"}
	resp, err := Get("http://httpbin.org/get", ro)

	if err != nil {
		log.Fatal("Oops something went wrong: ", err)
	}

	if resp.Ok != true {
		log.Println("Request did not return OK")
	}

	log.Println(resp.String())
}

func Example_basicAuth() {
	ro := &Request_options{Auth: []string{"Levi", "Bot"}}
	resp, err := Get("http://httpbin.org/get", ro)
	// Not the usual JSON so copy and paste from below

	if err != nil {
		log.Println("Unable to make request", err)
	}

	if resp.Ok != true {
		log.Println("Request did not return OK")
	}
}

func Example_customHTTPHeader() {
	ro := &Request_options{UserAgent: "LeviBot 0.1",
		Headers: map[string]string{"X-Wonderful-Header": "1"}}
	resp, err := Get("http://httpbin.org/get", ro)
	// Not the usual JSON so copy and paste from below

	if err != nil {
		log.Println("Unable to make request", err)
	}

	if resp.Ok != true {
		log.Println("Request did not return OK")
	}
}

func Example_acceptInvalidTLSCert() {
	ro := &Request_options{InsecureSkipVerify: true}
	resp, err := Get("https://www.pcwebshop.co.uk/", ro)

	if err != nil {
		log.Println("Unable to make request", err)
	}

	if resp.Ok != true {
		log.Println("Request did not return OK")
	}
}

func Example_urlQueryParams() {
	ro := &Request_options{
		Params: map[string]string{"Hello": "World", "Goodbye": "World"},
	}
	resp, err := Get("http://httpbin.org/get", ro)
	// url will now be http://httpbin.org/get?hello=world&goodbye=world

	if err != nil {
		log.Println("Unable to make request", err)
	}

	if resp.Ok != true {
		log.Println("Request did not return OK")
	}
}

func Example_downloadFile() {
	resp, err := Get("http://httpbin.org/get", nil)

	if err != nil {
		log.Println("Unable to make request", err)
	}

	if resp.Ok != true {
		log.Println("Request did not return OK")
	}

	if err := resp.DownloadToFile("randomFile"); err != nil {
		log.Println("Unable to download to file: ", err)
	}

	if err != nil {
		log.Println("Unable to download file", err)
	}

}

func Example_postForm() {
	resp, err := Post("http://httpbin.org/post",
		&Request_options{Data: map[string]string{"One": "Two"}})

	// This is the basic form POST. The request body will be `one=two`

	if err != nil {
		log.Println("Cannot post: ", err)
	}

	if resp.Ok != true {
		log.Println("Request did not return OK")
	}
}

func Example_postXML() {

	type XMLPostMessage struct {
		Name   string
		Age    int
		Height int
	}

	resp, err := Post("http://httpbin.org/post",
		&Request_options{Xml: XMLPostMessage{Name: "Human", Age: 1, Height: 1}})
	// The request body will contain the XML generated by the `XMLPostMessage` struct

	if err != nil {
		log.Println("Unable to make request", resp.Error)
	}

	if resp.Ok != true {
		log.Println("Request did not return OK")
	}
}

//func Example_postFileUpload() {
//
//	fd, err := FileUploadFromDisk("test_files/mypassword")
//
//	if err != nil {
//		log.Println("Unable to open file: ", err)
//	}
//
//	// This will upload the file as a multipart mime request
//	resp, err := grequests.Post("http://httpbin.org/post",
//		&grequests.RequestOptions{
//			Files: fd,
//			Data:  map[string]string{"One": "Two"},
//		})
//
//	if err != nil {
//		log.Println("Unable to make request", resp.Error)
//	}
//
//	if resp.Ok != true {
//		log.Println("Request did not return OK")
//	}
//}

func Example_postJSONAJAX() {
	resp, err := Post("http://httpbin.org/post",
		&Request_options{
			Json:   map[string]string{"One": "Two"},
			Is_ajax: true, // this adds the X-Requested-With: XMLHttpRequest header
		})

	if err != nil {
		log.Println("Unable to make request", resp.Error)
	}

	if resp.Ok != true {
		log.Println("Request did not return OK")
	}

}

func xmlASCIIDecoder(charset string, input io.Reader) (io.Reader, error) {
	return input, nil
}
