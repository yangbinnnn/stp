package stp

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type STPCmd struct {
	CmdType string `json:"type"`
	Data    []byte `json:"data"`
}

type STPLoginData struct {
	AuthKey string `json:"authKey"`
	Name    string `json:"name"`
}

type STPHBData struct {
	Msg string `json:"msg"`
}

type STPClient struct {
	authKey   string
	serverUrl string
	localPort string
	name      string

	conn   *websocket.Conn
	tunnel *SSHtunnel
}

func NewSTPClient(authKey, serverUrl, localPort, name string) *STPClient {
	client := &STPClient{
		authKey:   authKey,
		serverUrl: serverUrl,
		localPort: localPort,
		name:      name,
	}
	return client
}

func (s *STPClient) Login() error {
	loginCmd := &STPLoginData{AuthKey: s.authKey, Name: s.name}
	data, err := json.Marshal(loginCmd)
	if err != nil {
		return err
	}
	cmd := STPCmd{
		CmdType: "login",
		Data:    data,
	}
	err = s.connect()
	if err != nil {
		return err
	}
	err = s.conn.WriteJSON(cmd)
	if err != nil {
		return err
	}
	resp := &STPResp{}
	err = s.conn.ReadJSON(resp)
	if err != nil {
		return err
	}
	if resp.Status != 200 {
		log.Println(resp.ErrMsg, resp.Status)
		return fmt.Errorf("Status: %d, ErrMsg: %s", resp.Status, resp.ErrMsg)
	}
	sshUser, ok := resp.Data["sshUser"].(string)
	if !ok {
		return errors.New("invalid ssh user resp")
	}
	sshAddr, ok := resp.Data["sshAddr"].(string)
	if !ok {
		return errors.New("invalid ssh addr resp")
	}
	assginPort, ok := resp.Data["port"].(string)
	if !ok {
		return errors.New("invalid port resp")
	}
	privateKey, ok := resp.Data["privateKey"].(string)
	if !ok {
		return errors.New("invalid privateKey resp")
	}
	publicKey, ok := resp.Data["publicKey"].(string)
	if !ok {
		return errors.New("invalid publicKey resp")
	}

	log.Println("ssh user:", sshUser)
	log.Println("ssh addr:", sshAddr)
	log.Println("assgin port:", assginPort)
	err = AddAuthorizedKey(publicKey, "")
	if err != nil {
		log.Println("add authorized key error", err.Error())
	}

	go s.StartSSHTunnel(sshUser, sshAddr, assginPort, privateKey)
	return nil
}

func (s *STPClient) StartSSHTunnel(sshUser string, sshAddr string, assginPort string, privateKey string) {
	local := &Endpoint{
		"localhost",
		s.localPort,
	}
	items := strings.Split(sshAddr, ":")
	server := &Endpoint{
		items[0],
		items[1],
	}
	remote := &Endpoint{
		"0.0.0.0",
		assginPort,
	}
	authMethod, err := privateKeyAuthMethod(privateKey)
	if err != nil {
		log.Println("ssh load private auth key error", err.Error())
		return
	}
	sshConfig := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	s.tunnel = &SSHtunnel{
		Local:    local,
		Server:   server,
		Remote:   remote,
		Config:   sshConfig,
		StopConn: make(chan bool),
	}

	// block
	err = s.tunnel.Start()
	if err != nil {
		log.Println("start ssh tunnel error", err.Error())
	}
	log.Println("tunnel end")
	return
}

func (s *STPClient) connect() error {
	if s.conn != nil {
		s.conn.Close()
	}
	wsconn, _, err := websocket.DefaultDialer.Dial(s.serverUrl, nil)
	if err != nil {
		return err
	}
	s.conn = wsconn
	return nil
}

func (s *STPClient) Daemon() {
	for {
		cmd := &STPCmd{}
		err := s.conn.ReadJSON(cmd)
		if err != nil {
			if _, ok := err.(*websocket.CloseError); ok {
				s.Relogin()
				continue
			} else {
				log.Println("deamon", err.Error())
			}
		}
		// handler cmd
		switch cmd.CmdType {
		case "heartBeat":
			s.SendHeartBeat(s.conn)
		case "relogin":
			// remote server tunnel stop, handler server relogin cmd
			m := make(map[string]interface{})
			err := json.Unmarshal(cmd.Data, &m)
			msg := ""
			if err != nil {
				msg, _ = m["msg"].(string)
			}
			log.Println("recived the relogin cmd, msg", msg)
			s.Relogin()
			continue
		}
	}
}

