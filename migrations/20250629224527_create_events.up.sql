CREATE EXTENSION "pg_partman";

CREATE TABLE events(
    seq bigint GENERATED ALWAYS AS IDENTITY,
    uuid uuid NOT NULL,
    name varchar NOT NULL,
    actor_id uuid NOT NULL,
    aggregate_id uuid NOT NULL,

    payload jsonb NOT NULL,
    created_at timestamptz DEFAULT NOW() NOT NULL
) PARTITION BY range(created_at);

CREATE INDEX ON events(created_at);

CREATE TABLE events_table_template (LIKE events);

ALTER TABLE events_table_template ADD PRIMARY KEY (seq);

SELECT create_parent(
   p_parent_table := 'public.events',
   p_control := 'created_at',
   p_interval := '10 seconds',
   p_template_table := 'public.events_table_template'
);