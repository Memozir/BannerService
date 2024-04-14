package queries

var (
	QBannersAllFilters = `
	select *
	from (
         select
             distinct b.id as banner_id,
                      f.feature_id as feature_id,
                      b.content::text as content,
                      b.is_active as is_active,
                      cast(b.created_at::timestamp with time zone as text) as created_at,
                      coalesce(cast(b.updated_at::timestamp with time zone as text), '') as updated_at,
                      array((
                          select bn_tags.tag_id
                          from (
                                   select t1.tag_id
                                   from banner_tag bt1
                                            inner join tag t1 on bt1.tag_id = t1.id
                                   where bt1.banner_id = b.id
                               ) bn_tags
                          where $2 = any(array((
                              select t1.tag_id
                              from banner_tag bt1
                                       inner join tag t1 on bt1.tag_id = t1.id
                              where bt1.banner_id = b.id
                          )))
                      )) tag_ids
         from banner b
                  inner join feature f on b.banner_feature = f.id
                  inner join banner_tag bt on b.id = bt.banner_id
         where f.feature_id = $1
         group by b.id, bt.tag_id, f.feature_id
         limit $3
		 offset $4
     ) t1
	where array_length(t1.tag_ids, 1) > 0
	`

	QBannersTagIdFilter = `
	select *
	from (
         select
             distinct b.id as banner_id,
                      f.feature_id as feature_id,
                      b.content::text as content,
                      b.is_active as is_active,
                      cast(b.created_at::timestamp with time zone as text) as created_at,
                      coalesce(cast(b.updated_at::timestamp with time zone as text), '') as updated_at,
                      array((
                          select bn_tags.tag_id
                          from (
                                   select t1.tag_id
                                   from banner_tag bt1
                                            inner join tag t1 on bt1.tag_id = t1.id
                                   where bt1.banner_id = b.id
                               ) bn_tags
                          where $2 = any(array((
                              select t1.tag_id
                              from banner_tag bt1
                                       inner join tag t1 on bt1.tag_id = t1.id
                              where bt1.banner_id = b.id
                          )))
                      )) tag_ids
         from banner b
                  inner join feature f on b.banner_feature = f.id
                  inner join banner_tag bt on b.id = bt.banner_id
         group by b.id, bt.tag_id, f.feature_id
         limit $3
		 offset $4
     ) t1
	where array_length(t1.tag_ids, 1) > 0
	`

	QBannersFeatureIdFilter = `
	select *
	from (
         select
             distinct b.id as banner_id,
                      f.feature_id as feature_id,
                      b.content::text as content,
                      b.is_active as is_active,
                      cast(b.created_at::timestamp with time zone as text) as created_at,
                      coalesce(cast(b.updated_at::timestamp with time zone as text), '') as updated_at,
                      array((
                          select bn_tags.tag_id
                          from (
                                   select t1.tag_id
                                   from banner_tag bt1
                                            inner join tag t1 on bt1.tag_id = t1.id
                                   where bt1.banner_id = b.id
                               ) bn_tags
                          where $2 = any(array((
                              select t1.tag_id
                              from banner_tag bt1
                                       inner join tag t1 on bt1.tag_id = t1.id
                              where bt1.banner_id = b.id
                          )))
                      )) tag_ids
         from banner b
                  inner join feature f on b.banner_feature = f.id
                  inner join banner_tag bt on b.id = bt.banner_id
         where f.feature_id = $1
         group by b.id, bt.tag_id, f.feature_id
         limit $3
		 offset $4
     ) t1
	where array_length(t1.tag_ids, 1) > 0
	`
)
