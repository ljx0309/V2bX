package panel

import (
	"github.com/goccy/go-json"
	"regexp"
	"strconv"
	"strings"
)

type NodeInfo struct {
	Host            string            `json:"host"`
	ServerPort      int               `json:"server_port"`
	ServerName      string            `json:"server_name"`
	Network         string            `json:"network"`
	NetworkSettings json.RawMessage   `json:"networkSettings"`
	Cipher          string            `json:"cipher"`
	ServerKey       string            `json:"server_key"`
	Tls             int               `json:"tls"`
	Routes          []Route           `json:"routes"`
	BaseConfig      *BaseConfig       `json:"base_config"`
	Rules           []DestinationRule `json:"-"`
	localNodeConfig `json:"-"`
}
type Route struct {
	Id     int         `json:"id"`
	Match  interface{} `json:"match"`
	Action string      `json:"action"`
	//ActionValue interface{} `json:"action_value"`
}
type BaseConfig struct {
	PushInterval any `json:"push_interval"`
	PullInterval any `json:"pull_interval"`
}
type DestinationRule struct {
	ID      int
	Pattern *regexp.Regexp
}
type localNodeConfig struct {
	NodeId      int
	NodeType    string
	EnableVless bool
	EnableTls   bool
	SpeedLimit  int
	DeviceLimit int
}

func (c *Client) GetNodeInfo() (nodeInfo *NodeInfo, err error) {
	const path = "/api/v1/server/UniProxy/config"
	r, err := c.client.R().Get(path)
	if err = c.checkResponse(r, path, err); err != nil {
		return
	}
	err = json.Unmarshal(r.Body(), &nodeInfo)
	if err != nil {
		return
	}
	if c.etag == r.Header().Get("ETag") { // node info not changed
		return nil, nil
	}
	nodeInfo.NodeId = c.NodeId
	nodeInfo.NodeType = c.NodeType
	for i := range nodeInfo.Routes { // parse rules from routes
		if nodeInfo.Routes[i].Action == "block" {
			var matchs []string
			if _, ok := nodeInfo.Routes[i].Match.(string); ok {
				matchs = strings.Split(nodeInfo.Routes[i].Match.(string), ",")
			} else {
				matchs = nodeInfo.Routes[i].Match.([]string)
			}
			for _, v := range matchs {
				nodeInfo.Rules = append(nodeInfo.Rules, DestinationRule{
					ID:      nodeInfo.Routes[i].Id,
					Pattern: regexp.MustCompile(v),
				})
			}
		}
	}
	nodeInfo.Routes = nil
	if _, ok := nodeInfo.BaseConfig.PullInterval.(int); !ok {
		i, _ := strconv.Atoi(nodeInfo.BaseConfig.PullInterval.(string))
		nodeInfo.BaseConfig.PullInterval = i
	}
	if _, ok := nodeInfo.BaseConfig.PushInterval.(int); !ok {
		i, _ := strconv.Atoi(nodeInfo.BaseConfig.PushInterval.(string))
		nodeInfo.BaseConfig.PushInterval = i
	}
	c.etag = r.Header().Get("Etag")
	return
}
