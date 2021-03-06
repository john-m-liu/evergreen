package route

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/db"
	"github.com/evergreen-ci/evergreen/model/distro"
	"github.com/evergreen-ci/evergreen/model/task"
	"github.com/evergreen-ci/evergreen/model/user"
	"github.com/evergreen-ci/evergreen/rest/data"
	"github.com/evergreen-ci/evergreen/rest/model"
	"github.com/evergreen-ci/evergreen/testutil"
	"github.com/evergreen-ci/gimlet"
	"github.com/stretchr/testify/suite"
)

////////////////////////////////////////////////////////////////////////
//
// Tests for GET distro by id

type DistroByIDSuite struct {
	sc   *data.MockConnector
	data data.MockDistroConnector
	rm   gimlet.RouteHandler

	suite.Suite
}

func TestDistroSuite(t *testing.T) {
	db.SetGlobalSessionProvider(testutil.TestConfig().SessionFactory())
	suite.Run(t, new(DistroByIDSuite))
}

func (s *DistroByIDSuite) SetupSuite() {
	s.data = data.MockDistroConnector{
		CachedDistros: []*distro.Distro{
			{Id: "distro1"},
			{Id: "distro2"},
		},
		CachedTasks: []task.Task{
			{Id: "task1"},
			{Id: "task2"},
		},
	}
	s.sc = &data.MockConnector{
		MockDistroConnector: s.data,
	}
}

func (s *DistroByIDSuite) SetupTest() {
	s.rm = makeGetDistroByID(s.sc)
}

func (s *DistroByIDSuite) TestFindByIdFound() {
	s.rm.(*distroIDGetHandler).distroID = "distro1"

	resp := s.rm.Run(context.TODO())
	s.NotNil(resp)
	s.Equal(resp.Status(), http.StatusOK)
	s.NotNil(resp.Data())

	d, ok := (resp.Data()).(*model.APIDistro)
	s.True(ok)
	s.Equal(model.ToAPIString("distro1"), d.Name)
}

func (s *DistroByIDSuite) TestFindByIdFail() {
	s.rm.(*distroIDGetHandler).distroID = "distro3"

	resp := s.rm.Run(context.TODO())
	s.NotNil(resp)
	s.NotEqual(resp.Status(), http.StatusOK)
}

///////////////////////////////////////////////////////////////////////
//
// Tests for PUT distro by id

type DistroPutSuite struct {
	sc       *data.MockConnector
	data     data.MockDistroConnector
	rm       gimlet.RouteHandler
	settings *evergreen.Settings

	suite.Suite
}

func TestDistroPutSuite(t *testing.T) {
	db.SetGlobalSessionProvider(testutil.TestConfig().SessionFactory())
	suite.Run(t, new(DistroPutSuite))
}

func (s *DistroPutSuite) SetupTest() {
	s.data = data.MockDistroConnector{
		CachedDistros: []*distro.Distro{
			{
				Id: "distro1",
			},
			{
				Id: "distro2",
			},
			{
				Id: "distro3",
			},
		},
	}
	s.sc = &data.MockConnector{
		MockDistroConnector: s.data,
	}
	s.settings = &evergreen.Settings{}
	s.rm = makePutDistro(s.sc, s.settings)
}

func (s *DistroPutSuite) TestParse() {
	ctx := context.Background()
	json := []byte(`{"arch": "linux_amd64", "work_dir": "/data/mci", "ssh_key": "SSH string", "provider": "mock", "user": "tibor"}`)

	req, _ := http.NewRequest("PUT", "http://example.com/api/rest/v2/distros/distro4", bytes.NewBuffer(json))
	err := s.rm.Parse(ctx, req)
	s.NoError(err)
}

func (s *DistroPutSuite) TestRunNewWithValidEntity() {
	ctx := context.Background()
	json := []byte(`{"arch": "linux_amd64", "work_dir": "/data/mci", "ssh_key": "SSH Key", "provider": "mock", "user": "tibor"}`)
	h := s.rm.(*distroPutHandler)
	h.distroID = "distro4"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	s.Equal(apiDistro.Arch, model.ToAPIString("linux_amd64"))
	s.Equal(apiDistro.WorkDir, model.ToAPIString("/data/mci"))
	s.Equal(apiDistro.SSHKey, model.ToAPIString("SSH Key"))
	s.Equal(apiDistro.Provider, model.ToAPIString("mock"))
	s.Equal(apiDistro.User, model.ToAPIString("tibor"))
}

