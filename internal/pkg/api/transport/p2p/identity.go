package p2p

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/mitchellh/go-homedir"

	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
)

type Identity struct {
	nodeId     nodeSDK.NodeId
	privateKey crypto.PrivKey
}

func NewIdentity(filePath string) (*Identity, error) {
	filePath, err := homedir.Expand(filePath)
	if err != nil {
		return nil, fmt.Errorf("expand home directory: %w", err)
	}

	err = os.MkdirAll(path.Dir(filePath), 0700)
	if err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	privateKeyBuffer, err := os.ReadFile(filePath)
	if errors.Is(err, os.ErrNotExist) {
		privateKey, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 2048)
		if err != nil {
			return nil, fmt.Errorf("generate key pair: %w", err)
		}
		privateKeyBuffer, err = crypto.MarshalPrivateKey(privateKey)
		if err != nil {
			return nil, fmt.Errorf("marshal private key: %w", err)
		}

		err = os.WriteFile(filePath, privateKeyBuffer, 0600)
		if err != nil {
			return nil, fmt.Errorf("write file: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	privateKey, err := crypto.UnmarshalPrivateKey(privateKeyBuffer)
	if err != nil {
		return nil, fmt.Errorf("unmarshal private key: %w", err)
	}

	nodeId, err := peer.IDFromPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("id from pk: %w", err)
	}

	return &Identity{
		nodeId:     nodeSDK.NodeId(nodeId.String()),
		privateKey: privateKey,
	}, nil
}

func (identity *Identity) GetId() nodeSDK.NodeId {
	return identity.nodeId
}

func (identity *Identity) GetPrivateKey() crypto.PrivKey {
	return identity.privateKey
}
