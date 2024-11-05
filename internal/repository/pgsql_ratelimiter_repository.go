package repository

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

const (
	sqlCreate_table = `
			CREATE TABLE IF NOT EXISTS tb_limiters (
					chkey VARCHAR(100) PRIMARY KEY,
					chvalue VARCHAR(100) NOT NULL
				);`

	sqlInsert = `
				INSERT INTO tb_limiters (chkey, chvalue)
				VALUES ($1, $2)`

	sqlUpdate = `
				UPDATE tb_limiters
				SET chvalue = $1
				WHERE chkey = $2`

	sqlSelect = `
				SELECT chvalue 
				FROM tb_limiters
				WHERE chkey = $1`

	sqlDelete = `
				DELETE 
				FROM tb_limiters
				WHERE chkey = $1`
)

type PgsqlRepository struct {
	database *sql.DB
}

func NewPgSqlRepository(addr string) *PgsqlRepository {
	db, err := sql.Open("postgres", addr)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(sqlCreate_table)
	if err != nil {
		panic(err)
	}

	return &PgsqlRepository{
		database: db,
	}

}

func (repo *PgsqlRepository) Save(ctx context.Context, key, value string) {
	_, err := repo.database.ExecContext(ctx, sqlInsert, key, value)
	if err != nil {
		panic(err)
	}
}

func (repo *PgsqlRepository) Update(ctx context.Context, key, value string) {
	_, err := repo.database.ExecContext(ctx, sqlUpdate, value, key)
	if err != nil {
		panic(err)
	}

}

func (repo *PgsqlRepository) Find(ctx context.Context, key string) (string, error) {
	var value string
	err := repo.database.QueryRowContext(ctx, sqlSelect, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		panic(err)
	}
	return value, nil
}

func (repo *PgsqlRepository) Delete(ctx context.Context, key string) bool {
	_, err := repo.database.ExecContext(ctx, sqlDelete, key)
	if err != nil {
		panic(err)
	}
	return true
}
