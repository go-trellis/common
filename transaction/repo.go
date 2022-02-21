package transaction

type Repo interface {
	SetSession(interface{}) error
}
