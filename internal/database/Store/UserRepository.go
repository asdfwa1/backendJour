package Store

import (
	"database/sql"
	"errors"
	"fmt"
	"path"
	"sbitnev_back/internal/database/models"
)

var (
	internalServerErr = errors.New("internal server error")
	invalidUser       = errors.New("user not registered")
)

type UserRepository struct {
	store *Storage
}

func (u *UserRepository) GetUserByLogin(login string) (*models.User, error) {
	const op = "fc.userRep.GetUserByLogin"

	stmt, err := u.store.DB.Prepare("SELECT * FROM users WHERE login = $1")
	if err != nil {
		return nil, internalServerErr
	}
	defer stmt.Close()

	user := &models.User{}
	err = stmt.QueryRow(login).Scan(
		&user.UserID,
		&user.Login,
		&user.Password,
		&user.FullName,
		&user.Role,
	)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, invalidUser
	case err != nil:
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	return user, nil
}

func (u *UserRepository) GetUserByID(id int) (*models.User, error) {
	return nil, nil
}

func (u *UserRepository) GetUserByName(name string) (*models.User, error) {
	const op = "fc.userRep.GetUserByName"
	stmt, err := u.store.DB.Prepare("SELECT user_id, login, password, role FROM users WHERE full_name = $1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var user = models.User{
		FullName: name,
	}
	err = stmt.QueryRow(name).Scan(
		&user.UserID,
		&user.Login,
		&user.Password,
		&user.FullName,
		&user.Role,
	)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, invalidUser
	case err != nil:
		return nil, internalServerErr
	default:
		return &user, nil
	}
}

func (u *UserRepository) CreateUser(user *models.User) (int64, error) {
	const op = "fc.userRep.CreateUser"
	stmt, err := u.store.DB.Prepare("INSERT INTO users VALUES ($1, $2, $3, $4)")
	if err != nil {
		return 0, err
	}

	res, err := stmt.Exec(user.Role, user.Login, user.Password, user.Role)

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

/*func (u *UserRepository) DeleteUser(user *models.User) error {
	return nil
}

func (u *UserRepository) UpdateUser(user *models.User) error {
	return nil
}*/

func (u *UserRepository) CreateUserLink(userID int64, groupName string) error {
	const op = "fc.userRep.CreateUserLink"

	selectStmt, err := u.store.DB.Prepare("SELECT group_id FROM groups WHERE group_name = $1")
	if err != nil {
		return err
	}
	defer selectStmt.Close()

	insertStmt, err := u.store.DB.Prepare("INSERT INTO group_users VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	var groupID int
	if err := selectStmt.QueryRow(groupName).Scan(&groupID); err != nil {
		return err
	}

	_, err = insertStmt.Exec(groupID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserRepository) GetAllTeachers() ([]models.User, error) {
	const op = "fc.userRep.GetAllTeachers"

	stmt, err := u.store.DB.Prepare("SELECT * FROM users WHERE role = 'teacher'")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []models.User
	for rows.Next() {
		var teacher models.User
		err := rows.Scan(
			&teacher.UserID,
			&teacher.Login,
			&teacher.Password,
			&teacher.FullName,
			&teacher.Role,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, teacher)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	path.Join()

	return res, nil
}