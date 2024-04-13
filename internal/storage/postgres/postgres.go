package postgres

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/Memozir/BannerService/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

var (
	QueryExecuteError = "database error"
	TranBeginError    = "transaction start error"
	TranRollbackError = "transaction rollback error"
	TranCommitError   = "transaction commit error"
)

type DbError struct {
	error
	errStr string
}

func (dbErr *DbError) GetDbError() string { return dbErr.errStr }

func NewDbErr(errStr string) *DbError {
	return &DbError{errStr: errStr}
}

type Storage struct {
	conn *pgxpool.Pool
	log  *slog.Logger
}

func NewDb(ctx context.Context, logger *slog.Logger, cfg *config.Config) (*Storage, error) {
	const op = "storage.postgres.NewDb"

	postgresConfig, err := getPostgresConnConfig(cfg)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	connPool, err := pgxpool.NewWithConfig(ctx, postgresConfig)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	storage := &Storage{
		conn: connPool,
		log:  logger,
	}

	pingContext, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	err = storage.conn.Ping(pingContext)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return storage, nil
}

func getPostgresConnConfig(cfg *config.Config) (*pgxpool.Config, error) {
	connString := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s pool_max_conns=%s",
		cfg.PostgresUser,
		cfg.PostgresPass,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDb,
		cfg.PostgresSSLMode,
		cfg.PostgresMaxConns,
	)

	postgresConfig, err := pgxpool.ParseConfig(connString)

	if err != nil {
		return nil, err
	}

	return postgresConfig, nil
}

func (storage *Storage) Shutdown() {
	storage.conn.Close()
}

func (storage *Storage) createFeature(ctx context.Context, featureId int64) (int64, error) {
	const op = "storage.postgres.createFeature"

	query := `
		insert into feature(feature_id)
		values($1)
		on conflict do nothing
		returning id
	`

	var featurePk int64
	err := storage.conn.QueryRow(ctx, query, featureId).Scan(&featurePk)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return featurePk, nil
}

func (storage *Storage) getFeature(ctx context.Context, featureId int64) (int64, error) {
	const op = "storage.postgres.getFeature"

	query := `
		select coalesce(
			(select id
			from feature
			where feature_id=$1),
			0
		)
	`

	var featurePk int64
	err := storage.conn.QueryRow(ctx, query, featureId).Scan(&featurePk)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return featurePk, nil
}

type featureTagModel struct {
	Tag int64 `db:"tag"`
}

func (storage *Storage) validFeatureTags(ctx context.Context, featureId int64, tags []int64) (bool, error) {
	const op = "storage.postgres.validFeatureTags"

	query := `
		select distinct t.tag_id tag
		from banner_tag bt
			inner join banner b on b.id = bt.banner_id
			inner join feature f on b.banner_feature = f.id
			inner join tag t on bt.tag_id = t.id
		where f.feature_id=$1
	`

	rows, err := storage.conn.Query(ctx, query, featureId)

	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	featureTags, err := pgx.CollectRows(rows, pgx.RowToStructByName[featureTagModel])

	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if len(featureTags) == 0 {
		return true, nil
	}

	tagMap := make(map[int64]struct{})

	for _, tagId := range tags {
		tagMap[tagId] = struct{}{}
	}

	for _, item := range featureTags {
		if _, ok := tagMap[item.Tag]; ok {
			return false, nil
		}
	}

	return true, nil
}

type tagIdsModel struct {
	TagId int64 `db:"id"`
}

func (storage *Storage) setupBannerTags(ctx context.Context, tran pgx.Tx, bannerId int64, tags []int64) error {
	const op = "storage.postgres.setupBannerTags"

	getTagsPsql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	q, tagVals, err := getTagsPsql.Select("id").From("tag").Where(squirrel.Eq{"tag_id": tags}).ToSql()

	rows, err := tran.Query(ctx, q, tagVals...)
	tagIds, err := pgx.CollectRows(rows, pgx.RowToStructByName[tagIdsModel])
	fmt.Println(tagIds)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	sql := psql.Insert("banner_tag").Columns("banner_id", "tag_id")

	for _, tag := range tagIds {
		sql = sql.Values(bannerId, tag.TagId)
	}

	query, values, err := sql.ToSql()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tran.Exec(ctx, query, values...)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (storage *Storage) CreateBanner(ctx context.Context, tagIds []int64, featureId int64, content string, isActive bool) (int64, error) {
	const op = "storage.postgres.CreateBanner"

	valid, err := storage.validFeatureTags(ctx, featureId, tagIds)

	if err != nil {
		storage.log.Error(err.Error())
		return 0, NewDbErr(QueryExecuteError)
	}

	if !valid {
		// TODO: возвращать ошибку, указывающую, что нельзя создать баннер с такими feature и tags
		return 0, fmt.Errorf("%s: %s", op, "incorrect data")
	}

	featurePk, err := storage.getFeature(ctx, featureId)

	if err != nil {
		storage.log.Error(err.Error())
		return 0, NewDbErr(QueryExecuteError)
	}

	tran, err := storage.conn.Begin(ctx)

	if err != nil {
		tranErr := tran.Rollback(ctx)
		if tranErr != nil {
			storage.log.Error(fmt.Sprintf("%s: %w", op, err))
			return 0, NewDbErr(TranRollbackError)
		}
		storage.log.Error(fmt.Sprintf("%s: %w", op, err))
		return 0, NewDbErr(TranBeginError)
	}

	query := `
		insert into banner(banner_feature, content, is_active)
		values($1, $2, $3)
		returning id
	`

	var bannerId int64
	err = tran.QueryRow(ctx, query, featurePk, content, isActive).Scan(&bannerId)

	if err != nil {
		storage.log.Error(fmt.Sprintf("%s: %w", op, err))
		return 0, NewDbErr(QueryExecuteError)
	}

	err = storage.setupBannerTags(ctx, tran, bannerId, tagIds)

	if err != nil {
		tranErr := tran.Rollback(ctx)
		if tranErr != nil {
			storage.log.Error(fmt.Sprintf("%s: %w", op, err))
			return 0, NewDbErr(TranRollbackError)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	err = tran.Commit(ctx)

	if err != nil {
		tranErr := tran.Rollback(ctx)

		if tranErr != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
		storage.log.Error(fmt.Sprintf("%s: %w", op))
		return 0, NewDbErr(TranCommitError)
	}

	return bannerId, nil
}
