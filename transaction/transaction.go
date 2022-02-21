package transaction

//type NewTransactionFunc func() Transaction

type Transaction interface {
	Session() interface{}
	IsTransaction() bool

	Commit(fn interface{}, repos ...interface{}) error
}

// Derivative derive function
type Derivative interface {
	Derive() (repo interface{}, err error)
}

// Derive derive from developer function
func Derive(origin interface{}) (interface{}, error) {
	if d, ok := origin.(Derivative); ok {
		return d.Derive()
	}
	return nil, nil
}

// Inheritor inherit function
type Inheritor interface {
	Inherit(repo interface{}) error
}

// Inherit new repository from origin repository
func Inherit(new, origin interface{}) error {
	if inheritor, ok := new.(Inheritor); ok {
		return inheritor.Inherit(origin)
	}
	return nil
}
