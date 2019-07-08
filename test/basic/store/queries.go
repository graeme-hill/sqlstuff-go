// This code was generated by a tool (>^_^)>

package store

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type DBClient interface {
	GetUsers() ([]GetUsersResult, error)
	Close()
}

type SQLDBClient struct {
	db *sql.DB
}

func NewDBClient(connectionString string) (DBClient, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return SQLDBClient{}, err
	}

	return SQLDBClient{
		db: db,
	}, nil
}

func (client SQLDBClient) Close() {
	client.db.Close()
}

/******************************************************************************
 * get_users
 ****************************************************************************/

type GetUsersResult struct {
	Id        int32
	Email     string
	FirstName string
	LastName  string
	GroupName string
}

func (client SQLDBClient) GetUsers() (r1 []GetUsersResult, err error) {
	sql := "SELECT\n  u.id, u.email, u.first_name, u.last_name, g.name AS group_name\nFROM\n  users u\nLEFT JOIN user_groups ug ON u.id = ug.user_id\nLEFT JOIN groups g ON g.id = ug.group_id;"
	rows, err := client.db.Query(sql)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id        int32
			email     string
			firstName string
			lastName  string
			groupName string
		)
		err = rows.Scan(&id, &email, &firstName, &lastName, &groupName)
		if err != nil {
			return
		}

		r1 = append(r1, GetUsersResult{
			Id:        id,
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
			GroupName: groupName,
		})
	}

	return
}