package repository

// Transactor executes fn within a single database transaction.
// R defines exactly which repositories the caller needs — no assumptions are made
// about which repos are included. If fn returns an error, the transaction is rolled
// back; otherwise it is committed.
//
// Each use case declares its own R to receive only what it requires:
//
//	type CreateClientRepos struct {
//	    Client ClientRepository
//	    Outbox OutboxRepository
//	}
//	var _ Transactor[CreateClientRepos] = (*GORMTransactor[CreateClientRepos])(nil)
type Transactor[R any] interface {
	Execute(fn func(repos R) error) error
}
