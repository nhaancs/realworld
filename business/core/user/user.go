// Package user provides an example of a core business API. Right now these
// calls are just wrapping the data/data layer. But at some point you will
// want auditing or something that isn't specific to the data/store layer.
package user

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/nhaancs/bhms/foundation/logger"
	"golang.org/x/crypto/bcrypt"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("user not found")
	ErrUniqueEmail           = errors.New("email is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// =============================================================================

// Storer interface declares the behavior this package needs to perists and retrieve data.
type Storer interface {
	Create(ctx context.Context, usr UserEntity) error
	Update(ctx context.Context, usr UserEntity) error
	Delete(ctx context.Context, usr UserEntity) error
	QueryByID(ctx context.Context, userID uuid.UUID) (UserEntity, error)
	QueryByIDs(ctx context.Context, userID []uuid.UUID) ([]UserEntity, error)
	QueryByEmail(ctx context.Context, email mail.Address) (UserEntity, error)
}

// =============================================================================

// Core manages the set of APIs for user access.
type Core struct {
	store Storer
	log   *logger.Logger
}

// NewCore constructs a core for user api access.
func NewCore(log *logger.Logger, store Storer) *Core {
	return &Core{
		store: store,
		log:   log,
	}
}

// Register adds a new user to the system.
func (c *Core) Register(ctx context.Context, nu RegisterEntity) (UserEntity, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return UserEntity{}, fmt.Errorf("generatefrompassword: %w", err)
	}

	now := time.Now()

	usr := UserEntity{
		ID:           uuid.New(),
		Name:         nu.Name,
		Email:        nu.Email,
		PasswordHash: hash,
		Roles:        nu.Roles,
		Department:   nu.Department,
		Enabled:      true,
		DateCreated:  now,
		DateUpdated:  now,
	}

	if err := c.store.Create(ctx, usr); err != nil {
		return UserEntity{}, fmt.Errorf("create: %w", err)
	}

	return usr, nil
}

// QueryByID finds the user by the specified ID.
func (c *Core) QueryByID(ctx context.Context, userID uuid.UUID) (UserEntity, error) {
	user, err := c.store.QueryByID(ctx, userID)
	if err != nil {
		return UserEntity{}, fmt.Errorf("query: userID[%s]: %w", userID, err)
	}

	return user, nil
}

// QueryByEmail finds the user by a specified user email.
func (c *Core) QueryByEmail(ctx context.Context, email mail.Address) (UserEntity, error) {
	user, err := c.store.QueryByEmail(ctx, email)
	if err != nil {
		return UserEntity{}, fmt.Errorf("query: email[%s]: %w", email, err)
	}

	return user, nil
}

// =============================================================================

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Claims UserEntity representing this user. The claims can be
// used to generate a token for future authentication.
func (c *Core) Authenticate(ctx context.Context, email mail.Address, password string) (UserEntity, error) {
	usr, err := c.QueryByEmail(ctx, email)
	if err != nil {
		return UserEntity{}, fmt.Errorf("query: email[%s]: %w", email, err)
	}

	if err := bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(password)); err != nil {
		return UserEntity{}, fmt.Errorf("comparehashandpassword: %w", ErrAuthenticationFailure)
	}

	return usr, nil
}
