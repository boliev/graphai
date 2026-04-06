package vkapi

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/boliev/graphai/internal/handlers/me"
	"github.com/boliev/graphai/internal/handlers/vkHandler"
	"github.com/boliev/graphai/internal/pkg/config"
	"github.com/go-chi/chi/v5"
)

type VKApi struct {
}

func NewVKApi() *VKApi {
	return &VKApi{}
}

func (v *VKApi) Run() {

	cfg := config.New()
	v.startServer(cfg)
}

func (v *VKApi) startServer(cfg *config.Cfg) {
	r := chi.NewRouter()

	vk := vkHandler.NewHandler(cfg.VkSecureKey)
	meHandler := me.NewMeHandler()

	r.Post("/api/v1/vk", vk.Callback)
	r.Post("/api/v1/me/balance", meHandler.Balance)

	port, err := strconv.Atoi(cfg.VKApiPort)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(fmt.Sprintf("listen :%d", port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}
