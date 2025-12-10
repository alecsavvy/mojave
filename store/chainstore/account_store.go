package chainstore

import (
	accountv1 "github.com/alecsavvy/mojave/gen/account/v1"
	"google.golang.org/protobuf/proto"
)

func (c *ChainStore) StoreAccount(account *accountv1.Account) error {
	if err := c.RequireBatch(); err != nil {
		return err
	}

	writer := c.writer

	address := account.Address
	accountBytes, err := proto.Marshal(account)
	if err != nil {
		return err
	}

	key := accountKey(address)
	return writer.Set(key, accountBytes, nil)
}

func (c *ChainStore) GetAccount(address string) (*accountv1.Account, error) {
	reader := c.reader

	key := accountKey(address)
	accountBytes, closer, err := reader.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	account := &accountv1.Account{}
	err = proto.Unmarshal(accountBytes, account)
	if err != nil {
		return nil, err
	}

	return account, nil
}
