package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	agencyAPI "backend/api/agency"
	authAPI "backend/api/auth"
	invoiceAPI "backend/api/invoice"
	taskAPI "backend/api/task"
	userAPI "backend/api/user"
	authMiddleware "backend/internal/middleware"
	"backend/internal/repositories/memory"
	"backend/internal/repositories/postgres"
	agencyServices "backend/internal/services/agency"
	authServices "backend/internal/services/auth"
	invoiceServices "backend/internal/services/invoice"
	taskServices "backend/internal/services/task"
	userServices "backend/internal/services/user"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func cors(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func allowedOrigin() string {
	if o := os.Getenv("CORS_ORIGIN"); o != "" {
		return o
	}
	return "http://localhost:5173"
}

func main() {
	ctx := context.Background()

	// Required environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		log.Fatal("SUPABASE_URL is required")
	}
	anonKey := os.Getenv("SUPABASE_ANON_KEY")
	if anonKey == "" {
		log.Fatal("SUPABASE_ANON_KEY is required")
	}
	serviceRoleKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if serviceRoleKey == "" {
		log.Fatal("SUPABASE_SERVICE_ROLE_KEY is required")
	}
	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("SUPABASE_JWT_SECRET is required")
	}

	// Database
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("open db pool: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("connected to database")

	// Repositories
	invoiceRepo := memory.NewInvoiceRepo()
	taskRepo    := postgres.NewTaskRepo(pool)
	userRepo    := postgres.NewUserRepo(pool)
	agencyRepo  := postgres.NewAgencyRepo(pool)

	// Services
	invoiceSvc := invoiceServices.NewInvoiceService(invoiceRepo)
	taskSvc    := taskServices.NewTaskService(taskRepo)
	userSvc    := userServices.NewUserService(userRepo)
	agencySvc  := agencyServices.NewAgencyService(agencyRepo)
	authSvc    := authServices.NewAuthService(supabaseURL, anonKey, serviceRoleKey, agencyRepo, userRepo)

	// Handlers
	invoiceHandler := invoiceAPI.NewInvoiceHandler(invoiceSvc)
	taskHandler    := taskAPI.NewTaskHandler(taskSvc)
	userHandler    := userAPI.NewUserHandler(userSvc)
	agencyHandler  := agencyAPI.NewAgencyHandler(agencySvc)
	authHandler    := authAPI.NewAuthHandler(authSvc)

	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(cors(allowedOrigin()))

	// Health check — no auth required
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Public routes — no JWT required
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// Protected routes — JWT required + account must be activated
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Authenticate(jwtSecret, supabaseURL))
		r.Use(authMiddleware.RequireActivated(userRepo))

		r.Route("/api/invoices", func(r chi.Router) {
			r.Get("/", invoiceHandler.List)
			r.Post("/", invoiceHandler.Create)
			r.Get("/summary", invoiceHandler.Summary)
			r.Post("/{id}/pay", invoiceHandler.Pay)
		})

		r.Route("/api/tasks", func(r chi.Router) {
			r.Get("/", taskHandler.List)
			r.Post("/", taskHandler.Create)
			r.Get("/{id}", taskHandler.Get)
			r.Post("/{id}/assign", taskHandler.Assign)
			r.Post("/{id}/unassign", taskHandler.Unassign)
			r.Post("/{id}/complete", taskHandler.Complete)
			r.Post("/{id}/set-in-progress", taskHandler.SetInProgress)
			r.Patch("/{id}/due-date", taskHandler.UpdateDueDate)
			r.Patch("/{id}/description", taskHandler.UpdateDescription)
			r.Patch("/{id}/tags", taskHandler.UpdateTags)
		})

		r.Route("/api/users", func(r chi.Router) {
			r.Post("/", userHandler.Create)
			r.Get("/", userHandler.ListByAgency)
			r.Get("/{id}", userHandler.Get)
		})

		r.Route("/api/agencies", func(r chi.Router) {
			r.Post("/", agencyHandler.Create)
			r.Get("/", agencyHandler.List)
			r.Get("/{id}", agencyHandler.Get)
		})
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("server listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
