package node

import (
	"context"
	"fmt"
	"io"
)

type MethodName string

type ServiceId string

func (id ServiceId) WithArgument(argument string) ServiceId {
	return ServiceId(fmt.Sprintf("%s/%s", id, argument))
}

type Service interface {
	GetServiceId() ServiceId
	Handle(ctx context.Context, stream io.ReadWriteCloser)
}
