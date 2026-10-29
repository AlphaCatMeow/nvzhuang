package tryon

import (
	"context"

	"backend/api/tryon/v1"
)

type ITryonV1 interface {
	Tryon(ctx context.Context, req *v1.TryonReq) (res *v1.TryonRes, err error)
}
