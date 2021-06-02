package biz

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"testing"
)

var testBiz *GreeterUsecase
var ctx = context.Background()

func TestInit(t *testing.T) {
	var (
		err        error
		controller *gomock.Controller
	)
	controller, ctx = gomock.WithContext(ctx, t)
	repo := NewMockGreeterRepo(controller)
	testBiz, err = newBiz(repo)
	require.NoError(t, err)
}
