package gapi

import (
	"fmt"

	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/mhdna/kashi/pb"
	"github.com/mhdna/kashi/token"
	"github.com/mhdna/kashi/util"
)

type Server struct {
	pb.UnimplementedKashiServer
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %s", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	return server, nil
}
