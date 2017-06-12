package concordances

import (
	"encoding/json"
	"errors"
	"fmt"
	health "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rcrowley/go-metrics"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	db "github.com/Financial-Times/concordances-rw-dynamodb/dynamodb"

)

const (
	CONCORDANCES                 = "concordances"
	ContentTypeJson              = "application/json"
	UUID_Param                   = "uuid"
	ErrorMsgJson                 = "{\"message\":\"%s\"}"
	LogMsg503                    = "Error %s concordances"
	LogMsg404                    = "Concordances not found"
	ErrorMsg_BadPath             = "Invalid path."
	ErrorMsg_BadBody             = "Invalid payload."
	ErrorMsg_MismatchedConceptId = "Concept uuid in payload is different from uuid path parameter"
	ErrorMsg_MissingConcoredeIds = "Payload has no concorded uuids to store."
	ErrorMsg_BadJson             = "Corrupted JSON"
)

var uuidRegex = regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")

type ConcordancesRwHandler struct {
	srv           Service
	conf 	      AppConfig

}

func NewConcordanceRwHandler(router *mux.Router, conf AppConfig, srv Service) ConcordancesRwHandler {
	h := ConcordancesRwHandler{srv:srv, conf:conf}
	h.registerAdminHandlers(router)
	h.registerApiHandlers(router)
	return h
}

func (h *ConcordancesRwHandler) registerApiHandlers(router *mux.Router) {
	log.Info("Registering API handlers")

	rwHandler := handlers.MethodHandler{
		"GET":    http.HandlerFunc(h.HandleGet),
		"PUT":    http.HandlerFunc(h.HandlePut),
		"DELETE": http.HandlerFunc(h.HandleDelete),
	}
	countHandler := handlers.MethodHandler{
		"GET": http.HandlerFunc(h.HandleCount),
	}

	router.Handle("/concordances/__count", countHandler)
	router.Handle("/{concordances}/{uuid}", rwHandler)

	//Invalid path patterns
	router.Handle("/{concordances}/", http.HandlerFunc(h.HandleBadRequest))
	router.Handle("/{concordances}", http.HandlerFunc(h.HandleBadRequest))
	router.Handle("/", http.HandlerFunc(h.HandleBadRequest))
}

func (h *ConcordancesRwHandler) registerAdminHandlers(router *mux.Router) {
	log.Info("Registering admin handlers")
	healthService := newHealthService(&healthConfig{appSystemCode: h.conf.AppSystemCode, appName: h.conf.AppName, port: h.conf.Port,
		DynamoDbTable: h.conf.DynamoDbTableName,
		AwsRegion: h.conf.AWSRegion,
		SnsTopic: h.conf.SnsTopic,
	})
	hc := health.HealthCheck{SystemCode: h.conf.AppSystemCode, Name: h.conf.AppName, Description: "Stores concordances in cache and notifies downstream services", Checks: healthService.checks}
	router.HandleFunc(healthPath, health.Handler(hc))
	router.HandleFunc(status.PingPath, status.PingHandler)
	router.HandleFunc(status.GTGPath, status.NewGoodToGoHandler(healthService.gtgCheck))
	router.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)

	monitoringRouter := httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), router)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)
	http.Handle("/", monitoringRouter)
}

func (h *ConcordancesRwHandler) HandleGet(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	//400
	if h.invalidPath(vars) {
		h.HandleBadRequest(rw, r)
		return
	}
	uuid := vars[UUID_Param]
	model, err := h.srv.Read(uuid)

	//503
	if err != nil {
		h.handleServiceUnavailable("retrieving", err.Error(), rw, r)
		return
	}
	//404
	if model.ConcordedIds == nil {
		h.handleNotFound(uuid, rw, r)
		return
	}

	output, err := json.Marshal(&model)
	//404
	if err != nil {
		h.handleNotFound(uuid, rw, r)
		return
	}
	//200
	rw.Header().Set("Content-Type", ContentTypeJson)
	rw.WriteHeader(http.StatusOK)
	rw.Write(output)
}

