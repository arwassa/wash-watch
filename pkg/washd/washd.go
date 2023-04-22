package washd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	common "github.com/arwassa/wash-watch/pkg/common"
	"golang.org/x/net/html"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type LocalConfig struct {
	AuthToken     AuthToken     `json:"authToken"`
	LaundryClient LaundryClient `json:"laundryClient"`
}

type CsrfToken struct {
	CsrfToken string `json:"csrfToken"`
}

type User struct {
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	AccessToken   string `json:"accessToken"`
	Authenticated bool   `json:"authenticated"`
}

type LoginResponse struct {
	User    User   `json:"user"`
	Expires string `json:"expires"`
}

type AuthToken struct {
	AccessToken string `json:"accessToken"`
	Expires     string `json:"expires"`
}

type LaundryClient struct {
	*http.Client `json:"-"`
	Host         string `json:"host"`
	LoginUrl     string `json:"loginUrl"`
	Expires      string `json:"expires"`
}

func (client *LaundryClient) Expired() bool {
	d, e := time.Parse(time.RFC3339, client.Expires)
	if e != nil {
		log.Fatal(e)
	}
	return d.Before(time.Now())
}

func (laundryClient *LaundryClient) Renew() {
	jar, _ := cookiejar.New(nil)
	client := http.Client{Jar: jar}
	r, e := client.Get(laundryClient.LoginUrl)
	io.Copy(io.Discard, r.Body)
	r.Body.Close()

	if e != nil {
		fmt.Println(e)
	}
	laundryClient.Client = &client
}

func (authToken *AuthToken) Expired() bool {
	d, e := time.Parse(time.RFC3339, authToken.Expires)
	if e != nil {
		log.Fatal(e)
	}
	return d.Before(time.Now())
}

type Card struct {
}
type AccessibleLaundry struct {
	LocalSignInUrl      string
	LaundryGeoLongitude float64
	LaundryGeoLatitude  float64
	LaundryZipCode      string
	LaundryAddress      string
	LaundryName         string
	LaundryNumber       string
}
type LaundryResponse struct {
	ResultText          string
	ResultOK            bool
	ApartmentNumber     string
	Cards               []Card
	AccessibleLaundries []AccessibleLaundry
	WasherPrice         int
}

func GetLaundryClient(authToken AuthToken) LaundryClient {

	laundryClient := LaundryClient{}

	jar, _ := cookiejar.New(nil)
	client := http.Client{Jar: jar}
	url, _ := url.Parse("https://www.sio.no/api/vaskeri/secure/account")
	url.Query().Add("_", fmt.Sprint(time.Now().Unix()))
	req := http.Request{
		Method: "GET",
		URL:    url,
		Header: http.Header{
			"AccessToken": {authToken.AccessToken},
		},
	}

	resp, _ := client.Do(&req)
	print(resp.Status)
	var laundryResponse LaundryResponse
	json.NewDecoder(resp.Body).Decode(&laundryResponse)
	resp.Body.Close()
	r, e := client.Get(laundryResponse.AccessibleLaundries[0].LocalSignInUrl)
	io.Copy(io.Discard, r.Body)
	r.Body.Close()

	if e != nil {
		fmt.Println(e)
	}
	parsedUrl, _ := url.Parse(laundryResponse.AccessibleLaundries[0].LocalSignInUrl)
	laundryClient.Client = &client
	laundryClient.Host = fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host)
	laundryClient.LoginUrl = laundryResponse.AccessibleLaundries[0].LocalSignInUrl
	d, e := base64.StdEncoding.DecodeString(parsedUrl.Query().Get("session"))
	laundryClient.Expires = fmt.Sprintf("%sZ", d[6:6+19])
	return laundryClient
}

type LoginCredentials struct {
	Username string
	Password string
}

func Login(credentials LoginCredentials) AuthToken {
	jar, _ := cookiejar.New(nil)
	client := http.Client{
		Jar: jar,
	}
	r, _ := client.Get("https://www.sio.no/api/authorization/csrf")
	var csrfToken CsrfToken
	json.NewDecoder(r.Body).Decode(&csrfToken)
	r.Body.Close()
	loginData := url.Values{
		"username":  {credentials.Username},
		"password":  {credentials.Password},
		"remember":  {"false"},
		"csrfToken": {csrfToken.CsrfToken},
		"json":      {"true"},
	}

	r, e := client.PostForm("https://www.sio.no/api/authorization/callback/sio-user-pass-login", loginData)
	if e != nil {
		fmt.Println(e)
	}
	io.Copy(io.Discard, r.Body)
	r.Body.Close()

	r, e = client.Get("https://www.sio.no/api/authorization/session")
	if e != nil {
		fmt.Println(e)
	}
	fmt.Printf("Login status %s\n", r.Status)

	var loginResposne LoginResponse
	json.NewDecoder(r.Body).Decode(&loginResposne)
	r.Body.Close()
	return AuthToken{
		AccessToken: loginResposne.User.AccessToken,
		Expires:     loginResposne.Expires,
	}
}

type WashServiceServerHandle struct {
	UnimplementedWashServiceServer
	laundryClient *LaundryClient
	authToken     *AuthToken
	credentials   LoginCredentials
}

func CreateWashServiceServerHandle(credentials LoginCredentials) *WashServiceServerHandle {
	return &WashServiceServerHandle{
		credentials: credentials,
	}
}

func (handle *WashServiceServerHandle) CheckLogin(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
	}
	if handle.authToken == nil || handle.authToken.Expired() {
		token := Login(handle.credentials)
		handle.authToken = &token
	}
	select {
	case <-ctx.Done():
		return
	default:
	}
	if handle.laundryClient == nil || handle.laundryClient.Expired() {
		client := GetLaundryClient(*handle.authToken)
		handle.laundryClient = &client
	}
}

func (handle *WashServiceServerHandle) ListMachines(ctx context.Context, _ *MachineListRequest) (*MachineList, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	handle.CheckLogin(ctx)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	machineList := ListMachines(handle.laundryClient)

	machineProtoList := make([]*Machine, 0, len(machineList))
	for _, m := range machineList {
		machineProtoList = append(machineProtoList, &Machine{
			Name:   m.Name,
			Status: m.StateField,
		})
	}

	ret := &MachineList{
		Machines: machineProtoList,
	}

	return ret, nil
}

func ListMachines(client *LaundryClient) []common.Machine {
	r, _ := client.Get(fmt.Sprintf("%s/LaundryState", client.Host))
	transform := transform.NewReader(r.Body, charmap.ISO8859_1.NewDecoder())
	doc, err := goquery.NewDocumentFromReader(transform)
	if err != nil {
		log.Fatal(err)
	}
	r.Body.Close()

	out := strings.Builder{}
	machineList := make([]common.Machine, 0, 20)
	doc.Find("table.tb").Each(func(i int, s *goquery.Selection) {
		s.Find("td.p").Each(func(i2 int, s2 *goquery.Selection) {
			out.WriteString(fmt.Sprintln(s2.Text()))
			name := s2.Find("b").Text()
			stateFields := make([]string, 0, 10)
			for n := s2.Nodes[0].FirstChild; n != nil; n = n.NextSibling {
				if n.Type == html.TextNode {
					stateFields = append(stateFields, strings.Trim(n.Data, " "))
				}
			}

			m := common.Machine{
				Name:       name,
				StateField: strings.Join(stateFields, " "),
			}
			machineList = append(machineList, m)
		})
	})
	return machineList
}
