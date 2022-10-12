package ants

import (
	"context"
)

type Runner interface {
	Run(ctx context.Context)
}