func (h *ConcordancesRwHandler) HandlePut(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	//400
	if h.invalidPath(vars) {
		h.HandleBadRequest(rw, r)
		return
	}
	uuid := vars[UUID_Param]
	if r.ContentLength <= 0 {
		h.handleBadPayload(ErrorMsg_BadJson, rw, r)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	//400
	if err != nil {
		h.handleBadPayload(ErrorMsg_BadJson, rw, r)
		return
	}

	model := db.Model{}
	err = json.Unmarshal(body, &model)
	//400
	if err != nil {
		h.handleBadPayload(ErrorMsg_BadJson, rw, r)
		return
	}

	err = h.invalidModel(uuid, model)
	//400
	if err != nil {
		h.handleBadPayload(err.Error(), rw, r)
		return
	}

	created, err := h.srv.Write(model)

	//503
	if err != nil {
		h.handleServiceUnavailable("storing", err.Error(), rw, r)
		return
	}

	if created {
		rw.WriteHeader(http.StatusCreated)
		return
	} else {
		rw.WriteHeader(http.StatusOK)
	}
}

func (h *ConcordancesRwHandler) HandleDelete(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	//400
	if h.invalidPath(vars) {
		h.HandleBadRequest(rw, r)
		return
	}
	uuid := vars[UUID_Param]
	deleted, err := h.srv.Delete(uuid)
	//503
	if err != nil {
		h.handleServiceUnavailable("deleting", err.Error(), rw, r)
		return
	}
	//404
	if !deleted {
		h.handleNotFound(uuid, rw, r)
		return
	}
	//204
	rw.WriteHeader(http.StatusNoContent)
}

func (h *ConcordancesRwHandler) HandleCount(rw http.ResponseWriter, r *http.Request) {
	count, err := h.srv.Count()
	//503
	if err != nil {
		h.handleServiceUnavailable("counting", err.Error(), rw, r)
		return
	}
	//200
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
	b := []byte{}
	b = strconv.AppendInt(b, count, 10)
	rw.Write(b)
}

func (h *ConcordancesRwHandler) HandleBadRequest(rw http.ResponseWriter, r *http.Request) {
	log.Infof(fmt.Sprintf("%s: %s", ErrorMsg_BadPath, r.URL.Path))
	rw.Header().Set("Content-Type", ContentTypeJson)
	rw.WriteHeader(http.StatusBadRequest)
	rw.Write([]byte(fmt.Sprintf(ErrorMsgJson, ErrorMsg_BadPath)))
	return
}

func (h *ConcordancesRwHandler) handleBadPayload(errorMsg string, rw http.ResponseWriter, r *http.Request) {
	logMsg := fmt.Sprintf("%s Error: %s", ErrorMsg_BadBody, errorMsg)
	log.Infof(logMsg)
	rw.Header().Set("Content-Type", ContentTypeJson)
	rw.WriteHeader(http.StatusBadRequest)
	rw.Write([]byte(fmt.Sprintf(ErrorMsgJson, logMsg)))
	return
}

func (h *ConcordancesRwHandler) handleNotFound(uuid string, rw http.ResponseWriter, r *http.Request) {
	log.Infof("%s for %s", LogMsg404, uuid)
	rw.Header().Set("Content-Type", ContentTypeJson)
	rw.WriteHeader(http.StatusNotFound)
	rw.Write([]byte(fmt.Sprintf(ErrorMsgJson, LogMsg404)))
}

func (h *ConcordancesRwHandler) handleServiceUnavailable(op string, errorMsg string, rw http.ResponseWriter, r *http.Request) {
	logMsg := fmt.Sprintf(LogMsg503, op)
	log.Errorf("%s %s", logMsg, errorMsg)
	rw.Header().Set("Content-Type", ContentTypeJson)
	rw.WriteHeader(http.StatusServiceUnavailable)
	rw.Write([]byte(fmt.Sprintf(ErrorMsgJson, logMsg)))
}

func (h *ConcordancesRwHandler) invalidPath(vars map[string]string) bool {
	isUuid := uuidRegex.MatchString(vars[UUID_Param])
	if (vars[CONCORDANCES] != CONCORDANCES) || !isUuid {
		return true
	}
	return false
}

func (h *ConcordancesRwHandler) invalidModel(uuid string, model db.Model) error {
	if model.UUID != uuid {
		return errors.New(ErrorMsg_MismatchedConceptId)
	}
	if (model.ConcordedIds == nil) || (len(model.ConcordedIds) < 1) {
		return errors.New(ErrorMsg_MissingConcoredeIds)
	}
	return nil
}
