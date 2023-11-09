package propertydb

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nhaancs/bhms/business/core/property"
	db "github.com/nhaancs/bhms/business/data/dbsql/pgx"
	"github.com/nhaancs/bhms/business/data/transaction"
	"github.com/nhaancs/bhms/foundation/logger"
)

type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

func (s *Store) Create(ctx context.Context, core property.Property) error {
	const q = `
	INSERT INTO properties
		(id, manager_id, name, address_level_1_id, address_level_2_id, address_level_3_id, street, status, created_at, updated_at)
	VALUES
		(:id, :manager_id, :name, :address_level_1_id, :address_level_2_id, :address_level_3_id, :street, :status, :created_at, :updated_at)`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, toDBProperty(core)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Update(ctx context.Context, core property.Property) error {
	const q = `
	UPDATE
		properties
	SET 
		"name" = :name,
		"address_level_1_id" = :address_level_1_id,
		"address_level_2_id" = :address_level_2_id,
		"address_level_3_id" = :address_level_3_id,
		"street" = :street,
		"status" = :status,
		"updated_at" = :updated_at
	WHERE
		id = :id`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, toDBProperty(core)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, core property.Property) error {
	data := struct {
		ID string `db:"id"`
	}{
		ID: core.ID.String(),
	}

	const q = `
	DELETE FROM
		properties
	WHERE
		id = :id`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (property.Property, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `
	SELECT
        id, manager_id, address_level_1_id, address_level_2_id, address_level_3_id, street, status, created_at, updated_at
	FROM
		properties
	WHERE 
		id = :id`

	var row dbProperty
	if err := db.NamedQueryStruct(ctx, s.log, s.db, q, data, &row); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return property.Property{}, fmt.Errorf("namedquerystruct: %w", property.ErrNotFound)
		}
		return property.Property{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	prprty, err := toCoreProperty(row)
	if err != nil {
		return property.Property{}, err
	}

	return prprty, nil
}

func (s *Store) QueryByManagerID(ctx context.Context, id uuid.UUID) ([]property.Property, error) {
	data := struct {
		ManagerID string `db:"manager_id"`
	}{
		ManagerID: id.String(),
	}

	const q = `
	SELECT
        id, manager_id, address_level_1_id, address_level_2_id, address_level_3_id, street, status, created_at, updated_at
	FROM
		properties
	WHERE
		manager_id = :manager_id`

	var rows []dbProperty
	if err := db.NamedQuerySlice(ctx, s.log, s.db, q, data, &rows); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return nil, fmt.Errorf("namedqueryslice: %w", property.ErrNotFound)
		}
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	prprties, err := toCoreProperties(rows)
	if err != nil {
		return nil, err
	}

	return prprties, nil
}

// ExecuteUnderTransaction constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) ExecuteUnderTransaction(tx transaction.Transaction) (property.Storer, error) {
	ec, err := db.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	s = &Store{
		log: s.log,
		db:  ec,
	}

	return s, nil
}