func (s *STPClient) Relogin() {
	log.Println("relogin")
	if s.tunnel != nil {
		s.tunnel.Stop()
	}
	for {
		err := s.Login()
		if err == nil {
			break
		} else {
			log.Println("login error", err.Error())
		}
		time.Sleep(3 * time.Second)
	}
}

func (s *STPClient) SendHeartBeat(c *websocket.Conn) error {
	hb := STPHBData{
		Msg: "Ping",
	}
	hbdata, _ := json.Marshal(hb)
	cmd := STPCmd{
		CmdType: "heartBeat",
		Data:    hbdata,
	}
	return c.WriteJSON(cmd)
}

type Client struct {
	Name       string `json:"name"`
	Port       string `json:"port"`
	Addr       string `json:"addr"`
	LoginTime  int64  `json:"loginTime"`
	OnlineTime int64  `json:"onlineTime"`
	IsOnline   bool   `json:"isOnline"`
	conn       *websocket.Conn
}

type ClientManager struct {
	clients []*Client
	lock    sync.Mutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{clients: []*Client{}}
}

func (cm *ClientManager) AddClient(cli *Client) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	cm.clients = append(cm.clients, cli)
}

func (cm *ClientManager) DelClientByIdx(idx int) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	if idx >= len(cm.clients) || len(cm.clients) == 0 {
		return
	}
	lastIdx := len(cm.clients) - 1
	cm.clients[idx] = cm.clients[lastIdx]
	cm.clients[lastIdx] = nil
	cm.clients = cm.clients[:lastIdx]
}

type STPServer struct {
	authKey    string
	listenAddr string
	privateKey string
	publicKey  string
	sshUser    string
	sshAddr    string
	portMgr    *PortManager
	cliMgr     *ClientManager
}

func NewSTPServer(authKey, listenAddr, sshAddr, privateKey, publicKey, sshUser, portRange string) *STPServer {
	startEnd := strings.Split(portRange, "-")
	if len(startEnd) != 2 {
		log.Fatalln("invalid port range,", portRange)
	}
	startInt, err := strconv.Atoi(startEnd[0])
	if err != nil {
		log.Fatalln(err.Error())
	}
	endInt, err := strconv.Atoi(startEnd[1])
	if err != nil {
		log.Fatalln(err.Error())
	}
	portMgr := NewPortManager(startInt, endInt)
	cliMgr := NewClientManager()

	return &STPServer{
		authKey:    authKey,
		listenAddr: listenAddr,
		portMgr:    portMgr,
		privateKey: privateKey,
		publicKey:  publicKey,
		sshUser:    sshUser,
		sshAddr:    sshAddr,
		cliMgr:     cliMgr,
	}
}

func (s *STPServer) Start() {
	go s.checker()
	http.HandleFunc("/", s.WsHandler)
	http.HandleFunc("/showClient", s.ShowClientHandler)
	log.Println("listen on", s.listenAddr)
	log.Fatal(http.ListenAndServe(s.listenAddr, nil))
}

func (s *STPServer) checker() {
	for {
		time.Sleep(10 * time.Second)
		offlineIdxs := []int{}
		for idx, client := range s.cliMgr.clients {
			if client == nil {
				continue
			}
			port, err := strconv.Atoi(client.Port)
			if err != nil {
				continue
			}
			if online, err := s.portMgr.PingPort(port); !online {
				log.Printf("[check] %d port %s offline, err %s", idx, client.Port, err.Error())
				client.IsOnline = false
				s.portMgr.ReleasePort(client.Port)
				s.SendRelogin(client.conn, fmt.Sprintf("check port %s offline", client.Port))
				offlineIdxs = append(offlineIdxs, idx)
				continue
			} else {
				client.IsOnline = true
				client.OnlineTime = time.Now().Unix() - client.LoginTime
			}
			s.SendHeartBeat(client.conn)
		}

		// must iterate it backwards
		for idx := len(offlineIdxs) - 1; idx >= 0; idx-- {
			s.cliMgr.DelClientByIdx(offlineIdxs[idx])
		}
	}
}

type STPResp struct {
	Status int                    `json:"status"`
	ErrMsg string                 `json:"errMsg"`
	Data   map[string]interface{} `json:"data"`
}

func NewBadRequestError(msg string) STPResp {
	resp := STPResp{
		Status: 400,
		ErrMsg: msg,
	}
	return resp
}

func (s *STPServer) ShowClientHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.cliMgr.clients)
}

