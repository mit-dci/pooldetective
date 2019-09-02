--
-- PostgreSQL database dump
--

-- Dumped from database version 12.4
-- Dumped by pg_dump version 12.2

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA public;


--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA public IS 'standard public schema';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: algorithms; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.algorithms (
    id smallint NOT NULL,
    name character varying(20)
);


--
-- Name: algorithms_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.algorithms_id_seq
    AS smallint
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: algorithms_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.algorithms_id_seq OWNED BY public.algorithms.id;


--
-- Name: analysis_blocks_first_seen; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.analysis_blocks_first_seen (
    block_id bigint NOT NULL,
    first_observation timestamp without time zone
);


--
-- Name: analysis_competing_blocks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.analysis_competing_blocks (
    coin_id integer NOT NULL,
    block_hash bytea NOT NULL,
    previous_block_hash bytea NOT NULL,
    height integer NOT NULL
);


--
-- Name: analysis_fork_blocks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.analysis_fork_blocks (
    coin_id integer NOT NULL,
    fork_block_hash bytea NOT NULL,
    fork_height integer NOT NULL,
    next_fork_height integer
);


--
-- Name: analysis_fork_depth; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.analysis_fork_depth (
    coin_id integer NOT NULL,
    fork_block_hash bytea NOT NULL,
    fork_depth integer NOT NULL
);


--
-- Name: analysis_jobs_first_seen; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.analysis_jobs_first_seen (
    previous_block_id bigint NOT NULL,
    first_observation timestamp without time zone
);


--
-- Name: analysis_pool_output_scripts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.analysis_pool_output_scripts (
    pool_id integer,
    output_script bytea,
    address character varying(50),
    first_seen timestamp without time zone,
    last_seen timestamp without time zone
);


--
-- Name: analysis_pool_output_scripts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.analysis_pool_output_scripts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: analysis_wrong_work; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.analysis_wrong_work (
    pool_id integer NOT NULL,
    pool_observer_id integer NOT NULL,
    location_id integer NOT NULL,
    expected_coin_id integer NOT NULL,
    got_coin_id integer NOT NULL,
    total_jobs bigint NOT NULL,
    wrong_jobs bigint NOT NULL,
    observed_date date NOT NULL
);


--
-- Name: analysis_wrong_work_daily; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.analysis_wrong_work_daily (
    observed_on date NOT NULL,
    pool_id integer NOT NULL,
    pool_observer_id integer NOT NULL,
    location_id integer NOT NULL,
    expected_coin_id integer NOT NULL,
    got_coin_id integer NOT NULL,
    total_jobs bigint NOT NULL,
    total_time_msec bigint NOT NULL,
    wrong_jobs bigint NOT NULL,
    wrong_time_msec bigint NOT NULL
);


--
-- Name: block_coinbase_merkleproofs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.block_coinbase_merkleproofs (
    block_id bigint NOT NULL,
    coinbase_merklebranches bytea[] NOT NULL
);


--
-- Name: block_coinbase_outputs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.block_coinbase_outputs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: block_coinbase_outputs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.block_coinbase_outputs (
    id bigint DEFAULT nextval('public.block_coinbase_outputs_id_seq'::regclass) NOT NULL,
    block_id bigint NOT NULL,
    output_script bytea NOT NULL,
    value bigint NOT NULL,
    output_index integer
);


--
-- Name: block_observations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.block_observations (
    id bigint NOT NULL,
    location_id integer NOT NULL,
    block_id bigint NOT NULL,
    observed timestamp without time zone,
    peer_ip inet NOT NULL,
    peer_port integer NOT NULL,
    frommongo boolean
);


--
-- Name: block_observations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.block_observations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: block_observations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.block_observations_id_seq OWNED BY public.block_observations.id;


--
-- Name: blocks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blocks (
    id bigint NOT NULL,
    coin_id integer NOT NULL,
    block_hash bytea NOT NULL,
    previous_block_hash bytea,
    height integer,
    merkle_root bytea,
    "timestamp" integer,
    coinbase_data boolean
);


--
-- Name: blocks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.blocks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: blocks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.blocks_id_seq OWNED BY public.blocks.id;


--
-- Name: coin_algorithm; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.coin_algorithm (
    coin_id integer NOT NULL,
    algorithm_id smallint NOT NULL
);


--
-- Name: coins; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.coins (
    id integer NOT NULL,
    name character varying(64) NOT NULL,
    ticker character(8) NOT NULL,
    algorithm_id smallint
);


