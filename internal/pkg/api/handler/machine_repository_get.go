package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	stdhttp "net/http"
	"strings"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/http"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

type MachineRepositoryGet struct {
	machineRepository machineSDK.Repository
}

func NewMachineRepositoryGet(machineRepository machineSDK.Repository) *MachineRepositoryGet {
	return &MachineRepositoryGet{
		machineRepository: machineRepository,
	}
}

var _ http.ServerHandler = (*MachineRepositoryGet)(nil)

func (handler *MachineRepositoryGet) Register(router chi.Router) {
	router.Get("/machine/{machineIdentifier}", handler.ServeHTTP)
}

func (handler *MachineRepositoryGet) ServeHTTP(responseWriter stdhttp.ResponseWriter, request *stdhttp.Request) {
	if request.Method != stdhttp.MethodGet {
		responseWriter.WriteHeader(stdhttp.StatusMethodNotAllowed)
		_, _ = responseWriter.Write([]byte("method not allowed"))
		return
	}

	machineIdentifierPath := chi.URLParam(request, "machineIdentifier")
	if machineIdentifierPath == "" {
		responseWriter.WriteHeader(stdhttp.StatusBadRequest)
		_, _ = responseWriter.Write([]byte("machine identifier is required"))
		return
	}

	machineIdentifier, err := handler.parseMachineIdentifier(machineIdentifierPath)
	if err != nil {
		responseWriter.WriteHeader(stdhttp.StatusBadRequest)
		_, _ = responseWriter.Write([]byte(err.Error()))
		return
	}

	machine, err := handler.handle(request.Context(), machineIdentifier)
	if errors.Is(err, machineSDK.ErrMachineNotFound) {
		responseWriter.WriteHeader(stdhttp.StatusNotFound)
		_, _ = responseWriter.Write([]byte(err.Error()))
		return
	} else if err != nil {
		responseWriter.WriteHeader(stdhttp.StatusInternalServerError)
		_, _ = responseWriter.Write([]byte(err.Error()))
		return
	}

	responseBody := MachineResponse{
		MachineId:   machine.GetId(),
		MachineName: machine.GetName(),
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

type machineRepositoryIdentifierKind int

const (
	machineRepositoryIdentifierKindUnknown machineRepositoryIdentifierKind = iota
	machineRepositoryIdentifierKindId
	machineRepositoryIdentifierKindName
)

type machineRepositoryIdentifier struct {
	identifierKind machineRepositoryIdentifierKind
	machineId      machineSDK.MachineId
	machineName    machineSDK.MachineName
}

func (handler *MachineRepositoryGet) parseMachineIdentifier(identifierPath string) (machineRepositoryIdentifier, error) {
	const (
		idPrefix   = "id:"
		namePrefix = "name:"
	)

	switch {
	case strings.HasPrefix(identifierPath, idPrefix):
		rawIdentifier := strings.TrimPrefix(identifierPath, idPrefix)
		if strings.Contains(rawIdentifier, "/") {
			return machineRepositoryIdentifier{}, errors.New("machine identifier contains unexpected path segment")
		}

		machineId, err := machineSDK.NewMachineId(rawIdentifier)
		if err != nil {
			return machineRepositoryIdentifier{}, fmt.Errorf("invalid machine id: %w", err)
		}

		return machineRepositoryIdentifier{
			identifierKind: machineRepositoryIdentifierKindId,
			machineId:      machineId,
		}, nil
	case strings.HasPrefix(identifierPath, namePrefix):
		rawIdentifier := strings.TrimPrefix(identifierPath, namePrefix)
		if strings.Contains(rawIdentifier, "/") {
			return machineRepositoryIdentifier{}, errors.New("machine identifier contains unexpected path segment")
		}

		machineName, err := machineSDK.NewMachineName(rawIdentifier)
		if err != nil {
			return machineRepositoryIdentifier{}, fmt.Errorf("invalid machine name: %w", err)
		}

		return machineRepositoryIdentifier{
			identifierKind: machineRepositoryIdentifierKindName,
			machineName:    machineName,
		}, nil
	default:
		return machineRepositoryIdentifier{}, errors.New("machine identifier must start with id: or name:")
	}
}

func (handler *MachineRepositoryGet) handle(ctx context.Context, identifier machineRepositoryIdentifier) (machineSDK.Machine, error) {
	switch identifier.identifierKind {
	case machineRepositoryIdentifierKindId:
		machine, err := handler.machineRepository.GetMachineById(ctx, identifier.machineId)
		if err != nil {
			return nil, fmt.Errorf("get machine by id %s: %w", identifier.machineId, err)
		}
		return machine, nil
	case machineRepositoryIdentifierKindName:
		machine, err := handler.machineRepository.GetMachineByName(ctx, identifier.machineName)
		if err != nil {
			return nil, fmt.Errorf("get machine by name %s: %w", identifier.machineName, err)
		}
		return machine, nil
	default:
		return nil, fmt.Errorf("unsupported machine identifier kind: %d", identifier.identifierKind)
	}
}

type MachineResponse struct {
	MachineId   machineSDK.MachineId   `json:"machineId"`
	MachineName machineSDK.MachineName `json:"machineName"`
}
