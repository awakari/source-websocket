package main

import (
	"context"
	"fmt"
	apiGrpc "github.com/awakari/source-websocket/api/grpc"
	"github.com/awakari/source-websocket/api/http/pub"
	"github.com/awakari/source-websocket/config"
	"github.com/awakari/source-websocket/model"
	"github.com/awakari/source-websocket/service"
	"github.com/awakari/source-websocket/service/converter"
	"github.com/awakari/source-websocket/service/handler"
	"github.com/awakari/source-websocket/storage/mongo"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

func main() {

	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		panic(fmt.Sprintf("failed to load the config from env: %s", err))
	}

	opts := slog.HandlerOptions{
		Level: slog.Level(cfg.Log.Level),
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	log.Info("starting the update for the feeds")

	// determine the replica index
	replicaNameParts := strings.Split(cfg.Replica.Name, "-")
	if len(replicaNameParts) < 2 {
		panic("unable to parse the replica name: " + cfg.Replica.Name)
	}
	var replicaIndex int
	replicaIndex, err = strconv.Atoi(replicaNameParts[len(replicaNameParts)-1])
	if err != nil {
		panic(err)
	}
	if replicaIndex < 0 {
		panic(fmt.Sprintf("Negative replica index: %d", replicaIndex))
	}
	log.Info(fmt.Sprintf("Replica: %d", replicaIndex))

	svcPub := pub.NewService(http.DefaultClient, cfg.Api.Writer.Uri, cfg.Api.Token.Internal, cfg.Api.Writer.Timeout)
	svcPub = pub.NewLogging(svcPub, log)
	log.Info("initialized the Awakari publish API client")

	ctx := context.Background()
	stor, err := mongo.NewStorage(ctx, cfg.Db)
	if err != nil {
		panic(err)
	}
	defer stor.Close()

	conv := converter.NewService(cfg.Api.Events.Type)
	conv = converter.NewLogging(conv, log)

	handlersLock := &sync.Mutex{}
	handlerByUrl := make(map[string]handler.Handler)
	handlerFactory := handler.NewFactory(cfg.Api, conv, svcPub, log)

	svc := service.NewService(stor, uint32(replicaIndex), handlersLock, handlerByUrl, handlerFactory)
	svc = service.NewServiceLogging(svc, log)
	if replicaIndex > 0 {
		err = resumeHandlers(ctx, log, svc, uint32(replicaIndex), handlersLock, handlerByUrl, handlerFactory)
		if err != nil {
			panic(err)
		}
	}

	log.Info(fmt.Sprintf("starting to listen the gRPC API @ port #%d...", cfg.Api.Port))
	err = apiGrpc.Serve(cfg.Api.Port, svc)
	if err != nil {
		panic(err)
	}
}

func resumeHandlers(
	ctx context.Context,
	log *slog.Logger,
	svc service.Service,
	replicaIndex uint32,
	handlersLock *sync.Mutex,
	handlerByUrl map[string]handler.Handler,
	handlerFactory handler.Factory,
) (err error) {
	var cursor string
	var urls []string
	var str model.Stream
	for {
		urls, err = svc.List(ctx, 100, model.Filter{}, model.OrderAsc, cursor)
		if err == nil {
			if len(urls) == 0 {
				break
			}
			cursor = urls[len(urls)-1]
			for _, url := range urls {
				str, err = svc.Read(ctx, url)
				if err == nil && str.Replica == replicaIndex {
					resumeHandler(ctx, log, url, str, handlersLock, handlerByUrl, handlerFactory)
				}
				if err != nil {
					break
				}
			}
		}
		if err != nil {
			break
		}
	}
	return
}

func resumeHandler(
	ctx context.Context,
	log *slog.Logger,
	url string,
	str model.Stream,
	handlersLock *sync.Mutex,
	handlerByUrl map[string]handler.Handler,
	handlerFactory handler.Factory,
) {
	handlersLock.Lock()
	defer handlersLock.Unlock()
	h := handlerFactory(url, str)
	handlerByUrl[url] = h
	go h.Handle(ctx)
	log.Info(fmt.Sprintf("resumed handler for %s", url))
}
