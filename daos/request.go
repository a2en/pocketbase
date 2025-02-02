package daos

import (
	"fmt"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/types"
)

// RequestQuery returns a new Request logs select query.
func (dao *Dao) RequestQuery() *dbx.SelectQuery {
	return dao.ModelQuery(&models.Request{})
}

// FindRequestById finds a single Request log by its id.
func (dao *Dao) FindRequestById(id string) (*models.Request, error) {
	model := &models.Request{}

	err := dao.RequestQuery().
		AndWhere(dbx.HashExp{"id": id}).
		Limit(1).
		One(model)

	if err != nil {
		return nil, err
	}

	return model, nil
}

type RequestsStatsItem struct {
	Total int            `db:"total" json:"total"`
	Date  types.DateTime `db:"dated" json:"date"`
}

// RequestsStats returns hourly grouped requests logs statistics.
func (dao *Dao) RequestsStats(expr dbx.Expression) ([]*RequestsStatsItem, error) {
	result := []*RequestsStatsItem{}

	query := dao.RequestQuery().
		Select("count(id) as total", "_requests.created as dated").
		GroupBy("created")

	if expr != nil {
		query.AndWhere(expr)
	}

	err := query.All(&result)

	fmt.Println("req stat err ", query.Build().SQL())

	return result, err
}

// DeleteOldRequests delete all requests that are created before createdBefore.
func (dao *Dao) DeleteOldRequests(createdBefore time.Time) error {
	m := models.Request{}
	tableName := m.TableName()

	formattedDate := createdBefore.UTC().Format(types.DefaultDateLayout)
	expr := dbx.NewExp("[[created]] <= {:dated}", dbx.Params{"dated": formattedDate})

	_, err := dao.DB().Delete(tableName, expr).Execute()

	return err
}

// SaveRequest upserts the provided Request model.
func (dao *Dao) SaveRequest(request *models.Request) error {
	return dao.Save(request)
}
