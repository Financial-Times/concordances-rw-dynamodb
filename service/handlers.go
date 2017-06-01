package service

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	health "github.com/Financial-Times/go-fthealth/v1_1"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/rcrowley/go-metrics"
	log "github.com/Sirupsen/logrus"
)



type ServiceHandler interface {
	RegisterHandlers()
	RegisterAdminHandlers()
	HandleGet(rw http.ResponseWriter, r *http.Request)
	HandlePut(rw http.ResponseWriter, r *http.Request)
	HandleDelete(rw http.ResponseWriter, r *http.Request)
	HandleCount(rw http.ResponseWriter, r *http.Request)

}

type ConcordancesRwHandler struct {
	srv Service
	router *mux.Router
	appSystemCode string
	appName string
	port string
}

func NewHandler(conf AppConfig, router *mux.Router) ServiceHandler {

	s := NewConcordancesRwService(conf)
	h := ConcordancesRwHandler{srv: s, router : router, appSystemCode: conf.appSystemCode, appName: conf.appName, port: conf.port}

	return &h;
}

func (h *ConcordancesRwHandler) RegisterHandlers() {
	log.Info("Registering handlers")

	rwHandler := handlers.MethodHandler{
		"GET" : http.HandlerFunc(h.HandleGet),
		"PUT" : http.HandlerFunc(h.HandlePut),
		"DELETE" : http.HandlerFunc(h.HandleDelete),
	}
	countHandler := handlers.MethodHandler{
		"GET" : http.HandlerFunc(h.HandleCount),
	}

	h.router.Handle("/concordances/{uuid}", rwHandler)
	h.router.Handle("/concordances/__count", countHandler)
}

func (h *ConcordancesRwHandler) RegisterAdminHandlers()  {
	log.Info("Registering admin handlers")

	//TODO: write specific healthchecks
	healthService := newHealthService(&healthConfig{appSystemCode: h.appSystemCode, appName: h.appName, port: h.port})
	hc := health.HealthCheck{SystemCode: h.appSystemCode, Name: h.appName, Description: "appDescription", Checks: healthService.checks}
	h.router.HandleFunc(healthPath, health.Handler(hc))
	h.router.HandleFunc(status.PingPath, status.PingHandler)
	h.router.HandleFunc(status.GTGPath, status.NewGoodToGoHandler(healthService.gtgCheck))
	h.router.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)
	//
	monitoringRouter := httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), h.router)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)
	http.Handle("/", monitoringRouter)
}

func (h *ConcordancesRwHandler) HandleGet(rw http.ResponseWriter, r *http.Request) {
	h.srv.Read("uuid")
}

func (h *ConcordancesRwHandler) HandlePut(rw http.ResponseWriter, r *http.Request) {
	h.srv.Write(Model{})
}

func (h *ConcordancesRwHandler) HandleDelete(rw http.ResponseWriter, r *http.Request) {
	h.srv.Delete("uuid")
}

func (h *ConcordancesRwHandler) HandleCount(rw http.ResponseWriter, r *http.Request) {
	h.srv.Count()
}
