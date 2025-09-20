package server

import (
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/transistxr/coach-assignment-server/src/internal/clients"
	"github.com/transistxr/coach-assignment-server/src/internal/db"
	"github.com/transistxr/coach-assignment-server/src/internal/handlers"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Server struct {
	router             *chi.Mux
	DB                 *sql.DB
	AvailabilityClient *clients.AvailabilityClient
}

func New(sqlDB *sql.DB, rdb *db.RedisClient) *Server {
	r := chi.NewRouter()

	rateLimitString := db.GetValueString(rdb, "PROD_ENV_RATE_LIMIT")

	rateLimit, err := strconv.Atoi(rateLimitString)
	if err != nil {
		log.Printf("Invalid Rate Limit in Configuration: %s \n", rateLimitString)
	}

	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(httprate.Limit(
		rateLimit,
		time.Minute,
		httprate.WithResponseHeaders(httprate.ResponseHeaders{
			Limit:     "X-RateLimit-Limit",
			Remaining: "X-RateLimit-Remaining",
			Reset:     "X-RateLimit-Reset",
		}),
	))

	availabilityClient := clients.NewAvailabilityClient(os.Getenv("CALENDAR_API_URL"))
	crmClient := clients.NewCRMClient(os.Getenv("CRM_WEBHOOK_URL"))
	authClient := clients.NewAuthClient(os.Getenv("AUTH_SERVICE_URL"))

	deps := &handlers.HandlerDeps{
		DB:                 sqlDB,
		RDB:                rdb,
		AvailabilityClient: availabilityClient,
		CRMClient:          crmClient,
		AuthClient:         authClient,
	}

	schedulingHandler := &handlers.SchedulingHandler{Deps: deps}

	r.Get("/health", handlers.HealthCheck)

	r.Get("/api/availability", schedulingHandler.GetAvailability)
	r.Post("/api/appointments", schedulingHandler.BookAppointment)

	r.Post("/api/webhooks/calendar", schedulingHandler.WebhookHandler)
	r.Get("/api/coaches/distribution", schedulingHandler.GetCoachDistribution)

	return &Server{router: r}
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
