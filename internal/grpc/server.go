package grpc

import (
	"context"
	"io/ioutil"
	"net"

	"github.com/golang/protobuf/ptypes/empty"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/address"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/credit"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Config struct {
	Address              string `toml:"address"`
	JWTRSAPrivateKeyFile string `toml:"private_key_file"`
	JWTRSAPublicKeyFile  string `toml:"public_key_file"`
}

type Server struct {
	UnimplementedIPPSServer
	config         Config
	addressStorage address.Storage
	creditStorage  credit.Storage
	userStorage    user.Storage
	privateKey     []byte
	publicKey      []byte
}

func NewServer(config *Config, as address.Storage, cs credit.Storage, us user.Storage) (*Server, error) {
	sk, err := ioutil.ReadFile(config.JWTRSAPrivateKeyFile)
	if err != nil {
		return nil, err
	}
	pk, err := ioutil.ReadFile(config.JWTRSAPublicKeyFile)
	if err != nil {
		return nil, err
	}
	s := &Server{
		config:         *config,
		addressStorage: as,
		creditStorage:  cs,
		userStorage:    us,
		privateKey:     sk,
		publicKey:      pk,
	}

	return s, nil
}

func (s *Server) ListenAndServe() error {
	rpcSrv := grpc.NewServer(grpc.UnaryInterceptor(s.authenticate))
	RegisterIPPSServer(rpcSrv, s)
	sock, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return err
	}
	return rpcSrv.Serve(sock)
}

var ErrUserOrPasswordWrong = status.Error(codes.PermissionDenied,
	"user does not exist or password is wrong")

// Login is the RPC call that logs a user in, returning a JSON Web Token
// which is used to authenticate users by the GRPC API and other services.
func (s *Server) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	username := req.GetUsername()
	u, err := s.userStorage.ByUsername(username)
	if err == user.ErrUserNotExists {
		return nil, ErrUserOrPasswordWrong
	} else if err != nil {
		return nil, err
	}
	pw := req.GetPassword()
	if !u.PasswordEquals(string(pw)) {
		return nil, ErrUserOrPasswordWrong
	}
	tok, err := NewJWT(u.Username, "RSA", []byte(s.privateKey))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &LoginResponse{AuthToken: tok}, nil
}

// GetPublicKey returns the server's PEM encoded public RSA key which
// is used to validate the signature of JSON Web Tokens handed out by
// the GRPC API. It is mainly intended to be used by other company's
// web services that use JWT Tokens to authenticate our users.
func (s *Server) GetPublicKey(ctx context.Context, req *empty.Empty) (*PublicKey, error) {
	return &PublicKey{Key: string(s.publicKey)}, nil
}

// AddAddress adds addr to the user's addresses.
func (s *Server) AddAddress(ctx context.Context, addr *Address) (*empty.Empty, error) {
	u := user.MustFromContext(ctx)
	a, err := address.NewForUser(u)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	a.Street = addr.Street
	a.Zip = addr.Zip
	a.City = addr.City
	a.Country = addr.Country
	a.Planet = addr.Planet

	err = s.addressStorage.Insert(a)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &empty.Empty{}, nil
}

// GetAddress returns the current user's addresses.
func (s *Server) GetAddresses(ctx context.Context, req *empty.Empty) (*Addresses, error) {
	u := user.MustFromContext(ctx)
	aa, err := s.addressStorage.ByUser(u)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	addrs := make([]*Address, 0, len(aa))
	for _, a := range aa {
		addrs = append(addrs, &Address{
			Street:  a.Street,
			Zip:     a.Zip,
			City:    a.City,
			Country: a.Country,
			Planet:  a.Planet,
		})
	}

	return &Addresses{Addresses: addrs}, nil
}

// AddCreditCard adds a credit card to the current user's payment options.
func (s *Server) AddCreditCard(ctx context.Context, card *CreditCard) (*empty.Empty, error) {
	u := user.MustFromContext(ctx)
	c, err := credit.NewCard(u)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	c.Number = card.Number
	err = s.creditStorage.Insert(c)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &empty.Empty{}, nil
}

// GetCreditCards returns the current user's credit cards.
func (s *Server) GetCreditCards(ctx context.Context, req *empty.Empty) (*CreditCards, error) {
	u := user.MustFromContext(ctx)
	cc, err := s.creditStorage.ByUser(u)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	cards := make([]*CreditCard, 0, len(cc))
	for _, c := range cc {
		card := &CreditCard{
			Number: c.Number,
		}
		cards = append(cards, card)
	}

	return &CreditCards{Cards: cards}, nil
}
