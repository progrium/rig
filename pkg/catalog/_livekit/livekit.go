package livekit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/progrium/rig/pkg/catalog/host"
	"github.com/progrium/rig/pkg/catalog/obs"
	"github.com/progrium/rig/pkg/entity"
	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/node"
	"github.com/progrium/rig/pkg/util"
	"tractor.dev/hack/pkg/misc"

	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/webhook"
	lksdk "github.com/livekit/server-sdk-go"
	lksdk2 "github.com/livekit/server-sdk-go/v2"
)

func Instance() *node.Raw {
	webhookAddr := misc.ListenAddr()
	n := node.New("LiveKit Instance",
		node.Attrs{
			"activated": "false",
		},
		&host.Subprocess{Cmd: redisCmd()},
		&host.Subprocess{Cmd: serverCmd(webhookAddr)},
		&host.Subprocess{Cmd: ingressCmd()},
		&Provider{
			URL:         "http://localhost:7880",
			APIKey:      "devkey",
			SecretKey:   "secret",
			webhookAddr: webhookAddr,
		},
		&util.ObjectToggler{},
	)
	return n
}

func redisCmd() *exec.Cmd {
	cmdpath := "redis-server"
	binpath := os.Getenv("BINPATH")
	if binpath != "" {
		cmdpath = filepath.Join(binpath, "redis-server")
	}
	dirpath := "./"
	datapath := os.Getenv("DATAPATH")
	if datapath != "" {
		dirpath = datapath
	}
	cmd := exec.Command(cmdpath, "--dir", dirpath)
	// cmd.Stdout = &host.PrefixWriter{Prefix: "livekit-redis"}
	// cmd.Stdout = &host.PrefixWriter{Prefix: "livekit-redis"}
	cmd.Stderr = io.Discard
	cmd.Stderr = io.Discard
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}

func serverCmd(webhookAddr string) *exec.Cmd {
	config := fmt.Sprintf(`
bind_addresses:
  - "0.0.0.0"
redis:
  address: "localhost:6379"
ingress:
  rtmp_base_url: "rtmp://localhost:1935/live"
  whip_base_url: "http://localhost:8080/whip"
webhook:
  api_key: "devkey"
  urls:
    - "http://%s"
`, webhookAddr)
	cmdpath := "livekit-server"
	binpath := os.Getenv("BINPATH")
	if binpath != "" {
		cmdpath = filepath.Join(binpath, cmdpath)
	}
	cmd := exec.Command(cmdpath, "--config-body", config, "--dev")
	// cmd.Stdout = &host.PrefixWriter{Prefix: "livekit-server"}
	// cmd.Stderr = &host.PrefixWriter{Prefix: "livekit-server"}
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.Env = os.Environ()
	if binpath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", strings.TrimRight(binpath, "/")))
	} else {
		cmd.Env = append(cmd.Env, "PATH=/opt/homebrew/bin")
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}

func ingressCmd() *exec.Cmd {
	config := `
log_level: debug
api_key: devkey
api_secret: secret
ws_url: ws://localhost:7880
redis:
  address: localhost:6379
`
	cmdpath := "ingress"
	binpath := os.Getenv("BINPATH")
	if binpath != "" {
		cmdpath = filepath.Join(binpath, cmdpath)
	}
	cmd := exec.Command(cmdpath, "--config-body", config)
	// cmd.Stdout = &host.PrefixWriter{Prefix: "livekit-ingress"}
	// cmd.Stderr = &host.PrefixWriter{Prefix: "livekit-ingress"}
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.Env = os.Environ()
	if binpath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", strings.TrimRight(binpath, "/")))
	} else {
		cmd.Env = append(cmd.Env, "DYLD_LIBRARY_PATH=/opt/homebrew/lib")
		cmd.Env = append(cmd.Env, "PATH=/opt/homebrew/bin")
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}

type Provider struct {
	URL       string
	APIKey    string
	SecretKey string

	client *Client

	webhookAddr     string
	webhookReceiver *http.Server

	ingressProvider      *ingressProvider
	roomProvider         *roomProvider
	participantProviders map[string]*participantProvider
	chatRoom             *lksdk2.Room

	com entity.Node
}

type Client struct {
	*lksdk.RoomServiceClient
	*lksdk.IngressClient
}

func (c *Provider) Client() *Client {
	return c.client
}

func (c *Provider) ComponentAttached(com entity.Node) {
	c.com = com
}

