--
-- PostgreSQL database dump
--

-- Dumped from database version 11.6 (Debian 11.6-1.pgdg90+1)
-- Dumped by pg_dump version 12.1 (Debian 12.1-1.pgdg100+1)

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

SET default_tablespace = '';

--
-- Name: files; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.files (
    id integer NOT NULL,
    uuid uuid NOT NULL,
    url character varying(1024) NOT NULL,
    url_expiration timestamp without time zone NOT NULL,
    name character varying(255) NOT NULL,
    size integer NOT NULL,
    content_type character varying(255) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.files OWNER TO scrutinizer;

--
-- Name: files_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.files_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.files_id_seq OWNER TO scrutinizer;

--
-- Name: files_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.files_id_seq OWNED BY public.files.id;


--
-- Name: locations; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.locations (
    id integer NOT NULL,
    description character varying(255) DEFAULT ''::character varying NOT NULL,
    country character varying(2) DEFAULT ''::character varying NOT NULL,
    latitude numeric(8,5),
    longitude numeric(8,5)
);


ALTER TABLE public.locations OWNER TO scrutinizer;

--
-- Name: locations_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.locations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.locations_id_seq OWNER TO scrutinizer;

--
-- Name: locations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.locations_id_seq OWNED BY public.locations.id;


--
-- Name: meetings; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.meetings (
    id integer NOT NULL,
    uuid uuid NOT NULL,
    name character varying(80) NOT NULL,
    description character varying(4096),
    more_info_url character varying(255),
    image_file_id integer,
    created_by_id integer,
    location_id integer NOT NULL,
    start_date date NOT NULL,
    end_date date NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.meetings OWNER TO scrutinizer;

--
-- Name: meetings_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.meetings_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.meetings_id_seq OWNER TO scrutinizer;

--
-- Name: meetings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.meetings_id_seq OWNED BY public.meetings.id;


--
-- Name: messages; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.messages (
    id integer NOT NULL,
    uuid uuid NOT NULL,
    thread_id integer NOT NULL,
    sent_by_id integer NOT NULL,
    content character varying(4096) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.messages OWNER TO scrutinizer;

--
-- Name: messages_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.messages_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.messages_id_seq OWNER TO scrutinizer;

--
-- Name: messages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.messages_id_seq OWNED BY public.messages.id;


--
-- Name: organization_domains; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.organization_domains (
    id integer NOT NULL,
    organization_id integer NOT NULL,
    domain character varying(255) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.organization_domains OWNER TO scrutinizer;

--
-- Name: organization_domains_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.organization_domains_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.organization_domains_id_seq OWNER TO scrutinizer;

--
-- Name: organization_domains_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.organization_domains_id_seq OWNED BY public.organization_domains.id;


--
-- Name: organization_trusts; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.organization_trusts (
    id integer NOT NULL,
    primary_id integer NOT NULL,
    secondary_id integer NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.organization_trusts OWNER TO scrutinizer;

--
-- Name: organization_trusts_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.organization_trusts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.organization_trusts_id_seq OWNER TO scrutinizer;

--
-- Name: organization_trusts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.organization_trusts_id_seq OWNED BY public.organization_trusts.id;


--
-- Name: organizations; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.organizations (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    url character varying(255),
    auth_type character varying(255) NOT NULL,
    auth_config jsonb NOT NULL,
    uuid uuid NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    logo_file_id integer
);


ALTER TABLE public.organizations OWNER TO scrutinizer;

--
-- Name: organizations_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.organizations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.organizations_id_seq OWNER TO scrutinizer;

--
-- Name: organizations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.organizations_id_seq OWNED BY public.organizations.id;


--
-- Name: post_files; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.post_files (
    id integer NOT NULL,
    post_id integer NOT NULL,
    file_id integer NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.post_files OWNER TO scrutinizer;

--
-- Name: post_files_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.post_files_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.post_files_id_seq OWNER TO scrutinizer;

--
-- Name: post_files_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.post_files_id_seq OWNED BY public.post_files.id;


--
-- Name: post_histories; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.post_histories (
    id integer NOT NULL,
    post_id integer NOT NULL,
    receiver_id integer,
    provider_id integer,
    status character varying(255) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.post_histories OWNER TO scrutinizer;

--
-- Name: post_histories_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.post_histories_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.post_histories_id_seq OWNER TO scrutinizer;

--
-- Name: post_histories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.post_histories_id_seq OWNED BY public.post_histories.id;


--
-- Name: posts; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.posts (
    id integer NOT NULL,
    created_by_id integer NOT NULL,
    type character varying(255) NOT NULL,
    organization_id integer NOT NULL,
    status character varying(255) NOT NULL,
    title character varying(255) NOT NULL,
    size character varying(255) NOT NULL,
    uuid uuid NOT NULL,
    receiver_id integer,
    provider_id integer,
    description character varying(4096),
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    url character varying(255),
    photo_file_id integer,
    destination_id integer NOT NULL,
    origin_id integer,
    kilograms numeric(13,4) DEFAULT '0'::numeric NOT NULL,
    meeting_id integer,
    visibility character varying(255) DEFAULT 'SAME'::character varying NOT NULL
);


ALTER TABLE public.posts OWNER TO scrutinizer;

--
-- Name: posts_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.posts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.posts_id_seq OWNER TO scrutinizer;

--
-- Name: posts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.posts_id_seq OWNED BY public.posts.id;


--
-- Name: potential_providers; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.potential_providers (
    id integer NOT NULL,
    post_id integer NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.potential_providers OWNER TO scrutinizer;

--
-- Name: potential_providers_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.potential_providers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.potential_providers_id_seq OWNER TO scrutinizer;

--
-- Name: potential_providers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.potential_providers_id_seq OWNED BY public.potential_providers.id;


--
-- Name: schema_migration; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.schema_migration (
    version character varying(14) NOT NULL
);


ALTER TABLE public.schema_migration OWNER TO scrutinizer;

--
-- Name: thread_participants; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.thread_participants (
    id integer NOT NULL,
    thread_id integer NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    last_viewed_at timestamp without time zone DEFAULT '1900-01-01 00:00:00'::timestamp without time zone NOT NULL,
    last_notified_at timestamp without time zone DEFAULT '1900-01-01 00:00:00'::timestamp without time zone NOT NULL
);


ALTER TABLE public.thread_participants OWNER TO scrutinizer;

--
-- Name: thread_participants_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.thread_participants_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.thread_participants_id_seq OWNER TO scrutinizer;

--
-- Name: thread_participants_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.thread_participants_id_seq OWNED BY public.thread_participants.id;


--
-- Name: threads; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.threads (
    id integer NOT NULL,
    uuid uuid NOT NULL,
    post_id integer NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.threads OWNER TO scrutinizer;

--
-- Name: threads_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.threads_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.threads_id_seq OWNER TO scrutinizer;

--
-- Name: threads_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.threads_id_seq OWNED BY public.threads.id;


--
-- Name: user_access_tokens; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.user_access_tokens (
    id integer NOT NULL,
    user_id integer NOT NULL,
    user_organization_id integer NOT NULL,
    access_token character varying(255) NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.user_access_tokens OWNER TO scrutinizer;

--
-- Name: user_access_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.user_access_tokens_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_access_tokens_id_seq OWNER TO scrutinizer;

--
-- Name: user_access_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.user_access_tokens_id_seq OWNED BY public.user_access_tokens.id;


--
-- Name: user_organizations; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.user_organizations (
    id integer NOT NULL,
    organization_id integer NOT NULL,
    user_id integer NOT NULL,
    role character varying(255) NOT NULL,
    auth_id character varying(255) NOT NULL,
    auth_email character varying(255) NOT NULL,
    last_login timestamp without time zone NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.user_organizations OWNER TO scrutinizer;

--
-- Name: user_organizations_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.user_organizations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_organizations_id_seq OWNER TO scrutinizer;

--
-- Name: user_organizations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.user_organizations_id_seq OWNED BY public.user_organizations.id;


--
-- Name: user_preferences; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.user_preferences (
    id integer NOT NULL,
    uuid uuid NOT NULL,
    user_id integer NOT NULL,
    key character varying(4096) NOT NULL,
    value character varying(4096) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.user_preferences OWNER TO scrutinizer;

--
-- Name: user_preferences_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.user_preferences_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_preferences_id_seq OWNER TO scrutinizer;

--
-- Name: user_preferences_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.user_preferences_id_seq OWNED BY public.user_preferences.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.users (
    id integer NOT NULL,
    email character varying(255) NOT NULL,
    first_name character varying(255) NOT NULL,
    last_name character varying(255) NOT NULL,
    nickname character varying(255) NOT NULL,
    admin_role character varying(255) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    uuid uuid NOT NULL,
    photo_file_id integer,
    auth_photo_url character varying(255),
    location_id integer
);


ALTER TABLE public.users OWNER TO scrutinizer;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_id_seq OWNER TO scrutinizer;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: watches; Type: TABLE; Schema: public; Owner: scrutinizer
--

CREATE TABLE public.watches (
    id integer NOT NULL,
    uuid uuid NOT NULL,
    owner_id integer NOT NULL,
    location_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.watches OWNER TO scrutinizer;

--
-- Name: watches_id_seq; Type: SEQUENCE; Schema: public; Owner: scrutinizer
--

CREATE SEQUENCE public.watches_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.watches_id_seq OWNER TO scrutinizer;

--
-- Name: watches_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: scrutinizer
--

ALTER SEQUENCE public.watches_id_seq OWNED BY public.watches.id;


--
-- Name: files id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.files ALTER COLUMN id SET DEFAULT nextval('public.files_id_seq'::regclass);


--
-- Name: locations id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.locations ALTER COLUMN id SET DEFAULT nextval('public.locations_id_seq'::regclass);


--
-- Name: meetings id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.meetings ALTER COLUMN id SET DEFAULT nextval('public.meetings_id_seq'::regclass);


--
-- Name: messages id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.messages ALTER COLUMN id SET DEFAULT nextval('public.messages_id_seq'::regclass);


--
-- Name: organization_domains id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.organization_domains ALTER COLUMN id SET DEFAULT nextval('public.organization_domains_id_seq'::regclass);


--
-- Name: organization_trusts id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.organization_trusts ALTER COLUMN id SET DEFAULT nextval('public.organization_trusts_id_seq'::regclass);


--
-- Name: organizations id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.organizations ALTER COLUMN id SET DEFAULT nextval('public.organizations_id_seq'::regclass);


--
-- Name: post_files id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.post_files ALTER COLUMN id SET DEFAULT nextval('public.post_files_id_seq'::regclass);


--
-- Name: post_histories id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.post_histories ALTER COLUMN id SET DEFAULT nextval('public.post_histories_id_seq'::regclass);


--
-- Name: posts id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.posts ALTER COLUMN id SET DEFAULT nextval('public.posts_id_seq'::regclass);


--
-- Name: potential_providers id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.potential_providers ALTER COLUMN id SET DEFAULT nextval('public.potential_providers_id_seq'::regclass);


--
-- Name: thread_participants id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.thread_participants ALTER COLUMN id SET DEFAULT nextval('public.thread_participants_id_seq'::regclass);


--
-- Name: threads id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.threads ALTER COLUMN id SET DEFAULT nextval('public.threads_id_seq'::regclass);


--
-- Name: user_access_tokens id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.user_access_tokens ALTER COLUMN id SET DEFAULT nextval('public.user_access_tokens_id_seq'::regclass);


--
-- Name: user_organizations id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.user_organizations ALTER COLUMN id SET DEFAULT nextval('public.user_organizations_id_seq'::regclass);


--
-- Name: user_preferences id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.user_preferences ALTER COLUMN id SET DEFAULT nextval('public.user_preferences_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: watches id; Type: DEFAULT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.watches ALTER COLUMN id SET DEFAULT nextval('public.watches_id_seq'::regclass);


--
-- Name: files files_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_pkey PRIMARY KEY (id);


--
-- Name: locations locations_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.locations
    ADD CONSTRAINT locations_pkey PRIMARY KEY (id);


--
-- Name: meetings meetings_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.meetings
    ADD CONSTRAINT meetings_pkey PRIMARY KEY (id);


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);


--
-- Name: organization_domains organization_domains_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.organization_domains
    ADD CONSTRAINT organization_domains_pkey PRIMARY KEY (id);


--
-- Name: organization_trusts organization_trusts_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.organization_trusts
    ADD CONSTRAINT organization_trusts_pkey PRIMARY KEY (id);


--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);


--
-- Name: post_files post_files_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.post_files
    ADD CONSTRAINT post_files_pkey PRIMARY KEY (id);


--
-- Name: post_histories post_histories_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.post_histories
    ADD CONSTRAINT post_histories_pkey PRIMARY KEY (id);


--
-- Name: posts posts_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_pkey PRIMARY KEY (id);


--
-- Name: potential_providers potential_providers_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.potential_providers
    ADD CONSTRAINT potential_providers_pkey PRIMARY KEY (id);


--
-- Name: thread_participants thread_participants_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.thread_participants
    ADD CONSTRAINT thread_participants_pkey PRIMARY KEY (id);


--
-- Name: threads threads_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.threads
    ADD CONSTRAINT threads_pkey PRIMARY KEY (id);


--
-- Name: user_access_tokens user_access_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.user_access_tokens
    ADD CONSTRAINT user_access_tokens_pkey PRIMARY KEY (id);


--
-- Name: user_organizations user_organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.user_organizations
    ADD CONSTRAINT user_organizations_pkey PRIMARY KEY (id);


--
-- Name: user_preferences user_preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.user_preferences
    ADD CONSTRAINT user_preferences_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: watches watches_pkey; Type: CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.watches
    ADD CONSTRAINT watches_pkey PRIMARY KEY (id);


--
-- Name: files_uuid_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX files_uuid_idx ON public.files USING btree (uuid);


--
-- Name: meetings_location_id_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX meetings_location_id_idx ON public.meetings USING btree (location_id);


--
-- Name: meetings_uuid_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX meetings_uuid_idx ON public.meetings USING btree (uuid);


--
-- Name: messages_uuid_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX messages_uuid_idx ON public.messages USING btree (uuid);


--
-- Name: organization_domains_domain_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX organization_domains_domain_idx ON public.organization_domains USING btree (domain);


--
-- Name: organization_trusts_primary_id_secondary_id_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX organization_trusts_primary_id_secondary_id_idx ON public.organization_trusts USING btree (primary_id, secondary_id);


--
-- Name: organizations_logo_file_id_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX organizations_logo_file_id_idx ON public.organizations USING btree (logo_file_id);


--
-- Name: organizations_uuid_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX organizations_uuid_idx ON public.organizations USING btree (uuid);


--
-- Name: post_files_file_id_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX post_files_file_id_idx ON public.post_files USING btree (file_id);


--
-- Name: post_histories_created_at_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE INDEX post_histories_created_at_idx ON public.post_histories USING btree (created_at);


--
-- Name: posts_destination_id_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX posts_destination_id_idx ON public.posts USING btree (destination_id);


--
-- Name: posts_meeting_id_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX posts_meeting_id_idx ON public.posts USING btree (meeting_id);


--
-- Name: posts_origin_id_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX posts_origin_id_idx ON public.posts USING btree (origin_id);


--
-- Name: posts_photo_file_id_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX posts_photo_file_id_idx ON public.posts USING btree (photo_file_id);


--
-- Name: posts_uuid_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX posts_uuid_idx ON public.posts USING btree (uuid);


--
-- Name: schema_migration_version_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX schema_migration_version_idx ON public.schema_migration USING btree (version);


--
-- Name: threads_uuid_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX threads_uuid_idx ON public.threads USING btree (uuid);


--
-- Name: user_access_tokens_access_token_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX user_access_tokens_access_token_idx ON public.user_access_tokens USING btree (access_token);


--
-- Name: user_organizations_organization_id_auth_email_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX user_organizations_organization_id_auth_email_idx ON public.user_organizations USING btree (organization_id, auth_email);


--
-- Name: user_organizations_organization_id_auth_id_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX user_organizations_organization_id_auth_id_idx ON public.user_organizations USING btree (organization_id, auth_id);


--
-- Name: user_organizations_organization_id_user_id_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX user_organizations_organization_id_user_id_idx ON public.user_organizations USING btree (organization_id, user_id);


--
-- Name: user_preferences_uuid_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX user_preferences_uuid_idx ON public.user_preferences USING btree (uuid);


--
-- Name: users_email_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX users_email_idx ON public.users USING btree (email);


--
-- Name: users_location_id_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX users_location_id_idx ON public.users USING btree (location_id);


--
-- Name: users_nickname_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX users_nickname_idx ON public.users USING btree (nickname);


--
-- Name: users_photo_file_id_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX users_photo_file_id_idx ON public.users USING btree (photo_file_id);


--
-- Name: users_uuid_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX users_uuid_idx ON public.users USING btree (uuid);


--
-- Name: watches_location_id_idx; Type: INDEX; Schema: public; Owner: scrutinizer
--

CREATE UNIQUE INDEX watches_location_id_idx ON public.watches USING btree (location_id);


--
-- Name: posts meeting_fk; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT meeting_fk FOREIGN KEY (meeting_id) REFERENCES public.meetings(id) ON DELETE SET NULL;


--
-- Name: meetings meetings_created_by_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.meetings
    ADD CONSTRAINT meetings_created_by_id_fkey FOREIGN KEY (created_by_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: meetings meetings_image_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.meetings
    ADD CONSTRAINT meetings_image_file_id_fkey FOREIGN KEY (image_file_id) REFERENCES public.files(id) ON DELETE SET NULL;


--
-- Name: meetings meetings_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.meetings
    ADD CONSTRAINT meetings_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.locations(id);


--
-- Name: messages messages_sent_by_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_sent_by_id_fkey FOREIGN KEY (sent_by_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: messages messages_thread_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_thread_id_fkey FOREIGN KEY (thread_id) REFERENCES public.threads(id) ON DELETE CASCADE;


--
-- Name: organization_domains organization_domains_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.organization_domains
    ADD CONSTRAINT organization_domains_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: organization_trusts organization_trusts_primary_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.organization_trusts
    ADD CONSTRAINT organization_trusts_primary_id_fkey FOREIGN KEY (primary_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: organization_trusts organization_trusts_secondary_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.organization_trusts
    ADD CONSTRAINT organization_trusts_secondary_id_fkey FOREIGN KEY (secondary_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: organizations organizations_files_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_files_id_fk FOREIGN KEY (logo_file_id) REFERENCES public.files(id) ON DELETE SET NULL;


--
-- Name: posts post_destination_fk; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT post_destination_fk FOREIGN KEY (destination_id) REFERENCES public.locations(id);


--
-- Name: post_files post_files_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.post_files
    ADD CONSTRAINT post_files_file_id_fkey FOREIGN KEY (file_id) REFERENCES public.files(id) ON DELETE CASCADE;


--
-- Name: post_files post_files_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.post_files
    ADD CONSTRAINT post_files_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id) ON DELETE CASCADE;


--
-- Name: post_histories post_histories_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.post_histories
    ADD CONSTRAINT post_histories_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id) ON DELETE CASCADE;


--
-- Name: post_histories post_histories_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.post_histories
    ADD CONSTRAINT post_histories_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: post_histories post_histories_receiver_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.post_histories
    ADD CONSTRAINT post_histories_receiver_id_fkey FOREIGN KEY (receiver_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: posts post_origin_fk; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT post_origin_fk FOREIGN KEY (origin_id) REFERENCES public.locations(id) ON DELETE SET NULL;


--
-- Name: posts posts_created_by_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_created_by_id_fkey FOREIGN KEY (created_by_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: posts posts_files_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_files_id_fk FOREIGN KEY (photo_file_id) REFERENCES public.files(id) ON DELETE SET NULL;


--
-- Name: posts posts_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE SET NULL;


--
-- Name: posts posts_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: posts posts_receiver_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_receiver_id_fkey FOREIGN KEY (receiver_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: potential_providers potential_providers_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.potential_providers
    ADD CONSTRAINT potential_providers_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id) ON DELETE CASCADE;


--
-- Name: potential_providers potential_providers_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.potential_providers
    ADD CONSTRAINT potential_providers_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: thread_participants thread_participants_thread_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.thread_participants
    ADD CONSTRAINT thread_participants_thread_id_fkey FOREIGN KEY (thread_id) REFERENCES public.threads(id) ON DELETE CASCADE;


--
-- Name: thread_participants thread_participants_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.thread_participants
    ADD CONSTRAINT thread_participants_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: threads threads_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.threads
    ADD CONSTRAINT threads_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id) ON DELETE CASCADE;


--
-- Name: user_access_tokens user_access_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.user_access_tokens
    ADD CONSTRAINT user_access_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_access_tokens user_access_tokens_user_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.user_access_tokens
    ADD CONSTRAINT user_access_tokens_user_organization_id_fkey FOREIGN KEY (user_organization_id) REFERENCES public.user_organizations(id) ON DELETE CASCADE;


--
-- Name: user_organizations user_organizations_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.user_organizations
    ADD CONSTRAINT user_organizations_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: user_organizations user_organizations_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.user_organizations
    ADD CONSTRAINT user_organizations_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_preferences user_preferences_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.user_preferences
    ADD CONSTRAINT user_preferences_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: users users_files_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_files_id_fk FOREIGN KEY (photo_file_id) REFERENCES public.files(id) ON DELETE SET NULL;


--
-- Name: users users_locations_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_locations_id_fk FOREIGN KEY (location_id) REFERENCES public.locations(id) ON DELETE SET NULL;


--
-- Name: watches watches_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.watches
    ADD CONSTRAINT watches_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.locations(id) ON DELETE CASCADE;


--
-- Name: watches watches_owner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: scrutinizer
--

ALTER TABLE ONLY public.watches
    ADD CONSTRAINT watches_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

