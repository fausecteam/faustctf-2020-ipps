package grpc

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	ErrUnsupportedAlgorithm = errors.New("jwt: the signing method is not supported")
	ErrNoAuthHeader         = status.Error(codes.Unauthenticated, "no authorization header in request")
	ErrInvalidAuthHeader    = status.Error(codes.Unauthenticated, "authorization header is invalid")
	ErrJWTInvalid           = status.Error(codes.InvalidArgument, "authorization token is invalid")
)

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func NewJWT(username string, algorithm string, key []byte) (string, error) {

	var err error
	var k interface{}
	var signingMethod jwt.SigningMethod
	switch algorithm {
	case "HMAC":
		k = key
		signingMethod = jwt.SigningMethodHS256
	case "RSA":
		k, err = jwt.ParseRSAPrivateKeyFromPEM(key)
		if err != nil {
			return "", err
		}
		signingMethod = jwt.SigningMethodRS256
	default:
		return "", ErrUnsupportedAlgorithm
	}
	tok := jwt.NewWithClaims(signingMethod, &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour * 31).Unix(),
			Issuer:    "ipps",
			NotBefore: time.Now().Unix(),
		},
	})
	s, err := tok.SignedString(k)
	if err != nil {
		return "", err
	}

	return s, nil
}

// JWTCredentials is the type implementing the
// grpc/credentials.PerRPCCredentials interface.
type JWTCredentials struct {
	token string
}

func NewJWTCredentials(jwToken string) *JWTCredentials {
	return &JWTCredentials{token: jwToken}
}

func (c *JWTCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": c.token,
	}, nil
}

func (c *JWTCredentials) RequireTransportSecurity() bool {
	return false
}

func (s *Server) authenticate(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if info.FullMethod == "/grpc.IPPS/Login" ||
		info.FullMethod == "/grpc.IPPS/GetPublicKey" {
		return handler(ctx, req)
	}

	tok, err := extractJWT(ctx)
	if err != nil {
		log.Printf("grpc: %v\n", err)
		return nil, err
	}
	username, err := authenticateUser(tok, s.publicKey)
	if err != nil {
		log.Printf("grpc: %v\n", err)
		return nil, err
	}
	u, err := s.userStorage.ByUsername(username)
	if err != nil {
		log.Printf("ByUsername: %v\n", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	ctx = user.NewContext(ctx, u)

	return handler(ctx, req)
}

func authenticateUser(jwToken string, key []byte) (string, error) {
	tok, err := jwt.ParseWithClaims(jwToken, &Claims{},
		func(token *jwt.Token) (interface{}, error) {
			alg := token.Method.Alg()
			if len(alg) < 2 {
				return nil, status.Error(codes.InvalidArgument, ErrUnsupportedAlgorithm.Error())
			}

			alg = strings.ToUpper(alg[:2])
			switch alg {
			case "HS":
				return key, nil
			case "RS":
				k, err := jwt.ParseRSAPublicKeyFromPEM(key)
				if err != nil {
					return nil, status.Error(codes.Internal, err.Error())
				}

				return k, nil
			default:
				return nil, status.Error(codes.InvalidArgument, ErrUnsupportedAlgorithm.Error())
			}
		})
	if err != nil {
		return "", err
	}
	if !tok.Valid {
		return "", ErrJWTInvalid
	}
	c := tok.Claims.(*Claims)

	return c.Username, nil
}

func extractJWT(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrNoAuthHeader
	}
	aa, ok := md["authorization"]
	if !ok {
		return "", ErrNoAuthHeader
	}
	if len(aa) != 1 {
		return "", ErrInvalidAuthHeader
	}

	return aa[0], nil
}