--
-- Name: coins_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.coins_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: coins_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.coins_id_seq OWNED BY public.coins.id;


--
-- Name: exchange_confirmations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exchange_confirmations (
    exchange_id integer NOT NULL,
    coin_id integer NOT NULL,
    "timestamp" timestamp without time zone NOT NULL,
    confirmations integer,
    source_id smallint
);


--
-- Name: exchange_confirmations_source; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exchange_confirmations_source (
    id smallint NOT NULL,
    description character varying(100)
);


--
-- Name: exchange_confirmations_source_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.exchange_confirmations_source_id_seq
    AS smallint
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exchange_confirmations_source_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.exchange_confirmations_source_id_seq OWNED BY public.exchange_confirmations_source.id;


--
-- Name: exchanges; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exchanges (
    id integer NOT NULL,
    name character varying(60)
);


--
-- Name: exchanges_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.exchanges_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exchanges_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.exchanges_id_seq OWNED BY public.exchanges.id;


--
-- Name: job_coinbase_outputs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.job_coinbase_outputs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: job_coinbase_outputs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.job_coinbase_outputs (
    id bigint DEFAULT nextval('public.job_coinbase_outputs_id_seq'::regclass) NOT NULL,
    job_id bigint NOT NULL,
    output_script bytea NOT NULL,
    value bigint NOT NULL,
    output_index integer
);


--
-- Name: jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.jobs (
    id bigint NOT NULL,
    observed timestamp without time zone NOT NULL,
    pool_observer_id integer NOT NULL,
    pool_job_id bytea NOT NULL,
    previous_block_hash bytea NOT NULL,
    previous_block_id integer,
    generation_transaction_part_1 bytea,
    generation_transaction_part_2 bytea,
    merkle_branches bytea[],
    block_version bytea,
    difficulty_bits bytea,
    clean_jobs boolean,
    "timestamp" bigint,
    frommongo boolean,
    tempid bigint,
    next_job_id bigint,
    reserved bytea,
    previous_block_hash_corrected bytea,
    time_spent_msec bigint,
    probable_blockheight integer
);


--
-- Name: jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.jobs_id_seq OWNED BY public.jobs.id;


--
-- Name: locations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.locations (
    id integer NOT NULL,
    name character varying(64) NOT NULL,
    reload_coordinator boolean
);


--
-- Name: locations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.locations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: locations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.locations_id_seq OWNED BY public.locations.id;


--
-- Name: pool_observer_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.pool_observer_events (
    id bigint NOT NULL,
    pool_observer_id integer,
    event smallint,
    "timestamp" timestamp without time zone
);


--
-- Name: pool_observer_events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.pool_observer_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pool_observer_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.pool_observer_events_id_seq OWNED BY public.pool_observer_events.id;


--
-- Name: pool_observers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.pool_observers (
    id integer NOT NULL,
    location_id integer NOT NULL,
    pool_id integer NOT NULL,
    coin_id integer,
    stratum_host character varying(100),
    stratum_port smallint,
    stratum_username character varying(50),
    stratum_password character varying(30),
    stratum_difficulty numeric(30,20),
    stratum_extranonce1 bytea,
    stratum_extranonce2size smallint,
    disabled boolean DEFAULT false,
    frommongo boolean,
    last_job_received timestamp without time zone,
    last_share_found timestamp without time zone,
    last_job_id bigint,
    last_share_id bigint,
    stratum_protocol smallint DEFAULT 0,
    stratum_target bytea,
    algorithm_id integer
);


--
-- Name: pool_observers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.pool_observers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pool_observers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.pool_observers_id_seq OWNED BY public.pool_observers.id;


--
-- Name: pools; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.pools (
    id integer NOT NULL,
    name character varying(64) NOT NULL
);


--
-- Name: pools_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.pools_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pools_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.pools_id_seq OWNED BY public.pools.id;


--
-- Name: shares; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.shares (
    id bigint NOT NULL,
    job_id bigint,
    extranonce2 bytea NOT NULL,
    "timestamp" bigint NOT NULL,
    nonce bigint NOT NULL,
    found timestamp without time zone,
    submitted timestamp without time zone,
    responsereceived timestamp without time zone,
    accepted boolean,
    details text,
    stale boolean,
    frommongo boolean,
    jobtempid bigint,
    additional_solution_data bytea[]
);


--
-- Name: shares_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.shares_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: shares_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.shares_id_seq OWNED BY public.shares.id;


