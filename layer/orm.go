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
