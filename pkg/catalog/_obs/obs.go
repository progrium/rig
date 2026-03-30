package obs

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/config"
	"github.com/andreykaipov/goobs/api/requests/general"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/progrium/rig/pkg/catalog/host"
	"github.com/progrium/rig/pkg/entity"
	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/node"
	"github.com/progrium/rig/pkg/util"
	"gopkg.in/ini.v1"
	"tractor.dev/hack/pkg/debouncer"
	"tractor.dev/hack/pkg/misc"
)

func Instance() *node.Raw {
	prefix := os.Getenv("OBS_PATH")
	var cmd *exec.Cmd
	if prefix == "" {
		cmd = exec.Command("/Applications/OBS.app/Contents/MacOS/OBS", "--minimize-to-tray", "--disable-shutdown-check")
	} else {
		cmd = exec.Command(filepath.Join(prefix, "obs"), "--portable", "--minimize-to-tray", "--disable-shutdown-check")
	}
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	configpath := "/Users/progrium/Library/Application Support"
	datapath := os.Getenv("DATAPATH")
	if datapath != "" {
		configpath = datapath
	}
	cfg, err := ini.Load(filepath.Join(configpath, "obs-studio/global.ini"))
	if err != nil {
		log.Fatal(err)
	}
	enabled, err := cfg.Section("OBSWebSocket").Key("ServerEnabled").Bool()
	if err != nil {
		log.Fatal(err)
	}
	if !enabled {
		log.Fatal("obs websocket not enabled")
	}
	port := cfg.Section("OBSWebSocket").Key("ServerPort").String()
	passwd := cfg.Section("OBSWebSocket").Key("ServerPassword").String()
	client := &Client{
		Addr:     fmt.Sprintf(":%s", port),
		Password: passwd,
	}
	n := node.New("OBS Instance",
		node.Attrs{
			"activated": "false",
		},
		&host.Subprocess{Cmd: cmd},
		client,
		// &Configuration{
		// 	Client:          client,
		// 	Profile:         "debug",
		// 	SceneCollection: "debug",
		// },
		&util.ObjectToggler{},
	)
	return n
}

type Client struct {
	*goobs.Client
	Addr          string
	Password      string
	StudioVersion string
	ServerVersion string

	com entity.Node
}

func (c *Client) OnEnabled() {
	timeout := 5 * time.Second

	var err error
	if err = misc.DialUntilOpen(c.Addr, timeout); err != nil {
		log.Println("preconnect:", err)
		return
	}

	c.Client, err = goobs.New(c.Addr, goobs.WithPassword(c.Password))
	if err != nil {
		log.Println("connect:", err)
		return
	}

	var version *general.GetVersionResponse
	err = misc.TryUntilSuccess(timeout, 1*time.Second, func() error {
		var err error
		version, err = c.Client.General.GetVersion()
		return err
	})
	if err != nil {
		log.Println("postconnect:", err)
		return
	}

	debounce := debouncer.New(500 * time.Millisecond)
	go c.Client.Listen(func(event any) {
		if c.com != nil {
			p := entity.Parent(c.com) // we're assuming node layout
			if p != nil {
				debounce(func() {
					node.Send(p.(entity.Node), "obs-event", nil)
				})

			}

		}
		// switch e := event.(type) {
		// case *events.InputCreated,
		// 	*events.InputRemoved,
		// 	*events.InputNameChanged,
		// 	*events.InputSettingsChanged: // more
		// 	log.Println("input:", e)
		// case *events.SceneCreated,
		// 	*events.CurrentPreviewSceneChanged,
		// 	*events.CurrentProgramSceneChanged,
		// 	*events.SceneListChanged,
		// 	*events.SceneNameChanged,
		// 	*events.SceneRemoved: // more
		// 	log.Println("scene:", e)
		// case *events.SceneItemCreated,
		// 	*events.SceneItemEnableStateChanged,
		// 	*events.SceneItemListReindexed,
		// 	*events.SceneItemLockStateChanged,
		// 	*events.SceneItemRemoved,
		// 	*events.SceneItemSelected,
		// 	*events.SceneItemTransformChanged:
		// 	log.Println("sceneItem:", e)
		// default:
		// 	// no-op
		// }
	})

	c.StudioVersion = version.ObsVersion
	c.ServerVersion = version.ObsWebSocketVersion
}

