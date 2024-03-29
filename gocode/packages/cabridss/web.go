package cabridss

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type WebServerConfig struct {
	Addr              string // host[:port]
	HasLog            bool
	IsTls             bool   // https
	TlsCert           string // certificate file on https server or untrusted CA on https client
	TlsKey            string // certificate key file on https server
	TlsNoCheck        bool   // no check of certificate by https client
	BasicAuthUser     string
	BasicAuthPassword string
}

type WebServer interface {
	Serve() error
	Shutdown() error
	ConfigureApi(
		root string, customConfig interface{},
		shutdownCallback func(root string, customConfigs map[string]interface{}) error,
		ctor func(e *echo.Echo, root string, customConfigs map[string]interface{}) error,
	) error
	getEcho() *echo.Echo
}

type TlsConfig struct {
	cert              string
	key               string
	noClientCheck     bool
	basicAuthUser     string
	basicAuthPassword string
}

func getTlsClientConfig(tlsConfig *TlsConfig) (*tls.Config, error) {
	if tlsConfig == nil {
		return nil, nil
	}
	if tlsConfig.noClientCheck {
		return &tls.Config{InsecureSkipVerify: true}, nil
	}
	if tlsConfig.cert == "" {
		return &tls.Config{}, nil
	}
	caCert, err := os.ReadFile(tlsConfig.cert)
	if err != nil {
		return nil, fmt.Errorf("in getTlsConfig: %v", err)
	}
	caCertPool, _ := x509.SystemCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return &tls.Config{RootCAs: caCertPool}, nil
}

func getTlsServerConfig(wsConfig WebServerConfig) *TlsConfig {
	return &TlsConfig{
		cert:              wsConfig.TlsCert,
		key:               wsConfig.TlsKey,
		noClientCheck:     wsConfig.TlsNoCheck,
		basicAuthUser:     wsConfig.BasicAuthUser,
		basicAuthPassword: wsConfig.BasicAuthPassword,
	}
}

type eServer struct {
	e                 *echo.Echo
	tlsConfig         *TlsConfig
	customConfigs     map[string]interface{}
	shutdownCallbacks map[string]func(root string, customConfigs map[string]interface{}) error
	addr              string
	firstRoot         string
	shutReq           chan interface{}
	shutResp          chan interface{}
	closed            bool
}

type eCustomContext struct {
	echo.Context
	esv *eServer
}

func (esv *eServer) getEcho() *echo.Echo { return esv.e }

func (esv *eServer) getHP() (string, string) {
	host := "localhost"
	port := "3000"
	frags := strings.Split(esv.addr, ":")
	if len(frags) == 2 {
		if frags[0] != "" {
			host = frags[0]
		}
		if frags[1] != "" {
			port = frags[1]
		}
	}
	return host, port
}

func (esv *eServer) checkPort(serving bool) error {
	host, port := esv.getHP()
	for i := 0; i < 5; i++ {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 10*time.Millisecond)
		if err != nil {
			if !serving {
				return nil
			}
		} else {
			conn.Close()
			if serving {
				return nil
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("host:port %s:%s is not available", host, port)
}

func (esv *eServer) Serve() error {
	if err := esv.checkPort(false); err != nil {
		esv.closed = true
		return fmt.Errorf("in Serve: %v", err)
	}
	esv.shutReq = make(chan interface{})
	esv.shutResp = make(chan interface{})
	go func() {
		var err error
		if esv.tlsConfig != nil && esv.tlsConfig.basicAuthUser != "" {
			esv.e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
				if subtle.ConstantTimeCompare([]byte(username), []byte(esv.tlsConfig.basicAuthUser)) == 1 &&
					subtle.ConstantTimeCompare([]byte(password), []byte(esv.tlsConfig.basicAuthPassword)) == 1 {
					return true, nil
				}
				return false, nil
			}))
		}
		if esv.tlsConfig == nil {
			err = esv.e.Start(esv.addr)
		} else {
			err = esv.e.StartTLS(esv.addr, esv.tlsConfig.cert, esv.tlsConfig.key)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Start or StartTLS %v\n", err)
		}
		close(esv.shutResp)
	}()
	if err := esv.checkPort(true); err != nil {
		esv.e.Shutdown(context.Background())
		esv.closed = true
		return fmt.Errorf("in Serve: %v", err)
	}
	host, port := esv.getHP()
	protocol := "http"
	var (
		ht     *http.Transport
		client Client
	)
	if esv.tlsConfig != nil {
		protocol = "https"
		tlsClientConfig, err := getTlsClientConfig(esv.tlsConfig)
		if err != nil {
			return fmt.Errorf("in Serve: %v", err)
		}
		ht = &http.Transport{TLSClientConfig: tlsClientConfig}
		client = Client{Client: http.Client{Transport: ht}}
	} else {
		client = Client{Client: http.Client{}}
	}
	url := fmt.Sprintf("%s://%s:%s%scheck", protocol, host, port, esv.firstRoot)
	for i := 0; i < 5; i++ {
		var (
			req *http.Request
			rsp *http.Response
			err error
		)
		req, err = http.NewRequest("GET", url, nil)
		if err == nil {
			if esv.tlsConfig != nil && esv.tlsConfig.basicAuthUser != "" {
				req.SetBasicAuth(esv.tlsConfig.basicAuthUser, esv.tlsConfig.basicAuthPassword)
			}
			rsp, err = client.Do(req, nil)
			if err == nil && rsp.StatusCode == http.StatusOK {
				break
			}
		}
		if i == 4 {
			return fmt.Errorf("in Serve: check KO")
		}
		time.Sleep(100 * time.Millisecond)
	}

	go func() {
		<-esv.shutReq
		if err := esv.e.Shutdown(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "http Shutdown: %v\n", err) // FIXME
		}
	}()
	return nil
}

