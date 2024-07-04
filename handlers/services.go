package handlers

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/groups"
	"github.com/statping-ng/statping-ng/types/hits"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
)

type serviceOrder struct {
	Id    int64 `json:"service"`
	Order int   `json:"order"`
}

func isBelongToGroup(r *http.Request, service *services.Service) bool {
	isUser := IsUser(r)
	groupName := r.URL.Query().Get("group_name")
	if isUser {
		return true
	}

	group, err := groups.FindByName(groupName)
	if err != nil {
		return false
	}

	if int64(service.GroupId) != group.Id {
		return false
	}
	return true
}

func findService(r *http.Request) (*services.Service, error) {
	vars := mux.Vars(r)
	id := utils.ToInt(vars["id"])
	service, err := services.Find(id)
	if err != nil {
		return nil, err
	}

	if !isBelongToGroup(r, service) {
		return nil, errors.NotAuthenticated
	}

	if !service.Public.Bool && !IsReadAuthenticated(r) {
		return nil, errors.NotAuthenticated
	}
	return service, nil
}

func reorderServiceHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var newOrder []*serviceOrder
	if err := DecodeJSON(r, &newOrder); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	for _, s := range newOrder {
		service, err := services.Find(s.Id)
		if err != nil {
			sendErrorJson(err, w, r)
			return
		}
		service.Order = s.Order
		service.Update()
	}
	returnJson(newOrder, w, r)
}

func apiServiceHandler(r *http.Request) interface{} {
	srv, err := findService(r)
	if err != nil {
		return err
	}
	srv = srv.UpdateStats()
	return *srv
}

func apiCreateServiceHandler(w http.ResponseWriter, r *http.Request) {
	var service *services.Service
	if err := DecodeJSON(r, &service); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	if err := service.Create(); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	go services.ServiceCheckQueue(service, true)

	sendJsonAction(service, "create", w, r)
}

type servicePatchReq struct {
	Online  bool   `json:"online"`
	Issue   string `json:"issue,omitempty"`
	Latency int64  `json:"latency,omitempty"`
}

func apiServicePatchHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	var req servicePatchReq
	if err := DecodeJSON(r, &req); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	service.Online = req.Online
	service.Latency = req.Latency

	issueDefault := "Service was triggered to be offline"
	if req.Issue != "" {
		issueDefault = req.Issue
	}

	if !req.Online {
		services.RecordFailure(service, issueDefault, "trigger")
	} else {
		services.RecordSuccess(service)
	}

	if err := service.Update(); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	sendJsonAction(service, "update", w, r)
}

func apiServiceUpdateHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	if err := DecodeJSON(r, &service); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	if err := service.Update(); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	go service.CheckService(true)
	sendJsonAction(service, "update", w, r)
}

func apiServiceDataHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	if !isBelongToGroup(r, service) {
		sendErrorJson(errors.NotAuthenticated, w, r)
		return
	}

	groupQuery, err := database.ParseQueries(r, service.AllHits())
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	objs, err := groupQuery.GraphData(database.ByAverage("latency", 1000))
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	returnJson(objs, w, r)
}

func apiServiceFailureDataHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	if !isBelongToGroup(r, service) {
		sendErrorJson(errors.NotAuthenticated, w, r)
		return
	}

	groupQuery, err := database.ParseQueries(r, service.AllFailures())
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	objs, err := groupQuery.GraphData(database.ByCount)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	returnJson(objs, w, r)
}

func apiServicePingDataHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	if !isBelongToGroup(r, service) {
		sendErrorJson(errors.NotAuthenticated, w, r)
		return
	}

	groupQuery, err := database.ParseQueries(r, service.AllHits())
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	objs, err := groupQuery.GraphData(database.ByAverage("ping_time", 1000))
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	returnJson(objs, w, r)
}

func apiServiceTimeDataHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	if !isBelongToGroup(r, service) {
		sendErrorJson(errors.NotAuthenticated, w, r)
		return
	}

	groupHits, err := database.ParseQueries(r, service.AllHits())
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	groupFailures, err := database.ParseQueries(r, service.AllFailures())
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	var allFailures []*failures.Failure
	var allHits []*hits.Hit

	if err := groupHits.Find(&allHits); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	if err := groupFailures.Find(&allFailures); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	uptimeData, err := service.UptimeData(allHits, allFailures)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	returnJson(uptimeData, w, r)
}

func apiServiceHitsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	if err := service.AllHits().DeleteAll(); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	sendJsonAction(service, "delete", w, r)
}

func apiServiceDeleteHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	err = service.Delete()
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	sendJsonAction(service, "delete", w, r)
}

func apiAllServicesHandler(r *http.Request) interface{} {
	isUser := IsUser(r)

	if isUser {
		var srvs []services.Service
		for _, v := range services.AllInOrder() {
			srvs = append(srvs, v)
		}
		return srvs
	}

	// get group from query
	group := r.URL.Query().Get("group_name")

	if group == "" {
		return make([]services.Service, 0)
	}

	gr, err := groups.FindByName(group)
	if err != nil {
		log.Errorf("Failed to find group %s: %v", group, err)
		return make([]services.Service, 0)
	}

	var srvs []services.Service
	for _, v := range services.AllInGroupOrder(int(gr.Id)) {
		if !v.Public.Bool {
			continue
		}
		srvs = append(srvs, v)
	}
	return srvs
}

func servicesDeleteFailuresHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	if err := service.AllFailures().DeleteAll(); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	sendJsonAction(service, "delete_failures", w, r)
}

func apiServiceFailuresHandler(r *http.Request) interface{} {
	service, err := findService(r)
	if err != nil {
		return err
	}

	if !isBelongToGroup(r, service) {
		return errors.NotAuthenticated
	}

	var fails []*failures.Failure
	query, err := database.ParseQueries(r, service.AllFailures())
	if err != nil {
		return err
	}
	query.Find(&fails)
	return fails
}

func apiServiceHitsHandler(r *http.Request) interface{} {
	service, err := findService(r)
	if err != nil {
		return err
	}

	if !isBelongToGroup(r, service) {
		return errors.NotAuthenticated
	}

	var hts []*hits.Hit
	query, err := database.ParseQueries(r, service.AllHits())
	if err != nil {
		return err
	}
	query.Find(&hts)
	return hts
}
