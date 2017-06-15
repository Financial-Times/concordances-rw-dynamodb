package concordances

import (
	"encoding/json"
	"errors"
	"fmt"
	db "github.com/Financial-Times/concordances-rw-dynamodb/dynamodb"
	health "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rcrowley/go-metrics"
	"net/http"
)

const (
	ContentTypeJson              = "application/json"
	UUID_Param                   = "uuid"
	ErrorMsgJson                 = "{\"message\":\"%s\"}"
	LogMsg503                    = "Error %s concordances"
	LogMsg404                    = "Concordances not found"
	ErrorMsg_BadBody             = "Invalid payload."
	ErrorMsg_MismatchedConceptId = "Concept UUID in payload is different from UUID path parameter"
	ErrorMsg_MissingConcordedIds = "Payload has no concorded UUIDs to store."
	ErrorMsg_BadJson             = "Corrupted JSON"
)

type Handler struct {
	srv  Service
	conf AppConfig
}

func NewHandler(router *mux.Router, conf AppConfig, srv Service) Handler {
	h := Handler{srv: srv, conf: conf}
	healthcheckConfig := &healthConfig{appSystemCode: h.conf.AppSystemCode, appName: h.conf.AppName, port: h.conf.Port, srv: srv}
	h.registerAdminHandlers(router, healthcheckConfig)
	h.registerAPIHandlers(router)
	return h
}

func (h *Handler) registerAPIHandlers(router *mux.Router) {
	log.Info("Registering API handlers")

	rwHandler := handlers.MethodHandler{
		"GET":    http.HandlerFunc(h.HandleGet),
		"PUT":    http.HandlerFunc(h.HandlePut),
		"DELETE": http.HandlerFunc(h.HandleDelete),
	}

	router.Handle("/concordances/{uuid:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}", rwHandler)
}

func (h *Handler) registerAdminHandlers(router *mux.Router, config *healthConfig) {
	log.Info("Registering admin handlers")
	healthService := newHealthService(config)
	hc := health.HealthCheck{SystemCode: h.conf.AppSystemCode, Name: h.conf.AppName, Description: "Stores concordances in cache and notifies downstream services", Checks: healthService.checks}
	router.HandleFunc(healthPath, health.Handler(hc))
	router.HandleFunc(status.GTGPath, status.NewGoodToGoHandler(healthService.gtgCheck))
	router.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)

	monitoringRouter := httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), router)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)
}

func (h *Handler) HandleGet(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	uuid := vars[UUID_Param]
	model, err := h.srv.Read(uuid)

	//503
	if err != nil {
		logMsg := fmt.Sprintf(LogMsg503, "retrieving")
		log.WithFields(log.Fields{"UUID": uuid}).Errorf("%s %s", logMsg, err.Error())
		writeJSONError(rw, logMsg, http.StatusServiceUnavailable)
		return
	}
	//404
	if model.ConcordedIds == nil {
		log.WithFields(log.Fields{"UUID": uuid}).Infof("%s for %s", LogMsg404, uuid)
		writeJSONError(rw, LogMsg404, http.StatusNotFound)
		return
	}

	//200
	rw.Header().Set("Content-Type", ContentTypeJson)
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(&model)
}

func (h *Handler) HandlePut(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	//400
	uuid := vars[UUID_Param]
	if r.ContentLength <= 0 {
		logMsg := fmt.Sprintf("%s Error: %s", ErrorMsg_BadBody, ErrorMsg_BadJson)
		log.WithFields(log.Fields{"UUID": uuid}).Infof(logMsg)
		writeJSONError(rw, logMsg, http.StatusBadRequest)
		return
	}
	model := db.ConcordancesModel{}
	err := json.NewDecoder(r.Body).Decode(&model)
	defer r.Body.Close()

	//400
	if err != nil {
		logMsg := fmt.Sprintf("%s Error: %s", ErrorMsg_BadBody, ErrorMsg_BadJson)
		log.WithFields(log.Fields{"UUID": uuid}).Infof(logMsg)
		writeJSONError(rw, logMsg, http.StatusBadRequest)
		return
	}

	err = h.invalidModel(uuid, model)
	//400
	if err != nil {
		logMsg := fmt.Sprintf("%s Error: %s", ErrorMsg_BadBody, err.Error())
		log.WithFields(log.Fields{"UUID": uuid}).Infof(logMsg)
		writeJSONError(rw, logMsg, http.StatusBadRequest)
		return
	}

	created, err := h.srv.Write(model)

	//503
	if err != nil {
		logMsg := fmt.Sprintf(LogMsg503, "storing")
		log.WithFields(log.Fields{"UUID": uuid}).Errorf("%s %s", logMsg, err.Error())
		writeJSONError(rw, logMsg, http.StatusServiceUnavailable)
		return
	}

	if created {
		rw.WriteHeader(http.StatusCreated)
		return
	} else {
		rw.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) HandleDelete(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars[UUID_Param]
	deleted, err := h.srv.Delete(uuid)
	//503
	if err != nil {
		logMsg := fmt.Sprintf(LogMsg503, "deleting")
		log.WithFields(log.Fields{"UUID": uuid}).Errorf("%s %s", logMsg, err.Error())
		writeJSONError(rw, logMsg, http.StatusServiceUnavailable)
		return
	}
	//404
	if !deleted {
		log.WithFields(log.Fields{"UUID": uuid}).Infof("%s for %s", LogMsg404, uuid)
		writeJSONError(rw, LogMsg404, http.StatusNotFound)
		return
	}
	//204
	rw.WriteHeader(http.StatusNoContent)
}

func writeJSONError(rw http.ResponseWriter, logMsg string, statusCode int) {
	rw.Header().Set("Content-Type", ContentTypeJson)
	rw.WriteHeader(statusCode)
	rw.Write([]byte(fmt.Sprintf(ErrorMsgJson, logMsg)))
}

func (h *Handler) invalidModel(uuid string, model db.ConcordancesModel) error {
	if model.UUID != uuid {
		return errors.New(ErrorMsg_MismatchedConceptId)
	}
	if (model.ConcordedIds == nil) || (len(model.ConcordedIds) < 1) {
		return errors.New(ErrorMsg_MissingConcordedIds)
	}
	return nil
}
