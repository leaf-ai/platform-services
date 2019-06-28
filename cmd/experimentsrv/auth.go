package main

// This file contains the implementation of per request authentication
// for our gRPC server.  Authetication and validation are done using the Auth0
// platform.

import (
	"flag"
	"fmt"
	"strings"
	"sync"
	"time"

	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	"github.com/leaf-ai/platform-services/internal/platform"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/auth0-community/go-auth0"
	"gopkg.in/square/go-jose.v2"

	"github.com/golang/groupcache/lru"
)

var (
	auth0Enable   = flag.Bool("auth0-enable", true, "This can be used to enable auth checking when running in a mesh for example")
	auth0Scope    = flag.String("auth0-scope", "all:experiments", "The scope that must be claimed in order to be permitted access to the service")
	auth0Audience = flag.String("auth0-audience", "http://api.cognizant-ai.dev/experimentsrv", "The audience URL raw token string received from an invocation of {auth0-domain}/oauth/token}")
	auth0Domain   = flag.String("auth0-domain", "cognizant-ai.auth0.com", "The domain assigned to the server API by Auth0")
	jwksCache     = &jwksState{
		ok: false,
	}

	cacheSize = flag.Int("token-cache-size", 4096, "the number of entries to limit the JWT cache size to")
	cache     = &tokenCache{}
)

type jwksState struct {
	client *auth0.JWKClient
	ok     bool
	sync.Mutex
}

type tokenCache struct {
	// Golang teams cache that is a subset of a distributed memcached group cache
	// but sits within a single host for now, later we might add a distributed
	// in mesh cache if the system if the system grows to enourmas proportions
	lru *lru.Cache
	sync.Mutex
}

func initJwksUpdate(quitC <-chan struct{}) {

	// Initialize a cache for our auth0 JWT authorizations
	cache.Lock()
	cache.lru = lru.New(*cacheSize)
	cache.Unlock()

	// When starting set the auth module to be down and only when it load the JWKS successfully set the state to up
	serverModule := "jwks"
	modules := &Modules{}
	modules.SetModule(serverModule, false)

	go func() {
		// Used for recording the states of server components
		modules := &Modules{}

		updateCycle := time.Duration(5 * time.Second)
		for {
			select {
			case <-time.After(updateCycle):
				func() {
					jwksCache.Lock()
					defer jwksCache.Unlock()
				}()
				jwksCache.Lock()
				if jwksCache.client == nil {
					jwksCache.client = auth0.NewJWKClient(auth0.JWKClientOptions{
						URI: "https://" + *auth0Domain + "/.well-known/jwks.json",
					}, nil)
				}
				jwksCache.Unlock()
				modules.SetModule("jwks", true)
			case <-quitC:
				return
			}
		}
	}()
}

