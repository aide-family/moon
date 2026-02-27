package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"
)

type Message interface {
	AppendMessage(ctx context.Context, messageUID snowflake.ID) error
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
}