func (c *Client) ComponentAttached(com entity.Node) {
	c.com = com
}

func (c *Client) OnDisabled() {
	if c.Client != nil {
		c.Client.Disconnect()
		c.Client = nil
	}
}

func (c *Client) Nodes(com manifold.Node) entity.Nodes {
	if c.Client == nil {
		return entity.Nodes{}
	}
	return entity.Nodes{
		node.NewID(path.Join(com.ID(), "profiles"), "Profiles",
			node.Attrs{"view": "obs.ProfileList"},
			&ProfileList{Client: c},
		),
		node.NewID(path.Join(com.ID(), "scenes"), "Scene Collections",
			node.Attrs{"view": "obs.SceneCollectionList"},
			&SceneCollectionList{Client: c},
		),
		node.NewID(path.Join(com.ID(), "inputs"), "Input Types",
			node.Attrs{"view": "obs.InputTypeList"},
			&InputTypeList{Client: c},
		),
	}
}

type ProfileList struct {
	Client *Client
}

func (p *ProfileList) Nodes(com manifold.Node) (nodes entity.Nodes) {
	lst, err := p.Client.Config.GetProfileList()
	if err != nil {
		log.Println(err)
		return
	}
	for _, p := range lst.Profiles {
		if p == lst.CurrentProfileName {
			nodes = append(nodes, node.NewID(path.Join(com.ID(), p), p, node.Attrs{"desc": "active"}))
		} else {
			nodes = append(nodes, node.NewID(path.Join(com.ID(), p), p))
		}
	}
	return
}

type SceneCollectionList struct {
	Client *Client
}

func (sc *SceneCollectionList) Nodes(com manifold.Node) (nodes entity.Nodes) {
	lst, err := sc.Client.Config.GetSceneCollectionList()
	if err != nil {
		log.Println(err)
		return
	}
	for _, c := range lst.SceneCollections {
		if c == lst.CurrentSceneCollectionName {
			nodes = append(nodes, node.NewID(path.Join(com.ID(), c), c, node.Attrs{"desc": "active"}))
		} else {
			nodes = append(nodes, node.NewID(path.Join(com.ID(), c), c))
		}
	}
	return
}

type InputTypeList struct {
	Client *Client
}

func (it *InputTypeList) Nodes(com manifold.Node) (nodes entity.Nodes) {
	lst, err := it.Client.Inputs.GetInputKindList()
	if err != nil {
		log.Println(err)
		return
	}
	for _, i := range lst.InputKinds {
		nodes = append(nodes, node.NewID(path.Join(com.ID(), i), i))
	}
	return
}

type Configuration struct {
	Client          *Client
	Profile         string
	SceneCollection string
}

func (c *Configuration) OnEnabled() {
	params1 := config.NewSetCurrentProfileParams().WithProfileName(c.Profile)
	_, err := c.Client.Config.SetCurrentProfile(params1)
	if err != nil {
		log.Println("config:", err)
		return
	}
	params2 := config.NewSetCurrentSceneCollectionParams().WithSceneCollectionName(c.SceneCollection)
	_, err = c.Client.Config.SetCurrentSceneCollection(params2)
	if err != nil {
		log.Println("config:", err)
		return
	}
}

func (c *Configuration) OnDisabled() {
}

