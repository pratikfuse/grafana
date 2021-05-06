package libraryelements

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"gopkg.in/macaron.v1"

	"github.com/grafana/grafana/pkg/api/response"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/registry"
	"github.com/grafana/grafana/pkg/services/dashboards"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/setting"
)

const UserInDbName = "user_in_db"
const UserInDbAvatar = "/avatar/402d08de060496d6b6874495fe20f5ad"

type libraryElement struct {
	ID          int64  `json:"id"`
	OrgID       int64  `json:"orgId"`
	FolderID    int64  `json:"folderId"`
	UID         string `json:"uid"`
	Name        string `json:"name"`
	Type        string
	Description string
	Model       map[string]interface{} `json:"model"`
	Version     int64                  `json:"version"`
	Meta        LibraryElementDTOMeta  `json:"meta"`
}

type libraryElementResult struct {
	Result libraryElement `json:"result"`
}

type libraryElementsSearch struct {
	Result libraryElementsSearchResult `json:"result"`
}

type libraryElementsSearchResult struct {
	TotalCount int64            `json:"totalCount"`
	Elements   []libraryElement `json:"elements"`
	Page       int              `json:"page"`
	PerPage    int              `json:"perPage"`
}

func overrideLibraryElementServiceInRegistry(cfg *setting.Cfg) LibraryElementService {
	l := LibraryElementService{
		SQLStore: nil,
		Cfg:      cfg,
	}

	overrideServiceFunc := func(d registry.Descriptor) (*registry.Descriptor, bool) {
		descriptor := registry.Descriptor{
			Name:         "LibraryElementService",
			Instance:     &l,
			InitPriority: 0,
		}

		return &descriptor, true
	}

	registry.RegisterOverride(overrideServiceFunc)

	return l
}

func getCreatePanelCommand(folderID int64, name string) createLibraryElementCommand {
	command := getCreateCommandWithModel(folderID, name, Panel, []byte(`
			{
			  "datasource": "${DS_GDEV-TESTDATA}",
			  "id": 1,
			  "title": "Text - Library Panel",
			  "type": "text",
			  "description": "A description"
			}
		`))

	return command
}

func getCreateVariableCommand(folderID int64, name string) createLibraryElementCommand {
	command := getCreateCommandWithModel(folderID, name, Variable, []byte(`
			{
			  "datasource": "${DS_GDEV-TESTDATA}",
			  "name": "query0",
			  "type": "query",
			  "description": "A description"
			}
		`))

	return command
}

func getCreateCommandWithModel(folderID int64, name string, kind LibraryElementKind, model []byte) createLibraryElementCommand {
	command := createLibraryElementCommand{
		FolderID: folderID,
		Name:     name,
		Model:    model,
		Kind:     int64(kind),
	}

	return command
}

type scenarioContext struct {
	ctx           *macaron.Context
	service       *LibraryElementService
	reqContext    *models.ReqContext
	user          models.SignedInUser
	folder        *models.Folder
	initialResult libraryElementResult
	sqlStore      *sqlstore.SQLStore
}

type folderACLItem struct {
	roleType   models.RoleType
	permission models.PermissionType
}

func createFolderWithACL(t *testing.T, sqlStore *sqlstore.SQLStore, title string, user models.SignedInUser,
	items []folderACLItem) *models.Folder {
	t.Helper()

	s := dashboards.NewFolderService(user.OrgId, &user, sqlStore)
	t.Logf("Creating folder with title and UID %q", title)
	folder, err := s.CreateFolder(title, title)
	require.NoError(t, err)

	updateFolderACL(t, sqlStore, folder.Id, items)

	return folder
}

func updateFolderACL(t *testing.T, sqlStore *sqlstore.SQLStore, folderID int64, items []folderACLItem) {
	t.Helper()

	if len(items) == 0 {
		return
	}

	var aclItems []*models.DashboardAcl
	for _, item := range items {
		role := item.roleType
		permission := item.permission
		aclItems = append(aclItems, &models.DashboardAcl{
			DashboardID: folderID,
			Role:        &role,
			Permission:  permission,
			Created:     time.Now(),
			Updated:     time.Now(),
		})
	}

	err := sqlStore.UpdateDashboardACL(folderID, aclItems)
	require.NoError(t, err)
}

func validateAndUnMarshalResponse(t *testing.T, resp response.Response) libraryElementResult {
	t.Helper()

	require.Equal(t, 200, resp.Status())

	var result = libraryElementResult{}
	err := json.Unmarshal(resp.Body(), &result)
	require.NoError(t, err)

	return result
}

func scenarioWithPanel(t *testing.T, desc string, fn func(t *testing.T, sc scenarioContext)) {
	t.Helper()

	testScenario(t, desc, func(t *testing.T, sc scenarioContext) {
		command := getCreatePanelCommand(sc.folder.Id, "Text - Library Panel")
		resp := sc.service.createHandler(sc.reqContext, command)
		sc.initialResult = validateAndUnMarshalResponse(t, resp)

		fn(t, sc)
	})
}

// testScenario is a wrapper around t.Run performing common setup for library panel tests.
// It takes your real test function as a callback.
func testScenario(t *testing.T, desc string, fn func(t *testing.T, sc scenarioContext)) {
	t.Helper()

	t.Run(desc, func(t *testing.T) {
		t.Cleanup(registry.ClearOverrides)

		ctx := macaron.Context{
			Req: macaron.Request{Request: &http.Request{}},
		}
		orgID := int64(1)
		role := models.ROLE_ADMIN

		cfg := setting.NewCfg()
		// Everything in this service is behind the feature toggle "panelLibrary"
		cfg.FeatureToggles = map[string]bool{"panelLibrary": true}
		// Because the LibraryPanelService is behind a feature toggle, we need to override the service in the registry
		// with a Cfg that contains the feature toggle so migrations are run properly
		service := overrideLibraryElementServiceInRegistry(cfg)

		// We need to assign SQLStore after the override and migrations are done
		sqlStore := sqlstore.InitTestDB(t)
		service.SQLStore = sqlStore

		user := models.SignedInUser{
			UserId:     1,
			Name:       "Signed In User",
			Login:      "signed_in_user",
			Email:      "signed.in.user@test.com",
			OrgId:      orgID,
			OrgRole:    role,
			LastSeenAt: time.Now(),
		}

		// deliberate difference between signed in user and user in db to make it crystal clear
		// what to expect in the tests
		// In the real world these are identical
		cmd := models.CreateUserCommand{
			Email: "user.in.db@test.com",
			Name:  "User In DB",
			Login: UserInDbName,
		}
		_, err := sqlStore.CreateUser(context.Background(), cmd)
		require.NoError(t, err)

		sc := scenarioContext{
			user:     user,
			ctx:      &ctx,
			service:  &service,
			sqlStore: sqlStore,
			reqContext: &models.ReqContext{
				Context:      &ctx,
				SignedInUser: &user,
			},
		}

		sc.folder = createFolderWithACL(t, sc.sqlStore, "ScenarioFolder", sc.user, []folderACLItem{})

		fn(t, sc)
	})
}

func getCompareOptions() []cmp.Option {
	return []cmp.Option{
		cmp.Transformer("Time", func(in time.Time) int64 {
			return in.UTC().Unix()
		}),
	}
}