func (esv *eServer) Shutdown() error {
	if esv.closed {
		return nil
	}
	errs := ErrorCollector{}
	for root, shutdownCallback := range esv.shutdownCallbacks {
		if shutdownCallback != nil {
			if err := shutdownCallback(root, esv.customConfigs); err != nil {
				errs.Collect(err)
			}
		}
	}
	close(esv.shutReq)
	<-esv.shutResp
	esv.closed = true
	if errs.Any() {
		return fmt.Errorf("in Shutdown: %s", errs.Error())
	}
	return nil
}

func (esv *eServer) ConfigureApi(
	root string, customConfig interface{},
	shutdownCallback func(root string, customConfigs map[string]interface{}) error,
	ctor func(e *echo.Echo, root string, customConfigs map[string]interface{}) error,
) error {
	if root == "" {
		root = "/"
	} else if root[0] != '/' {
		root = "/" + root
	}
	if root[len(root)-1] != '/' {
		root += "/"
	}
	if esv.firstRoot == "" {
		esv.firstRoot = root
	}
	esv.customConfigs[root] = customConfig
	esv.shutdownCallbacks[root] = shutdownCallback
	if err := ctor(esv.e, root, esv.customConfigs); err != nil {
		return fmt.Errorf("in ConfigureApi: %v", err)
	}
	esv.e.GET(root+"check", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "check OK")
	})
	return nil
}

func NewEServer(addr string, hasLog bool, tlsConfig *TlsConfig) WebServer {
	e := echo.New()
	esv := &eServer{e: e, addr: addr, tlsConfig: tlsConfig,
		customConfigs:     map[string]interface{}{},
		shutdownCallbacks: map[string]func(root string, customConfigs map[string]interface{}) error{},
	}
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &eCustomContext{Context: c, esv: esv}
			return next(cc)
		}
	})
	e.HideBanner = true
	e.HidePort = true
	if hasLog {
		e.Use(middleware.Logger())
	}
	return esv
}

func GetCustomConfig(c echo.Context) interface{} {
	cct, ok := c.(*eCustomContext)
	var customConfig interface{}
	for root, ccf := range cct.esv.customConfigs {
		if strings.HasPrefix(c.Path(), root) {
			customConfig = ccf
			break
		}
	}
	if !ok {
		panic("here")
	}
	return customConfig
}

func NewServerErr(where string, err error) error {
	return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("in %s: %v", where, err))
}

func NewClientErr(where string, resp *http.Response, err error, bs []byte) error {
	pfx := ""
	if where != "" {
		pfx = fmt.Sprintf("in %s: ", where)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		if bs != nil {
			return fmt.Errorf("%serror status %s %s", pfx, resp.Status, string(bs))
		} else {
			return fmt.Errorf("%serror status %s", pfx, resp.Status)
		}
	}
	if err != nil {
		return fmt.Errorf("%s%v", pfx, err)
	}
	return nil
}

type WebApiClient interface {
	Url() string
	DoAsJson(request *http.Request, outBody any) (*http.Response, error)
	SimpleDoAsJson(method, url string, inBody any, outBody any) (*http.Response, error)
	GetConfig() interface{}
	SetCabriHeader(h string)
}

type Client struct {
	http.Client
	mux               sync.Mutex
	curRetries        int
	history           map[int64]bool
	nextId            int
	basicAuthUser     string
	basicAuthPassword string
	cabriHeader       string
}

