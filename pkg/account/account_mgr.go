package account

import (
	"net/http"

	"github.com/zhulingbiezhi/go12306/config"
	"github.com/zhulingbiezhi/go12306/tools/errors"
)

type AccountMgr struct {
	accounts map[string]*AccountHelper
}

func (mgr *AccountMgr) GetAccount(uuid string) (*AccountHelper, error) {
	acc, ok := mgr.accounts[uuid]
	if !ok {
		return nil, errors.Errorf(nil, "account uuid: %s is empty", uuid)
	}
	return acc, nil
}

func (mgr *AccountMgr) Init(cfgs []*config.Account) error {
	mgr.accounts = make(map[string]*AccountHelper)
	for _, cfg := range cfgs {
		acc := &Account{
			Name:            cfg.AccountName,
			AccountName:     cfg.AccountName,
			AccountPassword: cfg.AccountPassword,
			cookieMap:       make(map[string]*http.Cookie),
		}
		_, ok := mgr.accounts[cfg.UUID]
		if ok {
			return errors.Errorf(nil, "account uuid: %s is exist", cfg.UUID)
		}
		helper := &AccountHelper{Account: acc}
		helper.Init()
		mgr.accounts[cfg.UUID] = helper
	}
	return nil
}
