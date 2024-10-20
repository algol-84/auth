package auth

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/algol-84/auth/internal/repository"
	"github.com/algol-84/auth/internal/repository/auth/converter"
	"github.com/algol-84/auth/internal/repository/auth/model"
	desc "github.com/algol-84/auth/pkg/user_v1"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Представление БД сервиса Auth
const table = "chat_user"
const (
	fieldID        = "id"
	fieldName      = "name"
	fieldEmail     = "email"
	fieldRole      = "role"
	fieldPassword  = "password"
	fieldCreatedAt = "created_at"
	fieldUpdatedAt = "updated_at"
)

type repo struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) repository.AuthRepository {
	return &repo{db: db}
}

// Create создает нового юзера в БД
// Данные юзера передаются указателем на структуру protobuf
// Функция возвращает присвоенный в БД ID юзера или ошибку записи
func (r *repo) Create(ctx context.Context, user *desc.User) (int64, error) {
	var userID int64
	// Собрать запрос на вставку записи в таблицу
	builderQuery := sq.Insert(table).
		PlaceholderFormat(sq.Dollar).
		Columns(fieldName, fieldPassword, fieldEmail, fieldRole).
		Values(user.Name, user.Password, user.Email, user.Role.String()).
		Suffix("RETURNING " + fieldID)

	query, args, err := builderQuery.ToSql()
	if err != nil {
		return 0, err
	}

	err = r.db.QueryRow(ctx, query, args...).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// Get возвращает информацию о юзере по ID
func (r *repo) Get(ctx context.Context, id int64) (*desc.UserInfo, error) {
	builderQuery := sq.Select(fieldID, fieldName, fieldEmail, fieldRole, fieldCreatedAt, fieldUpdatedAt).
		From(table).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{fieldID: id})

	query, args, err := builderQuery.ToSql()
	if err != nil {
		return nil, err
	}

	var user model.User
	err = r.db.QueryRow(ctx, query, args...).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return converter.ToUserFromRepo(&user), nil
}

// Update обновляет данные юзера в БД
func (r *repo) Update(ctx context.Context, user *desc.UserUpdate) error {
	builderQuery := sq.Update(table).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{fieldID: user.Id})

	if user.Name != nil {
		builderQuery = builderQuery.Set(fieldName, user.Name.Value)
	}
	if user.Email != nil {
		builderQuery = builderQuery.Set(fieldEmail, user.Email.Value)
	}
	if user.Role != desc.Role_UNKNOWN {
		builderQuery = builderQuery.Set(fieldRole, user.Role.String())
	}
	builderQuery = builderQuery.Set(fieldUpdatedAt, time.Now())

	query, args, err := builderQuery.ToSql()
	if err != nil {
		return err
	}

	res, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user with ID=%d not found", user.Id)
	}

	return nil
}

// Delete удаляет юзера из БД
func (r *repo) Delete(ctx context.Context, id int64) error {
	builderQuery := sq.Delete(table).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{fieldID: id})

	query, args, err := builderQuery.ToSql()
	if err != nil {
		return err
	}

	res, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user with ID=%d not found", id)
	}

	return nil
}