func (s *DistroPutSuite) TestRunNewWithInValidEntity() {
	ctx := context.Background()
	json := []byte(`{"arch": "linux_amd64", "work_dir": "/data/mci", "ssh_key": "", "provider": "mock", "user": "tibor"}`)
	h := s.rm.(*distroPutHandler)
	h.distroID = "distro4"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusBadRequest)
	error := (resp.Data()).(gimlet.ErrorResponse)
	//s.Equal(error.Message, "distro 'ssh_key' cannot be blank")
	s.Equal(error.Message, "ERROR: distro 'ssh_key' cannot be blank")
}

func (s *DistroPutSuite) TestRunNewConflictingName() {
	ctx := context.Background()
	json := []byte(`{"name": "distro5", "arch": "linux_amd64", "work_dir": "/data/mci", "ssh_key": "", "provider": "mock", "user": "tibor"}`)
	h := s.rm.(*distroPutHandler)
	h.distroID = "distro4"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusBadRequest)
	error := (resp.Data()).(gimlet.ErrorResponse)
	s.Equal(error.Message, fmt.Sprintf("Distro name is immutable; cannot rename distro resource '%s'", h.distroID))
}

func (s *DistroPutSuite) TestRunExistingWithValidEntity() {
	ctx := context.Background()
	json := []byte(`{"arch": "linux_amd64", "work_dir": "/data/mci", "ssh_key": "SSH Key", "provider": "mock", "user": "tibor"}`)
	h := s.rm.(*distroPutHandler)
	h.distroID = "distro3"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	s.Equal(apiDistro.Arch, model.ToAPIString("linux_amd64"))
	s.Equal(apiDistro.WorkDir, model.ToAPIString("/data/mci"))
	s.Equal(apiDistro.SSHKey, model.ToAPIString("SSH Key"))
	s.Equal(apiDistro.Provider, model.ToAPIString("mock"))
	s.Equal(apiDistro.User, model.ToAPIString("tibor"))
}

func (s *DistroPutSuite) TestRunExistingWithInValidEntity() {
	ctx := context.Background()
	json := []byte(`{"arch": "", "work_dir": "/data/mci", "ssh_key": "SSH Key", "provider": "", "user": ""}`)
	h := s.rm.(*distroPutHandler)
	h.distroID = "distro3"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusBadRequest)
	error := (resp.Data()).(gimlet.ErrorResponse)
	//s.Equal(error.Message, "distro 'arch' cannot be blank, distro 'user' cannot be blank, distro 'provider' cannot be blank")
	s.Equal(error.Message, "ERROR: distro 'arch' cannot be blank\nERROR: distro 'user' cannot be blank\nERROR: distro 'provider' cannot be blank")
}

func (s *DistroPutSuite) TestRunExistingConflictingName() {
	ctx := context.Background()
	json := []byte(`{"name": "distro5", "arch": "linux_amd64", "work_dir": "/data/mci", "ssh_key": "", "provider": "mock", "user": "tibor"}`)
	h := s.rm.(*distroPutHandler)
	h.distroID = "distro3"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusBadRequest)
	error := (resp.Data()).(gimlet.ErrorResponse)
	s.Equal(error.Message, fmt.Sprintf("Distro name is immutable; cannot rename distro resource '%s'", h.distroID))
}

////////////////////////////////////////////////////////////////////////
//
// Tests for DELETE distro by id

type DistroDeleteByIDSuite struct {
	sc   *data.MockConnector
	data data.MockDistroConnector
	rm   gimlet.RouteHandler

	suite.Suite
}

func TestDistroDeleteSuite(t *testing.T) {
	db.SetGlobalSessionProvider(testutil.TestConfig().SessionFactory())
	suite.Run(t, new(DistroDeleteByIDSuite))
}

func (s *DistroDeleteByIDSuite) SetupTest() {
	s.data = data.MockDistroConnector{
		CachedDistros: []*distro.Distro{
			{
				Id: "distro1",
			},
			{
				Id: "distro2",
			},
			{
				Id: "distro3",
			},
		},
	}
	s.sc = &data.MockConnector{
		MockDistroConnector: s.data,
	}
	s.rm = makeDeleteDistroByID(s.sc)
}

