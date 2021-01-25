package main

import (
	"context"
	"net"

	"github.com/dgrijalva/jwt-go"
	"github.com/fausecteam/ctf-gameserver/go/checkerlib"
	"github.com/golang/protobuf/ptypes/empty"
	pb "gitlab.cs.fau.de/faust/faustctf-2020/ipps/internal/grpc"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/address"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/credit"
	"google.golang.org/grpc"
)

type GRPCClient struct {
	client    pb.IPPSClient
	address   string
	authToken string
}

const grpcPort = "8001"

func NewGRPCClient(ip string) (*GRPCClient, error) {
	ctx, cancel := newTimeoutContext()
	defer cancel()


	addr := net.JoinHostPort(ip, grpcPort)
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	c := pb.NewIPPSClient(conn)

	return &GRPCClient{
		client:  c,
		address: addr,
	}, nil
}

func (gc *GRPCClient) Login(username, password string) error {
	req := &pb.LoginRequest{
		Username: username,
		Password: []byte(password),
	}

	ctx, cancel1 := newTimeoutContext()
	defer cancel1()
	resp, err := gc.client.Login(ctx, req)
	if err == context.DeadlineExceeded {
		return err
	} else if err != nil {
		return ErrLoginFailed
	}
	cred := pb.NewJWTCredentials(resp.AuthToken)

	ctx, cancel2 := newTimeoutContext()
	defer cancel2()
	conn, err := grpc.DialContext(ctx, gc.address, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithPerRPCCredentials(cred))
	if err != nil {
		return err
	}
	gc.client = pb.NewIPPSClient(conn)
	gc.authToken = resp.AuthToken

	return nil
}

func (gc *GRPCClient) AddAddress(addr *address.Address) error {
	if !gc.isLoggedIn() {
		return &NotLoggedInError{action: "grpc: AddAddress"}
	}

	a := &pb.Address{
		Street:  addr.Street,
		Zip:     addr.Zip,
		City:    addr.City,
		Country: addr.Country,
		Planet:  addr.Planet,
	}
	ctx, cancel := newTimeoutContext()
	defer cancel()
	_, err := gc.client.AddAddress(ctx, a)
	if err == context.DeadlineExceeded {
		return err
	} else if err != nil {
		return ErrAddAddressFailed
	}

	return nil
}

func (gc *GRPCClient) HasAddress(addr *address.Address) (bool, error) {
	if !gc.isLoggedIn() {
		return false, &NotLoggedInError{action: "grpc: HasAddress"}
	}

	ctx, cancel := newTimeoutContext()
	defer cancel()
	aa, err := gc.client.GetAddresses(ctx, &empty.Empty{})
	if err == context.DeadlineExceeded {
		return false, err
	} else if err != nil {
		return false, nil
	}
	for _, a := range aa.Addresses {
		cur := &address.Address{
			Street:  a.Street,
			Zip:     a.Zip,
			City:    a.City,
			Country: a.Country,
			Planet:  a.Planet,
		}
		if isSameAddress(cur, addr) {
			return true, nil
		}
	}

	return false, nil
}

func (gc *GRPCClient) AddCreditCard(card *credit.Card) error {
	if !gc.isLoggedIn() {
		return &NotLoggedInError{action: "grpc: AddCreditCard"}
	}

	cc := &pb.CreditCard{Number: card.Number}
	ctx, cancel := newTimeoutContext()
	defer cancel()
	_, err := gc.client.AddCreditCard(ctx, cc)
	if err == context.DeadlineExceeded {
		return err
	} else if err != nil {
		return ErrAddCreditCardFailed
	}

	return nil
}

func (gc *GRPCClient) HasCreditCard(card *credit.Card) (bool, error) {
	if !gc.isLoggedIn() {
		return false, &NotLoggedInError{action: "grpc: HasCreditCard"}
	}

	ctx, cancel := newTimeoutContext()
	defer cancel()
	cc, err := gc.client.GetCreditCards(ctx, &empty.Empty{})
	if err == context.DeadlineExceeded {
		return false, err
	} else if err != nil {
		return false, nil
	}
	for _, c := range cc.Cards {
		cur := &credit.Card{
			Number: c.Number,
		}
		if isSameCreditCard(cur, card) {
			return true, nil
		}
	}

	return false, nil
}

func (gc *GRPCClient) CheckPublicKey() (bool, error) {
	if !gc.isLoggedIn() {
		return false, &NotLoggedInError{action: "grpc: CheckPublicKey"}
	}
	ctx, cancel := newTimeoutContext()
	defer cancel()
	k, err := gc.client.GetPublicKey(ctx, &empty.Empty{})
	if err == context.DeadlineExceeded {
		return false, err
	} else if err != nil {
		return false, nil
	}
	pk, err := jwt.ParseRSAPublicKeyFromPEM([]byte(k.Key))
	if err != nil {
		return false, nil
	}

	_, err = jwt.Parse(gc.authToken, func(token *jwt.Token) (interface{}, error) {
		return pk, nil
	})
	if err != nil {
		return false, nil
	}

	return true, nil
}

func (gc *GRPCClient) isLoggedIn() bool {
	return gc.authToken != ""
}

func newTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), checkerlib.Timeout)
}
