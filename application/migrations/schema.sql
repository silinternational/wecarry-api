--
-- PostgreSQL database dump
--

-- Dumped from database version 9.6.13
-- Dumped by pg_dump version 11.3 (Debian 11.3-1.pgdg90+1)

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

SET default_with_oids = false;

--
-- Name: organizations; Type: TABLE; Schema: public; Owner: handcarry
--

CREATE TABLE public.organizations (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    url character varying(255),
    auth_type character varying(255) NOT NULL,
    auth_config jsonb NOT NULL,
    uuid character varying(255) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.organizations OWNER TO handcarry;

--
-- Name: organizations_id_seq; Type: SEQUENCE; Schema: public; Owner: handcarry
--

CREATE SEQUENCE public.organizations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.organizations_id_seq OWNER TO handcarry;

--
-- Name: organizations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: handcarry
--

ALTER SEQUENCE public.organizations_id_seq OWNED BY public.organizations.id;


--
-- Name: schema_migration; Type: TABLE; Schema: public; Owner: handcarry
--

CREATE TABLE public.schema_migration (
    version character varying(14) NOT NULL
);


ALTER TABLE public.schema_migration OWNER TO handcarry;

--
-- Name: user_access_tokens; Type: TABLE; Schema: public; Owner: handcarry
--

CREATE TABLE public.user_access_tokens (
    id integer NOT NULL,
    user_id integer NOT NULL,
    access_token character varying(255) NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.user_access_tokens OWNER TO handcarry;

--
-- Name: user_access_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: handcarry
--

CREATE SEQUENCE public.user_access_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_access_tokens_id_seq OWNER TO handcarry;

--
-- Name: user_access_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: handcarry
--

ALTER SEQUENCE public.user_access_tokens_id_seq OWNED BY public.user_access_tokens.id;


--
-- Name: user_organizations; Type: TABLE; Schema: public; Owner: handcarry
--

CREATE TABLE public.user_organizations (
    id uuid NOT NULL,
    org_id integer NOT NULL,
    user_id integer NOT NULL,
    role character varying(255) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.user_organizations OWNER TO handcarry;

--
-- Name: users; Type: TABLE; Schema: public; Owner: handcarry
--

CREATE TABLE public.users (
    id integer NOT NULL,
    email character varying(255) NOT NULL,
    first_name character varying(255) NOT NULL,
    last_name character varying(255) NOT NULL,
    nickname character varying(255) NOT NULL,
    auth_org_id integer NOT NULL,
    auth_org_uid character varying(255) NOT NULL,
    admin_role character varying(255),
    uuid character varying(255) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.users OWNER TO handcarry;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: handcarry
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_id_seq OWNER TO handcarry;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: handcarry
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: organizations id; Type: DEFAULT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.organizations ALTER COLUMN id SET DEFAULT nextval('public.organizations_id_seq'::regclass);


--
-- Name: user_access_tokens id; Type: DEFAULT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.user_access_tokens ALTER COLUMN id SET DEFAULT nextval('public.user_access_tokens_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);


--
-- Name: user_access_tokens user_access_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.user_access_tokens
    ADD CONSTRAINT user_access_tokens_pkey PRIMARY KEY (id);


--
-- Name: user_organizations user_organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.user_organizations
    ADD CONSTRAINT user_organizations_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: schema_migration_version_idx; Type: INDEX; Schema: public; Owner: handcarry
--

CREATE UNIQUE INDEX schema_migration_version_idx ON public.schema_migration USING btree (version);


--
-- Name: user_access_tokens user_access_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.user_access_tokens
    ADD CONSTRAINT user_access_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_organizations user_organizations_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.user_organizations
    ADD CONSTRAINT user_organizations_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: user_organizations user_organizations_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.user_organizations
    ADD CONSTRAINT user_organizations_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: users users_auth_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_auth_org_id_fkey FOREIGN KEY (auth_org_id) REFERENCES public.organizations(id) ON DELETE SET NULL;


--
-- PostgreSQL database dump complete
--

