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
	auth0Domain = flag.String("auth0-domain", "cognizantai.auth0.com", "The domain assigned to the server API by Auth0")
	jwksCache   = &jwksState{
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
					})
				}
				jwksCache.Unlock()
				modules.SetModule("jwks", true)
			case <-quitC:
				return
			}
		}
	}()
}

func validateToken(token string, claimCheck string) (err errors.Error) {

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

	audience := []string{
		"http://api.sentient.ai/experimentsrv",
	}

	jwksCache.Lock()
	configuration := auth0.NewConfiguration(jwksCache.client, audience, "https://"+*auth0Domain+"/", jose.RS256)
	jwksCache.Unlock()

	validator := auth0.NewValidator(configuration)

	headerTokenRequest, errGo := http.NewRequest("", audience[0], nil)
	if errGo != nil {
		return errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime()).With("hint", "possibly the Bearer label is missing")
	}
	headerTokenRequest.Header.Add("Authorization", token)

	validResp, errGo := validator.ValidateRequest(headerTokenRequest)
	if errGo != nil {
		return errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime())
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

	if !strings.Contains(claims["scope"].(string), claimCheck) {
		return errors.New(fmt.Sprintf("the authenticated user does not have the appropriate '%s' scope", claimCheck)).With("stack", stack.Trace().TrimRuntime())
	}
	return nil
}

func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, grpc.Errorf(codes.Unauthenticated, "missing context metadata")
	}

	//if !authenticate(md["username"], md["password"]) {
	//	return nil, codes.Unauthenticated
	//}

	if len(meta["authorization"]) != 1 {
		return nil, grpc.Errorf(codes.Unauthenticated, "invalid security token")
	}
	if len(meta["authorization"][0]) == 0 {
		return nil, grpc.Errorf(codes.Unauthenticated, "empty security token")
	}
	if err := validateToken(meta["authorization"][0], ""); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, err.Error())
	}

	return handler(ctx, req)
}
