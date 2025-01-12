package test

import (
	"context"
	"log"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/arangodb/go-driver"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"github.com/SecurityBrewery/catalyst"
	"github.com/SecurityBrewery/catalyst/bus"
	"github.com/SecurityBrewery/catalyst/database"
	"github.com/SecurityBrewery/catalyst/database/busdb"
	"github.com/SecurityBrewery/catalyst/generated/models"
	"github.com/SecurityBrewery/catalyst/generated/restapi"
	"github.com/SecurityBrewery/catalyst/hooks"
	"github.com/SecurityBrewery/catalyst/index"
	"github.com/SecurityBrewery/catalyst/pointer"
	"github.com/SecurityBrewery/catalyst/service"
	"github.com/SecurityBrewery/catalyst/storage"
)

func Context() context.Context {
	w := httptest.NewRecorder()
	gctx, _ := gin.CreateTestContext(w)
	busdb.SetContext(gctx, Bob)
	return gctx
}

func Config(ctx context.Context) (*catalyst.Config, error) {
	config := &catalyst.Config{
		IndexPath: "index.bleve",
		DB: &database.Config{
			Host:     "http://localhost:8529",
			User:     "root",
			Password: "foobar",
		},
		Storage: &storage.Config{
			Host:     "http://localhost:9000",
			User:     "minio",
			Password: "minio123",
		},
		Bus: &bus.Config{
			Host:   "tcp://localhost:9001",
			Key:    "A9RysEsPJni8RaHeg_K0FKXQNfBrUyw-",
			APIUrl: "http://localhost:8002/api",
		},
		UISettings: &models.Settings{
			ArtifactStates: []*models.Type{
				{Icon: "mdi-help-circle-outline", ID: "unknown", Name: "Unknown", Color: pointer.String(models.TypeColorInfo)},
				{Icon: "mdi-skull", ID: "malicious", Name: "Malicious", Color: pointer.String(models.TypeColorError)},
				{Icon: "mdi-check", ID: "clean", Name: "Clean", Color: pointer.String(models.TypeColorSuccess)},
			},
			TicketTypes: []*models.TicketTypeResponse{
				{ID: "alert", Icon: "mdi-alert", Name: "Alerts"},
				{ID: "incident", Icon: "mdi-radioactive", Name: "Incidents"},
				{ID: "investigation", Icon: "mdi-fingerprint", Name: "Forensic Investigations"},
				{ID: "hunt", Icon: "mdi-target", Name: "Threat Hunting"},
			},
			Version:    "0.0.0-test",
			Tier:       models.SettingsTierCommunity,
			Timeformat: "YYYY-MM-DDThh:mm:ss",
		},
		Secret: []byte("4ef5b29539b70233dd40c02a1799d25079595565e05a193b09da2c3e60ada1cd"),
		Auth: &catalyst.AuthConfig{
			OIDCIssuer: "http://localhost:9002/auth/realms/catalyst",
			OAuth2: &oauth2.Config{
				ClientID:     "catalyst",
				ClientSecret: "13d4a081-7395-4f71-a911-bc098d8d3c45",
				RedirectURL:  "http://localhost:8002/callback",
				Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
			},
			// OIDCClaimUsername: "",
			// OIDCClaimEmail:    "",
			// OIDCClaimName:     "",
			// AuthBlockNew:      false,
			// AuthDefaultRoles:  nil,
		},
	}
	err := config.Auth.Load(ctx)
	if err != nil {
		return nil, err
	}

	return config, err
}

func Index(t *testing.T) (*index.Index, func(), error) {
	dir, err := os.MkdirTemp("", "catalyst-test-"+cleanName(t))
	if err != nil {
		return nil, nil, err
	}

	catalystIndex, err := index.New(path.Join(dir, "index.bleve"))
	if err != nil {
		return nil, nil, err
	}
	return catalystIndex, func() { os.RemoveAll(dir) }, nil
}

func Bus(t *testing.T) (context.Context, *catalyst.Config, *bus.Bus, error) {
	ctx := Context()

	config, err := Config(ctx)
	if err != nil {
		t.Fatal(err)
	}

	catalystBus, err := bus.New(config.Bus)
	if err != nil {
		t.Fatal(err)
	}
	return ctx, config, catalystBus, err
}

func DB(t *testing.T) (context.Context, *catalyst.Config, *bus.Bus, *index.Index, *storage.Storage, *database.Database, func(), error) {
	ctx, config, rbus, err := Bus(t)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	catalystStorage, err := storage.New(config.Storage)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	catalystIndex, cleanup, err := Index(t)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	c := config.DB
	c.Name = cleanName(t)
	db, err := database.New(ctx, catalystIndex, rbus, &hooks.Hooks{
		DatabaseAfterConnectFuncs: []func(ctx context.Context, client driver.Client, name string){Clear},
	}, c)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	_, err = db.JobCreate(ctx, "99cd67131b48", &models.JobForm{
		Automation: "hash.sha1",
		Payload: "test",
		Origin: nil,
	})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	return ctx, config, rbus, catalystIndex, catalystStorage, db, func() {
		err := db.Remove(context.Background())
		if err != nil {
			log.Println(err)
		}
		cleanup()
	}, err
}

func Service(t *testing.T) (context.Context, *catalyst.Config, *bus.Bus, *index.Index, *storage.Storage, *database.Database, *service.Service, func(), error) {
	ctx, config, rbus, catalystIndex, catalystStorage, db, cleanup, err := DB(t)
	if err != nil {
		t.Fatal(err)
	}

	catalystService, err := service.New(rbus, db, catalystStorage, config.UISettings)
	if err != nil {
		t.Fatal(err)
	}

	return ctx, config, rbus, catalystIndex, catalystStorage, db, catalystService, cleanup, err
}

func Server(t *testing.T) (context.Context, *catalyst.Config, *bus.Bus, *index.Index, *storage.Storage, *database.Database, *service.Service, *restapi.Server, func(), error) {
	ctx, config, rbus, catalystIndex, catalystStorage, db, catalystService, cleanup, err := Service(t)
	if err != nil {
		t.Fatal(err)
	}

	catalystServer := restapi.New(catalystService, &restapi.Config{Address: "0.0.0.0:8000", InsecureHTTP: true})

	return ctx, config, rbus, catalystIndex, catalystStorage, db, catalystService, catalystServer, cleanup, err
}

func cleanName(t *testing.T) string {
	name := t.Name()
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "/", "_")
	return strings.ReplaceAll(name, "#", "_")
}

func Clear(ctx context.Context, client driver.Client, name string) {
	if exists, _ := client.DatabaseExists(ctx, name); exists {
		if db, err := client.Database(ctx, name); err == nil {
			if exists, _ = db.GraphExists(ctx, database.TicketArtifactsGraphName); exists {
				if g, err := db.Graph(ctx, database.TicketArtifactsGraphName); err == nil {
					if err := g.Remove(ctx); err != nil {
						log.Println(err)
					}
				}
			}
			if err := db.Remove(ctx); err != nil {
				log.Println(err)
			}
		}
	}
}
