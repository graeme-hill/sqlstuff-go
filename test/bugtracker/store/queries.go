// This code was generated by a tool (>^_^)>

package store

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
	"time"
)

type DBClient interface {
	GetIssue() ([]GetIssueResult1, []GetIssueResult2, error)
	GetProjects() ([]GetProjectsResult, error)
	GetTags() ([]GetTagsResult, error)
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
 * get_issue
 ****************************************************************************/

type GetIssueResult1 struct {
	Id          string
	Name        string
	Fields      map[string]interface{}
	Created     time.Time
	Modified    time.Time
	ProjectName string
}

type GetIssueResult2 struct {
	TagKey  string
	Created time.Time
}

func (client SQLDBClient) GetIssue() (r1 []GetIssueResult1, r2 []GetIssueResult2, err error) {
	sql := "SELECT\n  i.id,\n  i.name,\n  i.fields,\n  i.created,\n  i.modified,\n  p.name AS project_name\nFROM issues i\nJOIN projects p ON p.tid = i.tid AND p.\"key\" = i.project_key\nWHERE i.tid = $tid AND i.id = $id\nLIMIT 1;\n\nSELECT\n  tag_key,\n  created\nFROM issue_tags\nWHERE tid = $tid AND issue_id = $id;\n"
	rows, err := client.db.Query(sql)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id          string
			name        string
			fields      map[string]interface{}
			created     time.Time
			modified    time.Time
			projectName string
		)
		err = rows.Scan(&id, &name, &fields, &created, &modified, &projectName)
		if err != nil {
			return
		}

		r1 = append(r1, GetIssueResult1{
			Id:          id,
			Name:        name,
			Fields:      fields,
			Created:     created,
			Modified:    modified,
			ProjectName: projectName,
		})
	}

	if !rows.NextResultSet() {
		err = fmt.Errorf("Expecting more result sets: %v", rows.Err())
		return
	}

	for rows.Next() {
		var (
			tagKey  string
			created time.Time
		)
		err = rows.Scan(&tagKey, &created)
		if err != nil {
			return
		}

		r2 = append(r2, GetIssueResult2{
			TagKey:  tagKey,
			Created: created,
		})
	}

	return
}

/******************************************************************************
 * get_projects
 ****************************************************************************/

type GetProjectsResult struct {
	Key      string
	Name     string
	Created  time.Time
	Modified time.Time
}

func (client SQLDBClient) GetProjects() (r1 []GetProjectsResult, err error) {
	sql := "SELECT \"key\", \"name\", created, modified FROM projects WHERE tid = $tid"
	rows, err := client.db.Query(sql)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			key      string
			name     string
			created  time.Time
			modified time.Time
		)
		err = rows.Scan(&key, &name, &created, &modified)
		if err != nil {
			return
		}

		r1 = append(r1, GetProjectsResult{
			Key:      key,
			Name:     name,
			Created:  created,
			Modified: modified,
		})
	}

	return
}

/******************************************************************************
 * get_tags
 ****************************************************************************/

type GetTagsResult struct {
	Key     string
	Created time.Time
}

func (client SQLDBClient) GetTags() (r1 []GetTagsResult, err error) {
	sql := "SELECT \"key\", created FROM tags WHERE tid = $tid"
	rows, err := client.db.Query(sql)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			key     string
			created time.Time
		)
		err = rows.Scan(&key, &created)
		if err != nil {
			return
		}

		r1 = append(r1, GetTagsResult{
			Key:     key,
			Created: created,
		})
	}

	return
}