--
-- Name: stratum_servers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stratum_servers (
    id integer NOT NULL,
    algorithm_id integer,
    location_id integer,
    port integer,
    stratum_protocol smallint DEFAULT 0
);


--
-- Name: stratum_servers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.stratum_servers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: stratum_servers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.stratum_servers_id_seq OWNED BY public.stratum_servers.id;


--
-- Name: algorithms id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.algorithms ALTER COLUMN id SET DEFAULT nextval('public.algorithms_id_seq'::regclass);


--
-- Name: block_observations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.block_observations ALTER COLUMN id SET DEFAULT nextval('public.block_observations_id_seq'::regclass);


--
-- Name: blocks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blocks ALTER COLUMN id SET DEFAULT nextval('public.blocks_id_seq'::regclass);


--
-- Name: coins id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.coins ALTER COLUMN id SET DEFAULT nextval('public.coins_id_seq'::regclass);


--
-- Name: exchange_confirmations_source id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exchange_confirmations_source ALTER COLUMN id SET DEFAULT nextval('public.exchange_confirmations_source_id_seq'::regclass);


--
-- Name: exchanges id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exchanges ALTER COLUMN id SET DEFAULT nextval('public.exchanges_id_seq'::regclass);


--
-- Name: jobs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs ALTER COLUMN id SET DEFAULT nextval('public.jobs_id_seq'::regclass);


--
-- Name: locations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locations ALTER COLUMN id SET DEFAULT nextval('public.locations_id_seq'::regclass);


--
-- Name: pool_observer_events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pool_observer_events ALTER COLUMN id SET DEFAULT nextval('public.pool_observer_events_id_seq'::regclass);


--
-- Name: pool_observers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pool_observers ALTER COLUMN id SET DEFAULT nextval('public.pool_observers_id_seq'::regclass);


--
-- Name: pools id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pools ALTER COLUMN id SET DEFAULT nextval('public.pools_id_seq'::regclass);


--
-- Name: shares id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.shares ALTER COLUMN id SET DEFAULT nextval('public.shares_id_seq'::regclass);


--
-- Name: stratum_servers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stratum_servers ALTER COLUMN id SET DEFAULT nextval('public.stratum_servers_id_seq'::regclass);


--
-- Name: algorithms algorithms_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.algorithms
    ADD CONSTRAINT algorithms_pkey PRIMARY KEY (id);


--
-- Name: analysis_blocks_first_seen analysis_blocks_first_seen_new_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.analysis_blocks_first_seen
    ADD CONSTRAINT analysis_blocks_first_seen_new_pkey PRIMARY KEY (block_id);


--
-- Name: analysis_competing_blocks analysis_competing_blocks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.analysis_competing_blocks
    ADD CONSTRAINT analysis_competing_blocks_pkey PRIMARY KEY (coin_id, block_hash);


--
-- Name: analysis_fork_blocks analysis_fork_blocks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.analysis_fork_blocks
    ADD CONSTRAINT analysis_fork_blocks_pkey PRIMARY KEY (coin_id, fork_block_hash);


--
-- Name: analysis_fork_depth analysis_fork_depth_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.analysis_fork_depth
    ADD CONSTRAINT analysis_fork_depth_pkey PRIMARY KEY (coin_id, fork_block_hash);


--
-- Name: analysis_jobs_first_seen analysis_jobs_first_seen_new_pkey1; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.analysis_jobs_first_seen
    ADD CONSTRAINT analysis_jobs_first_seen_new_pkey1 PRIMARY KEY (previous_block_id);


--
-- Name: analysis_wrong_work_daily analysis_wrong_work_daily_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.analysis_wrong_work_daily
    ADD CONSTRAINT analysis_wrong_work_daily_pkey PRIMARY KEY (observed_on, pool_observer_id, got_coin_id);


--
-- Name: analysis_wrong_work analysis_wrong_work_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.analysis_wrong_work
    ADD CONSTRAINT analysis_wrong_work_pkey PRIMARY KEY (pool_observer_id, got_coin_id, observed_date);


--
-- Name: block_coinbase_merkleproofs block_coinbase_merkleproofs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.block_coinbase_merkleproofs
    ADD CONSTRAINT block_coinbase_merkleproofs_pkey PRIMARY KEY (block_id);


--
-- Name: block_coinbase_outputs block_coinbase_outputs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.block_coinbase_outputs
    ADD CONSTRAINT block_coinbase_outputs_pkey PRIMARY KEY (id);


--
-- Name: block_observations block_observations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.block_observations
    ADD CONSTRAINT block_observations_pkey PRIMARY KEY (id);


