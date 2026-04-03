package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	invoiceAPI "backend/api/invoice"
	taskAPI "backend/api/task"
	"backend/internal/repositories/memory"
	invoiceServices "backend/internal/services/invoice"
	taskServices "backend/internal/services/task"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func cors(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	// Repositories
	invoiceRepo := memory.NewInvoiceRepo()
	taskRepo := memory.NewTaskRepo()

	// Services
	invoiceSvc := invoiceServices.NewInvoiceService(invoiceRepo)
	taskSvc := taskServices.NewTaskService(taskRepo)

	// Handlers
	invoiceHandler := invoiceAPI.NewInvoiceHandler(invoiceSvc)
	taskHandler := taskAPI.NewTaskHandler(taskSvc)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors("http://localhost:5173"))

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
		r.Post("/{id}/complete", taskHandler.Complete)
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
