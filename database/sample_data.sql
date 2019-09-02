--
-- PostgreSQL database dump
--

-- Dumped from database version 12.1
-- Dumped by pg_dump version 12.2

-- Started on 2020-07-20 14:21:04 UTC

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
-- TOC entry 3 (class 2615 OID 2200)
-- Name: public; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA sample;


ALTER SCHEMA sample OWNER TO postgres;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 203 (class 1259 OID 34256)
-- Name: algorithms; Type: TABLE; Schema: public; Owner: pooldetective
--

CREATE TABLE sample.algorithms (
    id smallint NOT NULL,
    name character varying(20)
);


ALTER TABLE sample.algorithms OWNER TO pooldetective;

--
-- TOC entry 205 (class 1259 OID 34261)
-- Name: block_coinbase_merkleproofs; Type: TABLE; Schema: public; Owner: pooldetective
--

CREATE TABLE sample.block_coinbase_merkleproofs (
    id bigint NOT NULL,
    block_id bigint NOT NULL,
    coinbase_merklebranches bytea[] NOT NULL
);


ALTER TABLE sample.block_coinbase_merkleproofs OWNER TO pooldetective;

--
-- TOC entry 206 (class 1259 OID 34267)
-- Name: block_coinbase_merkleproofs_id_seq; Type: SEQUENCE; Schema: public; Owner: pooldetective
--

CREATE TABLE sample.block_coinbase_outputs (
    id bigint NOT NULL,
    block_id bigint NOT NULL,
    output_script bytea NOT NULL,
    value bigint NOT NULL,
    output_index integer
);


ALTER TABLE sample.block_coinbase_outputs OWNER TO postgres;

--
-- TOC entry 207 (class 1259 OID 34277)
-- Name: block_observations; Type: TABLE; Schema: public; Owner: pooldetective
--

CREATE TABLE sample.block_observations (
    id bigint NOT NULL,
    location_id integer NOT NULL,
    block_id bigint NOT NULL,
    observed timestamp without time zone,
    peer_ip inet NOT NULL,
    peer_port integer NOT NULL,
    frommongo boolean
);


--
-- TOC entry 209 (class 1259 OID 34285)
-- Name: blocks; Type: TABLE; Schema: public; Owner: pooldetective
--

CREATE TABLE sample.blocks (
    id bigint NOT NULL,
    coin_id integer NOT NULL,
    block_hash bytea NOT NULL,
    previous_block_hash bytea,
    height integer,
    merkle_root bytea,
    "timestamp" integer
);


ALTER TABLE sample.blocks OWNER TO pooldetective;

--
-- TOC entry 211 (class 1259 OID 34293)
-- Name: coins; Type: TABLE; Schema: public; Owner: pooldetective
--

CREATE TABLE sample.coins (
    id integer NOT NULL,
    name character varying(64) NOT NULL,
    ticker character(5) NOT NULL,
    algorithm_id smallint
);


ALTER TABLE sample.coins OWNER TO pooldetective;

--
-- TOC entry 212 (class 1259 OID 34296)
-- Name: coins_id_seq; Type: SEQUENCE; Schema: public; Owner: pooldetective
--


CREATE TABLE sample.job_coinbase_outputs (
    id bigint NOT NULL,
    job_id bigint NOT NULL,
    output_script bytea NOT NULL,
    value bigint NOT NULL,
    output_index integer
);


ALTER TABLE sample.job_coinbase_outputs OWNER TO postgres;

--
-- TOC entry 213 (class 1259 OID 34306)
-- Name: jobs; Type: TABLE; Schema: public; Owner: pooldetective
--

CREATE TABLE sample.jobs (
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
    time_spent_msec bigint
);


ALTER TABLE sample.jobs OWNER TO pooldetective;

--
-- TOC entry 215 (class 1259 OID 34314)
-- Name: locations; Type: TABLE; Schema: public; Owner: pooldetective
--

CREATE TABLE sample.locations (
    id integer NOT NULL,
    name character varying(64) NOT NULL,
    reload_coordinator boolean
);

ALTER TABLE sample.locations OWNER TO pooldetective;


CREATE TABLE sample.pool_observers (
    id integer NOT NULL,
    location_id integer NOT NULL,
    pool_id integer NOT NULL,
    coin_id integer NOT NULL,
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
    stratum_target bytea
);


ALTER TABLE sample.pool_observers OWNER TO pooldetective;

--
-- TOC entry 221 (class 1259 OID 34335)
-- Name: pools; Type: TABLE; Schema: public; Owner: pooldetective
--

CREATE TABLE sample.pools (
    id integer NOT NULL,
    name character varying(64) NOT NULL
);