func validateToken(token string, audience []string, claimCheck string) (err errors.Error) {

	claims := map[string]interface{}{}
	cache.Lock()
	item, isPresent := cache.lru.Get(token)
	cache.Unlock()

	if isPresent {
		claims = item.(map[string]interface{})
		exp, isExpPresent := claims["exp"]
		if isExpPresent {
			// Check the time at which the claim expires and if it has reject the request BUT dont
			// clear the cache so that any further attempts wont result in a round trip to the
			// ID provider
			expires := time.Unix(int64(platform.Round(exp.(float64))), 0)
			if expires.Before(time.Now().UTC()) {
				return errors.New("token has expired").With("stack", stack.Trace().TrimRuntime())
			}
			logger.Debug("cache hit")
			return nil
		}
	}

	logger.Debug("cache miss")

	jwksCache.Lock()
	configuration := auth0.NewConfiguration(jwksCache.client, audience, "https://"+*auth0Domain+"/", jose.RS256)
	jwksCache.Unlock()

	validator := auth0.NewValidator(configuration, nil)

	headerTokenRequest, errGo := http.NewRequest("", audience[0], nil)
	if errGo != nil {
		return errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime()).With("hint", "possibly the Bearer label is missing")
	}
	headerTokenRequest.Header.Add("Authorization", token)

	validResp, errGo := validator.ValidateRequest(headerTokenRequest)
	if errGo != nil {
		return errors.Wrap(errGo).With("token", "..."+token[len(token)-6:], "stack", stack.Trace().TrimRuntime())
	}

	if len(claimCheck) == 0 {
		return nil
	}

	errGo = validator.Claims(headerTokenRequest, validResp, &claims)
	if errGo != nil {
		return errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime())
	}

	// Get ready to cache things, RFC 7519 nbf, exp, iat are also validated by the provider so
	// if we are here then we really only need to check the exp for caching purposes.a c.f.
	// https://tools.ietf.org/html/rfc7519#section-4.1.4

	exp, isExpPresent := claims["exp"]
	if !isExpPresent {
		return errors.New("token did not contain an expiry").With("stack", stack.Trace().TrimRuntime())
	}
	expires := time.Unix(int64(platform.Round(exp.(float64))), 0)
	if expires.Before(time.Now().UTC()) {
		return errors.New("token has expired").With("stack", stack.Trace().TrimRuntime())
	}
	cache.Lock()
	cache.lru.Add(token, claims)
	cache.Unlock()

	if _, isPresent := claims["scope"]; !isPresent {
		return errors.New(fmt.Sprintf("the authenticated user has no roles for this API, specifically the '%s' scope is missing", claimCheck)).With("stack", stack.Trace().TrimRuntime())
	}
	if !strings.Contains(claims["scope"].(string), claimCheck) {
		return errors.New(fmt.Sprintf("the authenticated user does not have the appropriate '%s' scope", claimCheck)).With("stack", stack.Trace().TrimRuntime())
	}
	return nil
}

func authUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if *auth0Enable {
		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, grpc.Errorf(codes.Unauthenticated, "missing context metadata "+stack.Trace().TrimRuntime().String())
		}

		auth, isPresent := meta["authorization"]
		if !isPresent {
			return nil, grpc.Errorf(codes.Unauthenticated, "missing security token "+stack.Trace().TrimRuntime().String())
		}
		if len(auth) != 1 {
			return nil, grpc.Errorf(codes.Unauthenticated, fmt.Sprint("unexpected security token", auth, stack.Trace().TrimRuntime().String()))
		}
		if len(auth[0]) == 0 {
			return nil, grpc.Errorf(codes.Unauthenticated, "empty security token "+stack.Trace().TrimRuntime().String())
		}

		if err := validateToken(auth[0], []string{*auth0Audience}, *auth0Scope); err != nil {
			return nil, grpc.Errorf(codes.Unauthenticated, err.Error())
		}
	}
	return handler(ctx, req)
}

func authStreamInterceptor(srv interface{}, strm grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (errGo error) {
	if *auth0Enable {
		meta, ok := metadata.FromIncomingContext(strm.Context())
		if !ok {
			return grpc.Errorf(codes.Unauthenticated, "missing context metadata "+stack.Trace().TrimRuntime().String())
		}

		logger.Debug(stack.Trace().TrimRuntime().String())
		logger.Debug(spew.Sdump(strm.Context()))
		logger.Debug(spew.Sdump(srv))
		auth, isPresent := meta["authorization"]
		if !isPresent {
			return grpc.Errorf(codes.Unauthenticated, "missing security token "+stack.Trace().TrimRuntime().String())
		}
		if len(auth) != 1 {
			return grpc.Errorf(codes.Unauthenticated, fmt.Sprint("unexpected security token", auth, stack.Trace().TrimRuntime().String()))
		}
		if len(auth[0]) == 0 {
			return grpc.Errorf(codes.Unauthenticated, "empty security token "+stack.Trace().TrimRuntime().String())
		}

		if err := validateToken(auth[0], []string{*auth0Audience}, *auth0Scope); err != nil {
			return grpc.Errorf(codes.Unauthenticated, err.Error())
		}
	}
	return handler(srv, strm)
}
