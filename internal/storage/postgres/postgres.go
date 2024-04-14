package postgres

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/Memozir/BannerService/config"
	"github.com/Memozir/BannerService/internal/storage/models"
	"github.com/Memozir/BannerService/internal/storage/postgres/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"strconv"
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

type Banner struct {
	Content  string `db:"content"`
	IsActive bool   `db:"is_active"`
}

func (storage *Storage) GetBanner(ctx context.Context, tagId string, featureId string) (string, bool, error) {
	const op = "storage.postgres.GetBanner"

	query := `
		select b.content , b.is_active
		from banner b
				 inner join feature f on b.banner_feature = f.id
		where f.feature_id = $1 and $2 in (
			select tag_id
			from banner_tag
			where banner_id=b.id
		)
		order by created_at
		limit 1
	`

	rows, err := storage.conn.Query(ctx, query, featureId, tagId)

	if err != nil {
		storage.log.Error(fmt.Sprintf("%s: %w", op))
		return "", false, NewDbErr(QueryExecuteError)
	}

	banner, err := pgx.CollectRows(rows, pgx.RowToStructByName[Banner])

	if err != nil {
		storage.log.Error(fmt.Sprintf("%s: %w", op))
		return "", false, NewDbErr(QueryExecuteError)
	}

	return banner[0].Content, banner[0].IsActive, nil
}

func (storage *Storage) GetAllBanners(
	ctx context.Context, featureId string, tagId string, limit string, offset string,
) ([]models.BannerFull, error) {
	const op = "storage.postgres.GetAllBanners"
	var query string

	if len(featureId) == 0 && len(tagId) == 0 {
		query = queries.QBannersAllFilters
	} else if len(featureId) == 0 {
		query = queries.QBannersTagIdFilter
	} else {
		query = queries.QBannersFeatureIdFilter
	}

	rows, err := storage.conn.Query(ctx, query, featureId, tagId, limit, offset)

	if err != nil {
		storage.log.Error(fmt.Sprintf("%s: %w", op))
		return []models.BannerFull{}, NewDbErr(QueryExecuteError)
	}

	banners, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.BannerFull])

	if err != nil {
		storage.log.Error(fmt.Sprintf("%s: %w", op))
		return []models.BannerFull{}, NewDbErr(QueryExecuteError)
	}

	return banners, nil
}

func (storage *Storage) getBannerByFeatureTags(ctx context.Context, featureId int64, tagIds []int64) (int64, error) {
	const op = "storage.postgres.getBannerByFeatureTags"

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	sql := psql.Select("b.id").From("banner b").InnerJoin("feature f on b.banner_feature = f.id")
	sql = sql.InnerJoin("banner_tag bt on b.id = bt.banner_id")
	sql = sql.InnerJoin("tag t on bt.tag_id = t.id")
	sql = sql.Where("f.feature_id = ", featureId)

	for _, tag := range tagIds {
		sql = sql.Where("t.tag_id = ", tag)
	}

	query, args, err := sql.ToSql()

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var bannerId int64
	err = storage.conn.QueryRow(ctx, query, args...).Scan(&bannerId)

	if err != nil {
		storage.log.Error(fmt.Sprintf("%s: %w", op))
		return 0, NewDbErr(QueryExecuteError)
	}

	return bannerId, nil
}

func (storage *Storage) getBannerFeatureId(ctx context.Context, bannerId int64) (int64, error) {
	const op = "storage.postgres.getBannerFeatureId"

	query := `
		select banner_feature
		from banner
		where id = $1
	`

	var featureId int64
	err := storage.conn.QueryRow(ctx, query, bannerId).Scan(&featureId)
	if err != nil {
		storage.log.Error(err.Error(), slog.String("op", op))
		return 0, NewDbErr(err.Error())
	}

	return featureId, nil
}

func (storage *Storage) UpdateBanner(
	ctx context.Context, bannerId string, tagIds *[]int64, featureId *int64, content string, isActive bool) error {

	const op = "storage.postgres.GetAllBanners"

	bannerIdInt, err := strconv.Atoi(bannerId)

	if err != nil {
		storage.log.Error(err.Error(), slog.String("op", op))
		return err
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	sql := psql.Update("banner")

	if len(content) != 0 {
		sql = sql.Set("content", content)
	}

	tran, err := storage.conn.Begin(ctx)

	if err != nil {
		tranErr := tran.Rollback(ctx)
		if tranErr != nil {
			storage.log.Error(fmt.Sprintf("%s: %w", op, err))
			return NewDbErr(TranRollbackError)
		}
		storage.log.Error(fmt.Sprintf("%s: %w", op, err))
		return NewDbErr(TranBeginError)
	}

	if featureId != nil && tagIds != nil {
		valid, err := storage.validFeatureTags(ctx, *featureId, *tagIds)
		if err != nil {
			storage.log.Error(err.Error(), slog.String("op", op))
			return NewDbErr(err.Error())
		}

		if valid {
			err = storage.setupBannerTags(ctx, tran, int64(bannerIdInt), *tagIds)

			if err != nil {
				storage.log.Error(err.Error(), slog.String("op", op))
				return NewDbErr(err.Error())
			}
		}

		featurePk, err := storage.getFeature(ctx, *featureId)

		if err != nil {
			storage.log.Error(err.Error(), slog.String("op", op))
			return NewDbErr(err.Error())
		}

		sql = sql.Set("banner_feature", featurePk)
	} else if featureId != nil {
		featurePk, err := storage.getFeature(ctx, *featureId)

		if err != nil {
			storage.log.Error(err.Error(), slog.String("op", op))
			return NewDbErr(err.Error())
		}

		sql = sql.Set("banner_feature", featurePk)
	} else if tagIds != nil {
		updatedFeatureId, err := storage.getBannerFeatureId(ctx, int64(bannerIdInt))

		if err != nil {
			return err
		}

		valid, err := storage.validFeatureTags(ctx, updatedFeatureId, *tagIds)

		if err != nil {
			storage.log.Error(err.Error(), slog.String("op", op))
			return NewDbErr(err.Error())
		}

		if valid {
			err = storage.setupBannerTags(ctx, tran, int64(bannerIdInt), *tagIds)

			if err != nil {
				storage.log.Error(err.Error(), slog.String("op", op))
				return NewDbErr(err.Error())
			}
		}
	}

	err = tran.Commit(ctx)

	if err != nil {
		tranErr := tran.Rollback(ctx)

		if tranErr != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		storage.log.Error(fmt.Sprintf("%s: %w", op))
		return NewDbErr(TranCommitError)
	}

	sql = sql.Set("is_active", isActive)
	sql = sql.Where("id", bannerId)
	query, args, err := sql.ToSql()

	if err != nil {
		dbErrStr := fmt.Sprintf("%s: %w", op, err)
		dbErr := NewDbErr(dbErrStr)
		storage.log.Error(err.Error(), slog.String("op", op))
		return dbErr
	}

	_, err = storage.conn.Exec(ctx, query, args...)

	if err != nil {
		dbErrStr := fmt.Sprintf("%s: %w", op, err)
		dbErr := NewDbErr(dbErrStr)
		storage.log.Error(err.Error(), slog.String("op", op))
		return dbErr
	}

	return nil
}

func (storage *Storage) DeleteBanner(ctx context.Context, bannerId int64) error {
	const op = "storage.postgres.DeleteBanner"

	query := `
		DELETE FROM banner WHERE id = $1
	`

	_, err := storage.conn.Exec(ctx, query, bannerId)

	if err != nil {
		storage.log.Error(err.Error(), slog.String("op", op))
		return NewDbErr(err.Error())
	}

	return nil
}