func (s *DistroDeleteByIDSuite) TestParse() {
	ctx := context.Background()

	req, _ := http.NewRequest("DELETE", "http://example.com/api/rest/v2/distros/distro1", nil)
	err := s.rm.Parse(ctx, req)
	s.NoError(err)
}

func (s *DistroDeleteByIDSuite) TestRunValidDistroId() {
	ctx := context.Background()
	h := s.rm.(*distroIDDeleteHandler)
	h.distroID = "distro1"

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)
}

func (s *DistroDeleteByIDSuite) TestRunInValidDistroId() {
	ctx := context.Background()
	h := s.rm.(*distroIDDeleteHandler)
	h.distroID = "distro4"

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusNotFound)
	error := (resp.Data()).(gimlet.ErrorResponse)
	s.Equal(error.Message, fmt.Sprintf("distro with id '%s' not found", h.distroID))
}

////////////////////////////////////////////////////////////////////////
//
// Tests for PATCH distro by id

type DistroPatchByIDSuite struct {
	sc       *data.MockConnector
	data     data.MockDistroConnector
	rm       gimlet.RouteHandler
	settings *evergreen.Settings

	suite.Suite
}

func TestDistroPatchSuite(t *testing.T) {
	db.SetGlobalSessionProvider(testutil.TestConfig().SessionFactory())
	suite.Run(t, new(DistroPatchByIDSuite))
}

func (s *DistroPatchByIDSuite) SetupTest() {
	s.data = data.MockDistroConnector{
		CachedDistros: []*distro.Distro{
			{
				Id:       "fedora8",
				Arch:     "linux_amd64",
				WorkDir:  "/data/mci",
				PoolSize: 30,
				Provider: "mock",
				ProviderSettings: &map[string]interface{}{
					"bid_price":      0.2,
					"instance_type":  "m3.large",
					"key_name":       "mci",
					"security_group": "mci",
					"ami":            "ami-2814683f",
					"mount_points": map[string]interface{}{
						"device_name":  "/dev/xvdb",
						"virtual_name": "ephemeral0"},
				},
				SetupAsSudo: true,
				Setup:       "Set-up string",
				Teardown:    "Tear-down string",
				User:        "root",
				SSHKey:      "SSH key string",
				SSHOptions: []string{
					"StrictHostKeyChecking=no",
					"BatchMode=yes",
					"ConnectTimeout=10"},
				SpawnAllowed: false,
				Expansions: []distro.Expansion{
					distro.Expansion{
						Key:   "decompress",
						Value: "tar zxvf"},
					distro.Expansion{
						Key:   "ps",
						Value: "ps aux"},
					distro.Expansion{
						Key:   "kill_pid",
						Value: "kill -- -$(ps opgid= %v)"},
					distro.Expansion{
						Key:   "scons_prune_ratio",
						Value: "0.8"},
				},
				Disabled:      false,
				ContainerPool: "",
			},
		},
	}
	s.settings = &evergreen.Settings{}
	s.sc = &data.MockConnector{
		MockDistroConnector: s.data,
	}
	s.rm = makePatchDistroByID(s.sc, s.settings)
}

func (s *DistroPatchByIDSuite) TestParse() {
	ctx := context.Background()
	json := []byte(`{"ssh_options":["StrictHostKeyChecking=no","BatchMode=yes","ConnectTimeout=10"]}`)
	req, _ := http.NewRequest("PATCH", "http://example.com/api/rest/v2/distros/fedora8", bytes.NewBuffer(json))

	err := s.rm.Parse(ctx, req)
	s.NoError(err)
	s.Equal(json, s.rm.(*distroIDPatchHandler).body)
}

