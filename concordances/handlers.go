package concordances

import (
	"encoding/json"
	"fmt"
	"net/http"

	db "github.com/Financial-Times/concordances-rw-dynamodb/dynamodb"
	health "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/Financial-Times/transactionid-utils-go"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rcrowley/go-metrics"
)

const (
	ContentTypeJson = "application/json"
	UUID_Param      = "uuid"
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
	tid := transactionidutils.GetTransactionIDFromRequest(r)

	model, err := h.srv.Read(uuid, tid)

	//503
	if err != nil {
		writeJSONError(rw, "Error retrieving concordances", http.StatusServiceUnavailable)
		return
	}
	//404
	if model.ConcordedIds == nil {
		log.WithFields(log.Fields{"UUID": uuid, "transaction_id": tid}).Info("Unable to find concordance")
		writeJSONError(rw, "Unable to find concordance", http.StatusNotFound)
		return
	}

	//200
	rw.Header().Set("Content-Type", ContentTypeJson)
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(&model)
}

func (h *Handler) HandlePut(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tid := transactionidutils.GetTransactionIDFromRequest(r)
	uuid := vars[UUID_Param]

	model := db.ConcordancesModel{}
	err := json.NewDecoder(r.Body).Decode(&model)
	defer r.Body.Close()

	//400
	if err != nil {
		log.WithFields(log.Fields{"UUID": uuid, "transaction_id": tid}).Error("Error decoding the JSON of the request body")
		writeJSONError(rw, "Error decoding the JSON of the request body", http.StatusBadRequest)
		return
	}

	//400
	if model.UUID == "" {
		log.WithFields(log.Fields{"UUID": uuid, "transaction_id": tid}).Error("Concept UUID in payload is different from UUID path parameter")
		writeJSONError(rw,"Concept UUID is missing from the Payload", http.StatusBadRequest)
		return
	}
	if model.UUID != uuid {
		log.WithFields(log.Fields{"UUID": uuid, "transaction_id": tid}).Error("Concept UUID in payload is different from UUID path parameter")
		writeJSONError(rw, fmt.Sprintf("Concept UUID (%s) in payload is different from UUID path parameter (%s)", model.UUID, uuid), http.StatusBadRequest)
		return
	}

	if (model.ConcordedIds == nil) || (len(model.ConcordedIds) < 1) {
		log.WithFields(log.Fields{"UUID": uuid, "transaction_id": tid}).Error("Payload has no concorded UUIDs to store")
		writeJSONError(rw, "Payload has no concorded UUIDs to store", http.StatusBadRequest)
		return
	}

	status, err := h.srv.Write(model, tid)

	//503
	if err != nil || status == db.CONCORDANCE_ERROR {
		writeJSONError(rw, "Error writing concordance", http.StatusServiceUnavailable)
		return
	}

	if status == db.CONCORDANCE_CREATED {
		rw.WriteHeader(http.StatusCreated)
		return
	} else {
		rw.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) HandleDelete(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars[UUID_Param]
	tid := transactionidutils.GetTransactionIDFromRequest(r)

	status, err := h.srv.Delete(uuid, tid)

	//503
	if err != nil || status == db.CONCORDANCE_ERROR {
		writeJSONError(rw, "Error deleting concordance", http.StatusServiceUnavailable)
		return
	}
	//404
	if status == db.CONCORDANCE_NOT_FOUND {
		log.WithFields(log.Fields{"UUID": uuid, "transaction_id": tid}).Info("Unable to find concordance")
		writeJSONError(rw, "Unable to find concordance", http.StatusNotFound)
		return
	}
	//204
	rw.WriteHeader(http.StatusNoContent)
}

func writeJSONError(rw http.ResponseWriter, logMsg string, statusCode int) {
	rw.Header().Set("Content-Type", ContentTypeJson)
	rw.WriteHeader(statusCode)
	rw.Write([]byte(fmt.Sprintf("{\"message\":\"%s\"}", logMsg)))
}
