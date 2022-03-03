package tgorm

import (
	"gorm.io/gorm"
	"trellis.tech/trellis/common.v1/errcode"
	"trellis.tech/trellis/common.v1/transaction"
)

type trans struct {
	isTrans bool
	engine  *gorm.DB
	session *gorm.DB
}

func (p *trans) Session() interface{} {
	if p.isTrans {
		return p.session
	}
	return p.engine
}

func (p *trans) IsTransaction() bool {
	return p.isTrans
}

func (p *trans) Commit(fun interface{}, repos ...interface{}) error {
	fn := transaction.GetLogicFunc(fun)
	if fn == nil || fn.Logic == nil {
		//return todo error
		return nil
	}

	var (
		_values   []interface{}
		_newRepos []interface{}
		err       error
	)

	if p.IsTransaction() {
		p.session = p.engine.Begin()
		if p.session.Error != nil {
			return p.session.Error
		}
		defer func() {
			if err != nil {
				_ = p.session.Rollback()
			}
		}()

		for _, repo := range repos {
			tRepo, ok := repo.(transaction.Repo)
			if !ok {
				return errcode.New("not transaction repo, check the repo implement transaction repo")
			}
			if err = tRepo.SetSession(p.session); err != nil {
				return err
			}
			_newRepos = append(_newRepos, tRepo)
		}
	} else {
		for _, repo := range repos {
			tRepo, ok := repo.(transaction.Repo)
			if !ok {
				return errcode.New("not transaction repo, check the repo implement transaction repo")
			}
			if err = tRepo.SetSession(p.engine); err != nil {
				return err
			}
			_newRepos = append(_newRepos, tRepo)
		}
	}

	defer func() {
		if err != nil {
			transaction.CallFunc(fn.OnError, err)
		}
	}()

	if _, err = transaction.CallFunc(fn.BeforeLogic, _newRepos...); err != nil {
		return err
	}

	if _values, err = transaction.CallFunc(fn.Logic, _newRepos...); err != nil {
		return err
	}

	if _, err = transaction.CallFunc(fn.AfterLogic, _newRepos...); err != nil {
		return err
	}

	if p.isTrans {
		_ = p.session.Commit()
	}

	if _, err = transaction.CallFunc(fn.AfterCommit, _values); err != nil {
		return err
	}

	return nil
}