func (s *DistroPatchByIDSuite) TestRunValidSpawnAllowed() {
	ctx := context.Background()
	json := []byte(`{"user_spawn_allowed": true}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	s.Equal(apiDistro.UserSpawnAllowed, true)
}

func (s *DistroPatchByIDSuite) TestRunValidProvider() {
	ctx := context.Background()
	json := []byte(`{"provider": "mock"}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	s.Equal(apiDistro.Provider, model.ToAPIString("mock"))
}

func (s *DistroPatchByIDSuite) TestRunValidProviderSettings() {
	ctx := context.Background()
	json := []byte(
		`{"settings" :{"bid_price": 0.15, "security_group": "password123"}}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	points := apiDistro.ProviderSettings["mount_points"]
	mapped := points.(map[string]interface{})
	s.Equal(mapped["device_name"], "/dev/xvdb")
	s.Equal(mapped["virtual_name"], "ephemeral0")
	s.Equal(apiDistro.ProviderSettings["bid_price"], 0.15)
	s.Equal(apiDistro.ProviderSettings["instance_type"], "m3.large")
	s.Equal(apiDistro.ProviderSettings["key_name"], "mci")
	s.Equal(apiDistro.ProviderSettings["security_group"], "password123")
	s.Equal(apiDistro.ProviderSettings["ami"], "ami-2814683f")
}

func (s *DistroPatchByIDSuite) TestRunValidArch() {
	ctx := context.Background()
	json := []byte(`{"arch": "linux_amd32"}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	s.Equal(apiDistro.Arch, model.ToAPIString("linux_amd32"))
}

func (s *DistroPatchByIDSuite) TestRunValidWorkDir() {
	ctx := context.Background()
	json := []byte(`{"work_dir": "/tmp"}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	s.Equal(apiDistro.WorkDir, model.ToAPIString("/tmp"))
}

func (s *DistroPatchByIDSuite) TestRunValidPoolSize() {
	ctx := context.Background()
	json := []byte(`{"pool_size": 50}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	s.Equal(apiDistro.PoolSize, 50)
}

func (s *DistroPatchByIDSuite) TestRunValidSetupAsSudo() {
	ctx := context.Background()
	json := []byte(`{"setup_as_sudo": false}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	s.Equal(apiDistro.SetupAsSudo, false)
}

func (s *DistroPatchByIDSuite) TestRunValidSetup() {
	ctx := context.Background()
	json := []byte(`{"setup": "New Set-up string"}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	s.Equal(apiDistro.Setup, model.ToAPIString("New Set-up string"))
}

func (s *DistroPatchByIDSuite) TestRunValidTearDown() {
	ctx := context.Background()
	json := []byte(`{"teardown": "New Tear-down string"}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	s.Equal(apiDistro.Teardown, model.ToAPIString("New Tear-down string"))
}

func (s *DistroPatchByIDSuite) TestRunValidUser() {
	ctx := context.Background()
	json := []byte(`{"user": "user101"}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	s.Equal(apiDistro.User, model.ToAPIString("user101"))
}

func (s *DistroPatchByIDSuite) TestRunValidSSHKey() {
	ctx := context.Background()
	json := []byte(`{"ssh_key": "New SSH key string"}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	s.Equal(apiDistro.SSHKey, model.ToAPIString("New SSH key string"))
}

func (s *DistroPatchByIDSuite) TestRunValidSSHOptions() {
	ctx := context.Background()
	json := []byte(`{"ssh_options":["BatchMode=no"]}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	s.Equal(apiDistro.SSHOptions, []string{"BatchMode=no"})
}

func (s *DistroPatchByIDSuite) TestRunValidExpansions() {
	ctx := context.Background()
	json := []byte(`{"expansions": [{"key": "key1", "value": "value1"}]}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	expansion := model.APIExpansion{Key: model.ToAPIString("key1"), Value: model.ToAPIString("value1")}
	s.Equal(apiDistro.Expansions, []model.APIExpansion{expansion})
}

func (s *DistroPatchByIDSuite) TestRunValidDisabled() {
	ctx := context.Background()
	json := []byte(`{"disabled": true}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	s.Equal(apiDistro.Disabled, true)
}

func (s *DistroPatchByIDSuite) TestRunValidContainer() {
	ctx := context.Background()
	json := []byte(`{"container_pool": ""}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusOK)

	apiDistro := (resp.Data()).(*model.APIDistro)
	s.Equal(apiDistro.ContainerPool, model.ToAPIString(""))
}

func (s *DistroPatchByIDSuite) TestRunInValidEmptyStringValues() {
	ctx := context.Background()
	json := []byte(`{"arch": "","user": "","work_dir": "","ssh_key": "","provider": ""}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusBadRequest)
	s.NotNil(resp.Data())

	errors := []string{
		"ERROR: distro 'arch' cannot be blank",
		"ERROR: distro 'user' cannot be blank",
		"ERROR: distro 'work_dir' cannot be blank",
		"ERROR: distro 'ssh_key' cannot be blank",
		"ERROR: distro 'provider' cannot be blank",
	}

	error := (resp.Data()).(gimlet.ErrorResponse)
	s.Equal(strings.Join(errors, "\n"), error.Message)
}

func (s *DistroPatchByIDSuite) TestValidFindAndReplaceFullDocument() {
	ctx := context.Background()
	json := []byte(
		`{
				"arch" : "~linux_amd64",
				"work_dir" : "~/data/mci",
				"pool_size" : 20,
				"provider" : "mock",
				"settings" : {
					"mount_points" : [{
						"device_name" : "~/dev/xvdb",
						"virtual_name" : "~ephemeral0"
					}],
					"ami" : "~ami-2814683f",
					"bid_price" : 0.1,
					"instance_type" : "~m3.large",
					"key_name" : "~mci",
					"security_group" : "~mci"
				},
				"setup_as_sudo" : false,
				"setup" : "~Set-up script",
				"teardown" : "~Tear-down script",
				"user" : "~root",
				"ssh_key" : "~SSH string",
				"ssh_options" : [
					"~StrictHostKeyChecking=no",
					"~BatchMode=no",
					"~ConnectTimeout=10"
				],
				"spawn_allowed" : false,
				"expansions" : [
					{
						"key" : "~decompress",
						"value" : "~tar zxvf"
					},
					{
						"key" : "~ps",
						"value" : "~ps aux"
					},
					{
						"key" : "~kill_pid",
						"value" : "~kill -- -$(ps opgid= %v)"
					},
					{
						"key" : "~scons_prune_ratio",
						"value" : "~0.8"
					}
				],
				"disabled" : false
	}`)

	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.Equal(resp.Status(), http.StatusOK)

	s.NotNil(resp.Data())
	apiDistro := (resp.Data()).(*model.APIDistro)
	s.Equal(apiDistro.Disabled, false)
	s.Equal(apiDistro.Name, model.ToAPIString("fedora8"))
	s.Equal(apiDistro.WorkDir, model.ToAPIString("~/data/mci"))
	s.Equal(apiDistro.PoolSize, 20)
	s.Equal(apiDistro.Provider, model.ToAPIString("mock"))

	points := apiDistro.ProviderSettings["mount_points"]
	typed := points.([]interface{})
	mapped := typed[0].(map[string]interface{})
	s.Equal(mapped["device_name"], "~/dev/xvdb")
	s.Equal(mapped["virtual_name"], "~ephemeral0")

	s.Equal(apiDistro.ProviderSettings["ami"], "~ami-2814683f")
	s.Equal(apiDistro.ProviderSettings["bid_price"], 0.1)
	s.Equal(apiDistro.ProviderSettings["instance_type"], "~m3.large")
	s.Equal(apiDistro.SetupAsSudo, false)
	s.Equal(apiDistro.Setup, model.ToAPIString("~Set-up script"))
	s.Equal(apiDistro.Teardown, model.ToAPIString("~Tear-down script"))
	s.Equal(apiDistro.User, model.ToAPIString("~root"))
	s.Equal(apiDistro.SSHKey, model.ToAPIString("~SSH string"))
	s.Equal(apiDistro.SSHOptions, []string{"~StrictHostKeyChecking=no", "~BatchMode=no", "~ConnectTimeout=10"})
	s.Equal(apiDistro.UserSpawnAllowed, false)

	s.Equal(apiDistro.Expansions, []model.APIExpansion{
		model.APIExpansion{Key: model.ToAPIString("~decompress"), Value: model.ToAPIString("~tar zxvf")},
		model.APIExpansion{Key: model.ToAPIString("~ps"), Value: model.ToAPIString("~ps aux")},
		model.APIExpansion{Key: model.ToAPIString("~kill_pid"), Value: model.ToAPIString("~kill -- -$(ps opgid= %v)")},
		model.APIExpansion{Key: model.ToAPIString("~scons_prune_ratio"), Value: model.ToAPIString("~0.8")},
	})
}

func (s *DistroPatchByIDSuite) TestRunInValidNameChange() {
	ctx := context.Background()
	ctx = gimlet.AttachUser(ctx, &user.DBUser{Id: "user1"})
	json := []byte(`{"name": "Updated distro name"}`)
	h := s.rm.(*distroIDPatchHandler)
	h.distroID = "fedora8"
	h.body = json

	resp := s.rm.Run(ctx)
	s.NotNil(resp.Data())
	s.Equal(resp.Status(), http.StatusBadRequest)

	gimlet := (resp.Data()).(gimlet.ErrorResponse)
	s.Equal(gimlet.Message, fmt.Sprintf("Distro name is immutable; cannot rename distro resource '%s'", h.distroID))
}
