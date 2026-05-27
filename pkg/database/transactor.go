package database

import "gorm.io/gorm"

// GORMTransactor implements port/repository.Transactor[R] using GORM transactions.
//
// R is caller-defined — it carries exactly the repositories a specific use case
// needs. The factory function receives the transactional *gorm.DB and constructs R,
// so every repository inside the transaction shares the same connection.
//
// Usage:
//
//	type CreateClientRepos struct {
//	    Client repository.ClientRepository
//	    Outbox repository.OutboxRepository
//	}
//
//	transactor := database.NewGORMTransactor(db, func(tx *gorm.DB) CreateClientRepos {
//	    return CreateClientRepos{
//	        Client: persistence.NewClientRepository(tx),
//	        Outbox: persistence.NewOutboxRepository(tx),
//	    }
//	})
type GORMTransactor[R any] struct {
	db      *gorm.DB
	factory func(tx *gorm.DB) R
}

// NewGORMTransactor returns a GORMTransactor wired with the given factory.
func NewGORMTransactor[R any](db *gorm.DB, factory func(*gorm.DB) R) *GORMTransactor[R] {
	return &GORMTransactor[R]{db: db, factory: factory}
}

// Execute runs fn inside a single database transaction.
// If fn returns an error the transaction is rolled back; otherwise it is committed.
func (t *GORMTransactor[R]) Execute(fn func(repos R) error) error {
	return t.db.Transaction(func(tx *gorm.DB) error {
		return fn(t.factory(tx))
	})
}
