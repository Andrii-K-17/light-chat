package router

import (
	"net/http"
	"time"

	"github.com/Andrii-K-17/light-chat/internal/config"
	"github.com/Andrii-K-17/light-chat/internal/handlers"
	"github.com/Andrii-K-17/light-chat/internal/middleware"
	"github.com/Andrii-K-17/light-chat/internal/repository"
	"github.com/Andrii-K-17/light-chat/internal/services"
	"github.com/Andrii-K-17/light-chat/internal/ws"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
)

// New initializes and configures the main application router.
func New(db *sqlx.DB, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.AllowedOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	userRepo := repository.NewUserRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)
	chatRepo := repository.NewChatRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	authSvc := services.NewAuthService(userRepo, refreshRepo)

	hub := ws.NewHub()

	authH := handlers.NewAuthHandler(
		authSvc,
		userRepo,
		cfg.JWTSecret,
		cfg.JWTExpiry,
		cfg.RefreshExpiry,
		cfg.IsProd(),
	)
	chatH := handlers.NewChatHandler(chatRepo, messageRepo, userRepo)
	wsH := ws.NewHandler(hub, messageRepo, chatRepo, cfg.JWTSecret)

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/register", authH.Register)
		r.Post("/auth/login", authH.Login)
		r.Post("/auth/logout", authH.Logout)
		r.Post("/auth/refresh", authH.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.JWTSecret))

			r.Get("/auth/me", authH.Me)
			r.Patch("/auth/me", authH.UpdateProfile)
			r.Get("/users/search", authH.SearchUser)

			r.Get("/chats", chatH.GetChats)
			r.Post("/chats", chatH.CreateChat)
			r.Get("/chats/{id}/messages/search", chatH.SearchMessages)
			r.Get("/chats/{id}/messages", chatH.GetMessages)
		})
	})

	r.Get("/ws", wsH.ServeWS)

	return r
}
