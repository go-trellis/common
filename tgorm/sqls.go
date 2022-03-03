package tgorm

import (
	"fmt"

	"gorm.io/gorm"
)

// TransactionDo to do transaction with customer function
func TransactionDo(engine *gorm.DB, fn func(db *gorm.DB) error) error {
	return TransactionDoWithSession(engine, fn)
}

// TransactionDoWithSession to do transaction with customer function
func TransactionDoWithSession(s *gorm.DB, fn func(db *gorm.DB) error) (err error) {
	db := s.Begin()
	if err = db.Error; err != nil {
		return
	}
	defer func() {
		if err != nil {
			ie := db.Rollback().Error
			if ie != nil {
				fmt.Println("rollback failed", ie)
			}
			return
		}
		err = db.Commit().Error
	}()
	err = fn(db)
	return
}

/// Get Execute

type GetOption func(*GetOptions)
type GetOptions struct {
	Wheres interface{}
	Args   []interface{}

	Limit, Offset int
	OrderBy       string
}

func GetWheres(wheres interface{}) GetOption {
	return func(options *GetOptions) {
		options.Wheres = wheres
	}
}

func GetArgs(args ...interface{}) GetOption {
	return func(options *GetOptions) {
		options.Args = args
	}
}

func GetLimit(limit, offset int) GetOption {
	return func(options *GetOptions) {
		options.Limit = limit
		options.Offset = offset
	}
}

func GetOrderBy(order string) GetOption {
	return func(options *GetOptions) {
		options.OrderBy = order
	}
}

func Get(session *gorm.DB, bean interface{}, opts ...GetOption) (bool, error) {
	getOptions := &GetOptions{}
	for _, opt := range opts {
		opt(getOptions)
	}

	session = session.Where(getOptions.Wheres, getOptions.Args...)
	if getOptions.Limit > 0 {
		session = session.Limit(getOptions.Limit)
		session = session.Offset(getOptions.Offset)
	}
	if len(getOptions.OrderBy) > 0 {
		session = session.Order(getOptions.OrderBy)
	}
	if err := session.First(bean).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

/// Update Execute

type UpdateOption func(*UpdateOptions)
type UpdateOptions struct {
	Wheres interface{}
	Args   []interface{}
	Cols   []string
}

func UpdateWheres(wheres interface{}) UpdateOption {
	return func(options *UpdateOptions) {
		options.Wheres = wheres
	}
}

func UpdateArgs(args ...interface{}) UpdateOption {
	return func(options *UpdateOptions) {
		options.Args = args
	}
}

func UpdateCols(cols ...string) UpdateOption {
	return func(options *UpdateOptions) {
		options.Cols = cols
	}
}

func Update(session *gorm.DB, bean interface{}, opts ...UpdateOption) (int64, error) {
	updateOptions := &UpdateOptions{}
	for _, opt := range opts {
		opt(updateOptions)
	}

	session = session.Where(updateOptions.Wheres, updateOptions.Args...)
	if len(updateOptions.Cols) > 0 {
		session = session.Omit(updateOptions.Cols...)
	}

	db := session.Model(bean).Updates(bean)
	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

/// CreateMulti Execute

type CreateMultiOption func(*CreateMultiOptions)
type CreateMultiOptions struct {
	StepNumber int
}

func InsertMultiStepNumber(number int) CreateMultiOption {
	return func(options *CreateMultiOptions) {
		options.StepNumber = number
	}
}

// CreateMulti create multi seperated slice data in a big slice with every step number
// default to insert the slice with no seperated.
func CreateMulti(session *gorm.DB, ones interface{}, opts ...CreateMultiOption) (int64, error) {
	options := &CreateMultiOptions{}
	for _, opt := range opts {
		opt(options)
	}
	if options.StepNumber <= 0 {
		db := session.Save(ones)
		if db.Error != nil {
			return 0, db.Error
		}
		return db.RowsAffected, nil
	}

	db := session.CreateInBatches(ones, options.StepNumber)
	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

/// Delete Execute

type DeleteOption func(*DeleteOptions)
type DeleteOptions struct {
	Wheres interface{}
	Args   []interface{}
}

func DeleteWheres(wheres interface{}) DeleteOption {
	return func(options *DeleteOptions) {
		options.Wheres = wheres
	}
}

func DeleteArgs(args ...interface{}) DeleteOption {
	return func(options *DeleteOptions) {
		options.Args = args
	}
}

func Delete(session *gorm.DB, bean interface{}, opts ...DeleteOption) (int64, error) {
	deleteOptions := &DeleteOptions{}
	for _, opt := range opts {
		opt(deleteOptions)
	}

	if deleteOptions.Wheres != nil {
		session = session.Where(deleteOptions.Wheres, deleteOptions.Args...)
	}

	db := session.Delete(bean)
	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
