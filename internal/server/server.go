package server

import (
	"encoding/base64"
	"github.com/go-redis/redis/v7"
	"github.com/nandiheath/graphql-gateway/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"net/http"
	"time"
)

const (
	GraphqlQueryRequestPath       = "/graphql"
	GraphqlIntrospectionQueryName = "IntrospectionQuery"
)

type Server struct {
	HttpClient
	redisClient *redis.Client
}

// New Server
func NewServer(client HttpClient) *Server {

	if client == nil {
		client = &fasthttp.Client{
			MaxConnsPerHost: 1024,
		}

	}

	server := &Server{
		HttpClient: client,
	}

	if config.EnableCache {
		server.redisClient = redis.NewClient(&redis.Options{
			Addr: config.RedisHost,
		})
	}

	return server
}

type HttpClient interface {
	Do(req *fasthttp.Request, resp *fasthttp.Response) error
}

func (s *Server) Start() {
	log.Info().Msgf("Server has started to listen on PORT %s", config.Port)
	if err := fasthttp.ListenAndServe(":"+config.Port, s.requestHandler); err != nil {
		log.Error().Err(err).Msg("Server has shutdown with an error")
	}
}

func (s *Server) requestHandler(ctx *fasthttp.RequestCtx) {

	method := string(ctx.Method())
	path := string(ctx.Path())

	if method == http.MethodPost && path == GraphqlQueryRequestPath {
		s.handlePostGraphqlRequest(ctx)
	} else if method == http.MethodOptions {
		//log.Debug().Msg("Processing CORS Options Request")
		s.handleCORSRequest(ctx)
	} else {
		ctx.Error("Request not found", http.StatusNotFound)
	}
}

func (s *Server) handleCORSRequest(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	// https://github.com/apollographql/apollo-tracing
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	ctx.Response.Header.Set("Access-Control-Max-Age", "1728000")

	ctx.Response.SetStatusCode(http.StatusNoContent)
}

func (s *Server) handlePostGraphqlRequest(ctx *fasthttp.RequestCtx) {

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(config.GraphqlServerUri)
	req.Header.SetMethod(http.MethodPost)
	req.Header.Set("x-hasura-admin-secret", config.GraphqlServerSecret)
	req.Header.Set("content-type", "application/json")

	// do the caching
	cacheKey := base64.StdEncoding.EncodeToString(ctx.Request.Body())

	if config.EnableCache {
		cacheResult, err := s.redisClient.Get(cacheKey).Result()
		if err != nil && cacheResult != "" {
			ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
			// https://github.com/apollographql/apollo-tracing
			ctx.Response.Header.Set("Access-Control-Allow-Headers", "Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,*")
			ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			ctx.Response.SetBodyString(cacheResult)
			return
		}
	}

	// cache is not there
	req.SetBody(ctx.Request.Body())

	err := s.HttpClient.Do(req, resp)
	if err != nil {
		log.Error().Msgf("error when calling hasura. errror: %+v", err)
		ctx.SetStatusCode(400)
	} else {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		// https://github.com/apollographql/apollo-tracing
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		ctx.Response.SetBody(resp.Body())

		if config.EnableCache {
			err := s.redisClient.Set(cacheKey, resp.Body(), time.Hour*24).Err()
			if err != nil {
				log.Info().Msgf("cannot cache the result. error: %+v ", err)
			}
		}
	}


}
