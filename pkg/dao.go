package tq

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

var (
	DB           *sql.DB
	singletonDao *dao
	onceDao      sync.Once
)

const (
	sqlInsert = `INSERT INTO plugin_time_quality (
        point_id,
        last_updated,
        start_time,
        end_time,
        evals,
        mean_wait,
        max_wait,
        min_wait,
        cv_wait,
        sd_wait,
        fill_factor,
        score
    )
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	sqlUpdate = `UPDATE plugin_time_quality
        SET last_updated = ?,
            start_time = ?,
            end_time = ?,
            evals = ?,
            mean_wait = ?,
            max_wait = ?,
            min_wait = ?,
            cv_wait = ?,
            sd_wait = ?,
            fill_factor = ?,
            score = ?
        WHERE point_id = ?`
	sqlSelectByPointId = `SELECT id, point_id, last_updated, start_time, end_time, evals, mean_wait, max_wait, min_wait, cv_wait, sd_wait, fill_factor, score
        FROM plugin_time_quality
        WHERE point_id = ?
    `
)

type dao struct {
	insertStmt          *sql.Stmt
	updateStmt          *sql.Stmt
	selectByPointIdStmt *sql.Stmt
}

func getDao() *dao {
	onceDao.Do(func() {
		singletonDao = &dao{}
		var err error

		singletonDao.createTableIfNotExists()

		// Prepare the SQL statements
		singletonDao.insertStmt, err = DB.Prepare(sqlInsert)
		if err != nil {
			panic(fmt.Sprintf("failed to prepare insert statement: %v", err))
		}

		singletonDao.updateStmt, err = DB.Prepare(sqlUpdate)
		if err != nil {
			panic(fmt.Sprintf("failed to prepare update statement: %v", err))
		}

		singletonDao.selectByPointIdStmt, err = DB.Prepare(sqlSelectByPointId)
		if err != nil {
			panic(fmt.Sprintf("failed to prepare select statement: %v", err))
		}

	})
	return singletonDao
}
func (dao *dao) createTableIfNotExists() {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS plugin_time_quality (
			id INTEGER PRIMARY KEY,
			point_id INTEGER,
			last_updated TIMESTAMP,
			start_time TIMESTAMP,
			end_time TIMESTAMP, 
			evals INTEGER,
			mean_wait FLOAT,
			max_wait FLOAT,
			min_wait FLOAT,
			cv_wait FLOAT,
			sd_wait FLOAT,
			fill_factor INTEGER,
			score INTEGER,
			FOREIGN KEY (point_id) REFERENCES point(id),
			UNIQUE(point_id)
		);
	`)
	if err != nil {
		panic(fmt.Sprintf("failed to create table: %v", err))
	}
}

func (dao *dao) insert(tq *TimeQuality) (int64, error) {
	result, err := dao.insertStmt.Exec(
		tq.PointId,
		time.Now(),
		tq.Start,
		tq.End,
		tq.Count,
		tq.MeanWait,
		tq.MaxWait,
		tq.MinWait,
		tq.WaitCoefficientOfVariation,
		tq.WaitStandardDeviation,
		tq.FillFactor,
		tq.Score,
	)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (dao *dao) update(tq *TimeQuality) (int64, error) {

	result, err := dao.updateStmt.Exec(
		time.Now(),
		tq.Start,
		tq.End,
		tq.Count,
		tq.MeanWait,
		tq.MaxWait,
		tq.MinWait,
		tq.WaitCoefficientOfVariation,
		tq.WaitStandardDeviation,
		tq.FillFactor,
		tq.Score,
		tq.PointId)

	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func (dao *dao) selectByPointId(pointID uint32) (*TimeQuality, error) {

	row := dao.selectByPointIdStmt.QueryRow(pointID)
	var tq TimeQuality

	err := row.Scan(
		&tq.Id,
		&tq.PointId,
		&tq.LastUpdated,
		&tq.Start,
		&tq.End,
		&tq.Count,
		&tq.MeanWait,
		&tq.MaxWait,
		&tq.MinWait,
		&tq.WaitCoefficientOfVariation,
		&tq.WaitStandardDeviation,
		&tq.FillFactor,
		&tq.Score,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &tq, nil
}
