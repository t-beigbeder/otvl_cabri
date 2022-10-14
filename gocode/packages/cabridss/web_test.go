package cabridss

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
)

type tVersion struct {
	Version string `json:"version"`
}

func cGetTVersion(apc WebApiClient) (*tVersion, error) {
	req, _ := http.NewRequest(http.MethodGet, apc.Url()+"version", nil)
	v := tVersion{}
	_, err := apc.DoAsJson(req, &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func sGetTVersion(c echo.Context) error {
	version := "0.0.13.42"
	if GetCustomConfig(c) != nil {
		version = GetCustomConfig(c).(string)
	}
	return c.JSON(http.StatusOK, tVersion{Version: version})
}

type EchoIn struct {
	Param string `param:"kparam" json:"kparam"`
	Query string `query:"kquery" json:"kquery"`
	Body  string `json:"kbody"`
}

type tEchoOut struct {
	Echo string `json:"echo"`
}

func cGetTEcho(apc WebApiClient, in EchoIn) (*tEchoOut, error) {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%secho/%s?kquery=%s", apc.Url(), in.Param, in.Query), nil)
	v := tEchoOut{}
	_, err := apc.DoAsJson(req, &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func sGetTEcho(c echo.Context) error {
	v := EchoIn{}
	err := c.Bind(&v)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, tEchoOut{Echo: fmt.Sprintf("%+v", v)})
}

func cPutTEcho(apc WebApiClient, in EchoIn) (*tEchoOut, error) {
	out := tEchoOut{}
	_, err := apc.SimpleDoAsJson(http.MethodPut, fmt.Sprintf("%secho/%s?kquery=%s", apc.Url(), in.Param, in.Query), in, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func cPutTEchoConfig(apc WebApiClient) (*tEchoOut, error) {
	in := EchoIn{Param: "cptecParam", Query: "cptecQuery", Body: apc.GetConfig().(string)}
	out := tEchoOut{}
	_, err := apc.SimpleDoAsJson(http.MethodPut, fmt.Sprintf("%secho/%s?kquery=%s", apc.Url(), in.Param, in.Query), in, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func cErrorApp(apc WebApiClient) (*mError, error) {
	out := mError{}
	_, err := apc.SimpleDoAsJson(http.MethodPost, apc.Url()+"error/app", nil, &out)
	return &out, err
}

func cErrorServer(apc WebApiClient) (*mError, error) {
	out := mError{}
	_, err := apc.SimpleDoAsJson(http.MethodPost, apc.Url()+"error/server", nil, &out)
	return &out, err
}

func sPutTEcho(c echo.Context) error {
	v := EchoIn{}
	err := c.Bind(&v)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, tEchoOut{Echo: fmt.Sprintf("%+v", v)})
}

func sErrorApp(c echo.Context) error {
	return c.JSON(http.StatusOK, &mError{Error: fmt.Errorf("an app error").Error()})
}

func sErrorServer(c echo.Context) error {
	return NewServerErr("sErrorApp", fmt.Errorf("a server error"))
}

func testEchoConfigurator(e *echo.Echo, root string, _ interface{}) error {
	e.GET(root+"version", sGetTVersion)
	e.GET(root+"echo/:kparam", sGetTEcho)
	e.PUT(root+"echo/:kparam", sPutTEcho)
	e.POST(root+"error/app", sErrorApp)
	e.POST(root+"error/server", sErrorServer)
	return nil
}

func TestNewWebApiClient(t *testing.T) {
	optionalSkip(t)
	s := NewEServer(":3000", true)
	resShutdown := ""
	s.ConfigureApi("/test", "0.0.90.90", testEchoConfigurator, func(config interface{}) error {
		customConfig := config.(string)
		resShutdown = fmt.Sprintf("Shutdown %s", customConfig)
		return nil
	})
	defer s.Shutdown()
	if err := s.Serve(); err != nil {
		t.Fatal(err)
	}

	apc := NewWebApiClient("", "", "3000", "test", "sConfigClient")
	v, err := cGetTVersion(apc)
	if err != nil {
		t.Fatal(err)
	}
	_ = v
	eo1, err := cGetTEcho(apc, EchoIn{Query: "vquery", Param: "vparam"})
	if err != nil {
		t.Fatal(err)
	}
	_ = eo1
	eo2, err := cPutTEcho(apc, EchoIn{Query: "vquery", Param: "vparam", Body: "vbody"})
	if err != nil {
		t.Fatal(err)
	}
	_ = eo2
	eo3, err := cPutTEchoConfig(apc)
	if err != nil || !strings.Contains(eo3.Echo, "Body:sConfigClient") {
		t.Fatal(err)
	}
	_ = eo3
	ea, err := cErrorApp(apc)
	if err == nil || err.Error() != "in DoAsJson: an app error" || ea.Error != "an app error" {
		t.Fatal(err)
	}
	_ = ea
	es, err := cErrorServer(apc)
	if err == nil || err.Error() != "in DoAsJson: error status 500 Internal Server Error" || es.Error != "" {
		t.Fatal(err)
	}
	_ = es
	if err = s.Shutdown(); err != nil || resShutdown != "Shutdown 0.0.90.90" {
		t.Fatalf("%v %s", err, resShutdown)
	}
}

func TestWebApiClientBurst(t *testing.T) {
	optionalSkip(t)
	s := NewEServer(":3000", true)
	resShutdown := ""
	s.ConfigureApi("/test", "0.0.90.90", testEchoConfigurator, func(config interface{}) error {
		customConfig := config.(string)
		resShutdown = fmt.Sprintf("Shutdown %s", customConfig)
		return nil
	})
	defer s.Shutdown()
	if err := s.Serve(); err != nil {
		t.Fatal(err)
	}

	apc := NewWebApiClient("", "", "3000", "test", "sConfigClient")
	var err error
	wg := sync.WaitGroup{}
	wg.Add(220)
	for i := 0; i < 220; i++ {
		go func() {
			_, err = cGetTVersion(apc)
			wg.Done()
		}()
	}
	wg.Wait()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Shutdown(); err != nil || resShutdown != "Shutdown 0.0.90.90" {
		t.Fatalf("%v %s", err, resShutdown)
	}
}

type dataGen struct {
	count int
}

func (d *dataGen) Read(p []byte) (n int, err error) {
	if d.count >= 1000000 {
		return 0, io.EOF
	}
	var i int
	for i = 0; i < len(p) && i+d.count < 1000000; i++ {
		p[i] = byte((i + d.count) % 256)
	}
	d.count += i
	//fmt.Printf("dataGen.Read %d\n", d.count)
	return i, nil
}

func (d *dataGen) Close() error {
	//fmt.Printf("dataGen.Close %d\n", d.count)
	return nil
}

type dataSink struct {
	count int
	sink  []byte
}

func (d *dataSink) Read(p []byte) (n int, err error) {
	if d.sink == nil || d.count >= 1000000 {
		//fmt.Printf("dataSink.Read %d EOF\n", d.count)
		return 0, io.EOF
	}
	var i int
	for i = 0; i < len(p) && i+d.count < 1000000; i++ {
		p[i] = d.sink[i+d.count]
	}
	d.count += i
	//fmt.Printf("dataSink.Read %d\n", d.count)
	return i, nil
}

func (d *dataSink) Write(p []byte) (n int, err error) {
	if d.sink == nil {
		d.sink = make([]byte, 1000000)
	}
	if d.count >= 1000000 {
		//fmt.Printf("dataSink.Write %d error\n", d.count)
		return 0, fmt.Errorf("count %d", d.count)
	}
	var i int
	for i = 0; i < len(p) && i+d.count < 1000000; i++ {
		d.sink[i+d.count] = p[i]
	}
	d.count += i
	//fmt.Printf("dataSink.Write %d\n", d.count)
	return i, nil
}

func (d *dataSink) Close() error {
	//fmt.Printf("dataSink.Close %d\n", d.count)
	d.count = 0
	return nil
}

func cPostStream(apc WebApiClient) (interface{}, error) {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%sin", apc.Url()), nil)
	req.Body = &dataGen{}
	resp, err := apc.(*apiClient).client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("in cPostStream: error status %s", resp.Status)
	}
	return resp, err
}

func cOutStream(apc WebApiClient) (interface{}, error) {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%sout", apc.Url()), nil)
	resp, err := apc.(*apiClient).client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("in cOutStream: error status %s", resp.Status)
	}
	cDs := dataSink{}
	io.Copy(&cDs, resp.Body)
	return resp, err
}

var sDs dataSink

func sPostStream(c echo.Context) error {
	req := c.Request()
	io.Copy(&sDs, req.Body)
	sDs.Close()
	return nil
}

func sOutStream(c echo.Context) error {
	resp := c.Response()
	io.Copy(resp.Writer, &sDs)
	return nil
}

func testStreamConfigurator(e *echo.Echo, root string, _ interface{}) error {
	e.POST(root+"in", sPostStream)
	e.POST(root+"out", sOutStream)
	return nil
}

func runTestWebApiStream(t *testing.T) {
	optionalSkip(t)
	s := NewEServer(":3000", true)
	err := s.ConfigureApi("", "0.0.90.90", testStreamConfigurator, func(config interface{}) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Shutdown()
	if err := s.Serve(); err != nil {
		t.Fatal(err)
	}
	apc := NewWebApiClient("", "", "3000", "", nil)
	_, err = cPostStream(apc)
	if err != nil {
		t.Fatal(err)
	}
	_, err = cOutStream(apc)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWebApiStream(t *testing.T) {
	optionalSkip(t)
	for i := 0; i < 20; i++ {
		runTestWebApiStream(t)
	}
}

func TestWebTestSleep(t *testing.T) {
	optionalSkip(t)
	if os.Getenv("CABRIDSS_KEEP_DEV_TESTS") == "" {
		t.Skip(fmt.Sprintf("Skipping %s because you didn't set CABRIDSS_KEEP_DEV_TESTS", t.Name()))
	}
	s := NewEServer(":3000", true)
	s.ConfigureApi("test", nil, WebTestServerConfigurator, func(customConfig interface{}) error { return nil })
	defer s.Shutdown()
	if err := s.Serve(); err != nil {
		t.Fatal(err)
	}

	apc := NewWebApiClient("", "", "3000", "test", nil)
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%ssleep/%d", apc.Url(), 500), nil)
	resp, err := apc.(*apiClient).client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		t.Fatal(err)
	}
	count := 1000000
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("%spostStream/%d/%d", apc.Url(), count, 5000), nil)
	req.Body = ioutil.NopCloser(strings.NewReader(strings.Repeat("0", count)))
	resp, err = apc.(*apiClient).client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		t.Fatal(err)
	}

}