--
-- Name: blocks blocks_coin_id_block_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blocks
    ADD CONSTRAINT blocks_coin_id_block_hash_key UNIQUE (coin_id, block_hash);


--
-- Name: blocks blocks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blocks
    ADD CONSTRAINT blocks_pkey PRIMARY KEY (id);


--
-- Name: coin_algorithm coin_algorithm_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.coin_algorithm
    ADD CONSTRAINT coin_algorithm_pkey PRIMARY KEY (coin_id, algorithm_id);


--
-- Name: coins coins_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.coins
    ADD CONSTRAINT coins_pkey PRIMARY KEY (id);


--
-- Name: exchange_confirmations exchange_confirmations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exchange_confirmations
    ADD CONSTRAINT exchange_confirmations_pkey PRIMARY KEY (exchange_id, coin_id, "timestamp");


--
-- Name: exchange_confirmations_source exchange_confirmations_source_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exchange_confirmations_source
    ADD CONSTRAINT exchange_confirmations_source_pkey PRIMARY KEY (id);


--
-- Name: exchanges exchanges_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exchanges
    ADD CONSTRAINT exchanges_pkey PRIMARY KEY (id);


--
-- Name: job_coinbase_outputs job_coinbase_outputs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_coinbase_outputs
    ADD CONSTRAINT job_coinbase_outputs_pkey PRIMARY KEY (id);


--
-- Name: jobs jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (id);


--
-- Name: locations locations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locations
    ADD CONSTRAINT locations_pkey PRIMARY KEY (id);


--
-- Name: pool_observer_events pool_observer_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pool_observer_events
    ADD CONSTRAINT pool_observer_events_pkey PRIMARY KEY (id);


--
-- Name: pool_observers pool_observers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pool_observers
    ADD CONSTRAINT pool_observers_pkey PRIMARY KEY (id);


--
-- Name: pools pools_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pools
    ADD CONSTRAINT pools_pkey PRIMARY KEY (id);


--
-- Name: shares shares_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.shares
    ADD CONSTRAINT shares_pkey PRIMARY KEY (id);


--
-- Name: stratum_servers stratum_servers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stratum_servers
    ADD CONSTRAINT stratum_servers_pkey PRIMARY KEY (id);


--
-- Name: idx_analysis_competing_blocks1; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_analysis_competing_blocks1 ON public.analysis_competing_blocks USING btree (coin_id, height);


--
-- Name: idx_block_coinbase_outputs_block_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_block_coinbase_outputs_block_id ON public.block_coinbase_outputs USING btree (block_id);


--
-- Name: idx_block_observations_block_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_block_observations_block_id ON public.block_observations USING btree (block_id);


--
-- Name: idx_blocks_coin; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_blocks_coin ON public.blocks USING btree (coin_id);


--
-- Name: idx_blocks_coin_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_blocks_coin_hash ON public.blocks USING btree (coin_id, block_hash);


--
-- Name: idx_blocks_height_null; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_blocks_height_null ON public.blocks USING btree (((height IS NOT NULL))) WHERE (height IS NOT NULL);


--
-- Name: idx_coin_height_coinbase; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_coin_height_coinbase ON public.blocks USING btree (coin_id, height) WHERE ((height IS NOT NULL) AND (coinbase_data IS NULL));


--
-- Name: idx_coins_ticker; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_coins_ticker ON public.coins USING btree (ticker);


--
-- Name: idx_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_hash ON public.blocks USING btree (block_hash);


--
-- Name: idx_hash_height_not_null; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_hash_height_not_null ON public.blocks USING btree (block_hash) WHERE (height IS NOT NULL);


--
-- Name: idx_jobs_pool_observer_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_jobs_pool_observer_id ON public.jobs USING btree (pool_observer_id);


--
-- Name: idx_jobs_pool_observer_observed; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_jobs_pool_observer_observed ON public.jobs USING btree (pool_observer_id, observed);


--
-- Name: idx_jobs_previous_block_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_jobs_previous_block_hash ON public.jobs USING btree (previous_block_hash);


--
-- Name: idx_jobs_previous_block_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_jobs_previous_block_id ON public.jobs USING btree (previous_block_id);


--
-- Name: idx_jobs_previous_block_null; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_jobs_previous_block_null ON public.jobs USING btree (((previous_block_id IS NULL))) WHERE (previous_block_id IS NULL);


--
-- Name: idx_jobs_time_spent_next_job_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_jobs_time_spent_next_job_id ON public.jobs USING btree (next_job_id, time_spent_msec);