var upgrader = websocket.Upgrader{}

func (s *STPServer) WsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err.Error())
		return
	}
	defer func() {
		log.Println("client disconnect:", c.RemoteAddr().String())
		c.Close()
	}()
	log.Println("client connect:", c.RemoteAddr().String())

	for {
		cmd := STPCmd{}
		err := c.ReadJSON(&cmd)
		if err != nil {
			c.WriteJSON(NewBadRequestError(err.Error()))
			return
		}
		switch cmd.CmdType {
		case "login":
			log.Println("on login")
			client, err := s.OnLogin(c, cmd.Data)
			if err != nil {
				log.Println("onlogin error:", err.Error())
				c.WriteJSON(NewBadRequestError(err.Error()))
			} else {
				s.cliMgr.AddClient(client)
			}
		case "heartBeat":
			// do nothing
		default:
			c.WriteJSON(NewBadRequestError("Unknow CMD"))
		}
	}
}

func (s *STPServer) SendHeartBeat(c *websocket.Conn) error {
	hb := STPHBData{
		Msg: "Ping",
	}
	hbdata, _ := json.Marshal(hb)
	cmd := STPCmd{
		CmdType: "heartBeat",
		Data:    hbdata,
	}
	return c.WriteJSON(cmd)
}

func (s *STPServer) SendRelogin(c *websocket.Conn, msg string) error {
	m := make(map[string]interface{})
	m["msg"] = msg
	data, _ := json.Marshal(m)
	cmd := STPCmd{
		CmdType: "relogin",
		Data:    data,
	}
	return c.WriteJSON(cmd)
}

func (s *STPServer) OnLogin(c *websocket.Conn, data []byte) (*Client, error) {
	loginData := STPLoginData{}
	err := json.Unmarshal(data, &loginData)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	if loginData.AuthKey != s.authKey {
		log.Println("invalid auth key")
		return nil, errors.New("invalid auth key")
	}

	port := s.portMgr.AssginPort()
	if port == "" {
		log.Println("port not enough")
		return nil, errors.New("port not enough")
	}

	log.Println("login 200")
	respData := make(map[string]interface{})
	respData["port"] = port
	respData["privateKey"] = s.privateKey
	respData["publicKey"] = s.publicKey
	respData["sshUser"] = s.sshUser
	respData["sshAddr"] = s.sshAddr
	resp := STPResp{
		Status: 200,
		Data:   respData,
	}
	err = c.WriteJSON(resp)
	if err != nil {
		return nil, err
	}
	cli := &Client{
		Name:      loginData.Name,
		Port:      port,
		Addr:      c.RemoteAddr().String(),
		LoginTime: time.Now().Unix(),
		conn:      c,
	}
	return cli, nil
}

type Port struct {
	port int
	used bool
}

type PortManager struct {
	StartPort int
	EndPort   int
	ports     []*Port
	idx       int
	lock      sync.Mutex
}

func NewPortManager(startPort, endPort int) *PortManager {
	ports := []*Port{}
	for i := startPort; i <= endPort; i++ {
		ports = append(ports, &Port{i, false})
	}
	log.Printf("port range %d-%d\n", startPort, endPort)
	return &PortManager{StartPort: startPort, EndPort: endPort, ports: ports, idx: 0}
}

// AssginPort 给远程客户端分配绑定端口
func (pm *PortManager) AssginPort() string {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	startIdx := pm.idx
	for i := pm.idx + 1; i != startIdx; i++ {
		// loop and keep in range
		i = i % len(pm.ports)
		p := pm.ports[i]
		if p.used {
			continue
		}
		// double check
		if online, _ := pm.PingPort(p.port); online {
			p.used = true
			continue
		}
		pm.idx = i
		p.used = true
		return strconv.Itoa(p.port)
	}
	// not enough
	return ""
}

// ReleasePort 释放端口
func (pm *PortManager) ReleasePort(port string) {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return
	}
	// out of range
	if portInt < pm.StartPort || portInt > pm.EndPort {
		return
	}
	idx := portInt - pm.StartPort
	p := pm.ports[idx]
	if p.port != portInt {
		log.Println("cacl port idx error", idx, portInt, p.port)
		return
	}
	// release
	p.used = false
	log.Println("port released", p.port)
	return
}

func (pm *PortManager) PingPort(port int) (online bool, err error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 3*time.Second)
	if err != nil {
		online = false
		return
	}
	defer conn.Close()
	online = true
	return
}
