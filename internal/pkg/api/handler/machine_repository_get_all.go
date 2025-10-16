package handler

import (
	"context"
	"encoding/json"
	"fmt"
	stdhttp "net/http"

	"github.com/go-chi/chi/v5"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/http"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

type MachineRepositoryGetAll struct {
	machineRepository machineSDK.Repository
}

func NewMachineRepositoryGetAll(machineRepository machineSDK.Repository) *MachineRepositoryGetAll {
	return &MachineRepositoryGetAll{
		machineRepository: machineRepository,
	}
}

var _ http.ServerHandler = (*MachineRepositoryGetAll)(nil)

func (handler *MachineRepositoryGetAll) Register(router chi.Router) {
	router.Get("/machines", handler.ServeHTTP)
}

func (handler *MachineRepositoryGetAll) ServeHTTP(responseWriter stdhttp.ResponseWriter, request *stdhttp.Request) {
	if request.Method != stdhttp.MethodGet {
		responseWriter.WriteHeader(stdhttp.StatusMethodNotAllowed)
		_, _ = responseWriter.Write([]byte("method not allowed"))
		return
	}

	machines, err := handler.handle(request.Context())
	if err != nil {
		responseWriter.WriteHeader(stdhttp.StatusInternalServerError)
		_, _ = responseWriter.Write([]byte(err.Error()))
		return
	}

	responseBody := make([]MachineResponse, 0, len(machines))
	for _, machine := range machines {
		responseBody = append(responseBody, MachineResponse{
			MachineId:   machine.GetId(),
			MachineName: machine.GetName(),
		})
	}

	responseData, err := json.Marshal(responseBody)
	if err != nil {
		responseWriter.WriteHeader(stdhttp.StatusInternalServerError)
		_, _ = responseWriter.Write([]byte(err.Error()))
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(stdhttp.StatusOK)
	_, _ = responseWriter.Write(responseData)
}

func (handler *MachineRepositoryGetAll) handle(ctx context.Context) ([]machineSDK.Machine, error) {
	machines, err := handler.machineRepository.GetAllMachines(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all machines: %w", err)
	}

	return machines, nil
}