func (c *Provider) OnEnabled() {
	timeout := 5 * time.Second

	var err error
	u, _ := url.Parse(c.URL)

	if err = misc.DialUntilOpen(u.Host, timeout); err != nil {
		log.Println("preconnect:", err)
		return
	}

	c.client = &Client{
		RoomServiceClient: lksdk.NewRoomServiceClient(c.URL, c.APIKey, c.SecretKey),
		IngressClient:     lksdk.NewIngressClient(c.URL, c.APIKey, c.SecretKey),
	}

	err = misc.TryUntilSuccess(timeout, 1*time.Second, func() error {
		_, err := c.client.ListRooms(context.TODO(), &livekit.ListRoomsRequest{})
		return err
	})
	if err != nil {
		log.Println("postconnect:", err)
		return
	}

	c.webhookReceiver = &http.Server{
		Addr:    c.webhookAddr,
		Handler: c,
	}
	go func() {
		if err := c.webhookReceiver.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println(err)
		}
	}()

	go func() {
		<-time.After(2 * time.Second)
		c.chatRoom, err = lksdk2.ConnectToRoom(c.URL, lksdk2.ConnectInfo{
			APIKey:              c.APIKey,
			APISecret:           c.SecretKey,
			RoomName:            "theater",
			ParticipantIdentity: "chat-bot",
		}, &lksdk2.RoomCallback{
			ParticipantCallback: lksdk2.ParticipantCallback{
				OnDataPacket: c.onDataPacket,
			},
		})
		if err != nil {
			log.Println("chat-bot:", err)
		}
	}()
}

func ptr(s string) *string {
	return &s
}

func (c *Provider) onDataPacket(data lksdk2.DataPacket, params lksdk2.DataReceiveParams) {
	m := make(map[string]any)
	err := json.Unmarshal(data.ToProto().Value.(*livekit.DataPacket_User).User.Payload, &m)
	if err != nil {
		panic(err)
	}
	log.Println("CHAT:", m["message"])
	if video, ok := detectYouTubeURL(m["message"].(string)); ok {
		client := node.Get[*obs.Client](entity.Root(c.com), node.Include{Children: true})
		_, err := client.Inputs.SetInputSettings(&inputs.SetInputSettingsParams{
			InputName: ptr("Browser"),
			InputSettings: map[string]any{
				"url": fmt.Sprintf("https://hopollo.github.io/OBS-Youtube-Player/?watch?v=%s&hideWhenStopped=true&quality=hd1080", video),
			},
		})
		if err != nil {
			log.Println(err)
		}
	}
}

func detectYouTubeURL(s string) (string, bool) {
	if !strings.HasPrefix(s, "https://www.youtube.com/") &&
		!strings.HasPrefix(s, "https://youtu.be/") &&
		!strings.HasPrefix(s, "https://youtube.com") {
		return "", false
	}
	s = strings.ReplaceAll(s, "https://www.youtube.com/watch?v=", "")
	s = strings.ReplaceAll(s, "https://youtu.be/", "")
	s = strings.ReplaceAll(s, "https://www.youtube.com/shorts/", "")
	s = strings.ReplaceAll(s, "https://youtube.com/shorts/", "")
	return s, true
}

func (c *Provider) OnDisabled() {
	c.client = nil
	if c.webhookReceiver != nil {
		if err := c.webhookReceiver.Shutdown(context.TODO()); err != nil {
			log.Println(err)
		}
		c.webhookReceiver = nil
	}
	if c.chatRoom != nil {
		c.chatRoom.Disconnect()
	}
}

func (c *Provider) IngressProvider() *ingressProvider {
	if c.ingressProvider == nil {
		c.ingressProvider = &ingressProvider{
			client: c.client.IngressClient,
		}
	}
	return c.ingressProvider
}

func (c *Provider) RoomProvider() *roomProvider {
	if c.roomProvider == nil {
		c.roomProvider = &roomProvider{
			client: c.client.RoomServiceClient,
		}
	}
	return c.roomProvider
}

func (c *Provider) ParticipantProvider(room string) *participantProvider {
	if c.participantProviders == nil {
		c.participantProviders = make(map[string]*participantProvider)
	}
	if p, ok := c.participantProviders[room]; ok {
		return p
	}
	p := &participantProvider{
		client: c.client.RoomServiceClient,
		room:   room,
	}
	c.participantProviders[room] = p
	return p
}

func (c *Provider) Nodes(com manifold.Node) entity.Nodes {
	if c.client == nil {
		return entity.Nodes{}
	}
	return entity.Nodes{
		node.NewID(path.Join(com.ID(), "rooms"), "Rooms", node.Attrs{"view": "livekit.RoomList"}, &RoomList{Provider: c}),
		node.NewID(path.Join(com.ID(), "ingress"), "Ingress", node.Attrs{"view": "livekit.IngressList"}, &IngressList{Provider: c}),
	}
}

func (c *Provider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authProvider := auth.NewSimpleKeyProvider(c.APIKey, c.SecretKey)
	_, err := webhook.ReceiveWebhookEvent(r, authProvider)
	if err != nil {
		log.Println(err)
		return
	}
	// TODO: route events!
	// log.Println("livekit:", event)
	if c.com != nil {
		p := entity.Parent(c.com) // we're assuming node layout
		if p != nil {
			node.Send(p.(entity.Node), "livekit-event", nil)
		}
	}
}
