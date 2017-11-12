package layer

import (
	log "github.com/Sirupsen/logrus"
	"github.com/go-xorm/builder"
)

func (el *Layer) Select(database, table string) ([]map[string]interface{}, error) {
	sql, _, _ := builder.Select("*").
		From(database + "." + table).
		ToSQL()
	data, err := el.Execute(sql)
	return data, err
}

func (el *Layer) Insert(database, table string, row builder.Eq) error {
	sql, args, _ := builder.
		Insert(row).
		Into(database + "." + table).
		ToSQL()
	log.Info(sql, args, row)
	_, err := el.ExecuteStmt(sql, args...)
	return err
}

func (el *Layer) Update(database, table string, id string, row builder.Eq) error {
	sql, args, _ := builder.
		Update(row).
		Where(builder.Eq{"id": id}).
		From(database + "." + table).
		ToSQL()
	log.Info(sql, args, row)
	_, err := el.ExecuteStmt(sql, args...)
	return err
}

func (el *Layer) Delete(database, table string, id string) error {
	sql, args, _ := builder.
		Delete(builder.Eq{"id": id}).
		From(database + "." + table).
		ToSQL()
	log.Info(sql, args, id)
	_, err := el.ExecuteStmt(sql, args...)
	return err
}