func (c *Configuration) Nodes(com manifold.Node) entity.Nodes {
	if c.Client.Client == nil {
		return entity.Nodes{}
	}
	stream, err := c.Client.Config.GetStreamServiceSettings()
	if err != nil {
		log.Println(err)
	}
	return entity.Nodes{
		node.NewID(path.Join(com.ID(), "scenes"), "Scenes",
			node.Attrs{"view": "obs.SceneList"},
			&SceneList{
				Client: c.Client,
				provider: &sceneProvider{
					client: c.Client.Client,
				},
			},
		),
		node.NewID(path.Join(com.ID(), "inputs"), "Sources",
			node.Attrs{"view": "obs.InputList"},
			&InputList{Client: c.Client},
		),
		node.NewID(path.Join(com.ID(), "outputs"), "Outputs",
			node.Attrs{"view": "obs.OutputList"},
			&OutputList{Client: c.Client},
		),
		node.NewID(path.Join(com.ID(), "stream"), "Stream",
			node.Attrs{"view": "fields"},
			stream.StreamServiceSettings,
		),
	}
}

type SceneItemList struct {
	Client    *Client
	SceneName string
}

func (sil *SceneItemList) Nodes(com manifold.Node) (nodes entity.Nodes) {
	params := sceneitems.NewGetSceneItemListParams().WithSceneName(sil.SceneName)
	lst, err := sil.Client.SceneItems.GetSceneItemList(params)
	if err != nil {
		log.Println(err)
		return
	}
	for _, s := range lst.SceneItems {
		nodes = append(nodes, node.NewID(path.Join(com.ID(), s.SourceName), s.SourceName,
			node.Attrs{"desc": s.InputKind},
			node.Children{
				node.NewID(path.Join(com.ID(), s.SourceName, "item"), "Item", node.Attrs{"view": "fields"}, s),
				node.NewID(path.Join(com.ID(), s.SourceName, "props"), "Source", node.Children{
					node.New("TODO"),
				}),
				node.NewID(path.Join(com.ID(), s.SourceName, "filters"), "Filters", node.Children{
					node.New("TODO"),
				}),
			},
		))
	}
	return
}

type InputList struct {
	Client *Client
}

func (il *InputList) Nodes(com manifold.Node) (nodes entity.Nodes) {
	lst, err := il.Client.Inputs.GetInputList()
	if err != nil {
		log.Println(err)
		return
	}
	for _, i := range lst.Inputs {
		uniquekeys := make(map[string]bool)
		dflt, err := il.Client.Inputs.GetInputDefaultSettings(inputs.NewGetInputDefaultSettingsParams().WithInputKind(i.InputKind))
		log.Println(string(dflt.GetRaw()))
		if err != nil {
			log.Println(err)
		}
		for k := range dflt.DefaultInputSettings {
			uniquekeys[k] = true
		}
		settings, err := il.Client.Inputs.GetInputSettings(inputs.NewGetInputSettingsParams().WithInputName(i.InputName))
		log.Println(string(settings.GetRaw()))
		if err != nil {
			log.Println(err)
		}
		for k := range settings.InputSettings {
			uniquekeys[k] = true
		}
		var keys []string
		for k := range uniquekeys {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var children node.Children
		for _, key := range keys {
			var v any
			if vv, ok := dflt.DefaultInputSettings[key]; ok {
				v = vv
			}
			if vv, ok := settings.InputSettings[key]; ok {
				v = vv
			}
			children = append(children, node.NewID(path.Join(com.ID(), i.InputName, key), key,
				node.Attrs{"desc": fmt.Sprintf("%#v", v)}))
		}
		nodes = append(nodes, node.NewID(path.Join(com.ID(), i.InputName), i.InputName,
			node.Attrs{"desc": i.InputKind},
			children,
		))
	}
	return
}

type OutputList struct {
	Client *Client
}

func (ol *OutputList) Nodes(com manifold.Node) (nodes entity.Nodes) {
	lst, err := ol.Client.Outputs.GetOutputList()
	if err != nil {
		log.Println(err)
		return
	}
	for _, o := range lst.Outputs {
		nodes = append(nodes, node.NewID(path.Join(com.ID(), o.Name), o.Name,
			node.Attrs{"desc": o.Kind, "view": "fields"},
			o,
		))
	}
	return
}