type ClientReqOpts struct {
	raiseError  bool
	errorRaised bool
	getRequest  func() (*http.Request, error)
}

func HasRaiseError() bool { return os.Getenv("CABRIDSS_WEB_RAISE_ERROR") != "" }

func (c *Client) addRetry() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.curRetries++
}

func (c *Client) removeRetry() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.curRetries--
	if c.curRetries < 0 {
		panic("logic")
	}
}

func (c *Client) historize() {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.history == nil {
		c.history = map[int64]bool{}
	}
	c.history[time.Now().UnixNano()] = true
}

func (c *Client) stats() (rps, curRetries int, du time.Duration) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.history == nil {
		c.history = map[int64]bool{}
	}
	now := time.Now().UnixNano()
	for dt, _ := range c.history {
		n := TimeResolution("s").NanoSeconds()
		if dt+n >= now {
			rps += 1
		} else {
			delete(c.history, dt)
		}
	}
	curRetries = c.curRetries
	ms := 0
	if rps > 1000 {
		ms = rps - 1000
		if ms > 1000 {
			ms = 1000
		}
	}
	if curRetries == 0 {
		du = time.Duration(uint(ms)) * time.Millisecond
		return
	}
	if curRetries > 10 {
		ms = 2000
		du = time.Duration(uint(ms)) * time.Millisecond
		return
	}
	ms2 := 200 * curRetries
	if ms < ms2 {
		ms = ms2
	}
	du = time.Duration(uint(ms)) * time.Millisecond
	return
}

func (c *Client) doDo(req *http.Request, opts *ClientReqOpts) (*http.Response, error, int, int, time.Duration) {
	ellapsed := time.Duration(0)
	rps, nbr, du := c.stats()
	du0 := du
	for ellapsed < du {
		time.Sleep(du)
		ellapsed += du
		rps, nbr, du = c.stats()
	}
	var err error
	if opts.getRequest != nil {
		req, err = opts.getRequest()
		if err != nil {
			return nil, err, rps, nbr, du
		}
	}
	if c.cabriHeader != "" {
		req.Header.Set("Cabri", c.cabriHeader)
	}
	if c.basicAuthUser != "" {
		req.SetBasicAuth(c.basicAuthUser, c.basicAuthPassword)
	}
	c.historize()
	rsp, err := c.Client.Do(req)
	if err == nil && opts.raiseError && !opts.errorRaised {
		err = fmt.Errorf("ClientReqOpts: raised")
		opts.errorRaised = true
		fmt.Fprintf(os.Stderr, "%v %v\n", *opts, err)
	}
	_ = du0
	//if du != 0 || len(c.history) > 2000 {
	//	fmt.Fprintf(os.Stderr, "%d, %d, %d, %d, %d, (%v)\n", rps, nbr, du, du0, len(c.history), err)
	//}
	return rsp, err, rps, nbr, du
}

func (c *Client) Do(req *http.Request, opts *ClientReqOpts) (*http.Response, error) {
	if opts == nil {
		opts = &ClientReqOpts{}
	}

	hasRetry := false
	defer func() {
		if hasRetry {
			c.removeRetry()
		}
	}()
	for i := 0; i < 2; i++ {
		rsp, err, rps, nbr, du := c.doDo(req, opts)
		if err == nil || err == io.EOF {
			return rsp, err
		}
		if !hasRetry {
			c.addRetry()
			hasRetry = true
		}
		//if err != nil {
		//	fmt.Fprintf(os.Stderr, "%d, %d, %d (%v)\n", rps, nbr, du, err)
		//}
		_, _, _ = rps, nbr, du
	}
	rsp, err, rps, nbr, du := c.doDo(req, opts)
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "%d, %d, %d (%v)\n", rps, nbr, du, err)
	//}
	_, _, _ = rps, nbr, du
	return rsp, err
}

type apiClient struct {
	client   *Client
	protocol string
	host     string
	port     string
	root     string
	config   interface{}
}

func (apc apiClient) Url() string {
	protocol := "http"
	host := "localhost"
	port := ""
	root := "/"
	if apc.protocol != "" {
		protocol = apc.protocol
	}
	if apc.host != "" {
		host = apc.host
	}
	if apc.port != "" {
		port = ":" + apc.port
	}
	if apc.root != "" {
		if strings.HasPrefix(apc.root, "/") {
			root = apc.root
		} else {
			root = "/" + apc.root
		}
		if !strings.HasSuffix(root, "/") {
			root += "/"
		}
	}
	return fmt.Sprintf("%s://%s%s%s", protocol, host, port, root)
}

type mError struct {
	Error string `json:"error"`
}

type mErrorer interface {
	GetError() string
}