--
-- Name: idx_previous_block_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_previous_block_hash ON public.blocks USING btree (previous_block_hash);


--
-- Name: idx_shares_job_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_shares_job_id ON public.shares USING btree (job_id);


--
-- Name: job_coinbase_outputs_idx1; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX job_coinbase_outputs_idx1 ON public.job_coinbase_outputs USING btree (job_id);


--
-- Name: jobs_observed; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX jobs_observed ON public.jobs USING btree (observed);


--
-- Name: analysis_fork_blocks analysis_fork_blocks_fkey_coin; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.analysis_fork_blocks
    ADD CONSTRAINT analysis_fork_blocks_fkey_coin FOREIGN KEY (coin_id) REFERENCES public.coins(id);


--
-- Name: analysis_fork_depth analysis_fork_depth_fkey_coin; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.analysis_fork_depth
    ADD CONSTRAINT analysis_fork_depth_fkey_coin FOREIGN KEY (coin_id) REFERENCES public.coins(id);


--
-- Name: block_coinbase_merkleproofs block_coinbase_merkleproofs_block_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.block_coinbase_merkleproofs
    ADD CONSTRAINT block_coinbase_merkleproofs_block_id_fkey FOREIGN KEY (block_id) REFERENCES public.blocks(id);


--
-- Name: block_observations block_observations_block_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.block_observations
    ADD CONSTRAINT block_observations_block_id_fkey FOREIGN KEY (block_id) REFERENCES public.blocks(id) ON DELETE CASCADE;


--
-- Name: block_observations block_observations_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.block_observations
    ADD CONSTRAINT block_observations_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.locations(id) ON DELETE CASCADE;


--
-- Name: blocks blocks_coin_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blocks
    ADD CONSTRAINT blocks_coin_id_fkey FOREIGN KEY (coin_id) REFERENCES public.coins(id) ON DELETE CASCADE;


--
-- Name: coin_algorithm coin_algorithm_algorithm_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.coin_algorithm
    ADD CONSTRAINT coin_algorithm_algorithm_id_fkey FOREIGN KEY (algorithm_id) REFERENCES public.algorithms(id);


--
-- Name: coin_algorithm coin_algorithm_coin_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.coin_algorithm
    ADD CONSTRAINT coin_algorithm_coin_id_fkey FOREIGN KEY (coin_id) REFERENCES public.coins(id);


--
-- Name: coins coins_algorithm_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.coins
    ADD CONSTRAINT coins_algorithm_id_fkey FOREIGN KEY (algorithm_id) REFERENCES public.algorithms(id) ON DELETE CASCADE;


--
-- Name: exchange_confirmations fk_exchange_confirmations_coin; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exchange_confirmations
    ADD CONSTRAINT fk_exchange_confirmations_coin FOREIGN KEY (coin_id) REFERENCES public.coins(id);


--
-- Name: exchange_confirmations fk_exchange_confirmations_exchange; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exchange_confirmations
    ADD CONSTRAINT fk_exchange_confirmations_exchange FOREIGN KEY (exchange_id) REFERENCES public.exchanges(id);


--
-- Name: exchange_confirmations fk_exchange_confirmations_source; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exchange_confirmations
    ADD CONSTRAINT fk_exchange_confirmations_source FOREIGN KEY (source_id) REFERENCES public.exchange_confirmations_source(id) NOT VALID;


--
-- Name: jobs jobs_next_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_next_job_id_fkey FOREIGN KEY (next_job_id) REFERENCES public.jobs(id);


--
-- Name: jobs jobs_previous_block_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_previous_block_id_fkey FOREIGN KEY (previous_block_id) REFERENCES public.blocks(id) ON DELETE SET NULL;


--
-- Name: pool_observers pool_observers_coinid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pool_observers
    ADD CONSTRAINT pool_observers_coinid_fkey FOREIGN KEY (coin_id) REFERENCES public.coins(id) ON DELETE CASCADE;


--
-- Name: pool_observers pool_observers_locationid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pool_observers
    ADD CONSTRAINT pool_observers_locationid_fkey FOREIGN KEY (location_id) REFERENCES public.locations(id) ON DELETE CASCADE;


--
-- Name: pool_observers pool_observers_poolid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pool_observers
    ADD CONSTRAINT pool_observers_poolid_fkey FOREIGN KEY (pool_id) REFERENCES public.pools(id) ON DELETE CASCADE;


--
-- Name: shares shares_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.shares
    ADD CONSTRAINT shares_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.jobs(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

