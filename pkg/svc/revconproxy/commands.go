package revconproxy

import (
	"context"
	"log/slog"

	"github.com/ksysoev/oneway/api"
)

func (s *Proxy) ConnectCommandHandler(ctx context.Context, cmd *api.ConnectCommand) {
	err := s.rcpServ.CreateConnection(ctx, s.config.NameSpace, cmd.ServiceName)
	if err != nil {
		slog.Error("failed to create connection", slog.Any("error", err))
	}
}