func (me mError) GetError() string { return me.Error }

func err2mError(err error) *mError {
	if err == nil {
		return nil
	}
	return &mError{Error: err.Error()}
}

func (apc *apiClient) DoAsJson(request *http.Request, outBody any) (*http.Response, error) {
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp, err := apc.client.Do(request, nil)
	if err = NewClientErr("DoAsJson", resp, err, nil); err != nil {
		return nil, err
	}
	bs, err := io.ReadAll(resp.Body)
	if err = NewClientErr("DoAsJson", resp, err, bs); err != nil {
		return nil, err
	}
	if outBody != nil {
		if err = json.Unmarshal(bs, outBody); err != nil {
			return nil, fmt.Errorf("in DoAsJson: %v", err)
		}
		me, ok := outBody.(mErrorer)
		if ok && me.GetError() != "" {
			return nil, fmt.Errorf("in DoAsJson: %s", me.GetError())
		}
	}
	return resp, nil
}

func (apc *apiClient) SimpleDoAsJson(method, url string, inBody any, outBody any) (*http.Response, error) {
	var (
		req *http.Request
		err error
	)
	if inBody == nil {
		req, err = http.NewRequest(method, url, nil)
	} else {
		reqBody, err := json.Marshal(inBody)
		if err != nil {
			return nil, fmt.Errorf("in SimpleDoAsJson: %v", err)
		}
		req, err = http.NewRequest(method, url, strings.NewReader(string(reqBody)))
	}
	if err != nil {
		return nil, fmt.Errorf("in SimpleDoAsJson: %v", err)
	}
	return apc.DoAsJson(req, outBody)
}

func (apc *apiClient) GetConfig() interface{} { return apc.config }

func (apc *apiClient) SetCabriHeader(h string) { apc.client.cabriHeader = h }

func NewWebApiClient(protocol string, host string, port string, tlsConfig *TlsConfig, root string, config interface{}, timeout time.Duration) (WebApiClient, error) {
	var (
		ht     *http.Transport
		client Client
	)
	if tlsConfig != nil {
		tlsClientConfig, err := getTlsClientConfig(tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("in NewWebApiClient: %v", err)
		}
		ht = &http.Transport{TLSClientConfig: tlsClientConfig}
		client = Client{
			Client:            http.Client{Timeout: timeout, Transport: ht},
			basicAuthUser:     tlsConfig.basicAuthUser,
			basicAuthPassword: tlsConfig.basicAuthPassword,
		}
	} else {
		client = Client{Client: http.Client{Timeout: timeout}}
	}
	return &apiClient{client: &client, protocol: protocol, host: host, port: port, root: root, config: config}, nil
}

func sTestSleep(c echo.Context) error {
	var msDuration uint
	echo.PathParamsBinder(c).Uint("msDuration", &msDuration)
	time.Sleep(time.Duration(msDuration) * time.Millisecond)
	return c.JSON(http.StatusOK, nil)
}

type testDataSink struct {
	size       int
	msDuration int
	count      int
	sink       []byte
}

func (d *testDataSink) Write(p []byte) (n int, err error) {
	if d.sink == nil {
		d.sink = make([]byte, d.size)
	}
	if d.count >= d.size {
		fmt.Printf("testDataSink.Write %d error\n", d.count)
		return 0, fmt.Errorf("count %d", d.count)
	}
	var i int
	for i = 0; i < len(p) && i+d.count < d.size; i++ {
		d.sink[i+d.count] = p[i]
	}
	d.count += i
	dn := time.Duration(d.msDuration) * time.Millisecond
	fmt.Printf("testDataSink.Write %d during %v\n", d.count, dn)
	time.Sleep(dn)
	return i, nil
}

func (d *testDataSink) Close() error {
	fmt.Printf("testDataSink.Close %d\n", d.count)
	d.count = 0
	d.sink = nil
	return nil
}

func sTestPostStream(c echo.Context) error {
	var size int
	echo.PathParamsBinder(c).Int("size", &size)
	var msDuration int
	echo.PathParamsBinder(c).Int("msDuration", &msDuration)
	sDs := testDataSink{size: size, msDuration: msDuration}
	req := c.Request()
	io.Copy(&sDs, req.Body)
	sDs.Close()
	return nil
}

func sTemplate(c echo.Context) error {
	return echo.NewHTTPError(http.StatusInternalServerError, "sTemplate: not yet implemented")
}

func WebTestServerConfigurator(e *echo.Echo, root string, _ map[string]interface{}) error {
	e.GET(root+"sleep/:msDuration", sTestSleep)
	e.POST(root+"postStream/:size/:msDuration", sTestPostStream)
	return nil
}
