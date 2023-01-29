package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/lucheng0127/narwhal/internal/pkg/log"
)

type TraceCtx interface {
	NewTraceContext() context.Context
	AddContextTraceID(ctx context.Context) context.Context
}

type TraceID struct {
	id string
}

func NewTraceID() *TraceID {
	tId := new(TraceID)
	tId.genUUID()
	return tId
}

func (tId *TraceID) genUUID() {
	uuid := "00000000"
	u := make([]byte, 4)
	_, err := rand.Read(u)
	if err == nil {
		uuid = hex.EncodeToString(u)
	}
	tId.id = uuid
}

func (tId *TraceID) AddContextTraceID(ctx context.Context) context.Context {
	tId.genUUID()
	return context.WithValue(ctx, log.MSG_ID, tId.id)
}

func (tId *TraceID) NewTraceContext() context.Context {
	ctx := context.Background()
	return tId.AddContextTraceID(ctx)
}
