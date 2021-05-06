package libraryelements

import (
	"errors"

	"github.com/go-macaron/binding"

	"github.com/grafana/grafana/pkg/api/response"
	"github.com/grafana/grafana/pkg/api/routing"
	"github.com/grafana/grafana/pkg/middleware"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/util"
)

func (l *LibraryElementService) registerAPIEndpoints() {
	if !l.IsEnabled() {
		return
	}

	l.RouteRegister.Group("/api/library-elements", func(entities routing.RouteRegister) {
		entities.Post("/", middleware.ReqSignedIn, binding.Bind(createLibraryElementCommand{}), routing.Wrap(l.createHandler))
		//entities.Post("/:uid/dashboards/:dashboardId", middleware.ReqSignedIn, routing.Wrap(l.connectHandler))
		entities.Delete("/:uid", middleware.ReqSignedIn, routing.Wrap(l.deleteHandler))
		//entities.Delete("/:uid/dashboards/:dashboardId", middleware.ReqSignedIn, routing.Wrap(l.disconnectHandler))
		//entities.Get("/", middleware.ReqSignedIn, routing.Wrap(l.getAllHandler))
		//entities.Get("/:uid", middleware.ReqSignedIn, routing.Wrap(l.getHandler))
		//entities.Get("/:uid/dashboards/", middleware.ReqSignedIn, routing.Wrap(l.getConnectedDashboardsHandler))
		//entities.Patch("/:uid", middleware.ReqSignedIn, binding.Bind(patchLibraryPanelCommand{}), routing.Wrap(l.patchHandler))
	})
}

// createHandler handles POST /api/library-elements.
func (l *LibraryElementService) createHandler(c *models.ReqContext, cmd createLibraryElementCommand) response.Response {
	panel, err := l.createLibraryElement(c, cmd)
	if err != nil {
		return toLibraryElementError(err, "Failed to create library element")
	}

	return response.JSON(200, util.DynMap{"result": panel})
}

// deleteHandler handles DELETE /api/library-elements/:uid.
func (l *LibraryElementService) deleteHandler(c *models.ReqContext) response.Response {
	err := l.deleteLibraryElement(c, c.Params(":uid"))
	if err != nil {
		return toLibraryElementError(err, "Failed to delete library element")
	}

	return response.Success("Library element deleted")
}

// getHandler handles GET  /api/library-elements/:uid.
func (l *LibraryElementService) getHandler(c *models.ReqContext) response.Response {
	element, err := l.getLibraryElement(c, c.Params(":uid"))
	if err != nil {
		return toLibraryElementError(err, "Failed to get library element")
	}

	return response.JSON(200, util.DynMap{"result": element})
}

func toLibraryElementError(err error, message string) response.Response {
	if errors.Is(err, errLibraryElementAlreadyExists) {
		return response.Error(400, errLibraryElementAlreadyExists.Error(), err)
	}
	if errors.Is(err, errLibraryElementNotFound) {
		return response.Error(404, errLibraryElementNotFound.Error(), err)
	}
	if errors.Is(err, errLibraryElementDashboardNotFound) {
		return response.Error(404, errLibraryElementDashboardNotFound.Error(), err)
	}
	if errors.Is(err, errLibraryElementHeaderUIDMissing) {
		return response.Error(412, errLibraryElementHeaderUIDMissing.Error(), err)
	}
	if errors.Is(err, models.ErrFolderNotFound) {
		return response.Error(404, models.ErrFolderNotFound.Error(), err)
	}
	if errors.Is(err, models.ErrFolderAccessDenied) {
		return response.Error(403, models.ErrFolderAccessDenied.Error(), err)
	}
	if errors.Is(err, errLibraryElementHasConnectedDashboards) {
		return response.Error(403, errLibraryElementHasConnectedDashboards.Error(), err)
	}
	return response.Error(500, message, err)
}