ALTER TABLE sample.pools OWNER TO pooldetective;


CREATE TABLE sample.shares (
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


ALTER TABLE sample.shares OWNER TO pooldetective;


CREATE TABLE sample.stratum_servers (
    id integer NOT NULL,
    algorithm_id integer,
    location_id integer,
    port integer,
    stratum_protocol smallint DEFAULT 0
);


ALTER TABLE sample.stratum_servers OWNER TO pooldetective;

--
-- TOC entry 2958 (class 2606 OID 34362)
-- Name: algorithms algorithms_pkey; Type: CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.algorithms
    ADD CONSTRAINT algorithms_pkey PRIMARY KEY (id);



ALTER TABLE ONLY sample.block_coinbase_merkleproofs
    ADD CONSTRAINT block_coinbase_merkleproofs_pkey PRIMARY KEY (id);


--
-- TOC entry 3020 (class 2606 OID 704564)
-- Name: block_coinbase_outputs block_coinbase_outputs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY sample.block_coinbase_outputs
    ADD CONSTRAINT block_coinbase_outputs_pkey PRIMARY KEY (id);


--
-- TOC entry 2962 (class 2606 OID 34368)
-- Name: block_observations block_observations_pkey; Type: CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.block_observations
    ADD CONSTRAINT block_observations_pkey PRIMARY KEY (id);


--
-- TOC entry 2965 (class 2606 OID 34370)
-- Name: blocks blocks_coin_id_block_hash_key; Type: CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.blocks
    ADD CONSTRAINT blocks_coin_id_block_hash_key UNIQUE (coin_id, block_hash);


--
-- TOC entry 2967 (class 2606 OID 34372)
-- Name: blocks blocks_pkey; Type: CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.blocks
    ADD CONSTRAINT blocks_pkey PRIMARY KEY (id);


--
-- TOC entry 2972 (class 2606 OID 34374)
-- Name: coins coins_pkey; Type: CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.coins
    ADD CONSTRAINT coins_pkey PRIMARY KEY (id);


--
-- TOC entry 3009 (class 2606 OID 492448)
-- Name: exchange_confirmations exchange_confirmations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

-- TOC entry 3018 (class 2606 OID 704550)
-- Name: job_coinbase_outputs job_coinbase_outputs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY sample.job_coinbase_outputs
    ADD CONSTRAINT job_coinbase_outputs_pkey PRIMARY KEY (id);


--
-- TOC entry 2980 (class 2606 OID 34378)
-- Name: jobs jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (id);


--
-- TOC entry 2982 (class 2606 OID 34380)
-- Name: locations locations_pkey; Type: CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.locations
    ADD CONSTRAINT locations_pkey PRIMARY KEY (id);


--
-- TOC entry 2986 (class 2606 OID 34386)
-- Name: pool_observers pool_observers_pkey; Type: CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.pool_observers
    ADD CONSTRAINT pool_observers_pkey PRIMARY KEY (id);


--
-- TOC entry 2988 (class 2606 OID 34388)
-- Name: pools pools_pkey; Type: CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.pools
    ADD CONSTRAINT pools_pkey PRIMARY KEY (id);


--
-- TOC entry 2991 (class 2606 OID 34390)
-- Name: shares shares_pkey; Type: CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.shares
    ADD CONSTRAINT shares_pkey PRIMARY KEY (id);


--
-- TOC entry 2993 (class 2606 OID 35162)
-- Name: stratum_servers stratum_servers_pkey; Type: CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.stratum_servers
    ADD CONSTRAINT stratum_servers_pkey PRIMARY KEY (id);


--
-- TOC entry 2963 (class 1259 OID 479938)
-- Name: idx_block_observations_block_id; Type: INDEX; Schema: public; Owner: pooldetective
--

CREATE INDEX idx_block_observations_block_id ON sample.block_observations USING btree (block_id);


--
-- TOC entry 2968 (class 1259 OID 35145)
-- Name: idx_blocks_coin; Type: INDEX; Schema: public; Owner: pooldetective
--

CREATE INDEX idx_blocks_coin ON sample.blocks USING btree (coin_id);


--
-- TOC entry 2969 (class 1259 OID 34391)
-- Name: idx_blocks_coin_hash; Type: INDEX; Schema: public; Owner: pooldetective
--

CREATE UNIQUE INDEX idx_blocks_coin_hash ON sample.blocks USING btree (coin_id, block_hash);


--
-- TOC entry 2973 (class 1259 OID 34392)
-- Name: idx_coins_ticker; Type: INDEX; Schema: public; Owner: pooldetective
--

CREATE UNIQUE INDEX idx_coins_ticker ON sample.coins USING btree (ticker);


--
-- TOC entry 2970 (class 1259 OID 478598)
-- Name: idx_hash; Type: INDEX; Schema: public; Owner: pooldetective
--

CREATE INDEX idx_hash ON sample.blocks USING btree (block_hash);


--
-- TOC entry 2974 (class 1259 OID 34393)
-- Name: idx_jobs_pool_observer_id; Type: INDEX; Schema: public; Owner: pooldetective
--

CREATE INDEX idx_jobs_pool_observer_id ON sample.jobs USING btree (pool_observer_id);


--
-- TOC entry 2975 (class 1259 OID 479975)
-- Name: idx_jobs_pool_observer_observed; Type: INDEX; Schema: public; Owner: pooldetective
--

CREATE INDEX idx_jobs_pool_observer_observed ON sample.jobs USING btree (pool_observer_id, observed);


--
-- TOC entry 2976 (class 1259 OID 479959)
-- Name: idx_jobs_previous_block_id; Type: INDEX; Schema: public; Owner: pooldetective
--

CREATE INDEX idx_jobs_previous_block_id ON sample.jobs USING btree (previous_block_id);


--
-- TOC entry 2977 (class 1259 OID 493799)
-- Name: idx_jobs_time_spent_next_job_id; Type: INDEX; Schema: public; Owner: pooldetective
--

CREATE INDEX idx_jobs_time_spent_next_job_id ON sample.jobs USING btree (next_job_id, time_spent_msec);


--
-- TOC entry 2989 (class 1259 OID 34394)
-- Name: idx_shares_job_id; Type: INDEX; Schema: public; Owner: pooldetective
--

CREATE INDEX idx_shares_job_id ON sample.shares USING btree (job_id);


--
-- TOC entry 3016 (class 1259 OID 704647)
-- Name: job_coinbase_outputs_idx1; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX job_coinbase_outputs_idx1 ON sample.job_coinbase_outputs USING btree (job_id);


--
-- TOC entry 2978 (class 1259 OID 475155)
-- Name: jobs_observed; Type: INDEX; Schema: public; Owner: pooldetective
--

CREATE INDEX jobs_observed ON sample.jobs USING btree (observed);


ALTER TABLE ONLY sample.block_coinbase_merkleproofs
    ADD CONSTRAINT block_coinbase_merkleproofs_block_id_fkey FOREIGN KEY (block_id) REFERENCES sample.blocks(id) ON DELETE CASCADE;


--
-- TOC entry 3022 (class 2606 OID 34406)
-- Name: block_observations block_observations_block_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.block_observations
    ADD CONSTRAINT block_observations_block_id_fkey FOREIGN KEY (block_id) REFERENCES sample.blocks(id) ON DELETE CASCADE;


--
-- TOC entry 3023 (class 2606 OID 34411)
-- Name: block_observations block_observations_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.block_observations
    ADD CONSTRAINT block_observations_location_id_fkey FOREIGN KEY (location_id) REFERENCES sample.locations(id) ON DELETE CASCADE;


--
-- TOC entry 3024 (class 2606 OID 34416)
-- Name: blocks blocks_coin_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.blocks
    ADD CONSTRAINT blocks_coin_id_fkey FOREIGN KEY (coin_id) REFERENCES sample.coins(id) ON DELETE CASCADE;


--
-- TOC entry 3025 (class 2606 OID 34421)
-- Name: coins coins_algorithm_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.coins
    ADD CONSTRAINT coins_algorithm_id_fkey FOREIGN KEY (algorithm_id) REFERENCES sample.algorithms(id) ON DELETE CASCADE;



--
-- TOC entry 3027 (class 2606 OID 475150)
-- Name: jobs jobs_next_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.jobs
    ADD CONSTRAINT jobs_next_job_id_fkey FOREIGN KEY (next_job_id) REFERENCES sample.jobs(id);


--
-- TOC entry 3026 (class 2606 OID 34426)
-- Name: jobs jobs_previous_block_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.jobs
    ADD CONSTRAINT jobs_previous_block_id_fkey FOREIGN KEY (previous_block_id) REFERENCES sample.blocks(id) ON DELETE SET NULL;


--
-- TOC entry 3028 (class 2606 OID 34441)
-- Name: pool_observers pool_observers_coinid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.pool_observers
    ADD CONSTRAINT pool_observers_coinid_fkey FOREIGN KEY (coin_id) REFERENCES sample.coins(id) ON DELETE CASCADE;


--
-- TOC entry 3029 (class 2606 OID 34446)
-- Name: pool_observers pool_observers_locationid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.pool_observers
    ADD CONSTRAINT pool_observers_locationid_fkey FOREIGN KEY (location_id) REFERENCES sample.locations(id) ON DELETE CASCADE;


--
-- TOC entry 3030 (class 2606 OID 34451)
-- Name: pool_observers pool_observers_poolid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.pool_observers
    ADD CONSTRAINT pool_observers_poolid_fkey FOREIGN KEY (pool_id) REFERENCES sample.pools(id) ON DELETE CASCADE;


--
-- TOC entry 3031 (class 2606 OID 34456)
-- Name: shares shares_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pooldetective
--

ALTER TABLE ONLY sample.shares
    ADD CONSTRAINT shares_job_id_fkey FOREIGN KEY (job_id) REFERENCES sample.jobs(id) ON DELETE CASCADE;


-- Completed on 2020-07-20 14:21:05 UTC

--
-- PostgreSQL database dump complete
--




INSERT INTO sample.algorithms(id, name) SELECT id, name from public.algorithms
INSERT INTO sample.coins(id, name, ticker, algorithm_id) SELECT id, name, ticker, algorithm_id FROM public.coins
INSERT INTO sample.locations(id, name) SELECT id, name FROM public.locations
INSERT INTO sample.pools(id, name) SELECT id, name FROM public.pools
INSERT INTO sample.stratum_servers(id, algorithm_id, location_id, port, stratum_protocol) SELECT id, algorithm_id, location_id, port, stratum_protocol from public.stratum_servers
INSERT INTO sample.pool_observers(id, location_id, pool_id, coin_id, stratum_host, stratum_port, stratum_username, stratum_password, disabled, frommongo, last_job_received, last_share_found, last_job_id, last_share_id, stratum_protocol) SELECT id, location_id, pool_id, coin_id, stratum_host, stratum_port, null, null, disabled, frommongo, null,null,null,null,stratum_protocol,null FROM public.pool_observers
INSERT INTO sample.jobs(id, observed, pool_observer_id, pool_job_id, previous_block_hash, generation_transaction_part_1, generation_transaction_part_2, merkle_branches, block_version, difficulty_bits, clean_jobs, timestamp, reserved) SELECT id, observed, pool_observer_id, pool_job_id, previous_block_hash, generation_transaction_part_1, generation_transaction_part_2, merkle_branches, block_version, difficulty_bits, clean_jobs, timestamp, reserved FROM public.jobs WHERE observed >= '2020-03-01' AND observed < '2020-03-15'
INSERT INTO sample.blocks(id, coin_id, block_hash, previous_block_hash, height, merkle_root, timestamp) SELECT id, coin_id, block_hash, previous_block_hash, height, merkle_root, timestamp FROM public.blocks WHERE id IN (SELECT block_id FROM block_observations WHERE observed >= '2020-03-01' AND observed < '2020-03-15');
INSERT INTO sample.blocks(id, coin_id, block_hash, previous_block_hash, height, merkle_root, timestamp) SELECT id, coin_id, block_hash, previous_block_hash, height, merkle_root, timestamp FROM public.blocks WHERE id NOT IN (SELECT id FROM sample.blocks) AND timestamp >= 1583020500 AND timestamp < 1584230700;
INSERT INTO sample.block_observations(id, location_id, block_id, observed,peer_ip, peer_port) SELECT id, location_id, block_id, observed,peer_ip, peer_port FROM public.block_observations WHERE observed >= '2020-03-01' AND observed < '2020-03-15';
INSERT INTO sample.shares(id, job_id, extranonce2, timestamp, nonce, found, submitted, responsereceived, accepted, details, stale, additional_solution_data) SELECT id, job_id, extranonce2, timestamp, nonce, found, submitted, responsereceived, accepted, details, stale, additional_solution_data FROM public.shares WHERE job_id IN (select id from sample.jobs);
INSERT INTO sample.job_coinbase_outputs(id, job_id, output_script, value, output_index) SELECT id, job_id, output_script, value, output_index FROM public.job_coinbase_outputs WHERE job_id in (select id from sample.jobs);
INSERT INTO sample.block_coinbase_outputs(id, block_id, output_script, value, output_index) SELECT id, block_id, output_script, value, output_index FROM public.block_coinbase_outputs WHERE block_id in (select id from sample.blocks);
INSERT INTO sample.block_coinbase_merkleproofs(id, block_id, coinbase_merklebranches) SELECT id, block_id, coinbase_merklebranches FROM public.block_coinbase_merkleproofs WHERE block_id in (SELECT id from sample.blocks);