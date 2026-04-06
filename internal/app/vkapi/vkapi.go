package vkapi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/boliev/graphai/internal/domain/user"
	"github.com/boliev/graphai/internal/handlers/me"
	"github.com/boliev/graphai/internal/handlers/vkHandler"
	"github.com/boliev/graphai/internal/infra/pg/repository"
	"github.com/boliev/graphai/internal/pkg/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VKApi struct {
}

func NewVKApi() *VKApi {
	return &VKApi{}
}

func (v *VKApi) Run() {

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}
	v.startServer(cfg)
}

func (v *VKApi) startServer(cfg *config.Cfg) {
	ctx := context.Background()
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"https://graphai-pay.ai128.ru",
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"Origin",
		},
		ExposedHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           int((10 * time.Minute).Seconds()),
	}))

	pool, err := pgxpool.New(ctx, cfg.PGConnect)
	if err != nil {
		panic(err)
	}

	vk := vkHandler.NewHandler(cfg.VkSecureKey)

	userRepo := repository.NewUserRepo(pool)
	userService := user.NewService(userRepo)
	meHandler := me.NewMeHandler(userService, cfg.VkSecureKey)

	r.Post("/api/v1/vk", vk.Callback)
	r.Get("/api/v1/me/balance", meHandler.Balance)

	port, err := strconv.Atoi(cfg.VKApiPort)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(fmt.Sprintf("listen :%d", port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}
